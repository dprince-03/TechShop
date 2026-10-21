//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// stockOf is the on-hand quantity of a variant and condition in the Ikeja warehouse.
func stockOf(t *testing.T, variant, condition string) float64 {
	t.Helper()
	r := get(t, "/api/v1/staff/inventory/levels?variantId="+variant, staffLogin(t, "warehouse_lead"))
	for _, l := range r.List {
		m := l.(map[string]any)
		if m["warehouseId"] == seed.Extra["warehouseIkejaId"] && m["condition"] == condition {
			return m["onHand"].(float64)
		}
	}
	return 0
}

func TestReturnInspectRestockRefund(t *testing.T) {
	t.Parallel()
	cust := newCustomer(t, "Ruth")
	number, _, line := deliveredByCarrier(t, cust, seed.Extra["listing_p21"], nil)
	base := "/api/v1/me/orders/" + number + "/returns"
	item := func(qty int) []any { return []any{map[string]any{"orderLineId": line, "quantity": qty}} }

	expectCode(t, post(t, base, cust, map[string]any{"items": item(1), "reason": "change_of_mind", "method": "store_dropoff"}), 422, "change_of_mind_unavailable", "change of mind")
	expectCode(t, post(t, base, cust, map[string]any{"items": item(2), "reason": "faulty", "method": "store_dropoff"}), 422, "too_many", "more than bought")
	r := post(t, base, cust, map[string]any{"items": item(1), "reason": "faulty", "details": "Two bulbs won't pair", "method": "store_dropoff"})
	expect(t, r, 201, "return")
	rma := r.str("rma")
	if r.str("status") != "approved" || r.get("autoApproved") != true || !strings.HasPrefix(rma, "RMA-") || r.num("refund") <= 0 {
		t.Fatalf("faulty return within 7 days should be auto-approved: %s", r.Raw)
	}
	expectCode(t, post(t, base, cust, map[string]any{"items": item(1), "reason": "faulty", "method": "store_dropoff"}), 422, "too_many", "returned twice")
	if l := get(t, "/api/v1/me/returns", cust); !strings.Contains(string(l.Raw), rma) {
		t.Fatalf("my returns: %s", l.Raw)
	}
	expect(t, get(t, "/api/v1/me/returns/"+rma, newCustomer(t, "Other")), 404, "someone else's return")

	// Support can't re-decide an approved return; the warehouse receives and inspects it.
	expectCode(t, post(t, "/api/v1/staff/returns/"+rma+"/decision", staffLogin(t, "support_lead"), map[string]any{"decision": "approve"}), 409, "bad_transition", "decide twice")
	wh := staffLogin(t, "warehouse_lead")
	expectCode(t, post(t, "/api/v1/staff/returns/"+rma+"/refund", wh, nil), 409, "bad_transition", "refund before inspection")
	expect(t, post(t, "/api/v1/staff/returns/"+rma+"/receive", wh, nil), 200, "receive")
	itemID := get(t, "/api/v1/staff/returns/"+rma, wh).get("items").([]any)[0].(map[string]any)["id"].(string)
	before := stockOf(t, seed.Extra["variant_p21"], "open_box")
	ins := post(t, "/api/v1/staff/returns/"+rma+"/items/"+itemID+"/inspection", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "grade": "A", "disposition": "restock_open_box", "notes": "Box opened, bulbs fine"})
	expect(t, ins, 200, "inspect")
	if ins.str("status") != "inspected" {
		t.Fatalf("inspected: %s", ins.Raw)
	}
	if after := stockOf(t, seed.Extra["variant_p21"], "open_box"); after != before+1 {
		t.Fatalf("open-box stock %v → %v, want +1", before, after)
	}
	expectCode(t, post(t, "/api/v1/staff/returns/"+rma+"/items/"+itemID+"/inspection", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "grade": "A", "disposition": "write_off"}), 409, "bad_transition", "inspect twice")

	// The refund request goes to finance (dual control); the staff member who raised it can't approve it.
	ref := post(t, "/api/v1/staff/returns/"+rma+"/refund", wh, nil)
	expect(t, ref, 200, "refund")
	if ref.str("status") != "refunded" {
		t.Fatalf("refund: %s", ref.Raw)
	}
	officer := staffLogin(t, "finance_officer")
	waitFor(t, "refund request for "+number, 15*time.Second, func() bool {
		return strings.Contains(string(get(t, "/api/v1/staff/finance/refunds?status=requested", officer).Raw), number)
	})
	trialBalanced(t)
}

// A returned phone whose IMEI isn't the one sold is refused and a risk case is opened.
func TestReturnSerialMismatchOpensRiskCase(t *testing.T) {
	t.Parallel()
	wh := staffLogin(t, "warehouse_lead")
	imei := luhnIMEI(fmt.Sprintf("35%012d", time.Now().UnixNano()%1e12))
	expect(t, post(t, "/api/v1/staff/inventory/device-units", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "variantId": seed.Extra["variant_p18"],
		"condition": "new", "units": []any{map[string]any{"imei": imei}}}), 201, "receive imei")
	cust := newCustomer(t, "Swapper")
	number, _, line := deliveredByCarrier(t, cust, seed.Extra["listing_p18"], map[string]string{"*": imei})
	r := post(t, "/api/v1/me/orders/"+number+"/returns", cust, map[string]any{"items": []any{map[string]any{"orderLineId": line, "quantity": 1}}, "reason": "faulty", "method": "pickup"})
	expect(t, r, 201, "phone return")
	if r.str("status") != "awaiting_pickup" {
		t.Fatalf("pickup return: %s", r.Raw)
	}
	rma := r.str("rma")
	expect(t, post(t, "/api/v1/staff/returns/"+rma+"/receive", wh, nil), 200, "receive")
	item := get(t, "/api/v1/staff/returns/"+rma, wh).get("items").([]any)[0].(map[string]any)
	if item["serialised"] != true {
		t.Fatalf("phone should be serialised: %v", item)
	}
	other := luhnIMEI(fmt.Sprintf("35%012d", (time.Now().UnixNano()+7)%1e12))
	expectCode(t, post(t, "/api/v1/staff/returns/"+rma+"/items/"+item["id"].(string)+"/inspection", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "scannedSerial": other, "grade": "faulty", "disposition": "repair"}), 409, "serial_mismatch", "swapped phone")
	back := get(t, "/api/v1/me/returns/"+rma, cust)
	if back.str("status") != "rejected" || !strings.Contains(back.str("rejectionReason"), "doesn’t match") {
		t.Fatalf("customer should see the rejection: %s", back.Raw)
	}
}
