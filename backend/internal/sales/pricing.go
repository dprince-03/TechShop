package sales

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/marketing"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/money"
)

// Pricing order (docs/backend.md §3): base price (tier price on wholesale) → flash price →
// the single best automatic promotion → at most one coupon (allocated across eligible lines
// by largest remainder) → delivery fee per seller → VAT (PLACEHOLDER rate) → priceHash.
// Consumer prices are VAT-inclusive; VAT is only computed on TechShop's own (first-party) lines.

// Line is one priced cart line.
type Line struct {
	CartItemID  uuid.UUID `json:"cartItemId"`
	ListingID   uuid.UUID `json:"listingId"`
	VariantID   uuid.UUID `json:"variantId"`
	ProductID   uuid.UUID `json:"productId"`
	ProductName string    `json:"productName"`
	ProductSlug string    `json:"productSlug"`
	VariantName string    `json:"variantName"`
	SKU         string    `json:"sku"`
	Condition   string    `json:"condition"`
	SellerID    uuid.UUID `json:"sellerId"`
	SellerName  string    `json:"sellerName"`
	FulfilledBy string    `json:"fulfilledBy"`
	Quantity    int32     `json:"quantity"`

	ListPriceKobo int64  `json:"listPriceKobo"`
	UnitPriceKobo int64  `json:"unitPriceKobo"`
	PriceSource   string `json:"priceSource" enum:"list,tier,flash"`
	GrossKobo     int64  `json:"grossKobo"`
	DiscountKobo  int64  `json:"discountKobo"`
	LineTotalKobo int64  `json:"lineTotalKobo"`
	VATKobo       int64  `json:"vatKobo"`

	Available int32  `json:"available"`
	Problem   string `json:"problem,omitempty" doc:"Why this line can't be bought as is (unavailable, sold_out, min_quantity, …)"`

	// Internal: not serialised.
	firstParty     bool
	categoryID     uuid.UUID
	categoryPath   []uuid.UUID
	flashPromo     *uuid.UUID
	flashPerUser   *int32
	techShopFunded int64
	sellerFunded   bool
	commissionBps  int32
	sellerState    string
	weightGrams    int32
}

// Group is the part of the order one seller handles (becomes a fulfilment).
type Group struct {
	SellerID        uuid.UUID `json:"sellerId"`
	SellerName      string    `json:"sellerName"`
	FulfilledBy     string    `json:"fulfilledBy"`
	DeliveryFeeKobo int64     `json:"deliveryFeeKobo"`
	ItemsKobo       int64     `json:"itemsKobo"`
	Lines           []int     `json:"-"`
}

// Quote is the full price of a cart, as shown on the cart and checkout pages.
type Quote struct {
	Lines         []Line   `json:"lines"`
	Groups        []Group  `json:"groups"`
	ItemsKobo     int64    `json:"itemsKobo" doc:"Sum of line gross amounts (VAT-inclusive)"`
	DiscountKobo  int64    `json:"discountKobo"`
	DeliveryKobo  int64    `json:"deliveryKobo"`
	VATKobo       int64    `json:"vatKobo" doc:"VAT included in the total (TechShop's own items)"`
	TotalKobo     int64    `json:"totalKobo"`
	Promotion     *string  `json:"promotion,omitempty" doc:"Name of the automatic promotion applied"`
	CouponCode    *string  `json:"couponCode,omitempty"`
	CouponKobo    int64    `json:"couponKobo"`
	FreeDelivery  bool     `json:"freeDelivery"`
	ShipState     string   `json:"shipState"`
	PriceHash     string   `json:"priceHash" doc:"Send back with POST /orders; a different hash means prices changed"`
	Problems      []string `json:"problems"`
	Purchasable   bool     `json:"purchasable"`
	coupon        *marketing.Coupon
	couponByLine  []int64
	deliverySaved int64
}

// PriceInput is what pricing needs besides the cart lines.
type PriceInput struct {
	UserID     *uuid.UUID
	Channel    string // market, app, wholesale
	ShipState  string // two-letter state code of the delivery address ("" = Lagos assumption)
	ShipLGA    string // local government area, for LGA-specific delivery zones
	CouponCode string
}

