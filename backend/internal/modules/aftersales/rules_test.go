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

func TestGoodStanding(t *testing.T) {
	cases := []struct {
		returns, orders, cases int
		want                   bool
	}{
		{0, 0, 0, true},   // first order, first return
		{0, 5, 0, true},   // no returns yet
		{1, 10, 0, true},  // 2 of 10 = 20 %
		{3, 10, 0, false}, // 4 of 10 = 40 %
		{0, 3, 1, false},  // open risk case
		{2, 0, 0, false},  // returns but no delivered orders in 90 days
	}
	for _, c := range cases {
		if got := goodStanding(c.returns, c.orders, c.cases); got != c.want {
			t.Errorf("goodStanding(%d, %d, %d) = %v, want %v", c.returns, c.orders, c.cases, got, c.want)
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
