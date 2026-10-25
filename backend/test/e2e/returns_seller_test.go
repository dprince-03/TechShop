//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"
)

// refundFor finds the requested refund raised for a return.
func refundFor(t *testing.T, rma string) string {
	t.Helper()
	l := get(t, "/api/v1/staff/finance/refunds?status=requested", staffLogin(t, "finance_manager"))
	expect(t, l, 200, "refund queue")
	items, _ := l.get("items").([]any)
	if items == nil {
		items = l.List
	}
	for _, it := range items {
		if m := it.(map[string]any); strings.Contains(m["reason"].(string), rma) {
			return m["id"].(string)
		}
	}
	t.Fatalf("no refund for %s: %s", rma, truncate(string(l.Raw)))
	return ""
}

// A seller's item comes back through TechShop: it can only go back to the seller, the seller
// sees their charge and may dispute within 3 days; an upheld dispute credits them, and the
// refund itself is charged to the seller, not split across the order.
func TestSellerReturnAndDispute(t *testing.T) {
	t.Parallel()
	seller, shopID := approvedShop(t, "Chidi", "Eze")
	c := post(t, "/api/v1/seller/shops/"+shopID+"/listings", seller, map[string]any{"variantId": seed.Extra["variant_p3"], "condition": "new", "price": 15_000_000, "sellerStock": 4})
	expect(t, c, 201, "listing")
	listing := c.str("listing", "id")
	expect(t, post(t, "/api/v1/staff/catalog/listings/"+listing+"/decision", staffLogin(t, "catalog_editor"), map[string]any{"decision": "approved"}), 200, "approve listing")
	cust := newCustomer(t, "Zainab")
	number := deliverSellerOrder(t, seller, shopID, cust, listing)
	line := get(t, "/api/v1/me/orders/"+number, cust).get("fulfilments", "0", "lines", "0", "id").(string)

	r := post(t, "/api/v1/me/orders/"+number+"/returns", cust, map[string]any{"items": []any{map[string]any{"orderLineId": line, "quantity": 1}}, "reason": "not_as_described", "method": "store_dropoff"})
	expect(t, r, 201, "seller-item return")
	rma := r.str("rma")
	if r.str("status") != "approved" {
		t.Fatalf("seller items return like any other: %s", r.Raw)
	}
	shop := "/api/v1/seller/shops/" + shopID + "/returns"
	if l := get(t, shop, seller); !strings.Contains(string(l.Raw), rma) {
		t.Fatalf("seller sees the return: %s", l.Raw)
	}
	sr := get(t, shop+"/"+rma, seller)
	if sr.get("canDispute") != false || sr.num("charge") <= 0 || sr.num("refund") != 0 {
		t.Fatalf("before inspection: no dispute, charge shown, customer total hidden: %s", sr.Raw)
	}
	other, otherShop := approvedShop(t, "Bayo", "Ade")
	expect(t, get(t, "/api/v1/seller/shops/"+otherShop+"/returns/"+rma, other), 404, "another shop")

	wh := staffLogin(t, "warehouse_lead")
	expect(t, post(t, "/api/v1/staff/returns/"+rma+"/receive", wh, nil), 200, "receive")
	itemID := get(t, "/api/v1/staff/returns/"+rma, wh).get("items", "0", "id").(string)
	insp := "/api/v1/staff/returns/" + rma + "/items/" + itemID + "/inspection"
	expectCode(t, post(t, insp, wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "grade": "A", "disposition": "restock_new"}), 422, "return_to_seller", "restock a seller's item")
	expect(t, post(t, insp, wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "grade": "B", "disposition": "return_to_seller", "notes": "Box crushed, item works"}), 200, "inspect")

	sr = get(t, shop+"/"+rma, seller)
	if sr.get("canDispute") != true || sr.get("disputeUntil") == nil {
		t.Fatalf("dispute open after inspection: %s", sr.Raw)
	}
	d := post(t, shop+"/"+rma+"/dispute", seller, map[string]any{"reason": "It left us sealed; the customer opened it, so this is change of mind."})
	expect(t, d, 200, "dispute")
	if d.str("dispute", "reason") == "" || d.get("canDispute") != false {
		t.Fatalf("disputed: %s", d.Raw)
	}
	expectCode(t, post(t, shop+"/"+rma+"/dispute", seller, map[string]any{"reason": "again"}), 409, "already_disputed", "dispute twice")
	if l := get(t, "/api/v1/staff/returns?disputed=true", staffLogin(t, "support_lead")); !strings.Contains(string(l.Raw), rma) {
		t.Fatalf("staff see open disputes: %s", truncate(string(l.Raw)))
	}
	lead := staffLogin(t, "support_lead")
	expectCode(t, post(t, "/api/v1/staff/returns/"+rma+"/dispute-decision", lead, map[string]any{"decision": "upheld", "note": "Seal photos from the seller check out"}), 409, "refund_first", "uphold before the refund")

	expect(t, post(t, "/api/v1/staff/returns/"+rma+"/refund", wh, nil), 200, "raise refund")
	expect(t, post(t, "/api/v1/staff/finance/refunds/"+refundFor(t, rma)+"/approve", staffLogin(t, "finance_manager"), nil), 200, "approve refund")
	expectCode(t, post(t, "/api/v1/staff/returns/"+rma+"/dispute-decision", lead, map[string]any{"decision": "upheld", "note": " "}), 422, "note_required", "no note")
	up := post(t, "/api/v1/staff/returns/"+rma+"/dispute-decision", lead, map[string]any{"decision": "upheld", "note": "Seal photos from the seller check out"})
	expect(t, up, 200, "uphold")
	if up.str("dispute", "decision") != "upheld" {
		t.Fatalf("upheld: %s", up.Raw)
	}
	expectCode(t, post(t, "/api/v1/staff/returns/"+rma+"/dispute-decision", lead, map[string]any{"decision": "rejected", "note": "x"}), 409, "no_open_dispute", "decide twice")
	trialBalanced(t)
}

