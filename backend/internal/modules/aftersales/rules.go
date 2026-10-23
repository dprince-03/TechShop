package aftersales

import (
	"slices"
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

// withinWindow reports whether a delivery is still inside the return window.
func withinWindow(deliveredAt *time.Time, now time.Time) bool {
	return deliveredAt != nil && now.Sub(*deliveredAt) <= returnWindow
}

// returnRate is the share of delivered orders in 90 days that were returned, counting this
// request: the "return" checkpoint's high_return_rate rule compares it with 30 %.
func returnRate(returns90d, orders90d int) float64 {
	if returns90d == 0 {
		return 0 // a first return is always in good standing
	}
	if orders90d == 0 {
		return 1
	}
	return float64(returns90d+1) / float64(orders90d)
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

// Warranty rules (docs/plan.md 2026-10-23).

// warrantyEnds is the last day of cover: the listing's months from delivery.
func warrantyEnds(deliveredAt time.Time, months int16) time.Time {
	return deliveredAt.AddDate(0, int(months), 0)
}

// underWarranty reports whether a delivered item is still covered.
func underWarranty(deliveredAt *time.Time, months *int16, provider *string, now time.Time) bool {
	if deliveredAt == nil || months == nil || *months <= 0 || provider == nil || *provider == "none" {
		return false
	}
	return !now.After(warrantyEnds(*deliveredAt, *months))
}

// repairNext lists where a repair job can go from each status.
var repairNext = map[string][]string{
	"queued":         {"diagnosing", "awaiting_parts", "repairing", "done", "unrepairable"},
	"diagnosing":     {"awaiting_parts", "repairing", "done", "unrepairable"},
	"awaiting_parts": {"repairing", "unrepairable"},
	"repairing":      {"awaiting_parts", "done", "unrepairable"},
}

func repairCanMove(from, to string) bool { return slices.Contains(repairNext[from], to) }

// Trade-in rules (docs/plan.md 2026-10-23).

// tradeInFactor is the share of today's new price offered per condition. Swappa's trade-in
// quotes run 43–68 % of used resale value, and used devices resell below new, so the bands sit
// a little under those figures.
var tradeInFactor = map[string]int64{"like_new": 55, "good": 45, "fair": 35, "faulty": 10}

// tradeInEstimate prices a device from today's new price and its condition, rounded down to
// the nearest ₦500 so offers look deliberate rather than computed.
func tradeInEstimate(newPrice int64, condition string) int64 {
	f, ok := tradeInFactor[condition]
	if !ok || newPrice <= 0 {
		return 0
	}
	v := newPrice * f / 100
	return v - v%50_000
}

// tradeInCategories maps a device type to the catalogue categories that price it.
var tradeInCategories = map[string][]string{
	"phone":   {"phones"},
	"laptop":  {"laptops", "workstations"},
	"console": {"gaming"},
	// Tablets and smartwatches have no catalogue category yet: staff estimate them by hand.
	"tablet":     {},
	"smartwatch": {},
}

// tradeInOfferWindow is how long a customer has to accept an offer.
const tradeInOfferWindow = 7 * 24 * time.Hour
