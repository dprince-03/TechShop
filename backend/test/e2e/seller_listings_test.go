//go:build e2e

package e2e

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// approvedShop creates a seller with a verified, active individual shop and returns the
// seller's token and the shop ID.
func approvedShop(t *testing.T, first, last string) (string, string) {
	t.Helper()
	tok := sellerSignup(t, first, last)
	shop := post(t, "/api/v1/seller/shops", tok, map[string]any{"type": "individual", "displayName": first + " Gadgets " + newUUID()[:6], "stateCode": "LA", "city": "Yaba"})
	expect(t, shop, 201, "shop")
	id := shop.str("id")
	base := "/api/v1/seller/shops/" + id
	docs := []map[string]any{{"kind": "id_front", "fileId": kycFile(t, tok, "id.jpg")}, {"kind": "selfie", "fileId": kycFile(t, tok, "selfie.jpg")}}
	k := post(t, base+"/kyc", tok, map[string]any{"idType": "passport", "idNumber": "B" + newUUID()[:8], "documents": docs})
	expect(t, k, 201, "kyc")
	expect(t, post(t, base+"/bank-accounts", tok, map[string]any{"bankCode": "058", "accountNumber": fmt.Sprintf("01%08d", rand.Intn(100000000)), "accountName": first + " " + last}), 201, "bank")
	expect(t, post(t, "/api/v1/staff/vendors/kyc/"+k.str("id")+"/review", staffLogin(t, "kyc_reviewer"), map[string]any{"decision": "approved"}), 200, "approve shop")
	return tok, id
}

// hasOffer reports whether a listing is among the offers on a product page.
func hasOffer(t *testing.T, slug, listingID string) bool {
	t.Helper()
	p := get(t, "/api/v1/products/"+slug, "")
	offers, _ := p.get("offers").([]any)
	for _, o := range offers {
		if o.(map[string]any)["listingId"] == listingID {
			return true
		}
	}
	return false
}

