// Package money does integer-kobo arithmetic: VAT extraction, basis points and
// largest-remainder allocation (allocated parts always sum exactly to the total).
package money

import (
	"errors"
	"fmt"
	"math"
)

// Currency is the only currency TechShop trades in.
const Currency = "NGN"

// ErrOverflow is returned when a calculation would overflow int64 kobo.
var ErrOverflow = errors.New("money: overflow")

// Amount is the wire shape for money: integer kobo plus currency.
type Amount struct {
	Amount   int64  `json:"amount" doc:"Integer kobo (1/100 naira)"`
	Currency string `json:"currency" example:"NGN"`
}

// NGN builds an Amount in naira kobo.
func NGN(kobo int64) Amount { return Amount{Amount: kobo, Currency: Currency} }

// Mul multiplies kobo by a quantity, failing on overflow.
func Mul(kobo int64, qty int64) (int64, error) {
	if kobo != 0 && qty != 0 && (kobo > math.MaxInt64/abs(qty) || kobo < math.MinInt64/abs(qty)) {
		return 0, ErrOverflow
	}
	return kobo * qty, nil
}

// Bps returns amount × bps / 10,000, rounded half up.
func Bps(amount int64, bps int) int64 {
	return (amount*int64(bps) + 5000) / 10000
}

// SplitVAT splits a VAT-inclusive gross amount into net and VAT at the given rate (basis points).
// net + vat == gross always holds.
func SplitVAT(gross int64, vatBps int) (net, vat int64) {
	net = (gross*10000 + int64(10000+vatBps)/2) / int64(10000+vatBps)
	return net, gross - net
}

// Allocate splits total across weights using the largest-remainder method.
// The parts always sum exactly to total. Zero total weight returns all zeros.
func Allocate(total int64, weights []int64) []int64 {
	parts := make([]int64, len(weights))
	var sum int64
	for _, w := range weights {
		sum += w
	}
	if sum == 0 || total == 0 {
		return parts
	}
	type rem struct {
		idx int
		r   int64
	}
	rems := make([]rem, len(weights))
	var given int64
	for i, w := range weights {
		parts[i] = total * w / sum
		rems[i] = rem{i, total*w - parts[i]*sum}
		given += parts[i]
	}
	// Hand out the leftover kobo to the largest remainders (stable by index).
	for left := total - given; left > 0; left-- {
		best := -1
		for i := range rems {
			if best == -1 || rems[i].r > rems[best].r {
				best = i
			}
		}
		parts[rems[best].idx]++
		rems[best].r = -1
	}
	return parts
}

func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// Format renders kobo as naira for messages, e.g. 123456789 → "₦1,234,567.89" and 500000 → "₦5,000".
func Format(kobo int64) string {
	neg := kobo < 0
	if neg {
		kobo = -kobo
	}
	naira, k := kobo/100, kobo%100
	s := fmt.Sprintf("%d", naira)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if k != 0 {
		s += fmt.Sprintf(".%02d", k)
	}
	if neg {
		return "-₦" + s
	}
	return "₦" + s
}
