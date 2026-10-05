// Package marketing runs promotions, flash deals, coupons, audiences, campaigns, journeys and
// tracked links (docs/messaging-marketing.md §4). Pricing (internal/sales) asks it which
// promotions apply; checkout asks it to claim flash units and redeem coupons atomically.
//
// Promotion values: percentage = whole percent (1–90); fixed_amount = kobo off the order.
package marketing

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// Service is the marketing module.
type Service struct {
	d *kit.Deps
	// Send delivers one message (wired to notify.Service.Send by the composition root).
	Send func(ctx context.Context, tx *uow.Tx, m Message) error
}

// Message is what campaigns and journeys hand to the notification service.
type Message struct {
	UserID              uuid.UUID
	To                  string // phone or email for sms/email; push resolves tokens from UserID
	Channel             string
	Template            string
	Data                map[string]any
	Key                 string
	CampaignID          *uuid.UUID
	JourneyEnrollmentID *uuid.UUID
	TemplateVersionID   *uuid.UUID
}

// New creates the marketing service.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "marketing" }

// Promotion is a live automatic promotion with its targets.
type Promotion struct {
	ID          uuid.UUID
	Name        string
	Kind        string // percentage, fixed_amount, free_delivery
	Value       int64
	FundedBy    string
	CategoryIDs []uuid.UUID
	ListingIDs  []uuid.UUID
}

// Applies reports whether the promotion covers a listing (by listing, by category or its
// parents, or sitewide when it has no targets).
func (p Promotion) Applies(listingID uuid.UUID, categoryPath []uuid.UUID) bool {
	if len(p.CategoryIDs) == 0 && len(p.ListingIDs) == 0 {
		return true
	}
	for _, id := range p.ListingIDs {
		if id == listingID {
			return true
		}
	}
	for _, c := range p.CategoryIDs {
		for _, a := range categoryPath {
			if c == a {
				return true
			}
		}
	}
	return false
}

// LivePromotions returns the automatic (non-coupon) promotions running now.
func (s *Service) LivePromotions(ctx context.Context, q *store.Queries) ([]Promotion, error) {
	rows, err := q.MarketingLiveAutoPromotions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Promotion, len(rows))
	for i, r := range rows {
		out[i] = Promotion{ID: r.ID, Name: r.Name, Kind: r.Kind, Value: val(r.Value), FundedBy: r.FundedBy, CategoryIDs: r.CategoryIds, ListingIDs: r.ListingIds}
	}
	return out, nil
}

// CategoryPath is a category and all its parents.
func (s *Service) CategoryPath(ctx context.Context, q *store.Queries, categoryID uuid.UUID) []uuid.UUID {
	ids, err := q.MarketingCategoryAncestors(ctx, categoryID)
	if err != nil {
		return []uuid.UUID{categoryID}
	}
	return ids
}

// Coupon is a validated coupon ready to price with.
type Coupon struct {
	Code         string
	CouponID     uuid.UUID
	CodeID       *uuid.UUID
	Promotion    Promotion
	MinOrderKobo int64
}

// ResolveCoupon validates a typed code for this customer. Errors are 422 problems the checkout
// shows next to the coupon field.
func (s *Service) ResolveCoupon(ctx context.Context, q *store.Queries, code string, userID *uuid.UUID) (*Coupon, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}
	bad := func(c, msg string) error { return httpx.Invalid(c, msg).Field("body.couponCode", msg, code) }
	r, err := q.MarketingFindCoupon(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, bad("coupon_invalid", "This coupon code doesn't exist.")
	}
	if err != nil {
		return nil, err
	}
	now := time.Now()
	switch {
	case r.PromotionStatus != "live" || now.Before(r.StartsAt) || !now.Before(r.EndsAt):
		return nil, bad("coupon_expired", "This coupon isn't active right now.")
	case r.CodeRedeemedAt != nil:
		return nil, bad("coupon_used", "This coupon code has already been used.")
	case r.CodeExpiresAt != nil && now.After(*r.CodeExpiresAt):
		return nil, bad("coupon_expired", "This coupon code has expired.")
	case r.MaxRedemptions != nil && r.RedeemedCount >= *r.MaxRedemptions:
		return nil, bad("coupon_exhausted", "This coupon has been fully claimed.")
	case r.Kind == "flash_price":
		return nil, bad("coupon_invalid", "This code can't be used as a coupon.")
	}
	if userID == nil {
		return nil, bad("sign_in_required", "Sign in to use a coupon.")
	}
	if r.IssuedToUserID != nil && *r.IssuedToUserID != *userID {
		return nil, bad("coupon_invalid", "This coupon code was issued to someone else.")
	}
	used, err := q.MarketingUserRedemptions(ctx, store.MarketingUserRedemptionsParams{CouponID: r.CouponID, UserID: userID})
	if err != nil {
		return nil, err
	}
	if used >= r.PerUserLimit {
		return nil, bad("coupon_used", "You've already used this coupon.")
	}
	c := &Coupon{Code: code, CouponID: r.CouponID, CodeID: r.CodeID,
		Promotion: Promotion{ID: r.PromotionID, Kind: r.Kind, Value: val(r.Value), FundedBy: r.FundedBy}}
	if r.MinOrderKobo != nil {
		c.MinOrderKobo = *r.MinOrderKobo
	}
	t, err := q.MarketingPromotionTargetsFor(ctx, r.PromotionID)
	if err == nil {
		c.Promotion.CategoryIDs, c.Promotion.ListingIDs = t.CategoryIds, t.ListingIds
	}
	return c, nil
}

