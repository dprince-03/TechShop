//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"
)

// A brand-new account's first order over ₦1m is held after payment: hidden from the seller,
// a case in the risk queue, and the customer is told we're confirming it. Clearing releases it
// to the seller; blocking cancels and refunds.
func TestCheckoutRiskHoldClearAndBlock(t *testing.T) {
	t.Parallel()
	seller, shopID := approvedShop(t, "Uche", "Obi")
	c := post(t, "/api/v1/seller/shops/"+shopID+"/listings", seller, map[string]any{"variantId": seed.Extra["variant_p5"], "condition": "new", "price": 160_000_000, "sellerStock": 5})
	expect(t, c, 201, "listing")
	listing := c.str("listing", "id")
	expect(t, post(t, "/api/v1/staff/catalog/listings/"+listing+"/decision", staffLogin(t, "catalog_editor"), map[string]any{"decision": "approved"}), 200, "approve listing")
	orders := "/api/v1/seller/shops/" + shopID + "/orders"
	ra := staffLogin(t, "risk_analyst")

	held := func(first string) (cust, number, ref string) {
		cust = newCustomer(t, first)
		number, _ = buyAndPay(t, cust, listing)
		o := get(t, "/api/v1/me/orders/"+number, cust)
		if o.get("reviewHoldAt") == nil || o.str("status") != "paid" {
			t.Fatalf("a new account's ₦1.6m first order is held: %s", truncate(string(o.Raw)))
		}
		ref = riskCaseAbout(t, number)
		cs := get(t, "/api/v1/staff/risk/cases/"+ref, ra)
		if cs.str("kind") != "order" || cs.str("subject") != number || len(cs.get("trace", "fired").([]any)) < 2 || cs.get("trace", "features", "order_total_kobo") == nil {
			t.Fatalf("case with its decision trace: %s", cs.Raw)
		}
		return cust, number, ref
	}
	inQueue := func(number string) bool {
		return strings.Contains(string(get(t, orders+"?status=pending", seller).Raw), number)
	}

	cust, number, ref := held("Tobi")
	if inQueue(number) {
		t.Fatal("a held order must not reach the seller")
	}
	expect(t, post(t, "/api/v1/staff/risk/cases/"+ref+"/claim", ra, nil), 200, "claim")
	expect(t, post(t, "/api/v1/staff/risk/cases/"+ref+"/notes", ra, map[string]any{"note": "Called the customer; ID matches the delivery name"}), 200, "note")
	expect(t, post(t, "/api/v1/staff/risk/cases/"+ref+"/resolution", ra, map[string]any{"decision": "clear", "note": "Verified by phone"}), 200, "clear")
	waitFor(t, "order released", 15*time.Second, func() bool { return get(t, "/api/v1/me/orders/"+number, cust).get("reviewHoldAt") == nil })
	waitFor(t, "seller sees the order", 15*time.Second, func() bool { return inQueue(number) })

	cust2, number2, ref2 := held("Femi")
	expect(t, post(t, "/api/v1/staff/risk/cases/"+ref2+"/escalate", ra, map[string]any{"note": "Card from another country, address doesn't exist"}), 200, "escalate")
	expect(t, post(t, "/api/v1/staff/risk/cases/"+ref2+"/resolution", ra, map[string]any{"decision": "block", "note": "Stolen card confirmed by the bank"}), 200, "block")
	waitOrderStatus(t, cust2, number2, "cancelled")
	if inQueue(number2) {
		t.Fatal("a blocked order never reaches the seller")
	}
	trialBalanced(t)
}

