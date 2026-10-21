//go:build e2e

package e2e

import (
	"math"
	"strings"
	"testing"
	"time"
)

// deliveredByCarrier buys a TechShop item for a Kano address (outside rider zones, so it ships
// by carrier) and drives it to delivered. Returns the order number, fulfilment and line IDs.
func deliveredByCarrier(t *testing.T, cust, listing string) (number, fulfilmentID, lineID string) {
	t.Helper()
	a := post(t, "/api/v1/me/addresses", cust, map[string]any{"recipientName": "Halima Bello", "phone": "08031234570", "line1": "5 Murtala Mohammed Way", "city": "Kano", "stateCode": "KN"})
	expect(t, a, 201, "Kano address")
	clearCart(t, cust)
	addToCart(t, cust, listing, 1)
	pv := preview(t, cust, a.str("id"), "")
	expect(t, pv, 200, "preview")
	o := placeOrder(t, cust, a.str("id"), pv.str("priceHash"), "paystack", newKey(), "")
	expect(t, o, 201, "order to Kano")
	number = o.str("orderNumber")
	fakePay(t, "paystack", o.str("payment", "reference"), "success")
	waitOrderStatus(t, cust, number, "paid")
	p := pickAndPack(t, number, nil)
	tracking, label := p.str("shipment", "trackingNumber"), p.str("parcel", "labelCode")
	wh := staffLogin(t, "warehouse_lead")
	carrierID := ""
	for _, c := range get(t, "/api/v1/staff/wms/carriers", wh).List {
		if c.(map[string]any)["code"] == "fake" {
			carrierID = c.(map[string]any)["id"].(string)
		}
	}
	m := post(t, "/api/v1/staff/wms/manifests", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "carrierId": carrierID, "labels": []string{label}})
	expect(t, m, 201, "manifest")
	expect(t, post(t, "/api/v1/staff/wms/manifests/"+m.str("id")+"/sign", wh, map[string]any{"signedByName": "Courier"}), 200, "sign manifest")
	for _, status := range []string{"in_transit", "delivered"} {
		ev := post(t, "/dev/carriers/fake/"+tracking+"/"+status, "", nil)
		expect(t, call(t, req{method: "POST", path: "/api/v1/webhooks/carriers/fake", headers: map[string]string{"x-carrier-signature": ev.str("signature")}, raw: ev.str("raw")}), 200, status)
	}
	d := waitOrderStatus(t, cust, number, "delivered")
	f := d.get("fulfilments").([]any)[0].(map[string]any)
	return number, f["id"].(string), f["lines"].([]any)[0].(map[string]any)["id"].(string)
}

