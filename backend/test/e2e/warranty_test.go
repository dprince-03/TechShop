//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// receivedPhone sells a tracked phone to a new customer, delivers it, and returns the customer,
// the order line and the IMEI sold.
func deliveredPhone(t *testing.T, first string, offset int64) (cust, line, imei string) {
	t.Helper()
	imei = luhnIMEI(fmt.Sprintf("35%012d", (time.Now().UnixNano()+offset)%1e12))
	expect(t, post(t, "/api/v1/staff/inventory/device-units", staffLogin(t, "warehouse_lead"), map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"],
		"variantId": seed.Extra["variant_p18"], "condition": "new", "units": []any{map[string]any{"imei": imei}}}), 201, "receive imei")
	cust = newCustomer(t, first)
	_, _, line = deliveredByCarrier(t, cust, seed.Extra["listing_p18"], map[string]string{"*": imei})
	return cust, line, imei
}

// A claim from the customer's side to collection: cover, serial check, drop-off, workshop, ready.
func TestWarrantyClaimRepairCollect(t *testing.T) {
	t.Parallel()
	cust, line, imei := deliveredPhone(t, "Femi", 101)

	items := get(t, "/api/v1/me/warranty/items", cust)
	expect(t, items, 200, "covered items")
	if len(items.List) != 1 || items.List[0].(map[string]any)["serialised"] != true || items.List[0].(map[string]any)["serialHint"] != imei[11:] {
		t.Fatalf("one tracked phone under warranty: %s", items.Raw)
	}
	other := imei[:14] + fmt.Sprint((int(imei[14]-'0')+1)%10)
	expectCode(t, post(t, "/api/v1/me/warranty-claims", cust, map[string]any{"orderLineId": line, "imeiOrSerial": other, "fault": "Won't charge at all"}), 422, "serial_mismatch", "wrong IMEI")
	c := post(t, "/api/v1/me/warranty-claims", cust, map[string]any{"orderLineId": line, "imeiOrSerial": imei, "fault": "Won't charge at all"})
	expect(t, c, 201, "claim")
	number := c.str("number")
	if !strings.HasPrefix(number, "WR-") || c.str("status") != "new" || len(c.get("dropOff").([]any)) == 0 || c.str("provider") == "" {
		t.Fatalf("new claim with drop-off stores: %s", c.Raw)
	}
	expectCode(t, post(t, "/api/v1/me/warranty-claims", cust, map[string]any{"orderLineId": line, "imeiOrSerial": imei, "fault": "Still won't charge"}), 409, "claim_open", "second open claim")
	expect(t, get(t, "/api/v1/me/warranty-claims/"+number, newCustomer(t, "Stranger")), 404, "someone else's claim")

	tech := staffLogin(t, "aftersales")
	store := seed.Extra["storeVIId"]
	expectCode(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/receive", tech, map[string]any{"storeId": store, "scannedSerial": other}), 422, "serial_mismatch", "receive another phone")
	expect(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/receive", tech, map[string]any{"storeId": store, "scannedSerial": imei}), 200, "receive")
	r := post(t, "/api/v1/staff/warranty/claims/"+number+"/repairs", tech, map[string]any{"diagnosis": "Charging port worn"})
	expect(t, r, 200, "start repair")
	if r.str("status") != "in_repair" || len(r.get("repairs").([]any)) != 1 {
		t.Fatalf("in repair with one job: %s", r.Raw)
	}
	expectCode(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/repairs", tech, map[string]any{}), 409, "repair_open", "second repair job")
	job := r.get("repairs").([]any)[0].(map[string]any)["id"].(string)
	expect(t, post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "awaiting_parts", "notes": "Port ordered"}), 200, "awaiting parts")
	expectCode(t, post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "done"}), 409, "bad_transition", "done before parts")
	expect(t, post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "repairing"}), 200, "repairing")
	done := post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "done", "notes": "Port replaced, charges at full speed"})
	expect(t, done, 200, "done")
	if done.str("status") != "done" || done.str("claimNumber") != number || done.str("purpose") != "warranty" {
		t.Fatalf("finished job: %s", done.Raw)
	}
	if c := get(t, "/api/v1/staff/warranty/claims/"+number, tech); c.str("status") != "ready" || c.str("resolution") != "repaired" {
		t.Fatalf("a finished repair is ready to collect: %s", c.Raw)
	}
	mine := get(t, "/api/v1/me/warranty-claims/"+number, cust)
	if mine.str("status") != "ready" || mine.str("store") != "Victoria Island store" {
		t.Fatalf("customer sees it's ready at the store: %s", mine.Raw)
	}
	expect(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/collect", tech, nil), 200, "collect")
	expectCode(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/collect", tech, nil), 409, "bad_transition", "collect twice")

	// A second fault later: the device's history shows the first claim.
	again := post(t, "/api/v1/me/warranty-claims", cust, map[string]any{"orderLineId": line, "imeiOrSerial": imei, "fault": "Screen flickers"})
	expect(t, again, 201, "second claim after the first is done")
	h := get(t, "/api/v1/staff/warranty/claims/"+again.str("number"), tech)
	if len(h.get("history").([]any)) != 1 {
		t.Fatalf("history of this IMEI: %s", h.Raw)
	}
	rej := post(t, "/api/v1/staff/warranty/claims/"+again.str("number")+"/reject", tech, map[string]any{"reason": "Liquid damage indicator is red: not covered."})
	expect(t, rej, 200, "reject")
	if rej.str("status") != "rejected" || rej.str("resolution") != "not_covered" {
		t.Fatalf("rejected: %s", rej.Raw)
	}
	expect(t, get(t, "/api/v1/staff/warranty/claims", staffLogin(t, "support_agent")), 403, "support can't run the workshop")
}

// A walk-in claim on a phone that was later reported stolen is refused and opens a risk case;
// an unrepairable one is refunded through finance.
func TestWarrantyWalkInStolenAndRefund(t *testing.T) {
	t.Parallel()
	_, _, stolen := deliveredPhone(t, "Kunle", 211)
	expect(t, post(t, "/api/v1/staff/risk/devices", staffLogin(t, "risk_analyst"), map[string]any{"imei": stolen, "reason": "reported_stolen", "source": "Police extract"}), 201, "block")
	tech := staffLogin(t, "aftersales")
	expectCode(t, post(t, "/api/v1/staff/warranty/claims", tech, map[string]any{"imeiOrSerial": stolen, "fault": "Cracked camera", "storeId": seed.Extra["storeVIId"]}), 409, "device_blocked", "stolen phone")
	riskCaseAbout(t, stolen)

	_, _, imei := deliveredPhone(t, "Bisi", 307)
	expectCode(t, post(t, "/api/v1/staff/warranty/claims", tech, map[string]any{"imeiOrSerial": "000000000000000", "fault": "Dead on arrival", "storeId": seed.Extra["storeVIId"]}), 422, "unknown_device", "unknown device")
	w := post(t, "/api/v1/staff/warranty/claims", tech, map[string]any{"imeiOrSerial": imei, "fault": "Won't turn on", "storeId": seed.Extra["storeVIId"]})
	expect(t, w, 201, "walk-in")
	number := w.str("number")
	if w.str("status") != "received" {
		t.Fatalf("walk-ins are checked in at once: %s", w.Raw)
	}
	r := post(t, "/api/v1/staff/warranty/claims/"+number+"/repairs", tech, map[string]any{})
	job := r.get("repairs").([]any)[0].(map[string]any)["id"].(string)
	expectCode(t, post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "unrepairable"}), 422, "diagnosis_required", "finish without a diagnosis")
	expect(t, post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "unrepairable", "diagnosis": "Board failure"}), 200, "unrepairable")
	expectCode(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/resolution", tech, map[string]any{"resolution": "replaced"}), 422, "replacement_required", "replace without a unit")
	ref := post(t, "/api/v1/staff/warranty/claims/"+number+"/resolution", tech, map[string]any{"resolution": "refunded", "note": "No stock to replace"})
	expect(t, ref, 200, "refund")
	if ref.str("status") != "closed" || ref.str("resolution") != "refunded" {
		t.Fatalf("refunded claim closes: %s", ref.Raw)
	}
	if s := get(t, "/api/v1/staff/warranty/summary", tech); s.num("open") < 0 {
		t.Fatalf("summary: %s", s.Raw)
	}
}