// Price computes a quote for cart lines. q is the read or transaction query set, so checkout
// prices with the same snapshot it writes with.
func (s *Service) Price(ctx context.Context, q *store.Queries, rows []store.SalesCartLinesRow, in PriceInput) (*Quote, error) {
	if in.ShipState == "" {
		in.ShipState = "LA"
	}
	qt := &Quote{ShipState: in.ShipState, Problems: []string{}, Lines: make([]Line, 0, len(rows)), Groups: []Group{}}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ListingID.String() < rows[j].ListingID.String() })

	for _, r := range rows {
		l := Line{CartItemID: r.ID, ListingID: r.ListingID, VariantID: r.VariantID, ProductID: r.ProductID, ProductName: r.ProductName, ProductSlug: r.ProductSlug,
			VariantName: r.VariantName, SKU: r.Sku, Condition: r.Condition, SellerID: r.SellerID, SellerName: r.SellerName, FulfilledBy: r.FulfilledBy,
			Quantity: r.Quantity, ListPriceKobo: r.PriceKobo, UnitPriceKobo: r.PriceKobo, PriceSource: "list",
			firstParty: r.SellerType == "first_party", categoryID: r.CategoryID}
		if r.SellerState != nil {
			l.sellerState = *r.SellerState
		}
		if r.WeightGrams != nil {
			l.weightGrams = *r.WeightGrams * r.Quantity
		}
		if r.SellerType == "first_party" {
			l.FulfilledBy = "techshop"
		}
		// Wholesale quantity breaks are for business accounts only.
		if in.Channel == "wholesale" {
			if tier, err := q.SalesTierPrice(ctx, store.SalesTierPriceParams{ListingID: r.ListingID, MinQuantity: r.Quantity}); err == nil && tier < l.UnitPriceKobo {
				l.UnitPriceKobo, l.PriceSource = tier, "tier"
			}
		}
		// Flash price: sign-in required, within the deal's stock and the per-person limit.
		if fl, err := q.SalesLiveFlashPrice(ctx, r.ListingID); err == nil && fl.PriceKobo < l.UnitPriceKobo {
			switch {
			case in.UserID == nil:
				// Flash prices need sign-in; guests see the list price.
			case fl.StockLimit != nil && fl.Claimed+r.Quantity > *fl.StockLimit:
				qt.Problems = append(qt.Problems, "The flash deal on "+r.ProductName+" has sold out; it's priced normally.")
			default:
				l.UnitPriceKobo, l.PriceSource = fl.PriceKobo, "flash"
				pid := fl.PromotionID
				l.flashPromo, l.flashPerUser = &pid, fl.PerUserLimit
			}
		} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		gross, err := money.Mul(l.UnitPriceKobo, int64(l.Quantity))
		if err != nil {
			return nil, err
		}
		l.GrossKobo = gross

		avail, err := q.SalesListingAvailable(ctx, r.ListingID)
		if err != nil {
			return nil, err
		}
		l.Available = avail
		switch {
		case r.ListingStatus != "active" || r.SellerStatus != "active":
			l.Problem = "unavailable"
		case avail <= 0:
			l.Problem = "sold_out"
		case avail < r.Quantity:
			l.Problem = "not_enough_stock"
		case r.MinOrderQuantity != nil && r.Quantity < *r.MinOrderQuantity:
			l.Problem = "min_quantity"
		}
		bps := int32(0)
		if !l.firstParty {
			bps, err = q.SalesCommissionBps(ctx, store.SalesCommissionBpsParams{SellerID: r.SellerID, CategoryID: r.CategoryID, DefaultBps: int32(s.d.Cfg.DefaultCommissionBps)})
			if err != nil {
				return nil, err
			}
		}
		l.commissionBps = bps
		l.categoryPath = s.mk.CategoryPath(ctx, q, r.CategoryID)
		qt.Lines = append(qt.Lines, l)
	}

	// Best automatic promotion (no stacking of automatic promotions).
	promos, err := s.mk.LivePromotions(ctx, q)
	if err != nil {
		return nil, err
	}
	var best *marketing.Promotion
	var bestParts []int64
	var bestValue int64
	for i := range promos {
		p := promos[i]
		if p.Kind == "free_delivery" {
			continue
		}
		parts := discountParts(qt.Lines, p)
		if v := sum(parts); v > bestValue {
			best, bestParts, bestValue = &promos[i], parts, v
		}
	}
	if best != nil {
		qt.Promotion = &best.Name
		for i := range qt.Lines {
			qt.Lines[i].DiscountKobo += bestParts[i]
			if best.FundedBy == "techshop" {
				qt.Lines[i].techShopFunded += bestParts[i]
			} else if bestParts[i] > 0 {
				qt.Lines[i].sellerFunded = true
			}
		}
	}
	for _, p := range promos {
		if p.Kind == "free_delivery" && anyApplies(qt.Lines, p) {
			qt.FreeDelivery = true
		}
	}

	// Coupon (one per order), on what is left after the automatic promotion.
	if in.CouponCode != "" {
		c, err := s.mk.ResolveCoupon(ctx, q, in.CouponCode, in.UserID)
		if err != nil {
			return nil, err
		}
		var after int64
		for _, l := range qt.Lines {
			after += l.GrossKobo - l.DiscountKobo
		}
		if after < c.MinOrderKobo {
			return nil, couponMinOrder(c.MinOrderKobo)
		}
		parts := discountParts(qt.Lines, c.Promotion)
		if c.Promotion.Kind == "free_delivery" {
			if !anyApplies(qt.Lines, c.Promotion) {
				return nil, couponNotApplicable()
			}
			qt.FreeDelivery = true
		} else if sum(parts) == 0 {
			return nil, couponNotApplicable()
		}
		qt.coupon, qt.couponByLine, qt.CouponKobo = c, parts, sum(parts)
		code := c.Code
		qt.CouponCode = &code
		for i := range qt.Lines {
			qt.Lines[i].DiscountKobo += parts[i]
			if c.Promotion.FundedBy == "techshop" {
				qt.Lines[i].techShopFunded += parts[i]
			}
		}
	}

	// Line totals and VAT (VAT-inclusive; first-party lines only — PLACEHOLDER treatment).
	byGroup := map[uuid.UUID]int{}
	for i := range qt.Lines {
		l := &qt.Lines[i]
		l.LineTotalKobo = l.GrossKobo - l.DiscountKobo
		if l.firstParty {
			_, l.VATKobo = money.SplitVAT(l.LineTotalKobo, s.d.Cfg.VATBps)
		}
		qt.ItemsKobo += l.GrossKobo
		qt.DiscountKobo += l.DiscountKobo
		qt.VATKobo += l.VATKobo
		gi, ok := byGroup[l.SellerID]
		if !ok {
			gi = len(qt.Groups)
			byGroup[l.SellerID] = gi
			qt.Groups = append(qt.Groups, Group{SellerID: l.SellerID, SellerName: l.SellerName, FulfilledBy: l.FulfilledBy})
		}
		g := &qt.Groups[gi]
		if l.FulfilledBy == "seller" {
			g.FulfilledBy = "seller" // a seller shipping any line ships the parcel
		}
		g.ItemsKobo += l.LineTotalKobo
		g.Lines = append(g.Lines, i)
		if l.Problem != "" {
			qt.Problems = append(qt.Problems, l.ProductName+": "+problemText(l.Problem))
		}
	}
	for i := range qt.Groups {
		g := &qt.Groups[i]
		var weight int32
		for _, i := range g.Lines {
			weight += qt.Lines[i].weightGrams
		}
		fee := s.deliveryFee(ctx, q, g.FulfilledBy, qt.Lines[g.Lines[0]].sellerState, in.ShipState, in.ShipLGA, weight)
		if qt.FreeDelivery {
			qt.deliverySaved += fee
		} else {
			g.DeliveryFeeKobo = fee
		}
		qt.DeliveryKobo += g.DeliveryFeeKobo
	}
	qt.TotalKobo = qt.ItemsKobo - qt.DiscountKobo + qt.DeliveryKobo
	qt.Purchasable = len(qt.Lines) > 0 && !hasLineProblem(qt.Lines)
	qt.PriceHash = priceHash(qt)
	return qt, nil
}

