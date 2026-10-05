package sales

import (
	"testing"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/marketing"
)

func TestDiscountPartsPercentageSumsExactly(t *testing.T) {
	cat := uuid.New()
	lines := []Line{
		{ListingID: uuid.New(), GrossKobo: 3333300, categoryPath: []uuid.UUID{cat}},
		{ListingID: uuid.New(), GrossKobo: 1000001, categoryPath: []uuid.UUID{cat}},
		{ListingID: uuid.New(), GrossKobo: 500000, categoryPath: []uuid.UUID{uuid.New()}}, // not targeted
	}
	p := marketing.Promotion{Kind: "percentage", Value: 10, CategoryIDs: []uuid.UUID{cat}}
	parts := discountParts(lines, p)
	if parts[2] != 0 {
		t.Fatalf("untargeted line got %d", parts[2])
	}
	want := (3333300 + 1000001) / 10 // Bps rounds half up: 433330.1 → 433330
	if got := parts[0] + parts[1]; got != int64(want) {
		t.Fatalf("total discount %d, want %d", got, want)
	}
}

func TestDiscountPartsSkipsFlashAndSellerFunded(t *testing.T) {
	lines := []Line{
		{ListingID: uuid.New(), GrossKobo: 100000, PriceSource: "flash"},
		{ListingID: uuid.New(), GrossKobo: 100000, sellerFunded: true, DiscountKobo: 5000},
		{ListingID: uuid.New(), GrossKobo: 100000},
	}
	parts := discountParts(lines, marketing.Promotion{Kind: "fixed_amount", Value: 999999})
	if parts[0] != 0 || parts[1] != 0 {
		t.Fatalf("flash or seller-funded line discounted: %v", parts)
	}
	if parts[2] != 100000 {
		t.Fatalf("fixed amount must be capped at the eligible total, got %d", parts[2])
	}
}

func TestPriceHashChangesWithAnything(t *testing.T) {
	q := &Quote{Lines: []Line{{ListingID: uuid.New(), Quantity: 1, UnitPriceKobo: 1000}}, DeliveryKobo: 250000, TotalKobo: 251000}
	h1 := priceHash(q)
	q.Lines[0].Quantity = 2
	if h1 == priceHash(q) {
		t.Fatal("quantity change must change the hash")
	}
	q.Lines[0].Quantity = 1
	if h1 != priceHash(q) {
		t.Fatal("hash must be deterministic")
	}
}