func TestReviewsRatingsAndModeration(t *testing.T) {
	t.Parallel()
	slug := "lumen-smart-bulb-4-pack"
	ada := newCustomer(t, "Adaora")
	_, fulfilment, line := deliveredByCarrier(t, ada, seed.Extra["listing_p21"])

	// A verified review publishes straight away; one review per product.
	v := post(t, "/api/v1/products/"+slug+"/reviews", ada, map[string]any{"rating": 5, "title": "Bright and easy", "body": "Set up in two minutes with the app and the colours are lovely. Ref " + newUUID()[:8], "orderLineId": line})
	expect(t, v, 201, "verified review")
	if v.str("status") != "published" || v.get("verified") != true {
		t.Fatalf("verified review: %s", v.Raw)
	}
	reviewID := v.str("id")
	expectCode(t, post(t, "/api/v1/products/"+slug+"/reviews", ada, map[string]any{"rating": 4}), 409, "already_reviewed", "second review")

	// Someone else can't claim Ada's purchase; their review is unverified, and one with a phone
	// number is held for moderation (reasons visible to staff only).
	ben := newCustomer(t, "Bennett")
	expectCode(t, post(t, "/api/v1/products/"+slug+"/reviews", ben, map[string]any{"rating": 1, "orderLineId": line}), 422, "not_your_delivery", "claim another's purchase")
	h := post(t, "/api/v1/products/"+slug+"/reviews", ben, map[string]any{"rating": 1, "body": "Rubbish. Buy cheaper from me, call 0803 123 4567"})
	expect(t, h, 201, "held review")
	if h.str("status") != "pending" || h.get("verified") != false || strings.Contains(string(h.Raw), "contact") {
		t.Fatalf("held review: %s", h.Raw)
	}
	pub := get(t, "/api/v1/products/"+slug+"/reviews", "")
	expect(t, pub, 200, "public reviews")
	if strings.Contains(string(pub.Raw), h.str("id")) || !strings.Contains(string(pub.Raw), reviewID) || !strings.Contains(string(pub.Raw), "Adaora C.") {
		t.Fatalf("public list: %s", truncate(string(pub.Raw)))
	}

	expect(t, get(t, "/api/v1/staff/reviews/queue", staffLogin(t, "dispatcher")), 403, "dispatcher review queue")
	mod := staffLogin(t, "catalog_editor")
	q := get(t, "/api/v1/staff/reviews/queue", mod)
	if !strings.Contains(string(q.Raw), h.str("id")) || !strings.Contains(string(q.Raw), "contact details") {
		t.Fatalf("queue should hold the review with its reason: %s", truncate(string(q.Raw)))
	}
	before := get(t, "/api/v1/products/"+slug+"/reviews", "")
	expect(t, post(t, "/api/v1/staff/reviews/"+h.str("id")+"/decision", mod, map[string]any{"decision": "publish"}), 200, "publish held review")

	// The unverified 1-star review counts half: new average = (old average × old weight + 1 × 0.5)
	// / (old weight + 0.5), where verified reviews weigh 1 and unverified 0.5.
	total, verified := before.num("summary", "total"), before.num("summary", "verified")
	weight := verified + (total-verified)*0.5
	want := (before.num("summary", "average")*weight + 0.5) / (weight + 0.5)
	sum := get(t, "/api/v1/products/"+slug+"/reviews", "")
	if math.Abs(sum.num("summary", "average")-want) > 0.011 || sum.num("summary", "total") != total+1 {
		t.Fatalf("summary: want average %.2f, got %s", want, truncate(string(sum.Raw)))
	}
	if only := get(t, "/api/v1/products/"+slug+"/reviews?verified=true", ""); strings.Contains(string(only.Raw), h.str("id")) {
		t.Fatalf("verified filter: %s", truncate(string(only.Raw)))
	}
	waitFor(t, "rating on the product page", 20*time.Second, func() bool {
		return get(t, "/api/v1/products/"+slug, "").num("ratingCount") == sum.num("summary", "total")
	})

	// Reports: once per person, not your own; publishing settles them (dismissed).
	expect(t, post(t, "/api/v1/reviews/"+reviewID+"/report", ben, map[string]any{"reason": "fake"}), 204, "report")
	expectCode(t, post(t, "/api/v1/reviews/"+reviewID+"/report", ben, map[string]any{"reason": "fake"}), 409, "already_reported", "report twice")
	expectCode(t, post(t, "/api/v1/reviews/"+reviewID+"/report", ada, map[string]any{"reason": "fake"}), 422, "own_review", "report own review")
	if q := get(t, "/api/v1/staff/reviews/queue", mod); !strings.Contains(string(q.Raw), reviewID) {
		t.Fatalf("reported review should be in the queue")
	}
	expect(t, post(t, "/api/v1/staff/reviews/"+reviewID+"/decision", mod, map[string]any{"decision": "publish"}), 200, "dismiss reports")
	if q := get(t, "/api/v1/staff/reviews/queue", mod); strings.Contains(string(q.Raw), reviewID) {
		t.Fatalf("settled review should leave the queue")
	}

	// Seller ratings: once per delivered part of your own order.
	expect(t, post(t, "/api/v1/me/deliveries/"+fulfilment+"/rating", ben, map[string]any{"rating": 1}), 404, "rate someone else's delivery")
	expect(t, post(t, "/api/v1/me/deliveries/"+fulfilment+"/rating", ada, map[string]any{"rating": 4, "comment": "Fast"}), 204, "rate seller")
	expectCode(t, post(t, "/api/v1/me/deliveries/"+fulfilment+"/rating", ada, map[string]any{"rating": 5}), 409, "already_rated", "rate twice")

	// A marketplace seller can't answer a review of an item it didn't ship.
	seller, shop := approvedShop(t, "Yemi", "Coker")
	expect(t, post(t, "/api/v1/seller/shops/"+shop+"/reviews/"+reviewID+"/response", seller, map[string]any{"text": "Thanks!"}), 404, "answer another seller's review")
	if l := get(t, "/api/v1/seller/shops/"+shop+"/reviews", seller); l.Status != 200 || len(l.get("items").([]any)) != 0 {
		t.Fatalf("new seller's reviews: %d %s", l.Status, l.Raw)
	}
}
