package aftersales

import (
	"testing"
	"time"
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
