package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Marketplace listing lifecycle after creation (search-catalogue.md §3): sellers reprice,
// pause and resume, and fix listings moderators sent back. Callers run these inside the seller's
// scoped transaction (sellers.inShop) and have already checked the listing is theirs.

// RepriceMarketplace changes a marketplace listing's price. A live listing whose new price is
// more than 40 % below the median for its variant and condition goes back to review: the same
// scam signal as at creation, so approval can't be used to slip a bait price in afterwards.
func (s *Service) RepriceMarketplace(ctx context.Context, tx *uow.Tx, l store.CatalogListing, price int64, compareAt *int64, actor *uuid.UUID) (store.CatalogListing, error) {
	out, err := s.UpdatePrice(ctx, tx, l.ID, price, compareAt, actor)
	if err != nil || out.Status != "active" {
		return out, err
	}
	median, err := tx.Q.CatalogMedianPriceOthers(ctx, store.CatalogMedianPriceOthersParams{VariantID: l.VariantID, Condition: l.Condition, ID: l.ID})
	if err != nil || median <= 0 || float64(median-price)/float64(median) <= 0.4 {
		return out, err
	}
	risk := int16(86)
	if out, err = tx.Q.CatalogSetListingStatus(ctx, store.CatalogSetListingStatusParams{ID: l.ID, Status: "in_review", RiskScore: &risk}); err != nil {
		return out, err
	}
	reason := fmt.Sprintf("Price cut to more than 40%% below the median (₦%d) — scam signal", median/100)
	if err := recordCheck(ctx, tx.Q, l.ID, CheckResult{Status: "in_review", Risk: risk, Reasons: []string{reason}}); err != nil {
		return out, err
	}
	tx.Emit("listing", l.ID, "listing.changed", map[string]any{"listingId": l.ID, "status": out.Status})
	return out, nil
}

// SetPaused pauses a live listing or puts a paused one back on sale.
func (s *Service) SetPaused(ctx context.Context, tx *uow.Tx, l store.CatalogListing, paused bool) (store.CatalogListing, error) {
	from, to := "active", "paused"
	if !paused {
		from, to = "paused", "active"
	}
	if l.Status != from {
		return l, httpx.Conflict("bad_transition", fmt.Sprintf("Only %s listings can be %s.", from, map[bool]string{true: "paused", false: "resumed"}[paused]))
	}
	out, err := tx.Q.CatalogSetListingStatus(ctx, store.CatalogSetListingStatusParams{ID: l.ID, Status: to})
	if err != nil {
		return out, err
	}
	tx.Emit("listing", l.ID, "listing.changed", map[string]any{"listingId": l.ID, "status": to})
	return out, nil
}

// ListingEdit is what a seller may change while a listing is a draft or has changes requested.
type ListingEdit struct {
	Grade            string `json:"grade,omitempty" enum:"A,B,C,"`
	ConditionNotes   string `json:"conditionNotes,omitempty" maxLength:"500"`
	BatteryHealthPct *int16 `json:"batteryHealthPct,omitempty" minimum:"1" maximum:"100"`
	WarrantyProvider string `json:"warrantyProvider" enum:"manufacturer,techshop,seller,none"`
	WarrantyMonths   int16  `json:"warrantyMonths" minimum:"0" maximum:"60"`
	HandlingDays     int16  `json:"handlingDays" minimum:"0" maximum:"10"`
}

// EditAndResubmit applies a seller's fixes and runs the automatic checks again: the listing
// goes live, back to review, or stays as it is with the reasons if a check rejects it.
func (s *Service) EditAndResubmit(ctx context.Context, tx *uow.Tx, l store.CatalogListing, sellerType string, e ListingEdit) (store.CatalogListing, CheckResult, error) {
	if l.Status != "draft" && l.Status != "changes_requested" {
		return l, CheckResult{}, httpx.Conflict("bad_transition", "Only drafts and listings with changes requested can be edited and resubmitted.")
	}
	in := ListingInput{VariantID: l.VariantID, Condition: l.Condition, Grade: e.Grade, Price: l.PriceKobo, CompareAt: l.CompareAtKobo,
		FulfilledBy: l.FulfilledBy, SellerStock: l.SellerStock, ConditionNotes: e.ConditionNotes, BatteryHealthPct: e.BatteryHealthPct,
		WarrantyProvider: e.WarrantyProvider, WarrantyMonths: e.WarrantyMonths, HandlingDays: e.HandlingDays}
	if err := validateListing(in); err != nil {
		return l, CheckResult{}, err
	}
	out, err := tx.Q.CatalogUpdateListingDetails(ctx, store.CatalogUpdateListingDetailsParams{ID: l.ID, Grade: strOrNil(e.Grade),
		ConditionNotes: strOrNil(e.ConditionNotes), BatteryHealthPct: e.BatteryHealthPct, WarrantyProvider: e.WarrantyProvider,
		WarrantyMonths: e.WarrantyMonths, HandlingDays: e.HandlingDays})
	if err != nil {
		return out, CheckResult{}, httpx.DB(err, "listing")
	}
	// The "was" price was checked against history when it was set; a resubmission keeps it.
	in.CompareAt = nil
	chk := s.autoCheck(ctx, tx.Q, l.SellerID, sellerType, in)
	if chk.Status == "rejected" {
		return out, chk, httpx.Invalid("listing_rejected", strings.Join(chk.Reasons, "; "))
	}
	risk := chk.Risk
	if out, err = tx.Q.CatalogSetListingStatus(ctx, store.CatalogSetListingStatusParams{ID: l.ID, Status: chk.Status, RiskScore: &risk}); err != nil {
		return out, chk, err
	}
	if err := recordCheck(ctx, tx.Q, l.ID, chk); err != nil {
		return out, chk, err
	}
	tx.Emit("listing", l.ID, "listing.changed", map[string]any{"listingId": l.ID, "status": chk.Status})
	return out, chk, nil
}