func TestSellerListingsAndModeration(t *testing.T) {
	t.Parallel()
	// An unverified shop can't list.
	pending := sellerSignup(t, "Tunde", "Ade")
	ps := post(t, "/api/v1/seller/shops", pending, map[string]any{"type": "individual", "displayName": "Tunde Pending " + newUUID()[:6], "stateCode": "OY", "city": "Ibadan"})
	expect(t, ps, 201, "pending shop")
	expectCode(t, post(t, "/api/v1/seller/shops/"+ps.str("id")+"/listings", pending, map[string]any{"variantId": seed.Extra["variant_p1"], "condition": "new", "price": 88_000_000, "sellerStock": 5}), 409, "shop_not_active", "unverified shop lists")

	tok, shopID := approvedShop(t, "Kemi", "Balogun")
	base := "/api/v1/seller/shops/" + shopID + "/listings"
	variant := seed.Extra["variant_p1"] // Nova X5 Pro 5G: one live first-party offer at ₦899,000

	// Automatic checks: no "was" price on a new listing (no 30-day history, FCCPA); used items
	// need a condition description; stock is required (schema) because sellers ship their own.
	expectCode(t, post(t, base, tok, map[string]any{"variantId": variant, "condition": "new", "price": 88_000_000, "compareAt": 95_000_000, "sellerStock": 5}), 422, "listing_rejected", "fake was price")
	expectCode(t, post(t, base, tok, map[string]any{"variantId": variant, "condition": "uk_used", "grade": "A", "price": 60_000_000, "sellerStock": 1}), 422, "condition_notes_required", "used without notes")
	expect(t, post(t, base, tok, map[string]any{"variantId": variant, "condition": "new", "price": 88_000_000}), 422, "no stock")

	c := post(t, base, tok, map[string]any{"variantId": variant, "condition": "new", "price": 88_000_000, "sellerStock": 5, "warrantyMonths": 12})
	expect(t, c, 201, "create listing")
	listing := c.str("listing", "id")
	if c.str("listing", "status") != "in_review" || !strings.Contains(string(c.Raw), "New seller") || c.str("listing", "productSlug") != "nova-x5-pro-5g" {
		t.Fatalf("new seller's listing should go to review: %s", c.Raw)
	}
	expectCode(t, post(t, base, tok, map[string]any{"variantId": variant, "condition": "new", "price": 87_000_000, "sellerStock": 2}), 409, "already_exists", "same variant and condition twice")

	// Moderation: only catalogue moderators; the queue shows the automatic reasons.
	expect(t, get(t, "/api/v1/staff/catalog/moderation", staffLogin(t, "dispatcher")), 403, "dispatcher moderation")
	mod := staffLogin(t, "catalog_editor")
	q := get(t, "/api/v1/staff/catalog/moderation", mod)
	expect(t, q, 200, "queue")
	if !strings.Contains(string(q.Raw), listing) {
		t.Fatalf("listing not in the moderation queue")
	}
	expect(t, post(t, "/api/v1/staff/catalog/listings/"+listing+"/decision", mod, map[string]any{"decision": "approved"}), 200, "approve listing")
	waitFor(t, "offer on the product page", 20*time.Second, func() bool { return hasOffer(t, "nova-x5-pro-5g", listing) })

	// Another seller can see the offer on the market but not manage it.
	other, _ := approvedShop(t, "Ifeanyi", "Obi")
	otherShops := get(t, "/api/v1/seller/shops", other)
	otherBase := "/api/v1/seller/shops/" + otherShops.List[0].(map[string]any)["id"].(string) + "/listings/"
	expect(t, get(t, otherBase+listing, other), 404, "other seller reads listing")
	expect(t, post(t, otherBase+listing+"/pause", other, nil), 404, "other seller pauses listing")
	expect(t, get(t, base+"/"+listing, other), 404, "other seller uses this shop")

	// A deep price cut sends a live listing back to review (bait-price guard).
	cut := call(t, req{method: "PATCH", path: base + "/" + listing + "/price", token: tok, body: map[string]any{"price": 40_000_000}})
	expect(t, cut, 200, "price cut")
	if cut.str("status") != "in_review" {
		t.Fatalf("deep cut should go to review: %s", cut.Raw)
	}
	d := get(t, base+"/"+listing, tok)
	if rv, _ := d.get("reviews").([]any); len(rv) == 0 || rv[0].(map[string]any)["decision"] != "auto_flagged" {
		t.Fatalf("seller should see the flag: %s", d.Raw)
	}
	// A "was" price above the 30-day maximum is refused.
	expectCode(t, call(t, req{method: "PATCH", path: base + "/" + listing + "/price", token: tok, body: map[string]any{"price": 85_000_000, "compareAt": 99_000_000}}), 422, "fake_was_price", "was above 30-day max")
	expect(t, post(t, "/api/v1/staff/catalog/listings/"+listing+"/decision", mod, map[string]any{"decision": "approved"}), 200, "approve again")

	// Pause and resume; stock.
	expect(t, post(t, base+"/"+listing+"/pause", tok, nil), 200, "pause")
	expectCode(t, post(t, base+"/"+listing+"/pause", tok, nil), 409, "bad_transition", "pause twice")
	waitFor(t, "paused offer hidden", 20*time.Second, func() bool { return !hasOffer(t, "nova-x5-pro-5g", listing) })
	expect(t, post(t, base+"/"+listing+"/resume", tok, nil), 200, "resume")
	if s := call(t, req{method: "PUT", path: base + "/" + listing + "/stock", token: tok, body: map[string]any{"stock": 0}}); s.Status != 200 || s.num("sellerStock") != 0 {
		t.Fatalf("stock: %d %s", s.Status, s.Raw)
	}

	// Changes requested → the seller fixes and resubmits.
	used := post(t, base, tok, map[string]any{"variantId": variant, "condition": "uk_used", "grade": "B", "conditionNotes": "Light scratches", "batteryHealthPct": 88, "price": 60_000_000, "sellerStock": 1})
	expect(t, used, 201, "used listing")
	usedID := used.str("listing", "id")
	expect(t, post(t, "/api/v1/staff/catalog/listings/"+usedID+"/decision", mod, map[string]any{"decision": "changes_requested", "reasons": []string{"photos"}, "note": "Describe the scratches and the box contents."}), 200, "request changes")
	if ld := get(t, base+"/"+usedID, tok); ld.str("status") != "changes_requested" || !strings.Contains(string(ld.Raw), "box contents") {
		t.Fatalf("seller should see the request: %s", ld.Raw)
	}
	re := call(t, req{method: "PUT", path: base + "/" + usedID, token: tok, body: map[string]any{"grade": "B", "conditionNotes": "Light scratches on the back; original box and charger", "batteryHealthPct": 88, "warrantyProvider": "seller", "warrantyMonths": 3, "handlingDays": 2}})
	expect(t, re, 200, "resubmit")
	if re.str("listing", "status") != "in_review" || re.str("listing", "warrantyProvider") != "seller" {
		t.Fatalf("resubmit: %s", re.Raw)
	}
	expectCode(t, call(t, req{method: "PUT", path: base + "/" + usedID, token: tok, body: map[string]any{"warrantyProvider": "seller", "warrantyMonths": 3, "handlingDays": 2}}), 409, "bad_transition", "resubmit while in review")

	// The listing list is the shop's own only, filterable by status.
	l := get(t, base+"?status=in_review", tok)
	expect(t, l, 200, "list")
	items, _ := l.get("items").([]any)
	if len(items) != 1 || items[0].(map[string]any)["id"] != usedID {
		t.Fatalf("in-review list: %s", truncate(string(l.Raw)))
	}
}
