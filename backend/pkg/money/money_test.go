package money

import "testing"

func TestSplitVATAlwaysSums(t *testing.T) {
	for _, gross := range []int64{1, 99, 7900000, 139000000, 107500000, 3} {
		net, vat := SplitVAT(gross, 750)
		if net+vat != gross {
			t.Fatalf("gross %d: net %d + vat %d != gross", gross, net, vat)
		}
	}
	net, vat := SplitVAT(10750000, 750)
	if net != 10000000 || vat != 750000 {
		t.Fatalf("₦107,500 incl. 7.5%% VAT: got net %d vat %d", net, vat)
	}
}

func TestAllocateSumsExactly(t *testing.T) {
	cases := []struct {
		total   int64
		weights []int64
	}{{100, []int64{1, 1, 1}}, {1000, []int64{3, 7}}, {7, []int64{1, 1, 1, 1}}, {0, []int64{1}}, {5, []int64{0, 0}}}
	for _, c := range cases {
		parts := Allocate(c.total, c.weights)
		var sum int64
		for _, p := range parts {
			sum += p
		}
		var w int64
		for _, x := range c.weights {
			w += x
		}
		if w > 0 && sum != c.total {
			t.Fatalf("Allocate(%d, %v) = %v sums to %d", c.total, c.weights, parts, sum)
		}
	}
}

func TestBpsAndMul(t *testing.T) {
	if got := Bps(10000000, 1000); got != 1000000 {
		t.Fatalf("10%% of ₦100,000 = %d", got)
	}
	if _, err := Mul(1<<62, 4); err != ErrOverflow {
		t.Fatal("expected overflow")
	}
}

func TestFormat(t *testing.T) {
	cases := map[int64]string{0: "₦0", 500000: "₦5,000", 123456789: "₦1,234,567.89", 99: "₦0.99", -250000: "-₦2,500", 100000000: "₦1,000,000"}
	for in, want := range cases {
		if got := Format(in); got != want {
			t.Errorf("Format(%d) = %q, want %q", in, got, want)
		}
	}
}
