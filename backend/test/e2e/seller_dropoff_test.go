//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"
)

// warehouseByCode finds a seeded location's ID from the staff list.
func warehouseByCode(t *testing.T, code string) string {
	t.Helper()
	for _, w := range get(t, "/api/v1/staff/inventory/warehouses", staffLogin(t, "warehouse_lead")).List {
		if m := w.(map[string]any); m["code"] == code {
			return m["id"].(string)
		}
	}
	t.Fatalf("no warehouse %s", code)
	return ""
}

// acceptAndPack takes a new order through the seller's accept and pack; it returns the packed view.
func acceptAndPack(t *testing.T, seller, orders, number string) resp {
	t.Helper()
	fid := pendingOrderFor(t, seller, orders, number)
	expect(t, post(t, orders+"/"+fid+"/accept", seller, nil), 200, "accept")
	p := post(t, orders+"/"+fid+"/pack", seller, nil)
	expect(t, p, 200, "pack")
	return p
}

// Seller hand-over at a TechShop hub (docs/plan.md 2026-10-21): packing gives a TSS- label and a
// 24-hour drop-off deadline; hub staff scan the label in; TechShop's rider or carrier delivers,
// and a parcel never dropped off is cancelled and refunded.
func TestSellerDropoffToDelivery(t *testing.T) {
	t.Parallel()
	seller, shopID := approvedShop(t, "Ngozi", "Eze")
	c := post(t, "/api/v1/seller/shops/"+shopID+"/listings", seller, map[string]any{"variantId": seed.Extra["variant_p1"], "condition": "new", "price": 88_000_000, "sellerStock": 5})
	expect(t, c, 201, "listing")
	listing := c.str("listing", "id")
	expect(t, post(t, "/api/v1/staff/catalog/listings/"+listing+"/decision", staffLogin(t, "catalog_editor"), map[string]any{"decision": "approved"}), 200, "approve listing")
	orders := "/api/v1/seller/shops/" + shopID + "/orders"
	cust := newCustomer(t, "Tobi")

	// Drop-off points are warehouses and hubs, never retail stores.
	pts := get(t, "/api/v1/seller/dropoff-points", seller)
	expect(t, pts, 200, "drop-off points")
	if s := string(pts.Raw); !strings.Contains(s, "IKJ") || !strings.Contains(s, "WUSE") || strings.Contains(s, "VI-STORE") {
		t.Fatalf("drop-off points: %s", s)
	}

	// 1. Lagos order: pack → label and deadline → hub scan → rider job.
	number, _ := buyAndPay(t, cust, listing)
	p := acceptAndPack(t, seller, orders, number)
	label, fid := p.str("label"), p.str("id")
	by, _ := time.Parse(time.RFC3339, p.str("handoverBy"))
	if p.str("status") != "packed" || !strings.HasPrefix(label, "TSS-") || time.Until(by) < 23*time.Hour || time.Until(by) > 49*time.Hour {
		t.Fatalf("packed view needs a TSS label and a ~24h drop-off deadline: %s", p.Raw)
	}
	if again := get(t, orders+"/"+fid, seller); again.str("label") != label {
		t.Fatalf("label must stay the same: %s", again.Raw)
	}
	wh := staffLogin(t, "warehouse_lead")
	ikeja := seed.Extra["warehouseIkejaId"]
	expectCode(t, post(t, "/api/v1/staff/wms/dropoffs", wh, map[string]any{"warehouseId": warehouseByCode(t, "VI-STORE"), "label": label}), 422, "not_a_dropoff_point", "store isn't a hub")
	expect(t, post(t, "/api/v1/staff/wms/dropoffs", wh, map[string]any{"warehouseId": ikeja, "label": "TSS-NOTREAL1"}), 404, "unknown label")
	expect(t, post(t, "/api/v1/staff/wms/dropoffs", seller, map[string]any{"warehouseId": ikeja, "label": label}), 401, "seller can't scan in")
	d := post(t, "/api/v1/staff/wms/dropoffs", wh, map[string]any{"warehouseId": ikeja, "label": strings.ToLower(label)})
	expect(t, d, 201, "drop-off")
	if d.str("handoff") != "rider" || d.str("jobId") == "" || d.str("orderNumber") != number {
		t.Fatalf("a Lagos parcel should get a rider job: %s", d.Raw)
	}
	expectCode(t, post(t, "/api/v1/staff/wms/dropoffs", wh, map[string]any{"warehouseId": ikeja, "label": label}), 409, "parcel_not_expected", "scanned twice")
	if s := get(t, orders+"/"+fid, seller).str("status"); s != "handed_over" {
		t.Fatalf("after drop-off the seller sees %s", s)
	}
	expect(t, post(t, "/dev/sla/"+fid+"/handover", "", nil), 200, "late hand-over after drop-off")
	if s := get(t, orders+"/"+fid, seller).str("status"); s != "handed_over" {
		t.Fatalf("a dropped-off parcel must not be cancelled by the hand-over SLA, is %s", s)
	}

	// 2. Kano order: dropped at the Wuse hub, carrier booked, delivered on carrier tracking.
	a := post(t, "/api/v1/me/addresses", cust, map[string]any{"recipientName": "Tobi Musa", "phone": "08031234571", "line1": "5 Murtala Mohammed Way", "city": "Kano", "stateCode": "KN"})
	expect(t, a, 201, "Kano address")
	clearCart(t, cust)
	addToCart(t, cust, listing, 1)
	pv := preview(t, cust, a.str("id"), "")
	expect(t, pv, 200, "preview")
	o := placeOrder(t, cust, a.str("id"), pv.str("priceHash"), "paystack", newKey(), "")
	expect(t, o, 201, "order to Kano")
	number2 := o.str("orderNumber")
	fakePay(t, "paystack", o.str("payment", "reference"), "success")
	waitOrderStatus(t, cust, number2, "paid")
	p2 := acceptAndPack(t, seller, orders, number2)
	d2 := post(t, "/api/v1/staff/wms/dropoffs", wh, map[string]any{"warehouseId": warehouseByCode(t, "WUSE"), "label": p2.str("label")})
	expect(t, d2, 201, "drop-off at Wuse")
	tracking := d2.str("shipment", "trackingNumber")
	if d2.str("handoff") != "carrier" || tracking == "" {
		t.Fatalf("Kano parcel should be booked with a carrier: %s", d2.Raw)
	}
	for _, status := range []string{"in_transit", "delivered"} {
		ev := post(t, "/dev/carriers/fake/"+tracking+"/"+status, "", nil)
		expect(t, call(t, req{method: "POST", path: "/api/v1/webhooks/carriers/fake", headers: map[string]string{"x-carrier-signature": ev.str("signature")}, raw: ev.str("raw")}), 200, status)
	}
	waitOrderStatus(t, cust, number2, "delivered")
	if s := get(t, orders+"/"+p2.str("id"), seller).str("status"); s != "delivered" {
		t.Fatalf("seller sees %s after delivery", s)
	}

	// 3. Packed but never dropped off: cancelled as a late hand-over and refunded.
	number3, _ := buyAndPay(t, cust, listing)
	p3 := acceptAndPack(t, seller, orders, number3)
	expect(t, post(t, "/dev/sla/"+p3.str("id")+"/handover", "", nil), 200, "late hand-over")
	if l := get(t, orders+"/"+p3.str("id"), seller); l.str("status") != "cancelled" || !strings.Contains(l.str("cancelledReason"), "hub") {
		t.Fatalf("late hand-over: %s", l.Raw)
	}
	waitOrderStatus(t, cust, number3, "cancelled")
	officer := staffLogin(t, "finance_officer")
	waitFor(t, "refund request for "+number3, 15*time.Second, func() bool {
		return strings.Contains(string(get(t, "/api/v1/staff/finance/refunds?status=requested", officer).Raw), number3)
	})
	expectCode(t, post(t, "/api/v1/staff/wms/dropoffs", wh, map[string]any{"warehouseId": ikeja, "label": p3.str("label")}), 409, "parcel_not_expected", "cancelled parcel scanned")
	trialBalanced(t)
}
