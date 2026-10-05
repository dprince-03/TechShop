//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func names(r resp) []string {
	var out []string
	items, _ := r.get("items").([]any)
	for _, it := range items {
		out = append(out, it.(map[string]any)["name"].(string))
	}
	return out
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func TestCatalogBrowseAndSearch(t *testing.T) {
	cats := get(t, "/api/v1/categories", "")
	expect(t, cats, 200, "categories")
	if len(cats.List) < 8 {
		t.Fatalf("want ≥8 categories, got %d", len(cats.List))
	}
	phones := get(t, "/api/v1/categories/phones/products?limit=50", "")
	expect(t, phones, 200, "phones")
	if !contains(names(phones), "Nova X5 Pro 5G") {
		t.Fatalf("phones missing Nova X5: %v", names(phones))
	}
	// Synonym + price intent: tokunbo = UK-used, iphone = Orbit One, under 700k.
	r := get(t, "/api/v1/search?q=tokunbo%20iphone%20under%20700k", "")
	expect(t, r, 200, "search synonyms")
	if n := names(r); len(n) != 1 || n[0] != "Orbit One 13 (UK-used)" {
		t.Fatalf("tokunbo iphone under 700k → %v (understood %v)", n, r.get("understood"))
	}
	// Typo fallback.
	ty := get(t, "/api/v1/search?q=lapptop", "")
	expect(t, ty, 200, "typo")
	if ty.num("total") == 0 || ty.str("didYouMean") == "" {
		t.Fatalf("typo fallback failed: %s", truncate(string(ty.Raw)))
	}
	// Facets.
	if r2 := get(t, "/api/v1/search?q=nova", ""); r2.get("facets", "brand") == nil {
		t.Fatalf("facets missing: %s", truncate(string(r2.Raw)))
	}
	// Redirect.
	if rd := get(t, "/api/v1/search?q=cars", ""); rd.str("redirect") != "/c/cars" {
		t.Fatalf("redirect: %s", rd.Raw)
	}
	// Zero results are logged for staff insights.
	q := fmt.Sprintf("starlink%d", time.Now().Unix()%1000)
	z := get(t, "/api/v1/search?q="+q, "")
	expect(t, z, 200, "zero")
	staff := staffLogin(t, "catalog_editor")
	ins := get(t, "/api/v1/staff/search/insights", staff)
	expect(t, ins, 200, "insights")
	found := false
	for _, x := range ins.get("zeroResults").([]any) {
		if x.(map[string]any)["queryNorm"] == q {
			found = true
		}
	}
	if !found {
		t.Fatalf("zero-result query %q not in insights", q)
	}
	// Adding a synonym fixes it immediately.
	expect(t, post(t, "/api/v1/staff/search/synonyms", staff, map[string]any{"terms": []string{q}, "target": "power bank", "kind": "one_way"}), 201, "add synonym")
	time.Sleep(300 * time.Millisecond)
	if fx := get(t, "/api/v1/search?q="+q, ""); fx.num("total") == 0 {
		t.Fatalf("synonym not applied: %s", truncate(string(fx.Raw)))
	}
	// Deals are sorted by discount; best sellers respond.
	expect(t, get(t, "/api/v1/deals", ""), 200, "deals")
	expect(t, get(t, "/api/v1/best-sellers", ""), 200, "best sellers")
	sug := get(t, "/api/v1/search/suggest?q=nov", "")
	expect(t, sug, 200, "suggest")
	if len(sug.get("products").([]any)) == 0 {
		t.Fatalf("suggest products empty: %s", sug.Raw)
	}
}

func TestProductBuyBoxAndPriceGuard(t *testing.T) {
	staff := staffLogin(t, "catalog_editor")
	// A TechShop offer on the vendor's Nova A3 Lite: two offers → buy box decides.
	variant := seed.Extra["variant_p3"]
	l := post(t, "/api/v1/staff/catalog/listings", staff, map[string]any{"variantId": variant, "condition": "new", "price": 18_500_000, "fulfilledBy": "techshop", "warrantyMonths": 12})
	if l.Status != 201 && l.str("code") != "already_exists" {
		t.Fatalf("first-party listing: %s", l.Raw)
	}
	pd := get(t, "/api/v1/products/nova-a3-lite?shipTo=LA", "")
	expect(t, pd, 200, "product")
	offers := pd.get("offers").([]any)
	if len(offers) < 2 || offers[0].(map[string]any)["buyBox"] != true {
		t.Fatalf("buy box: %s", truncate(string(pd.Raw)))
	}
	// FCCPA: a "was" price above the 30-day maximum is rejected.
	if l.Status == 201 {
		id := l.str("id")
		bad := call(t, req{method: "PATCH", path: "/api/v1/staff/catalog/listings/" + id + "/price", token: staff, body: map[string]any{"price": 18_000_000, "compareAt": 99_000_000}})
		expectCode(t, bad, 422, "fake_was_price", "fake was price")
		ok := call(t, req{method: "PATCH", path: "/api/v1/staff/catalog/listings/" + id + "/price", token: staff, body: map[string]any{"price": 18_000_000, "compareAt": 18_500_000}})
		expect(t, ok, 200, "genuine was price")
	}
	expect(t, get(t, "/api/v1/products/does-not-exist", ""), 404, "unknown product")
}

