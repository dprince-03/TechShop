//go:build e2e

package e2e

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// deliverSellerOrder buys one unit to a Kano address and takes it from the seller's pack to
// carrier delivery through the Wuse hub. It returns the order number.
func deliverSellerOrder(t *testing.T, seller, shopID, cust, listing string) string {
	t.Helper()
	a := post(t, "/api/v1/me/addresses", cust, map[string]any{"recipientName": "Musa Bello", "phone": "08031234572", "line1": "5 Murtala Mohammed Way", "city": "Kano", "stateCode": "KN"})
	expect(t, a, 201, "Kano address")
	clearCart(t, cust)
	addToCart(t, cust, listing, 1)
	pv := preview(t, cust, a.str("id"), "")
	expect(t, pv, 200, "preview")
	o := placeOrder(t, cust, a.str("id"), pv.str("priceHash"), "paystack", newKey(), "")
	expect(t, o, 201, "order")
	number := o.str("orderNumber")
	fakePay(t, "paystack", o.str("payment", "reference"), "success")
	waitOrderStatus(t, cust, number, "paid")
	p := acceptAndPack(t, seller, "/api/v1/seller/shops/"+shopID+"/orders", number)
	d := post(t, "/api/v1/staff/wms/dropoffs", staffLogin(t, "warehouse_lead"), map[string]any{"warehouseId": warehouseByCode(t, "WUSE"), "label": p.str("label")})
	expect(t, d, 201, "drop-off")
	for _, status := range []string{"in_transit", "delivered"} {
		ev := post(t, "/dev/carriers/fake/"+d.str("shipment", "trackingNumber")+"/"+status, "", nil)
		expect(t, call(t, req{method: "POST", path: "/api/v1/webhooks/carriers/fake", headers: map[string]string{"x-carrier-signature": ev.str("signature")}, raw: ev.str("raw")}), 200, status)
	}
	waitOrderStatus(t, cust, number, "delivered")
	return number
}

// payoutFor finds a seller's payout in a batch detail.
func payoutFor(b resp, shopID string) map[string]any {
	ps, _ := b.get("payouts").([]any)
	for _, p := range ps {
		if m := p.(map[string]any); m["sellerId"] == shopID {
			return m
		}
	}
	return nil
}

