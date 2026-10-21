package payouts

import (
	"slices"
	"testing"
	"time"
)

func TestNextPayoutDay(t *testing.T) {
	none := func(time.Time) bool { return false }
	at := func(s string) time.Time { v, _ := time.Parse(time.RFC3339, s); return v }
	cases := []struct {
		now, want string
		holiday   func(time.Time) bool
	}{
		{"2026-10-20T09:00:00+01:00", "2026-10-20", none},                                                               // Tuesday → same day
		{"2026-10-21T09:00:00+01:00", "2026-10-23", none},                                                               // Wednesday → Friday
		{"2026-10-24T09:00:00+01:00", "2026-10-27", none},                                                               // Saturday → Tuesday
		{"2026-10-22T23:30:00Z", "2026-10-23", none},                                                                    // Thursday 23:30 UTC is Friday in Lagos
		{"2026-12-25T08:00:00+01:00", "2026-12-29", func(d time.Time) bool { return d.Month() == 12 && d.Day() == 25 }}, // Christmas Friday → Tuesday
	}
	for _, c := range cases {
		if got := nextPayoutDay(at(c.now), c.holiday).Format("2006-01-02"); got != c.want {
			t.Errorf("%s: got %s, want %s", c.now, got, c.want)
		}
	}
}

func TestHoldReasons(t *testing.T) {
	now := time.Now()
	ok := sellerState{Status: "active", HasBank: true, BankVerified: true, BankAddedAt: now.Add(-48 * time.Hour), Available: 100}
	if r := holdReasons(ok, now); len(r) != 0 {
		t.Fatalf("clean seller held: %v", r)
	}
	bad := ok
	bad.Status, bad.BankAddedAt, bad.CredentialMoves, bad.OpenRiskCases, bad.Available, bad.StoredHolds = "pending_kyc", now.Add(-time.Hour), 1, 2, -5, []string{"manual"}
	want := []string{HoldKYC, HoldBankChanged, HoldCredentials, HoldRiskCase, HoldNegative, "manual"}
	if r := holdReasons(bad, now); !slices.Equal(r, want) {
		t.Fatalf("got %v, want %v", r, want)
	}
	noBank := ok
	noBank.HasBank = false
	if r := holdReasons(noBank, now); !slices.Equal(r, []string{HoldNoBank}) {
		t.Fatalf("no bank: %v", r)
	}
}

func TestSplit(t *testing.T) {
	// Nothing netted: the full eligible earnings.
	if a, p := split([]int64{6000, 4000}, 25000, 0, 15000); a != 10000 || !slices.Equal(p, []int64{6000, 4000}) {
		t.Fatalf("full: %d %v", a, p)
	}
	// A ₦3,000 refund was booked against the seller: the payout shrinks, spread pro rata.
	if a, p := split([]int64{6000, 4000}, 22000, 0, 15000); a != 7000 || p[0]+p[1] != 7000 || p[0] != 4200 {
		t.Fatalf("netted: %d %v", a, p)
	}
	// Money already in another payout isn't paid twice.
	if a, _ := split([]int64{10000}, 10000, 10000, 0); a != 0 {
		t.Fatalf("in flight: %d", a)
	}
	// A negative balance pays nothing.
	if a, _ := split([]int64{5000}, -100, 0, 0); a != 0 {
		t.Fatalf("negative: %d", a)
	}
}
