package sales

import (
	"testing"
	"time"
)

func TestFulfilmentFlow(t *testing.T) {
	ok := [][2]string{{"pending", "picking"}, {"picking", "packed"}, {"packed", "handed_over"}, {"packed", "in_transit"},
		{"handed_over", "in_transit"}, {"in_transit", "delivered"}, {"in_transit", "failed"}, {"failed", "in_transit"}, {"failed", "returned"}, {"pending", "cancelled"}, {"packed", "cancelled"}}
	for _, m := range ok {
		if !CanMove(m[0], m[1]) {
			t.Errorf("%s → %s should be allowed", m[0], m[1])
		}
	}
	bad := [][2]string{{"pending", "delivered"}, {"delivered", "in_transit"}, {"handed_over", "cancelled"}, {"cancelled", "pending"}, {"in_transit", "returned"}}
	for _, m := range bad {
		if CanMove(m[0], m[1]) {
			t.Errorf("%s → %s must be refused", m[0], m[1])
		}
	}
}

func TestHandoverDeadline(t *testing.T) {
	cases := []struct{ packed, want string }{
		{"2026-10-21T10:00:00+01:00", "2026-10-22T10:00:00+01:00"}, // Wednesday → Thursday
		{"2026-10-24T15:00:00+01:00", "2026-10-26T15:00:00+01:00"}, // Saturday → Sunday → Monday
		{"2026-10-24T23:30:00+00:00", "2026-10-25T23:30:00+00:00"}, // Sunday 00:30 WAT → Monday 00:30 WAT
	}
	for _, c := range cases {
		p, _ := time.Parse(time.RFC3339, c.packed)
		w, _ := time.Parse(time.RFC3339, c.want)
		if got := handoverDeadline(p); !got.Equal(w) {
			t.Errorf("packed %s: deadline %s, want %s", c.packed, got, c.want)
		}
	}
}
