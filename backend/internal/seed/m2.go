package seed

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

func init() { Hooks = append(Hooks, seedCommerceM2) }

// seedCommerceM2 adds checkout and marketing sample data (development only): delivery
// addresses, marketing consent, a welcome coupon, a live flash deal, a scheduled promotion,
// a Lagos audience and an abandoned-cart journey.
func seedCommerceM2(ctx context.Context, d *kit.Deps, out *Output, ids map[string]uuid.UUID) error {
	q := d.Q
	acct := map[string]Account{}
	for _, a := range out.Accounts {
		acct[a.Role] = a
	}
	mm, ok := acct["marketing_manager"]
	if !ok {
		return errors.New("seed m2: marketing_manager account missing")
	}

	// Default Lagos delivery addresses (idempotent: only when the customer has none).
	for _, role := range []string{"customer", "customer_b"} {
		a, ok := acct[role]
		if !ok {
			continue
		}
		uid := a.UserID
		existing, err := q.IdentityListAddresses(ctx, &uid)
		if err != nil {
			return err
		}
		if len(existing) == 0 {
			if _, err := q.IdentityCreateAddress(ctx, store.IdentityCreateAddressParams{UserID: &uid, Label: ptr("Home"), RecipientName: "Dev " + role,
				Phone: a.Phone, Line1: "12 Allen Avenue", City: "Ikeja", Lga: ptr("Ikeja"), StateCode: "LA", IsDefault: true}); err != nil {
				return err
			}
		}
	}
	// The first customer opted in to marketing push and email (so campaigns reach someone).
	if c, ok := acct["customer"]; ok {
		uid := c.UserID
		for _, p := range []string{"marketing_push", "marketing_email"} {
			if has, _ := q.IdentityHasConsent(ctx, store.IdentityHasConsentParams{UserID: &uid, Purpose: p}); !has {
				if err := q.IdentityInsertConsent(ctx, store.IdentityInsertConsentParams{UserID: &uid, Purpose: p, Granted: true, PolicyVersion: "2026-10", Source: "import"}); err != nil {
					return err
				}
			}
		}
	}

	now := time.Now()
	// Welcome coupon: 10% off, minimum ₦20,000, once per customer.
	if _, err := q.MarketingFindCoupon(ctx, "WELCOME10"); errors.Is(err, pgx.ErrNoRows) {
		v := int32(10)
		p, err := q.MarketingCreatePromotion(ctx, store.MarketingCreatePromotionParams{Name: "Welcome 10% off", Kind: "percentage", Value: &v, FundedBy: "techshop",
			StartsAt: now.Add(-time.Hour), EndsAt: now.AddDate(1, 0, 0), Status: "live", CreatedBy: mm.UserID})
		if err != nil {
			return err
		}
		code := "WELCOME10"
		if _, err := q.MarketingCreateCoupon(ctx, store.MarketingCreateCouponParams{Code: &code, PromotionID: p.ID, PerUserLimit: 1, MinOrderKobo: ptr(int64(2000000))}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	existing, err := q.MarketingListPromotions(ctx, store.MarketingListPromotionsParams{Limit: 200})
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, p := range existing {
		have[p.Name] = true
	}
	// Live flash deal on the p5 listing: 5 units, 1 per person, 20% below list.
	if lid, ok := ids["p5:listing"]; ok && !have["Flash: power hour"] {
		l, err := q.CatalogGetListing(ctx, lid)
		if err != nil {
			return err
		}
		p, err := q.MarketingCreatePromotion(ctx, store.MarketingCreatePromotionParams{Name: "Flash: power hour", Kind: "flash_price", FundedBy: "techshop",
			StartsAt: now.Add(-time.Minute), EndsAt: now.AddDate(0, 1, 0), Status: "live", CreatedBy: mm.UserID})
		if err != nil {
			return err
		}
		if _, err := q.MarketingSetListingPrice(ctx, store.MarketingSetListingPriceParams{PromotionID: p.ID, ListingID: lid, PriceKobo: l.PriceKobo * 8 / 10,
			StockLimit: ptr(int32(5)), PerUserLimit: ptr(int32(1))}); err != nil {
			return err
		}
		out.Extra["flashListingId"] = lid.String()
	}
	// A scheduled sale next month (shows the countdown and the scheduler).
	if !have["December gadget sale"] {
		v := int32(15)
		start := time.Date(now.Year(), now.Month()+1, 1, 8, 0, 0, 0, time.FixedZone("WAT", 3600))
		if _, err := q.MarketingCreatePromotion(ctx, store.MarketingCreatePromotionParams{Name: "December gadget sale", Kind: "percentage", Value: &v, FundedBy: "techshop",
			StartsAt: start, EndsAt: start.AddDate(0, 0, 7), Status: "scheduled", CreatedBy: mm.UserID}); err != nil {
			return err
		}
	}
	// Audience and abandoned-cart journey.
	segs, err := q.MarketingListSegments(ctx)
	if err != nil {
		return err
	}
	if len(segs) == 0 {
		rules, _ := json.Marshal(map[string]any{"states": []string{"LA"}})
		if _, err := q.MarketingCreateSegment(ctx, store.MarketingCreateSegmentParams{Name: "Lagos shoppers", Rules: rules, CreatedBy: mm.UserID}); err != nil {
			return err
		}
	}
	js, err := q.MarketingListJourneys(ctx)
	if err != nil {
		return err
	}
	if len(js) == 0 {
		rd := int32(7)
		goal := "order_placed"
		j, err := q.MarketingCreateJourney(ctx, store.MarketingCreateJourneyParams{Name: "Abandoned cart reminder", Trigger: "cart_abandoned", EntryRules: []byte(`{}`),
			ReentryDays: &rd, GoalEvent: &goal, CreatedBy: mm.UserID})
		if err != nil {
			return err
		}
		steps := []struct{ kind, cfg string }{
			{"condition", `{"if":"cart_active"}`},
			{"send", `{"channel":"push","body":"Hi {{first_name}}, your cart is waiting. Complete your order before stock runs out."}`},
			{"wait", `{"hours":24}`},
			{"condition", `{"if":"not_ordered"}`},
			{"send", `{"channel":"email","body":"Still thinking it over, {{first_name}}? Your cart is saved."}`},
		}
		for i, st := range steps {
			if _, err := q.MarketingAddJourneyStep(ctx, store.MarketingAddJourneyStepParams{JourneyID: j.ID, Position: int32(i), Kind: st.kind, Config: []byte(st.cfg)}); err != nil {
				return err
			}
		}
		if _, err := q.MarketingSetJourneyStatus(ctx, store.MarketingSetJourneyStatusParams{ID: j.ID, Status: "active"}); err != nil {
			return err
		}
	}
	return nil
}