// An unrepairable phone is swapped for a unit from stock, which leaves stock with its own reason.
func TestWarrantyReplacement(t *testing.T) {
	t.Parallel()
	_, _, imei := deliveredPhone(t, "Tolu", 419)
	spare := luhnIMEI(fmt.Sprintf("35%012d", (time.Now().UnixNano()+523)%1e12))
	expect(t, post(t, "/api/v1/staff/inventory/device-units", staffLogin(t, "warehouse_lead"), map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"],
		"variantId": seed.Extra["variant_p18"], "condition": "new", "units": []any{map[string]any{"imei": spare}}}), 201, "spare unit")
	tech := staffLogin(t, "aftersales")
	w := post(t, "/api/v1/staff/warranty/claims", tech, map[string]any{"imeiOrSerial": imei, "fault": "Speaker crackles", "storeId": seed.Extra["storeVIId"]})
	expect(t, w, 201, "walk-in")
	number := w.str("number")
	expectCode(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/resolution", tech, map[string]any{"resolution": "replaced", "replacementSerial": imei}), 422, "replacement_unavailable", "the faulty unit itself")
	rep := post(t, "/api/v1/staff/warranty/claims/"+number+"/resolution", tech, map[string]any{"resolution": "replaced", "replacementSerial": spare})
	expect(t, rep, 200, "replace")
	if rep.str("status") != "ready" || rep.str("resolution") != "replaced" {
		t.Fatalf("replacement ready to collect: %s", rep.Raw)
	}
	// The spare left stock: it can't be handed over a second time.
	_, _, imei2 := deliveredPhone(t, "Tolu2", 631)
	w2 := post(t, "/api/v1/staff/warranty/claims", tech, map[string]any{"imeiOrSerial": imei2, "fault": "Speaker crackles", "storeId": seed.Extra["storeVIId"]})
	expectCode(t, post(t, "/api/v1/staff/warranty/claims/"+w2.str("number")+"/resolution", tech, map[string]any{"resolution": "replaced", "replacementSerial": spare}), 422, "replacement_unavailable", "spare already used")
	expect(t, post(t, "/api/v1/staff/warranty/claims/"+number+"/collect", tech, nil), 200, "collect")
}
