//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---- helpers

func waitFor(t *testing.T, what string, timeout time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(400 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func defaultAddress(t *testing.T, tok string) string {
	t.Helper()
	r := get(t, "/api/v1/me/addresses", tok)
	expect(t, r, 200, "addresses")
	for _, a := range r.List {
		m := a.(map[string]any)
		if m["isDefault"] == true {
			return m["id"].(string)
		}
	}
	t.Fatal("no default address")
	return ""
}

func cartLines(r resp) []any {
	l, _ := r.get("lines").([]any)
	return l
}

func clearCart(t *testing.T, tok string) {
	t.Helper()
	c := get(t, "/api/v1/cart", tok)
	expect(t, c, 200, "cart")
	for _, l := range cartLines(c) {
		id := l.(map[string]any)["cartItemId"].(string)
		expect(t, call(t, req{method: "DELETE", path: "/api/v1/cart/items/" + id, token: tok}), 200, "remove line")
	}
}

func addToCart(t *testing.T, tok, listing string, qty int) resp {
	t.Helper()
	r := post(t, "/api/v1/cart/items", tok, map[string]any{"listingId": listing, "quantity": qty})
	expect(t, r, 200, "add to cart")
	return r
}

func preview(t *testing.T, tok, address, coupon string) resp {
	t.Helper()
	return post(t, "/api/v1/checkout/preview", tok, map[string]any{"addressId": address, "couponCode": coupon})
}

func newKey() string { return fmt.Sprintf("k-%d", time.Now().UnixNano()) }

func placeOrder(t *testing.T, tok, address, hash, provider, key, coupon string) resp {
	t.Helper()
	return call(t, req{method: "POST", path: "/api/v1/orders", token: tok, headers: map[string]string{"Idempotency-Key": key},
		body: map[string]any{"addressId": address, "priceHash": hash, "provider": provider, "couponCode": coupon}})
}

// buy puts one listing in an empty cart and places the order.
func buy(t *testing.T, tok, listing, provider string) resp {
	t.Helper()
	clearCart(t, tok)
	addToCart(t, tok, listing, 1)
	addr := defaultAddress(t, tok)
	pv := preview(t, tok, addr, "")
	expect(t, pv, 200, "preview")
	o := placeOrder(t, tok, addr, pv.str("priceHash"), provider, newKey(), "")
	expect(t, o, 201, "place order")
	return o
}

// fakePay makes the fake provider send a signed webhook (development-only route).
func fakePay(t *testing.T, provider, reference, outcome string) resp {
	t.Helper()
	r := post(t, "/dev/pay/"+provider+"/"+reference+"/"+outcome, "", nil)
	expect(t, r, 200, "fake provider "+outcome)
	return r
}

// buyAndPay places an order for one listing and pays it by card.
func buyAndPay(t *testing.T, tok, listing string) (number, paymentID string) {
	t.Helper()
	o := buy(t, tok, listing, "paystack")
	fakePay(t, "paystack", o.str("payment", "reference"), "success")
	number, paymentID = o.str("orderNumber"), o.str("payment", "paymentId")
	waitOrderStatus(t, tok, number, "paid")
	return number, paymentID
}

func waitOrderStatus(t *testing.T, tok, number, want string) resp {
	t.Helper()
	var last resp
	waitFor(t, "order "+number+" "+want, 20*time.Second, func() bool {
		last = get(t, "/api/v1/me/orders/"+number, tok)
		return last.str("status") == want
	})
	return last
}

func trialBalanced(t *testing.T) {
	t.Helper()
	fin := staffLogin(t, "finance_manager")
	tb := get(t, "/api/v1/staff/finance/trial-balance", fin)
	expect(t, tb, 200, "trial balance")
	if tb.get("balanced") != true {
		t.Fatalf("ledger out of balance: %s", truncate(string(tb.Raw)))
	}
}

// ---- tests

// Guest cart → merge on sign-in → preview → idempotent order → fake card payment → webhook
// dedupe → paid order with one sale journal; the ledger stays balanced.
func TestCheckoutPaymentAndLedger(t *testing.T) {
	g := post(t, "/api/v1/cart/items", "", map[string]any{"listingId": seed.Extra["listing_p14"], "quantity": 1})
	expect(t, g, 200, "guest add")
	token := g.Header.Get("X-Cart-Token")
	if token == "" || g.str("cartToken") != token {
		t.Fatalf("guest cart token missing: %s", truncate(string(g.Raw)))
	}
	gc := call(t, req{method: "GET", path: "/api/v1/cart", headers: map[string]string{"X-Cart-Token": token}})
	if n := len(cartLines(gc)); n != 1 {
		t.Fatalf("guest cart lines = %d", n)
	}

	cust := customerLogin(t, "customer")
	clearCart(t, cust)
	m := post(t, "/api/v1/cart/merge", cust, map[string]any{"cartToken": token})
	expect(t, m, 200, "merge")
	if n := len(cartLines(m)); n != 1 {
		t.Fatalf("merged cart lines = %d", n)
	}
	addToCart(t, cust, seed.Extra["listing_p3"], 1) // a marketplace seller → a second fulfilment

	addr := defaultAddress(t, cust)
	pv := preview(t, cust, addr, "")
	expect(t, pv, 200, "preview")
	items, disc, del, total := pv.num("itemsKobo"), pv.num("discountKobo"), pv.num("deliveryKobo"), pv.num("totalKobo")
	if total != items-disc+del || del <= 0 || len(pv.get("groups").([]any)) < 2 || pv.get("purchasable") != true {
		t.Fatalf("preview totals: %s", truncate(string(pv.Raw)))
	}

	expectCode(t, placeOrder(t, cust, addr, "deadbeefdeadbeef", "paystack", newKey(), ""), 409, "price_changed", "stale price hash")
	key := newKey()
	o := placeOrder(t, cust, addr, pv.str("priceHash"), "paystack", key, "")
	expect(t, o, 201, "place order")
	number, paymentID, ref := o.str("orderNumber"), o.str("payment", "paymentId"), o.str("payment", "reference")
	if !strings.Contains(o.str("payment", "checkoutUrl"), "/dev/pay/paystack/") || o.num("totalKobo") != total {
		t.Fatalf("order: %s", truncate(string(o.Raw)))
	}
	re := placeOrder(t, cust, addr, pv.str("priceHash"), "paystack", key, "")
	if re.str("orderNumber") != number || re.Header.Get("Idempotent-Replayed") != "true" {
		t.Fatalf("replay: %d replayed=%q %s", re.Status, re.Header.Get("Idempotent-Replayed"), truncate(string(re.Raw)))
	}
	if n := len(cartLines(get(t, "/api/v1/cart", cust))); n != 0 {
		t.Fatalf("cart should be empty after checkout, has %d lines", n)
	}

	forged := call(t, req{method: "POST", path: "/api/v1/webhooks/paystack", headers: map[string]string{"x-paystack-signature": "00"},
		raw: `{"event":"charge.success","data":{"id":1,"reference":"` + ref + `","status":"success","amount":1}}`})
	expectCode(t, forged, 401, "bad_signature", "forged webhook")

	sent := fakePay(t, "paystack", ref, "success")
	waitOrderStatus(t, cust, number, "paid")
	if st := get(t, "/api/v1/me/payments/"+paymentID, cust); st.str("status") != "succeeded" {
		t.Fatalf("payment status: %s", st.Raw)
	}
	// The provider retries the same delivery: acknowledged, nothing changes.
	dup := call(t, req{method: "POST", path: "/api/v1/webhooks/paystack", headers: map[string]string{sent.str("signatureHeader"): sent.str("signature")}, raw: sent.str("raw")})
	expect(t, dup, 200, "duplicate webhook")
	if dup.str("status") != "duplicate" {
		t.Fatalf("duplicate webhook not detected: %s", dup.Raw)
	}

	d := get(t, "/api/v1/me/orders/"+number, cust)
	if len(d.get("fulfilments").([]any)) < 2 || d.num("vatKobo") <= 0 {
		t.Fatalf("order detail: %s", truncate(string(d.Raw)))
	}
	fin := staffLogin(t, "finance_manager")
	js := get(t, "/api/v1/staff/finance/journals?kind=sale&limit=100", fin)
	expect(t, js, 200, "journals")
	count := 0
	for _, j := range js.get("items").([]any) {
		if j.(map[string]any)["postingKey"] == "payment:"+paymentID+":captured" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("want exactly one sale journal for the payment, got %d", count)
	}
	trialBalanced(t)

	// Staff see the order with the customer's phone masked.
	ops := staffLogin(t, "ops_manager")
	so := get(t, "/api/v1/staff/orders/"+number, ops)
	expect(t, so, 200, "staff order")
	if strings.Contains(so.str("contactPhone"), "1234567") {
		t.Fatalf("staff view should mask the phone: %s", so.str("contactPhone"))
	}
	// Guest tracking needs the matching phone.
	_, phone, _, _, _ := account(t, "customer")
	expect(t, get(t, "/api/v1/track/"+number+"?phone="+urlEncode(phone), ""), 200, "track with phone")
	expect(t, get(t, "/api/v1/track/"+number+"?phone=08000000000", ""), 404, "track with wrong phone")
}

// Two customers race for the last unit: exactly one order wins; the loser gets 409.
// Cancelling the winner's unpaid order releases the unit.
func TestLastUnitRace(t *testing.T) {
	editor := staffLogin(t, "catalog_editor")
	wh := staffLogin(t, "warehouse_lead")
	suffix := fmt.Sprint(time.Now().UnixNano())
	p := post(t, "/api/v1/staff/catalog/products", editor, map[string]any{"categorySlug": "accessories", "brandName": "Volt", "slug": "race-cable-" + suffix,
		"name": "Race test cable " + suffix, "status": "active", "variants": []any{map[string]any{"sku": "RACE-" + suffix, "name": "1m"}}})
	expect(t, p, 201, "create product")
	variant := p.get("variants").([]any)[0].(map[string]any)["id"].(string)
	l := post(t, "/api/v1/staff/catalog/listings", editor, map[string]any{"variantId": variant, "condition": "new", "price": 1_500_000, "fulfilledBy": "techshop"})
	expect(t, l, 201, "create listing")
	listing := l.str("id")
	if l.str("status") != "active" {
		t.Fatalf("listing not active: %s", truncate(string(l.Raw)))
	}
	expect(t, post(t, "/api/v1/staff/inventory/adjustments", wh, map[string]any{"warehouseId": seed.Extra["warehouseIkejaId"], "variantId": variant,
		"condition": "new", "delta": 1, "reason": "adjustment", "note": "race test"}), 200, "one unit")

	a, b := customerLogin(t, "customer"), customerLogin(t, "customer_b")
	type buyer struct{ tok, addr, hash string }
	var buyers []buyer
	for _, tok := range []string{a, b} {
		clearCart(t, tok)
		addToCart(t, tok, listing, 1)
		addr := defaultAddress(t, tok)
		pv := preview(t, tok, addr, "")
		expect(t, pv, 200, "preview")
		buyers = append(buyers, buyer{tok, addr, pv.str("priceHash")})
	}
	results := make([]resp, 2)
	var wg sync.WaitGroup
	for i, by := range buyers {
		wg.Add(1)
		go func(i int, by buyer) {
			defer wg.Done()
			results[i] = placeOrder(t, by.tok, by.addr, by.hash, "paystack", newKey(), "")
		}(i, by)
	}
	wg.Wait()
	won, lost := -1, -1
	for i, r := range results {
		switch {
		case r.Status == 201:
			won = i
		case r.Status == 409 && r.str("code") == "out_of_stock":
			lost = i
		default:
			t.Fatalf("unexpected result %d: %s", r.Status, truncate(string(r.Raw)))
		}
	}
	if won < 0 || lost < 0 {
		t.Fatalf("want one winner and one 409; got %d and %d", results[0].Status, results[1].Status)
	}
	// The winner cancels before paying → the unit is back and the other buyer gets it.
	number := results[won].str("orderNumber")
	expect(t, post(t, "/api/v1/me/orders/"+number+"/cancel", buyers[won].tok, map[string]any{"reason": "changed my mind"}), 200, "cancel unpaid")
	loser := buyers[lost]
	pv := preview(t, loser.tok, loser.addr, "")
	expect(t, pv, 200, "preview again")
	expect(t, placeOrder(t, loser.tok, loser.addr, pv.str("priceHash"), "paystack", newKey(), ""), 201, "released unit can be bought")
}

// Coupons: validation, per-customer limit, release on cancellation. Flash prices: signed-in only.
func TestCouponsAndFlashDeals(t *testing.T) {
	cust := customerLogin(t, "customer_b")
	clearCart(t, cust)
	addToCart(t, cust, seed.Extra["listing_p7"], 1)
	addr := defaultAddress(t, cust)
	expectCode(t, preview(t, cust, addr, "NOPE-NOT-A-CODE"), 422, "coupon_invalid", "unknown coupon")
	pv := preview(t, cust, addr, "WELCOME10")
	expect(t, pv, 200, "preview with coupon")
	if pv.num("couponKobo") <= 0 || pv.num("discountKobo") < pv.num("couponKobo") {
		t.Fatalf("coupon not applied: %s", truncate(string(pv.Raw)))
	}
	o := placeOrder(t, cust, addr, pv.str("priceHash"), "paystack", newKey(), "WELCOME10")
	expect(t, o, 201, "order with coupon")
	addToCart(t, cust, seed.Extra["listing_p7"], 1)
	expectCode(t, preview(t, cust, addr, "WELCOME10"), 422, "coupon_used", "coupon used once per customer")
	expect(t, post(t, "/api/v1/me/orders/"+o.str("orderNumber")+"/cancel", cust, map[string]any{"reason": "testing coupon release"}), 200, "cancel")
	expect(t, preview(t, cust, addr, "WELCOME10"), 200, "coupon usable again after cancellation")
	clearCart(t, cust)

	// Flash deal: guests see the list price, signed-in customers the flash price.
	flash := seed.Extra["flashListingId"]
	g := post(t, "/api/v1/cart/items", "", map[string]any{"listingId": flash, "quantity": 1})
	expect(t, g, 200, "guest flash add")
	if src := g.get("lines").([]any)[0].(map[string]any)["priceSource"]; src != "list" {
		t.Fatalf("guest should see list price, got %v", src)
	}
	c := addToCart(t, cust, flash, 1)
	line := cartLines(c)[0].(map[string]any)
	if line["priceSource"] != "flash" || line["unitPriceKobo"].(float64) >= line["listPriceKobo"].(float64) {
		t.Fatalf("signed-in flash price: %v", line)
	}
	clearCart(t, cust)
}

// Refunds: a cancelled paid order raises a system refund that finance approves; staff can't
// approve their own request (separation of duties); the ledger stays balanced.
func TestRefundsDualControl(t *testing.T) {
	cust := customerLogin(t, "customer")
	number, _ := buyAndPay(t, cust, seed.Extra["listing_p2"])
	expect(t, post(t, "/api/v1/me/orders/"+number+"/cancel", cust, map[string]any{"reason": "found it cheaper"}), 200, "cancel paid order")

	officer := staffLogin(t, "finance_officer")
	var refundID string
	waitFor(t, "system refund request", 15*time.Second, func() bool {
		rs := get(t, "/api/v1/staff/finance/refunds?status=requested", officer)
		for _, r := range rs.List {
			m := r.(map[string]any)
			if m["orderNumber"] == number {
				refundID = m["id"].(string)
				return true
			}
		}
		return false
	})
	// ₦2.15m is above the ₦500k officer threshold: a finance manager must approve it.
	expect(t, post(t, "/api/v1/staff/finance/refunds/"+refundID+"/approve", officer, nil), 403, "officer above threshold")
	expect(t, post(t, "/api/v1/staff/finance/refunds/"+refundID+"/approve", staffLogin(t, "finance_manager"), nil), 200, "manager approves system refund")
	waitOrderStatus(t, cust, number, "refunded")
	mine := get(t, "/api/v1/me/refunds", cust)
	if !strings.Contains(string(mine.Raw), number) || !strings.Contains(string(mine.Raw), "succeeded") {
		t.Fatalf("customer refunds: %s", truncate(string(mine.Raw)))
	}

	// Separation of duties on a staff-requested partial refund.
	number2, _ := buyAndPay(t, cust, seed.Extra["listing_p2"])
	req1 := call(t, req{method: "POST", path: "/api/v1/staff/finance/refunds", token: officer, headers: map[string]string{"Idempotency-Key": newKey()},
		body: map[string]any{"orderNumber": number2, "amountKobo": 100_000, "reason": "goodwill for a late delivery"}})
	expect(t, req1, 201, "request refund")
	id := req1.str("id")
	expectCode(t, post(t, "/api/v1/staff/finance/refunds/"+id+"/approve", officer, nil), 403, "separation_of_duties", "approve own request")
	expect(t, post(t, "/api/v1/staff/finance/refunds/"+id+"/approve", staffLogin(t, "finance_manager"), nil), 200, "manager approves")
	waitOrderStatus(t, cust, number2, "partially_refunded")
	// More than what's left can't be requested.
	over := call(t, req{method: "POST", path: "/api/v1/staff/finance/refunds", token: officer, headers: map[string]string{"Idempotency-Key": newKey()},
		body: map[string]any{"orderNumber": number2, "amountKobo": 999_999_999_999, "reason": "too much money"}})
	expectCode(t, over, 422, "refund_too_large", "refund above captured")
	// Support can't approve refunds at all.
	expect(t, post(t, "/api/v1/staff/finance/refunds/"+id+"/approve", staffLogin(t, "support_agent"), nil), 403, "support can't approve")
	trialBalanced(t)
}

// Pay by transfer: a short transfer keeps the order pending; the rest completes it.
func TestTransferUnderpayment(t *testing.T) {
	cust := customerLogin(t, "customer_b")
	o := buy(t, cust, seed.Extra["listing_p16"], "moniepoint")
	acct := o.str("payment", "transfer", "accountNumber")
	if len(acct) != 10 {
		t.Fatalf("transfer account: %s", truncate(string(o.Raw)))
	}
	number, ref := o.str("orderNumber"), o.str("payment", "reference")
	fakePay(t, "moniepoint", ref, "underpay")
	time.Sleep(2 * time.Second)
	if s := get(t, "/api/v1/me/orders/"+number, cust).str("status"); s != "pending_payment" {
		t.Fatalf("underpaid order should stay pending, is %s", s)
	}
	fakePay(t, "moniepoint", ref, "success")
	waitOrderStatus(t, cust, number, "paid")
	trialBalanced(t)
}

// Unpaid orders expire (after asking the provider) and their stock is released. The e2e
// stack runs with ORDER_PAYMENT_TTL=100s; this test runs in parallel with the other slow one.
func TestUnpaidOrderExpires(t *testing.T) {
	t.Parallel()
	cust := customerLogin(t, "customer_b")
	o := buy(t, cust, seed.Extra["listing_p8"], "paystack")
	number := o.str("orderNumber")
	waitFor(t, "order expiry", 4*time.Minute, func() bool {
		return get(t, "/api/v1/me/orders/"+number, cust).str("status") == "cancelled"
	})
	d := get(t, "/api/v1/me/orders/"+number, cust)
	pays := d.get("payments").([]any)
	if pays[0].(map[string]any)["status"] != "cancelled" {
		t.Fatalf("payment should be cancelled: %s", truncate(string(d.Raw)))
	}
}

// Campaigns: the creator can't approve their own campaign; an approved one is sent to a
// recipient snapshot with a holdout.
func TestCampaignApproval(t *testing.T) {
	t.Parallel()
	mm := staffLogin(t, "marketing_manager")
	cnt := post(t, "/api/v1/staff/marketing/segments/preview", mm, map[string]any{"states": []string{"LA"}})
	expect(t, cnt, 200, "segment preview")
	if cnt.num("count") < 1 {
		t.Fatalf("Lagos audience should include the seeded customers: %s", cnt.Raw)
	}
	seg := post(t, "/api/v1/staff/marketing/segments", mm, map[string]any{"name": "e2e Lagos " + newKey(), "rules": map[string]any{"states": []string{"LA"}}})
	expect(t, seg, 201, "create segment")
	c := post(t, "/api/v1/staff/marketing/campaigns", mm, map[string]any{"name": "e2e campaign " + newKey(), "segmentId": seg.str("id"), "channels": []string{"push"},
		"holdoutPct": 0, "linkUrl": "https://techshop.ng/deals", "exclusions": map[string]any{},
		"variants": []any{map[string]any{"label": "A", "channel": "push", "body": "Hi {{first_name}}, deals inside: {{link}}", "sharePct": 100}}})
	expect(t, c, 201, "create campaign")
	id := c.str("id")
	expect(t, post(t, "/api/v1/staff/marketing/campaigns/"+id+"/submit", mm, nil), 200, "submit")
	expectCode(t, post(t, "/api/v1/staff/marketing/campaigns/"+id+"/approve", mm, nil), 403, "separation_of_duties", "creator approves")
	expect(t, post(t, "/api/v1/staff/marketing/campaigns/"+id+"/approve", staffLogin(t, "marketing_exec"), nil), 403, "exec lacks marketing.approve")
	expect(t, post(t, "/api/v1/staff/marketing/campaigns/"+id+"/approve", staffLogin(t, "admin"), nil), 200, "admin approves")
	var v resp
	waitFor(t, "campaign sent", 3*time.Minute, func() bool {
		v = get(t, "/api/v1/staff/marketing/campaigns/"+id, mm)
		return v.str("status") == "sent"
	})
	if v.num("stats", "recipients") < 1 {
		t.Fatalf("no recipients snapshotted: %s", truncate(string(v.Raw)))
	}
	// Unsigned unsubscribe links are refused.
	_, _, _, _, uid := account(t, "customer")
	expect(t, post(t, "/api/v1/unsubscribe?u="+uid+"&c=email&s=bad", "", nil), 403, "forged unsubscribe")
}
