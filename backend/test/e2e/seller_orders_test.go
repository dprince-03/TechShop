//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"
)

// newCustomer signs up a fresh market customer with a default Lagos address (its own cart, so
// parallel tests don't share one).
func newCustomer(t *testing.T, first string) string {
	t.Helper()
	phone := randomPhone()
	r := post(t, "/api/v1/auth/otp/request", "", map[string]any{"phone": phone})
	expect(t, r, 200, "customer otp")
	v := post(t, "/api/v1/auth/otp/verify", "", map[string]any{"challengeId": r.str("challengeId"), "code": latestCode(t, phone, ""), "audience": "market"})
	expect(t, v, 200, "customer verify")
	s := post(t, "/api/v1/auth/otp/signup", "", map[string]any{"signupToken": v.str("signupToken"), "firstName": first, "lastName": "Customer", "audience": "market", "deviceName": "e2e"})
	expect(t, s, 200, "customer signup")
	tok := s.str("accessToken")
	expect(t, post(t, "/api/v1/me/addresses", tok, map[string]any{"recipientName": first + " Adeyemi", "phone": phone, "line1": "15 Herbert Macaulay Way",
		"city": "Yaba", "stateCode": "LA", "isDefault": true}), 201, "address")
	return tok
}

// pendingOrderFor returns the fulfilment ID of an order in the shop's queue.
func pendingOrderFor(t *testing.T, tok, base, number string) string {
	t.Helper()
	var id string
	waitFor(t, "order "+number+" in the seller queue", 15*time.Second, func() bool {
		l := get(t, base+"?status=pending", tok)
		items, _ := l.get("items").([]any)
		for _, it := range items {
			if m := it.(map[string]any); m["orderNumber"] == number {
				id = m["id"].(string)
				return true
			}
		}
		return false
	})
	return id
}

func TestSellerOrdersAcceptPackCancelAndSLA(t *testing.T) {
	t.Parallel()
	seller, shopID := approvedShop(t, "Funmi", "Alade")
	listings := "/api/v1/seller/shops/" + shopID + "/listings"
	c := post(t, listings, seller, map[string]any{"variantId": seed.Extra["variant_p1"], "condition": "new", "price": 89_000_000, "sellerStock": 3})
	expect(t, c, 201, "listing")
	listing := c.str("listing", "id")
	expect(t, post(t, "/api/v1/staff/catalog/listings/"+listing+"/decision", staffLogin(t, "catalog_editor"), map[string]any{"decision": "approved"}), 200, "approve listing")

	cust := newCustomer(t, "Bola")
	orders := "/api/v1/seller/shops/" + shopID + "/orders"

	// 1. Accept and pack. The seller sees the items, the recipient's first name and city only.
	number, _ := buyAndPay(t, cust, listing)
	fid := pendingOrderFor(t, seller, orders, number)
	d := get(t, orders+"/"+fid, seller)
	expect(t, d, 200, "order detail")
	if d.str("recipientFirstName") != "Bola" || d.str("city") != "Yaba" || strings.Contains(string(d.Raw), "Herbert") || strings.Contains(string(d.Raw), "Adeyemi") {
		t.Fatalf("seller order view: %s", d.Raw)
	}
	if lines, _ := d.get("lines").([]any); len(lines) != 1 || d.get("acceptBy") == nil {
		t.Fatalf("lines and deadline: %s", d.Raw)
	}
	expectCode(t, post(t, orders+"/"+fid+"/pack", seller, nil), 409, "bad_transition", "pack before accept")
	a := post(t, orders+"/"+fid+"/accept", seller, nil)
	expect(t, a, 200, "accept")
	packBy, _ := time.Parse(time.RFC3339, a.str("packBy"))
	if a.str("status") != "accepted" || time.Until(packBy) < 20*time.Hour || time.Until(packBy) > 25*time.Hour {
		t.Fatalf("accept should set a 1-day packing deadline (handling days 1): %s", a.Raw)
	}
	expectCode(t, post(t, orders+"/"+fid+"/accept", seller, nil), 409, "bad_transition", "accept twice")
	expect(t, post(t, orders+"/"+fid+"/pack", seller, nil), 200, "pack")
	expectCode(t, post(t, orders+"/"+fid+"/cancel", seller, map[string]any{"reason": "out_of_stock"}), 409, "too_late_to_cancel", "cancel after packing")
	// Another seller can't see it.
	other, otherShop := approvedShop(t, "Segun", "Ojo")
	expect(t, get(t, "/api/v1/seller/shops/"+otherShop+"/orders/"+fid, other), 404, "other seller's order")

	// 2. The seller cancels: the customer's order is cancelled, a refund request is raised and
	// the stock comes back.
	number2, _ := buyAndPay(t, cust, listing)
	fid2 := pendingOrderFor(t, seller, orders, number2)
	cx := post(t, orders+"/"+fid2+"/cancel", seller, map[string]any{"reason": "out_of_stock", "note": "Last unit damaged"})
	expect(t, cx, 200, "seller cancel")
	if cx.str("status") != "cancelled" || !strings.Contains(cx.str("cancelledReason"), "out of stock") {
		t.Fatalf("seller cancel: %s", cx.Raw)
	}
	waitOrderStatus(t, cust, number2, "cancelled")
	officer := staffLogin(t, "finance_officer")
	waitFor(t, "refund request for "+number2, 15*time.Second, func() bool {
		return strings.Contains(string(get(t, "/api/v1/staff/finance/refunds?status=requested", officer).Raw), number2)
	})
	if l := get(t, listings+"/"+listing, seller); l.num("sellerStock") != 2 {
		t.Fatalf("stock after one packed and one cancelled order: %s", l.Raw)
	}

	// 3. Missed deadlines cancel automatically (dev route moves the deadline into the past).
	// Three breaches in 30 days send the shop's live listings back to review.
	number3, _ := buyAndPay(t, cust, listing)
	fid3 := pendingOrderFor(t, seller, orders, number3)
	expect(t, post(t, "/dev/sla/"+fid3+"/accept", "", nil), 200, "late accept")
	if o := get(t, orders+"/"+fid3, seller); o.str("status") != "cancelled" || !strings.Contains(o.str("cancelledReason"), "accept") {
		t.Fatalf("late accept: %s", o.Raw)
	}
	waitOrderStatus(t, cust, number3, "cancelled")
	for i, which := range []string{"accept", "pack"} {
		n, _ := buyAndPay(t, cust, listing)
		f := pendingOrderFor(t, seller, orders, n)
		if which == "pack" {
			expect(t, post(t, orders+"/"+f+"/accept", seller, nil), 200, "accept before late pack")
		}
		expect(t, post(t, "/dev/sla/"+f+"/"+which, "", nil), 200, "breach "+which)
		if st := get(t, listings+"/"+listing, seller).str("status"); (i == 0 && st != "active") || (i == 1 && st != "in_review") {
			t.Fatalf("after %d breaches the listing is %s", i+2, st)
		}
	}
	trialBalanced(t)
}