// DeliveryQuoter prices delivery from logistics zones and weight-band rates. ok=false means no
// zone covers the address, and pricing falls back to the flat config fees.
type DeliveryQuoter interface {
	Quote(ctx context.Context, q *store.Queries, fulfilledBy, state, lga string, weightGrams int32) (fee int64, ok bool)
}

// deliveryFee uses the zone rate for the destination when there is one; otherwise flat fees by
// who ships and whether the parcel crosses a state line.
func (s *Service) deliveryFee(ctx context.Context, q *store.Queries, fulfilledBy, fromState, toState, lga string, weight int32) int64 {
	if s.Delivery != nil {
		if fee, ok := s.Delivery.Quote(ctx, q, fulfilledBy, toState, lga, weight); ok {
			return fee
		}
	}
	if fromState == "" {
		fromState = "LA"
	}
	switch {
	case fromState != toState:
		return s.d.Cfg.DeliveryFeeInterstateKobo
	case fulfilledBy == "seller":
		return s.d.Cfg.DeliveryFeeVendorKobo
	default:
		return s.d.Cfg.DeliveryFeeTechShopKobo
	}
}

// discountParts computes a promotion's discount per line (0 for lines it doesn't cover).
// Flash-priced lines get no further discount, and a line with a seller-funded discount gets
// no TechShop-funded one on top (each line's discount has a single funder).
func discountParts(lines []Line, p marketing.Promotion) []int64 {
	parts := make([]int64, len(lines))
	weights := make([]int64, len(lines))
	var eligible int64
	for i, l := range lines {
		if l.PriceSource == "flash" || l.Problem != "" || l.sellerFunded || !p.Applies(l.ListingID, l.categoryPath) {
			continue
		}
		weights[i] = l.GrossKobo - l.DiscountKobo
		eligible += weights[i]
	}
	if eligible == 0 {
		return parts
	}
	var total int64
	switch p.Kind {
	case "percentage":
		total = money.Bps(eligible, int(p.Value*100))
	case "fixed_amount":
		total = min(p.Value, eligible)
	default:
		return parts
	}
	return money.Allocate(total, weights)
}