func TestSavedContentFilesTracking(t *testing.T) {
	tok := customerLogin(t, "customer")
	pid := seed.Extra["product_p1"]
	s1 := post(t, "/api/v1/me/saved/"+pid, tok, nil)
	expect(t, s1, 200, "save")
	saved := get(t, "/api/v1/me/saved", tok)
	expect(t, saved, 200, "saved list")

	expect(t, get(t, "/api/v1/content/pages/market/delivery", ""), 200, "cms page")
	expect(t, get(t, "/api/v1/content/pages/market/nope", ""), 404, "missing page")
	expect(t, get(t, "/api/v1/content/help/market/track-order", ""), 200, "help")

	// Files: wrong type rejected; presigned upload then complete.
	expectCode(t, post(t, "/api/v1/files/uploads", tok, map[string]any{"purpose": "return_photo", "contentType": "application/x-msdownload", "sizeBytes": 100}), 422, "file_type_not_allowed", "bad type")
	up := post(t, "/api/v1/files/uploads", tok, map[string]any{"purpose": "return_photo", "contentType": "image/jpeg", "sizeBytes": 2048})
	expect(t, up, 201, "upload")
	expect(t, post(t, "/api/v1/files/"+up.str("fileId")+"/complete", tok, map[string]any{}), 200, "complete")
	expect(t, get(t, "/api/v1/files/"+up.str("fileId")+"/url", tok), 200, "owner url")
	other := customerLogin(t, "customer_b")
	expect(t, get(t, "/api/v1/files/"+up.str("fileId")+"/url", other), 404, "other customer can't read")

	// Tracking: dropped without consent, accepted with it.
	anon := uuid.NewString()
	ev := map[string]any{"anonymousId": anon, "app": "market", "events": []any{map[string]any{"type": "product_view", "occurredAt": time.Now().Format(time.RFC3339), "productId": pid}}}
	r := post(t, "/api/v1/events", "", ev)
	expect(t, r, 200, "events no consent")
	if r.num("dropped") != 1 {
		t.Fatalf("without consent the event must be dropped: %s", r.Raw)
	}
	expect(t, post(t, "/api/v1/consents", "", map[string]any{"anonymousId": anon, "purpose": "analytics", "granted": true, "policyVersion": "2026-10", "source": "web"}), 204, "consent")
	r = post(t, "/api/v1/events", "", ev)
	if r.num("accepted") != 1 {
		t.Fatalf("with consent the event must be accepted: %s", r.Raw)
	}
}

func TestInventoryOps(t *testing.T) {
	wh := staffLogin(t, "warehouse_lead")
	lead2 := staffLogin(t, "admin")
	variant := seed.Extra["variant_p1"]
	whID := seed.Extra["warehouseIkejaId"]
	lv := get(t, "/api/v1/staff/inventory/levels?variantId="+variant, wh)
	expect(t, lv, 200, "levels")
	// Can't go below zero.
	expectCode(t, post(t, "/api/v1/staff/inventory/adjustments", wh, map[string]any{"warehouseId": whID, "variantId": variant, "condition": "new", "delta": -100000, "reason": "write_off", "note": "test"}), 409, "insufficient_stock", "negative stock")
	expect(t, post(t, "/api/v1/staff/inventory/adjustments", wh, map[string]any{"warehouseId": whID, "variantId": variant, "condition": "new", "delta": 2, "reason": "adjustment", "note": "found 2"}), 200, "adjust +2")
	// Serialised receiving: invalid IMEI check digit rejected; valid accepted.
	expectCode(t, post(t, "/api/v1/staff/inventory/device-units", wh, map[string]any{"warehouseId": whID, "variantId": variant, "condition": "new", "units": []any{map[string]any{"imei": "356789104452211"}}}), 422, "invalid_imei", "bad imei")
	imei := luhnIMEI(fmt.Sprintf("35%012d", time.Now().UnixNano()%1e12))
	expect(t, post(t, "/api/v1/staff/inventory/device-units", wh, map[string]any{"warehouseId": whID, "variantId": variant, "condition": "new", "units": []any{map[string]any{"imei": imei}}}), 201, "receive imei")
	// Cycle count: counter can't approve their own variance.
	c := post(t, "/api/v1/staff/inventory/counts", wh, map[string]any{"warehouseId": whID, "lines": []any{map[string]any{"variantId": variant, "condition": "new", "counted": 1}}})
	expect(t, c, 201, "count")
	id := c.str("count", "id")
	expectCode(t, post(t, "/api/v1/staff/inventory/counts/"+id+"/approve", wh, nil), 403, "forbidden", "self approve count")
	expect(t, post(t, "/api/v1/staff/inventory/counts/"+id+"/approve", lead2, nil), 200, "other approves count")
}

func luhnIMEI(first14 string) string {
	sum := 0
	for i := 0; i < 14; i++ {
		d := int(first14[13-i] - '0')
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return first14 + fmt.Sprint((10-sum%10)%10)
}
