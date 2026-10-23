package aftersales

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/modules/ledger"
)

func TestWithinWindow(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	six, eight := now.Add(-6*24*time.Hour), now.Add(-8*24*time.Hour)
	if !withinWindow(&six, now) || withinWindow(&eight, now) || withinWindow(nil, now) {
		t.Fatal("7-day window")
	}
}

func TestReturnRate(t *testing.T) {
	// The return checkpoint flags a rate above 30 % (see migration 00032).
	cases := []struct {
		returns, orders int
		flagged         bool
	}{
		{0, 0, false},  // first order, first return
		{0, 5, false},  // no returns yet
		{1, 10, false}, // 2 of 10 = 20 %
		{3, 10, true},  // 4 of 10 = 40 %
		{2, 0, true},   // returns but no delivered orders in 90 days
	}
	for _, c := range cases {
		if got := returnRate(c.returns, c.orders) > 0.30; got != c.flagged {
			t.Errorf("returnRate(%d, %d) flagged = %v, want %v", c.returns, c.orders, got, c.flagged)
		}
	}
}

func TestLineRefundAndSerial(t *testing.T) {
	if lineRefund(90_000, 3, 2) != 60_000 || lineRefund(90_000, 3, 0) != 0 {
		t.Fatal("lineRefund")
	}
	imei := "356789104452214"
	if !serialMatches(" 3567 8910 4452 214 ", &imei, nil) || serialMatches("356789104452215", &imei, nil) || serialMatches("", &imei, nil) {
		t.Fatal("serialMatches")
	}
	if c, ok := stockCondition("restock_open_box"); !ok || c != "open_box" {
		t.Fatal("stockCondition")
	}
	if _, ok := stockCondition("write_off"); ok {
		t.Fatal("write_off isn't restocked")
	}
}

func TestChangeOfMindAndDeliveryRefund(t *testing.T) {
	if !changeOfMindAllowed("lighting", "smart-home", true) || changeOfMindAllowed("lighting", "smart-home", false) ||
		changeOfMindAllowed("audio", "accessories", true) {
		t.Fatal("changeOfMindAllowed")
	}
	if deliveryRefund("faulty", 2500, 1, 1) != 2500 || deliveryRefund("faulty", 2500, 1, 2) != 0 || deliveryRefund("change_of_mind", 2500, 1, 1) != 0 {
		t.Fatal("deliveryRefund")
	}
}

func TestUnderWarranty(t *testing.T) {
	d := time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC)
	m12, m0 := int16(12), int16(0)
	tech, none := "techshop", "none"
	if !underWarranty(&d, &m12, &tech, d.AddDate(0, 11, 0)) {
		t.Fatal("11 months into a 12-month warranty is covered")
	}
	if underWarranty(&d, &m12, &tech, d.AddDate(1, 0, 1)) {
		t.Fatal("past the end isn't covered")
	}
	if underWarranty(&d, &m0, &tech, d) || underWarranty(&d, &m12, &none, d) || underWarranty(nil, &m12, &tech, d) || underWarranty(&d, nil, &tech, d) {
		t.Fatal("no months, no provider or not delivered isn't covered")
	}
}

func TestRepairTransitions(t *testing.T) {
	if !repairCanMove("queued", "diagnosing") || !repairCanMove("awaiting_parts", "repairing") || !repairCanMove("repairing", "done") {
		t.Fatal("normal path")
	}
	if repairCanMove("done", "repairing") || repairCanMove("awaiting_parts", "done") || repairCanMove("unrepairable", "queued") {
		t.Fatal("finished jobs don't move; parts must arrive before done")
	}
}

func TestTradeInEstimate(t *testing.T) {
	cases := []struct {
		price int64
		cond  string
		want  int64
	}{
		{100_000_000, "like_new", 55_000_000}, // ₦1,000,000 → ₦550,000
		{100_000_000, "good", 45_000_000},
		{100_000_000, "fair", 35_000_000},
		{100_000_000, "faulty", 10_000_000},
		{89_999_900, "good", 40_450_000}, // ₦404,999.55 rounds down to ₦404,500
		{100_000_000, "mint", 0},
		{0, "good", 0},
	}
	for _, c := range cases {
		if got := tradeInEstimate(c.price, c.cond); got != c.want {
			t.Errorf("tradeInEstimate(%d, %s) = %d, want %d", c.price, c.cond, got, c.want)
		}
	}
}

func sum(ls []ledger.Line) (total int64, by map[string]int64) {
	by = map[string]int64{}
	for _, l := range ls {
		total += l.Amount
		by[l.Code] += l.Amount
	}
	return total, by
}

func TestRefundAllocation(t *testing.T) {
	seller := uuid.New()
	earning := int64(220_000) // the 2-unit line: ₦2,000 net + ₦200 TechShop-funded discount
	items := []refundShare{
		{FirstParty: true, LineTotal: 215_000, VAT: 15_000, LineQty: 1, Qty: 1},
		{SellerID: seller, LineTotal: 200_000, SellerEarning: &earning, CommissionBps: 1000, LineQty: 2, Qty: 1, CommissionPosted: true},
	}
	total, by := sum(refundAllocation(items, 50_000))
	if total != 215_000+100_000+50_000 {
		t.Fatalf("allocation must sum to the refund: %d", total)
	}
	if by[ledger.RevenueFirstParty] != 200_000 || by[ledger.LiabilityVAT] != 15_000 || by[ledger.RevenueDelivery] != 50_000 {
		t.Fatalf("TechShop's share: %v", by)
	}
	// One of two units: the seller gives back 110,000 earned less the 10,000 commission they paid;
	// TechShop's 10,000 discount on that unit comes off the promotion expense.
	if by[ledger.SellerPayable(seller)] != 100_000 || by[ledger.ExpensePromotions] != -10_000 || by[ledger.RevenueCommission] != 10_000 {
		t.Fatalf("seller's share: %v", by)
	}
}