// Redeem records a coupon use inside the checkout transaction. The counter update is
// conditional, so concurrent checkouts can never exceed max_redemptions.
func (s *Service) Redeem(ctx context.Context, tx *uow.Tx, c *Coupon, orderID, userID uuid.UUID, discount int64) error {
	if discount <= 0 {
		return httpx.Invalid("coupon_not_applicable", "This coupon doesn't apply to anything in your cart.")
	}
	if _, err := tx.Q.MarketingIncrementCoupon(ctx, c.CouponID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.Conflict("coupon_exhausted", "This coupon was just fully claimed. Remove it to continue.")
		}
		return err
	}
	if c.CodeID != nil {
		n, err := tx.Q.MarketingRedeemCode(ctx, *c.CodeID)
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.Conflict("coupon_used", "This coupon code has already been used.")
		}
	}
	return tx.Q.MarketingInsertRedemption(ctx, store.MarketingInsertRedemptionParams{CouponID: c.CouponID, CouponCodeID: c.CodeID, OrderID: orderID, UserID: &userID, DiscountKobo: discount})
}

// ClaimFlash takes flash-deal units for an order: per-user limit, then the conditional
// counter (claimed + qty <= stock_limit). 409 when the deal sold out.
func (s *Service) ClaimFlash(ctx context.Context, tx *uow.Tx, promotionID, listingID, userID, orderID uuid.UUID, qty int32, perUserLimit *int32) error {
	if perUserLimit != nil {
		had, err := tx.Q.MarketingUserFlashClaimed(ctx, store.MarketingUserFlashClaimedParams{PromotionID: promotionID, ListingID: listingID, UserID: userID})
		if err != nil {
			return err
		}
		if had+qty > *perUserLimit {
			return httpx.Conflict("flash_limit_reached", "You've reached the limit for this flash deal.")
		}
	}
	if _, err := tx.Q.MarketingClaimFlash(ctx, store.MarketingClaimFlashParams{Qty: qty, PromotionID: promotionID, ListingID: listingID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.Conflict("flash_sold_out", "This flash deal just sold out.")
		}
		return err
	}
	return tx.Q.MarketingInsertFlashClaim(ctx, store.MarketingInsertFlashClaimParams{PromotionID: promotionID, ListingID: listingID, UserID: userID, OrderID: orderID, Quantity: qty})
}

// ReleaseOrder gives back flash units and the coupon use when an order is cancelled or expires.
func (s *Service) ReleaseOrder(ctx context.Context, tx *uow.Tx, orderID uuid.UUID) error {
	if _, err := tx.Q.MarketingReleaseFlashClaims(ctx, orderID); err != nil {
		return err
	}
	red, err := tx.Q.MarketingReleaseRedemption(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := tx.Q.MarketingDecrementCoupon(ctx, red.CouponID); err != nil {
		return err
	}
	if red.CouponCodeID != nil {
		return tx.Q.MarketingUnredeemCode(ctx, *red.CouponCodeID)
	}
	return nil
}

// GoalReached exits a person from journeys waiting for this event (e.g. "order_placed").
func (s *Service) GoalReached(ctx context.Context, tx *uow.Tx, userID uuid.UUID, event string) error {
	return tx.Q.MarketingGoalReached(ctx, store.MarketingGoalReachedParams{UserID: userID, GoalEvent: &event})
}

func val(v *int32) int64 {
	if v == nil {
		return 0
	}
	return int64(*v)
}