// A pickup return in a delivery zone gets a rider job: the rider photographs the item at
// collection and at the hub, and handing it in marks the return received.
func TestReturnRiderPickup(t *testing.T) {
	cust := newCustomer(t, "Kola")
	number, _ := buyAndPay(t, cust, seed.Extra["listing_p21"])
	warehouseMu.Lock()
	p := pickAndPack(t, number, nil)
	warehouseMu.Unlock()
	if p.str("handoff") != "rider" {
		t.Fatalf("a Lagos order gets a rider: %s", truncate(string(p.Raw)))
	}
	jobID, label := p.str("jobId"), p.str("parcel", "labelCode")
	disp := staffLogin(t, "dispatcher")
	_, _, _, _, riderID := account(t, "rider")
	rider := riderLogin(t)
	_ = post(t, "/api/v1/rider/shift/end", rider, nil)
	expect(t, post(t, "/api/v1/rider/shift/start", rider, map[string]any{"locationConsent": true, "latitude": 6.6018, "longitude": 3.3515}), 201, "start shift")
	defer post(t, "/api/v1/rider/shift/end", rider, nil)
	act := func(job, kind string, extra map[string]any) resp {
		body := map[string]any{"clientEventId": newUUID(), "kind": kind, "occurredAt": time.Now().UTC().Format(time.RFC3339)}
		for k, v := range extra {
			body[k] = v
		}
		return post(t, "/api/v1/rider/jobs/"+job+"/actions", rider, body)
	}
	expect(t, post(t, "/api/v1/dispatch/jobs/"+jobID+"/assign", disp, map[string]any{"riderId": riderID}), 200, "assign delivery")
	expect(t, act(jobID, "picked_up", map[string]any{"labelCode": label}), 200, "picked up")
	phone := get(t, "/api/v1/me", cust).str("phone")
	expect(t, act(jobID, "delivered", map[string]any{"otp": latestCodeContaining(t, phone, "delivery code")}), 200, "delivered")
	waitOrderStatus(t, cust, number, "delivered")

	line := get(t, "/api/v1/me/orders/"+number, cust).get("fulfilments", "0", "lines", "0", "id").(string)
	r := post(t, "/api/v1/me/orders/"+number+"/returns", cust, map[string]any{"items": []any{map[string]any{"orderLineId": line, "quantity": 1}}, "reason": "faulty", "method": "pickup"})
	expect(t, r, 201, "pickup return")
	rma := r.str("rma")
	if r.str("status") != "awaiting_pickup" || r.str("pickup", "status") != "unassigned" {
		t.Fatalf("a pickup return gets a rider job: %s", r.Raw)
	}
	var job string
	for _, j := range get(t, "/api/v1/dispatch/board?status=unassigned", disp).get("jobs").([]any) {
		if m := j.(map[string]any); m["rma"] == rma {
			job = m["id"].(string)
		}
	}
	if job == "" {
		t.Fatal("return pickup missing from the dispatch board")
	}
	expect(t, post(t, "/api/v1/dispatch/jobs/"+job+"/assign", disp, map[string]any{"riderId": riderID}), 200, "assign pickup")
	if j := get(t, "/api/v1/rider/jobs", rider); !strings.Contains(string(j.Raw), rma) {
		t.Fatalf("rider sees the return pickup: %s", truncate(string(j.Raw)))
	}
	expectCode(t, act(job, "picked_up", nil), 422, "photo_required", "collect without a photo")
	photo := func(name string) string {
		u := post(t, "/api/v1/files/uploads", rider, map[string]any{"purpose": "delivery_proof", "contentType": "image/jpeg", "sizeBytes": 90000, "name": name})
		expect(t, u, 201, "photo upload")
		expect(t, post(t, "/api/v1/files/"+u.str("fileId")+"/complete", rider, map[string]any{}), 200, "photo complete")
		return u.str("fileId")
	}
	expect(t, act(job, "picked_up", map[string]any{"photoFileId": photo("collected.jpg")}), 200, "collected")
	expect(t, act(job, "delivered", map[string]any{"photoFileId": photo("handed-in.jpg")}), 200, "handed in at the hub")
	waitFor(t, "return received", 15*time.Second, func() bool { return get(t, "/api/v1/me/returns/"+rma, cust).str("status") == "received" })
	if got := get(t, "/api/v1/me/returns/"+rma, cust); got.str("pickup", "status") != "delivered" || got.get("pickup", "handedInAt") == nil {
		t.Fatalf("pickup finished: %s", got.Raw)
	}
}
