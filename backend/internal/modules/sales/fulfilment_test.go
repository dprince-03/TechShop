package sales

import "testing"

func TestFulfilmentFlow(t *testing.T) {
	ok := [][2]string{{"pending", "picking"}, {"picking", "packed"}, {"packed", "handed_over"}, {"packed", "in_transit"},
		{"handed_over", "in_transit"}, {"in_transit", "delivered"}, {"in_transit", "failed"}, {"failed", "in_transit"}, {"failed", "returned"}, {"pending", "cancelled"}}
	for _, m := range ok {
		if !CanMove(m[0], m[1]) {
			t.Errorf("%s → %s should be allowed", m[0], m[1])
		}
	}
	bad := [][2]string{{"pending", "delivered"}, {"delivered", "in_transit"}, {"packed", "cancelled"}, {"cancelled", "pending"}, {"in_transit", "returned"}}
	for _, m := range bad {
		if CanMove(m[0], m[1]) {
			t.Errorf("%s → %s must be refused", m[0], m[1])
		}
	}
}
