package aftersales

import (
	"strings"
	"time"
)

// Return rules (orders-fulfilment.md §7). Pure, so they're unit-tested without a database.

// returnWindow is how long after delivery a faulty, damaged, wrong or not-as-described item can
// be returned. Later problems go to warranty.
const returnWindow = 7 * 24 * time.Hour

// ourFault are the reasons that are TechShop's (or the seller's) fault: refunded in full,
// including the delivery fee, and eligible for auto-approval (FCCPA s.122; docs/plan.md 2026-10-21).
var ourFault = map[string]bool{"faulty": true, "damaged": true, "not_as_described": true, "wrong_item": true}

// noChangeOfMind lists category slugs that can't be returned on a change of mind for hygiene
// reasons (in-ear audio here; underwear and cosmetics if those categories are ever added).
var noChangeOfMind = map[string]bool{"audio": true, "underwear": true, "cosmetics": true}

// changeOfMindAllowed: sealed items outside the hygiene categories (category or its parent).
func changeOfMindAllowed(category, parent string, unopened bool) bool {
	return unopened && !noChangeOfMind[category] && !noChangeOfMind[parent]
}

// deliveryRefund is the delivery fee returned when a whole parcel comes back for an our-fault
// reason; a partial return keeps it (the parcel was still delivered for the other items).
func deliveryRefund(reason string, fee int64, returnedUnits, parcelUnits int32) int64 {
	if !ourFault[reason] || returnedUnits < parcelUnits {
		return 0
	}
	return fee
}

// highReturnRate is the share of delivered orders returned in 90 days above which a customer
// isn't auto-approved (trust-safety.md §3: "returns > 30 % of orders in 90 days").
const highReturnRate = 0.30

// withinWindow reports whether a delivery is still inside the return window.
func withinWindow(deliveredAt *time.Time, now time.Time) bool {
	return deliveredAt != nil && now.Sub(*deliveredAt) <= returnWindow
}

// goodStanding: no open risk case, and returns at most 30 % of delivered orders in 90 days
// (counting this request).
func goodStanding(returns90d, orders90d, openCases int) bool {
	if openCases > 0 {
		return false
	}
	if orders90d == 0 {
		return returns90d == 0
	}
	return float64(returns90d+1)/float64(orders90d) <= highReturnRate || returns90d == 0
}

// lineRefund is the money back for qty of a line: its price share (line total already net of
// discounts), so a discounted item refunds what was paid, not the list price.
func lineRefund(lineTotal int64, lineQty, qty int32) int64 {
	if lineQty <= 0 || qty <= 0 {
		return 0
	}
	return lineTotal * int64(qty) / int64(lineQty)
}

// serialMatches compares a scanned IMEI or serial with the unit sold (case and spaces ignored).
func serialMatches(scanned string, imei, serial *string) bool {
	norm := func(s string) string { return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", "")) }
	got := norm(scanned)
	if got == "" {
		return false
	}
	return (imei != nil && norm(*imei) == got) || (serial != nil && norm(*serial) == got)
}

// stockCondition is the condition a restocked item is sold as.
func stockCondition(disposition string) (string, bool) {
	switch disposition {
	case "restock_new":
		return "new", true
	case "restock_open_box":
		return "open_box", true
	default:
		return "", false
	}
}
