//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func freshIMEI(offset int64) string {
	return luhnIMEI(fmt.Sprintf("86%012d", (time.Now().UnixNano()+offset)%1e12))
}

// tradeInDevice returns the catalogue entry for the seeded phone variant.
func tradeInDevice(t *testing.T) map[string]any {
	t.Helper()
	l := get(t, "/api/v1/trade-ins/devices?type=phone", "")
	expect(t, l, 200, "trade-in devices")
	for _, d := range l.List {
		if m := d.(map[string]any); m["variantId"] == seed.Extra["variant_p1"] {
			return m
		}
	}
	t.Fatalf("phone p1 isn't priced: %s", truncate(string(l.Raw)))
	return nil
}

// A trade-in paid as store credit, refurbished into stock, and the credit spent at checkout.
func TestTradeInStoreCreditRefurbishAndSpend(t *testing.T) {
	t.Parallel()
	dev := tradeInDevice(t)
	est := dev["estimates"].(map[string]any)
	if est["like_new"].(float64) <= est["good"].(float64) || est["fair"].(float64) <= est["faulty"].(float64) {
		t.Fatalf("estimates fall with condition: %v", est)
	}
	cust := newCustomer(t, "Adaeze")
	imei := freshIMEI(17)
	expectCode(t, post(t, "/api/v1/me/trade-ins", cust, map[string]any{"deviceType": "phone", "variantId": seed.Extra["variant_p1"], "condition": "good", "imei": "123"}), 422, "imei_invalid", "bad imei")
	expectCode(t, post(t, "/api/v1/me/trade-ins", cust, map[string]any{"deviceType": "laptop", "variantId": seed.Extra["variant_p1"], "condition": "good"}), 422, "wrong_type", "phone as laptop")
	r := post(t, "/api/v1/me/trade-ins", cust, map[string]any{"deviceType": "phone", "variantId": seed.Extra["variant_p1"], "condition": "good", "imei": imei})
	expect(t, r, 201, "trade-in")
	ref := r.str("reference")
	if !strings.HasPrefix(ref, "TI-") || r.str("status") != "estimated" || r.num("estimate") != est["good"].(float64) || len(r.get("dropOff").([]any)) == 0 {
		t.Fatalf("instant estimate with drop-off stores: %s", r.Raw)
	}

	tech := staffLogin(t, "aftersales")
	base := "/api/v1/staff/trade-ins/" + ref
	store := seed.Extra["storeVIId"]
	expectCode(t, post(t, base+"/receive", tech, map[string]any{"storeId": store}), 422, "imei_required", "no imei")
	expectCode(t, post(t, base+"/receive", tech, map[string]any{"storeId": store, "imei": freshIMEI(29)}), 422, "imei_mismatch", "other phone")
	expect(t, post(t, base+"/receive", tech, map[string]any{"storeId": store, "imei": imei}), 200, "receive")
	offer := post(t, base+"/inspection", tech, map[string]any{"condition": "fair", "notes": "Scratched back, battery 81%, Google account removed"})
	expect(t, offer, 200, "inspect")
	if offer.str("status") != "offer_sent" || offer.num("offer") != est["fair"].(float64) || offer.get("offerExpiresAt") == nil {
		t.Fatalf("offer follows the inspected grade: %s", offer.Raw)
	}
	expect(t, post(t, "/api/v1/me/trade-ins/"+ref+"/accept", newCustomer(t, "Stranger"), map[string]any{"payoutMethod": "store_credit"}), 404, "someone else's trade-in")
	acc := post(t, "/api/v1/me/trade-ins/"+ref+"/accept", cust, map[string]any{"payoutMethod": "store_credit"})
	expect(t, acc, 200, "accept")
	if acc.str("status") != "paid" {
		t.Fatalf("store credit is paid at once: %s", acc.Raw)
	}
	credit := get(t, "/api/v1/me/store-credit", cust)
	if credit.num("balanceKobo") != offer.num("offer") {
		t.Fatalf("store credit = offer: %s", credit.Raw)
	}

	// The phone is refurbished before it can be sold.
	var job string
	for _, j := range get(t, "/api/v1/staff/warranty/repairs?purpose=refurbish&status=active", tech).List {
		if m := j.(map[string]any); m["tradeIn"] == ref {
			job = m["id"].(string)
		}
	}
	if job == "" {
		t.Fatal("no refurbishment job for the trade-in")
	}
	before := stockAt(t, seed.Extra["variant_p1"], "refurbished", store)
	expect(t, post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "repairing", "notes": "New back glass"}), 200, "refurbishing")
	done := post(t, "/api/v1/staff/warranty/repairs/"+job, tech, map[string]any{"status": "done"})
	expect(t, done, 200, "refurbished")
	if done.str("purpose") != "refurbish" || done.str("status") != "done" {
		t.Fatalf("refurb job: %s", done.Raw)
	}
	if after := stockAt(t, seed.Extra["variant_p1"], "refurbished", store); after != before+1 {
		t.Fatalf("refurbished phone in stock at the store: before %v, after %v", before, after)
	}

	// Spend the credit: a small order is paid at once; one bigger than the balance is refused.
	o := buy(t, cust, seed.Extra["listing_p21"], "store_credit")
	if o.str("payment", "status") != "succeeded" {
		t.Fatalf("store credit settles at once: %s", o.Raw)
	}
	waitOrderStatus(t, cust, o.str("orderNumber"), "paid")
	left := get(t, "/api/v1/me/store-credit", cust).num("balanceKobo")
	if left >= credit.num("balanceKobo") {
		t.Fatalf("credit should go down: before %v, after %v", credit.num("balanceKobo"), left)
	}
	clearCart(t, cust)
	addToCart(t, cust, seed.Extra["listing_p18"], 3)
	addr := defaultAddress(t, cust)
	pv := preview(t, cust, addr, "")
	expectCode(t, placeOrder(t, cust, addr, pv.str("priceHash"), "store_credit", newKey(), ""), 422, "store_credit_insufficient", "more than the credit")
	none := newCustomer(t, "NoCredit")
	addToCart(t, none, seed.Extra["listing_p21"], 1)
	naddr := defaultAddress(t, none)
	npv := preview(t, none, naddr, "")
	expectCode(t, placeOrder(t, none, naddr, npv.str("priceHash"), "store_credit", newKey(), ""), 422, "store_credit_insufficient", "no credit at all")
	trialBalanced(t)
}