func anyApplies(lines []Line, p marketing.Promotion) bool {
	for _, l := range lines {
		if p.Applies(l.ListingID, l.categoryPath) {
			return true
		}
	}
	return false
}

func sum(xs []int64) int64 {
	var t int64
	for _, x := range xs {
		t += x
	}
	return t
}

func hasLineProblem(lines []Line) bool {
	for _, l := range lines {
		if l.Problem != "" {
			return true
		}
	}
	return false
}

func problemText(code string) string {
	switch code {
	case "unavailable":
		return "no longer available"
	case "sold_out":
		return "sold out"
	case "not_enough_stock":
		return "not enough stock for that quantity"
	case "min_quantity":
		return "below the minimum order quantity"
	}
	return code
}

// priceHash fingerprints everything the customer agreed to pay. Checkout reprices inside the
// transaction and rejects the order if the hash differs.
func priceHash(q *Quote) string {
	type l struct {
		L uuid.UUID
		Q int32
		U int64
		D int64
	}
	v := struct {
		Lines    []l
		Delivery int64
		Total    int64
		Coupon   *string
	}{Delivery: q.DeliveryKobo, Total: q.TotalKobo, Coupon: q.CouponCode}
	for _, x := range q.Lines {
		v.Lines = append(v.Lines, l{x.ListingID, x.Quantity, x.UnitPriceKobo, x.DiscountKobo})
	}
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:16])
}