// Seller payouts (payments-finance.md §9): earnings wait for delivery and the return window;
// finance prepares a batch, someone else approves it (two managers above the threshold); the
// transfer pays the seller, posts the payout journal and the books stay balanced. A refused
// transfer gives the earnings back; holds keep a seller out.
func TestSellerPayoutBatch(t *testing.T) {
	seller, shopID := approvedShop(t, "Kemi", "Adeyemi")
	other, otherShop := approvedShop(t, "Uche", "Obi")
	cust := newCustomer(t, "Ada")
	var listings []string
	for _, s := range []struct{ tok, shop string }{{seller, shopID}, {other, otherShop}} {
		c := post(t, "/api/v1/seller/shops/"+s.shop+"/listings", s.tok, map[string]any{"variantId": seed.Extra["variant_p1"], "condition": "new", "price": 40_000_000, "sellerStock": 3})
		expect(t, c, 201, "listing")
		expect(t, post(t, "/api/v1/staff/catalog/listings/"+c.str("listing", "id")+"/decision", staffLogin(t, "catalog_editor"), map[string]any{"decision": "approved"}), 200, "approve listing")
		listings = append(listings, c.str("listing", "id"))
	}
	number := deliverSellerOrder(t, seller, shopID, cust, listings[0])
	deliverSellerOrder(t, other, otherShop, cust, listings[1])

	// The other shop's payout account is one the "bank" refuses (dev numbers starting with 8).
	acct := post(t, "/api/v1/seller/shops/"+otherShop+"/bank-accounts", other, map[string]any{"bankCode": "044", "accountNumber": fmt.Sprintf("8%09d", rand.Intn(1e9)), "accountName": "Uche Obi"})
	expect(t, acct, 201, "second account")
	expect(t, post(t, "/api/v1/seller/shops/"+otherShop+"/bank-accounts/"+acct.str("id")+"/default", other, nil), 204, "make default")

	// Delivered but inside the return window (and a new bank account): pending, nothing available.
	bal := "/api/v1/seller/shops/" + shopID + "/balance"
	b := get(t, bal, seller)
	expect(t, b, 200, "balance")
	if b.num("pending") <= 0 || b.num("available") != 0 || !strings.Contains(string(b.Raw), "bank_account_changed") {
		t.Fatalf("fresh delivery should be pending and the new account held: %s", b.Raw)
	}
	expect(t, get(t, "/api/v1/seller/shops/"+shopID+"/balance", other), 404, "another seller's balance")

	expect(t, post(t, "/dev/payouts/age/"+shopID, "", nil), 200, "age shop")
	expect(t, post(t, "/dev/payouts/age/"+otherShop, "", nil), 200, "age other shop")
	b = get(t, bal, seller)
	earned := b.num("available")
	// ₦400,000 less commission: something, never more than the price.
	if earned <= 0 || earned > 40_000_000 || b.num("pending") != 0 || b.get("onHold") != false {
		t.Fatalf("after the return window the earnings are available: %s", b.Raw)
	}

	// A hold keeps a seller out of the batch until it's released.
	admin := staffLoginFresh(t, "admin")
	officer := staffLogin(t, "finance_officer")
	h := post(t, "/api/v1/staff/finance/sellers/"+shopID+"/payout-holds", officer, map[string]any{"reason": "manual", "note": "Checking a complaint"})
	expect(t, h, 201, "hold")
	if hb := get(t, bal, seller); hb.get("onHold") != true || !strings.Contains(string(hb.Raw), "manual") {
		t.Fatalf("hold not shown: %s", hb.Raw)
	}
	held := post(t, "/api/v1/staff/finance/payout-batches", officer, nil)
	expect(t, held, 201, "prepare while held")
	if payoutFor(held, shopID) != nil || !strings.Contains(string(held.Raw), shopID) {
		t.Fatalf("held shop must be skipped with a reason: %s", truncate(string(held.Raw)))
	}
	expect(t, post(t, "/api/v1/staff/finance/payout-batches/"+held.str("batch", "id")+"/cancel", officer, nil), 200, "cancel held batch")
	expectCode(t, post(t, "/api/v1/staff/finance/payout-holds/"+h.str("id")+"/release", officer, nil), 403, "forbidden", "officer can't release holds")
	expect(t, post(t, "/api/v1/staff/finance/payout-holds/"+h.str("id")+"/release", admin, nil), 200, "release hold")

	// Prepare → submit → approve (never by the preparer).
	bt := post(t, "/api/v1/staff/finance/payout-batches", officer, nil)
	expect(t, bt, 201, "prepare")
	batch := bt.str("batch", "id")
	mine, theirs := payoutFor(bt, shopID), payoutFor(bt, otherShop)
	if mine == nil || theirs == nil || mine["amount"].(float64) != earned || strings.Contains(string(bt.Raw), "accountNumber\"") {
		t.Fatalf("batch should pay both shops their available balance, showing last 4 only: %s", truncate(string(bt.Raw)))
	}
	expectCode(t, post(t, "/api/v1/staff/finance/payout-batches/"+batch+"/approve", staffLoginFresh(t, "finance_manager"), nil), 409, "batch_not_submitted", "approve a draft")
	expect(t, post(t, "/api/v1/staff/finance/payout-batches/"+batch+"/submit", officer, nil), 200, "submit")
	expectCode(t, post(t, "/api/v1/staff/finance/payout-batches/"+batch+"/approve", officer, nil), 403, "forbidden", "officer can't approve")
	ap := post(t, "/api/v1/staff/finance/payout-batches/"+batch+"/approve", admin, nil)
	expect(t, ap, 200, "first approval")
	if ap.get("batch", "needsTwoApprovers") == true {
		if ap.get("released") != false {
			t.Fatalf("above the threshold one approval isn't enough: %s", ap.Raw)
		}
		expectCode(t, post(t, "/api/v1/staff/finance/payout-batches/"+batch+"/approve", admin, nil), 403, "separation_of_duties", "same approver twice")
		ap = post(t, "/api/v1/staff/finance/payout-batches/"+batch+"/approve", staffLoginFresh(t, "finance_manager"), nil)
		expect(t, ap, 200, "second approval")
	}
	if ap.get("released") != true {
		t.Fatalf("approved batch should release the transfers: %s", ap.Raw)
	}

	// The transfers run in the background: one paid, one refused by the bank.
	var detail resp
	waitFor(t, "batch "+batch+" done", 20*time.Second, func() bool {
		detail = get(t, "/api/v1/staff/finance/payout-batches/"+batch, officer)
		return detail.str("batch", "status") == "done"
	})
	if p := payoutFor(detail, shopID); p["status"] != "paid" {
		t.Fatalf("shop payout: %v", p)
	}
	if p := payoutFor(detail, otherShop); p["status"] != "failed" || p["failureReason"] == nil {
		t.Fatalf("refused payout: %v", p)
	}
	if b := get(t, bal, seller); b.num("available") != 0 {
		t.Fatalf("paid out: nothing left available: %s", b.Raw)
	}
	if b := get(t, "/api/v1/seller/shops/"+otherShop+"/balance", other); b.num("available") <= 0 {
		t.Fatalf("a refused payout gives the money back: %s", b.Raw)
	}
	hist := get(t, "/api/v1/seller/shops/"+shopID+"/payouts", seller)
	expect(t, hist, 200, "payout history")
	items, _ := hist.get("items").([]any)
	if len(items) != 1 || items[0].(map[string]any)["status"] != "paid" {
		t.Fatalf("history: %s", hist.Raw)
	}
	stmt := get(t, "/api/v1/seller/shops/"+shopID+"/payouts/"+items[0].(map[string]any)["id"].(string), seller)
	expect(t, stmt, 200, "statement")
	if !strings.Contains(string(stmt.Raw), number) {
		t.Fatalf("statement should list order %s: %s", number, stmt.Raw)
	}
	expect(t, get(t, "/api/v1/seller/shops/"+otherShop+"/payouts/"+items[0].(map[string]any)["id"].(string), other), 404, "another shop's statement")

	// A transfer webhook only prompts a re-check: replaying it twice pays and books nothing more.
	paidID := items[0].(map[string]any)["id"].(string)
	for i := 0; i < 2; i++ {
		expect(t, post(t, "/dev/transfers/"+paidID+"/success", "", nil), 200, "transfer webhook")
	}
	waitFor(t, "transfer webhook processed", 15*time.Second, func() bool {
		return get(t, "/api/v1/seller/shops/"+shopID+"/payouts/"+paidID, seller).str("status") == "paid"
	})
	if b := get(t, bal, seller); b.num("available") != 0 {
		t.Fatalf("a replayed webhook must not change the balance: %s", b.Raw)
	}

	// A second batch now has nothing to pay this shop twice.
	again := post(t, "/api/v1/staff/finance/payout-batches", officer, nil)
	expect(t, again, 201, "prepare again")
	if payoutFor(again, shopID) != nil {
		t.Fatalf("paid earnings offered twice: %s", truncate(string(again.Raw)))
	}
	expect(t, post(t, "/api/v1/staff/finance/payout-batches/"+again.str("batch", "id")+"/cancel", officer, nil), 200, "cancel")
	trialBalanced(t)
}