// A trade-in paid by bank transfer: finance records it (not whoever priced the device). A device
// on the stolen list is refused and opens a case. Unpriced devices wait for a manual estimate.
func TestTradeInBankPayoutStolenAndManual(t *testing.T) {
	t.Parallel()
	cust := newCustomer(t, "Ibrahim")
	tech := staffLogin(t, "aftersales")
	store := seed.Extra["storeVIId"]

	m := post(t, "/api/v1/me/trade-ins", cust, map[string]any{"deviceType": "tablet", "brand": "Orbit", "model": "Tab 11", "storage": "128GB", "condition": "like_new"})
	expect(t, m, 201, "manual trade-in")
	if m.str("status") != "submitted" || m.get("estimate") != nil {
		t.Fatalf("no catalogue match waits for staff: %s", m.Raw)
	}
	expectCode(t, post(t, "/api/v1/staff/trade-ins/"+m.str("reference")+"/receive", tech, map[string]any{"storeId": store, "imei": freshIMEI(3)}), 409, "needs_estimate", "receive before estimate")
	e := post(t, "/api/v1/staff/trade-ins/"+m.str("reference")+"/estimate", tech, map[string]any{"estimate": 12_000_000})
	if e.str("status") != "estimated" || e.num("estimate") != 12_000_000 {
		t.Fatalf("manual estimate: %s", e.Raw)
	}
	expect(t, post(t, "/api/v1/me/trade-ins/"+m.str("reference")+"/cancel", cust, nil), 200, "cancel before drop-off")

	imei := freshIMEI(41)
	r := post(t, "/api/v1/me/trade-ins", cust, map[string]any{"deviceType": "phone", "variantId": seed.Extra["variant_p1"], "condition": "like_new", "imei": imei})
	ref := r.str("reference")
	base := "/api/v1/staff/trade-ins/" + ref
	expect(t, post(t, base+"/receive", tech, map[string]any{"storeId": store, "imei": imei}), 200, "receive")
	expect(t, post(t, base+"/inspection", tech, map[string]any{"condition": "good", "offer": 30_000_000}), 200, "inspect with an offer")
	expectCode(t, post(t, "/api/v1/me/trade-ins/"+ref+"/accept", cust, map[string]any{"payoutMethod": "bank"}), 422, "bank_required", "bank without details")
	acc := post(t, "/api/v1/me/trade-ins/"+ref+"/accept", cust, map[string]any{"payoutMethod": "bank", "bankCode": "044", "accountNumber": "0123456789", "accountName": "Ibrahim Customer"})
	expect(t, acc, 200, "accept by bank")
	if acc.str("status") != "accepted" || acc.str("accountLast4") != "6789" || acc.str("accountName") == "" {
		t.Fatalf("accepted, waiting for finance: %s", acc.Raw)
	}
	due := get(t, "/api/v1/staff/trade-ins?status=payout_due", tech)
	if !strings.Contains(string(due.Raw), ref) {
		t.Fatalf("payout due list: %s", truncate(string(due.Raw)))
	}
	expect(t, post(t, base+"/paid", tech, map[string]any{"reference": "FT123"}), 403, "technician can't pay out")
	fin := staffLoginFresh(t, "finance_manager")
	expectCode(t, post(t, base+"/paid", fin, map[string]any{"reference": " "}), 422, "reference_required", "no reference")
	paid := post(t, base+"/paid", fin, map[string]any{"reference": "FT2610231234"})
	expect(t, paid, 200, "mark paid")
	if paid.str("status") != "paid" || paid.str("payoutReference") != "FT2610231234" {
		t.Fatalf("paid: %s", paid.Raw)
	}

	stolen := freshIMEI(53)
	expect(t, post(t, "/api/v1/staff/risk/devices", staffLogin(t, "risk_analyst"), map[string]any{"imei": stolen, "reason": "reported_stolen", "source": "Police extract"}), 201, "block")
	s := post(t, "/api/v1/me/trade-ins", cust, map[string]any{"deviceType": "phone", "variantId": seed.Extra["variant_p1"], "condition": "good", "imei": stolen})
	expectCode(t, post(t, "/api/v1/staff/trade-ins/"+s.str("reference")+"/receive", tech, map[string]any{"storeId": store, "imei": stolen}), 409, "device_blocked", "stolen phone")
	if st := get(t, "/api/v1/staff/trade-ins/"+s.str("reference"), tech); st.str("status") != "blocked" {
		t.Fatalf("blocked trade-in: %s", st.Raw)
	}
	riskCaseAbout(t, s.str("reference"))
	trialBalanced(t)
}

// stockAt is the on-hand quantity of a variant and condition at one location.
func stockAt(t *testing.T, variant, condition, warehouse string) float64 {
	t.Helper()
	r := get(t, "/api/v1/staff/inventory/levels?variantId="+variant, staffLogin(t, "warehouse_lead"))
	for _, l := range r.List {
		m := l.(map[string]any)
		if m["warehouseId"] == warehouse && m["condition"] == condition {
			return m["onHand"].(float64)
		}
	}
	return 0
}
