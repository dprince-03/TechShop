package ledger

import "testing"

func TestMergeCombinesAndDropsZero(t *testing.T) {
	got := merge([]Line{{"a", 100}, {"b", -60}, {"a", -100}, {"c", -40}, {"b", 0}})
	if len(got) != 2 || got[0] != (Line{"b", -60}) || got[1] != (Line{"c", -40}) {
		t.Fatalf("merge = %v", got)
	}
}

func TestValidCode(t *testing.T) {
	for _, c := range []string{"bank:operating", "seller_payable:123", "liability:vat"} {
		if !validCode(c) {
			t.Errorf("%s should be valid", c)
		}
	}
	for _, c := range []string{"bank:", "cash", "foo:bar", ""} {
		if validCode(c) {
			t.Errorf("%s should be invalid", c)
		}
	}
}