// The rules editor: a draft from the active rules, a weight change, the test bench over recent
// decisions, and publishing with a fresh MFA. A risk analyst can't change rules.
// Not parallel: a published weight briefly changes checkout scoring for everyone.
func TestRiskRulesEditor(t *testing.T) {
	expect(t, post(t, "/api/v1/staff/risk/rule-sets/drafts", staffLogin(t, "risk_analyst"), map[string]any{"checkpoint": "checkout"}), 403, "analyst can't edit rules")
	// At least one checkout decision to replay (paid is enough; a warehouse wave may move it on).
	bench := newCustomer(t, "Bench")
	o := buy(t, bench, seed.Extra["listing_p21"], "paystack")
	fakePay(t, "paystack", o.str("payment", "reference"), "success")
	waitFor(t, "order paid", 20*time.Second, func() bool {
		return get(t, "/api/v1/me/orders/"+o.str("orderNumber"), bench).str("status") != "pending_payment"
	})
	admin := staffLoginFresh(t, "admin")
	sets := get(t, "/api/v1/staff/risk/rule-sets", admin)
	expect(t, sets, 200, "rule sets")
	var active map[string]any
	for _, s := range sets.List {
		if m := s.(map[string]any); m["checkpoint"] == "checkout" && m["status"] == "active" {
			active = m
		}
	}
	if active == nil || len(active["rules"].([]any)) < 2 {
		t.Fatalf("active checkout rules: %s", truncate(string(sets.Raw)))
	}
	d := post(t, "/api/v1/staff/risk/rule-sets/drafts", admin, map[string]any{"checkpoint": "checkout", "notes": "Trust repeat customers a little less"})
	expect(t, d, 201, "draft")
	id := d.str("id")
	defer post(t, "/api/v1/staff/risk/rule-sets/"+id+"/discard", admin, nil) // leave nothing behind if a step fails
	if d.str("status") != "draft" || d.num("version") <= active["version"].(float64) {
		t.Fatalf("draft is a newer version: %s", d.Raw)
	}
	expectCode(t, post(t, "/api/v1/staff/risk/rule-sets/drafts", admin, map[string]any{"checkpoint": "checkout"}), 409, "draft_exists", "second draft")
	var rule string
	for _, r := range d.get("rules").([]any) {
		if m := r.(map[string]any); m["key"] == "repeat_customer" {
			rule = m["id"].(string)
		}
	}
	u := post(t, "/api/v1/staff/risk/rule-sets/"+id+"/rules/"+rule, admin, map[string]any{"weight": -15, "enabled": true})
	expect(t, u, 200, "change weight")
	expectCode(t, post(t, "/api/v1/staff/risk/rule-sets/"+id+"/thresholds", admin, map[string]any{"step": 70, "hold": 40, "block": 90}), 422, "thresholds_invalid", "bands out of order")
	b := post(t, "/api/v1/staff/risk/rule-sets/"+id+"/test", admin, nil)
	expect(t, b, 200, "test bench")
	if b.num("decisions") < 1 || b.get("before") == nil || b.get("after") == nil {
		t.Fatalf("bench replays recent checkout decisions: %s", b.Raw)
	}
	p := post(t, "/api/v1/staff/risk/rule-sets/"+id+"/publish", admin, nil)
	expect(t, p, 200, "publish")
	if p.str("status") != "active" {
		t.Fatalf("published: %s", p.Raw)
	}
	expectCode(t, post(t, "/api/v1/staff/risk/rule-sets/"+id+"/rules/"+rule, admin, map[string]any{"weight": 10, "enabled": true}), 409, "not_a_draft", "edit a live rule set")
	// Put the launch weight back so other tests see the documented rule.
	d2 := post(t, "/api/v1/staff/risk/rule-sets/drafts", admin, map[string]any{"checkpoint": "checkout", "notes": "Restore"})
	for _, r := range d2.get("rules").([]any) {
		if m := r.(map[string]any); m["key"] == "repeat_customer" {
			expect(t, post(t, "/api/v1/staff/risk/rule-sets/"+d2.str("id")+"/rules/"+m["id"].(string), admin, map[string]any{"weight": -20, "enabled": true}), 200, "restore weight")
		}
	}
	expect(t, post(t, "/api/v1/staff/risk/rule-sets/"+d2.str("id")+"/publish", admin, nil), 200, "republish")
}
