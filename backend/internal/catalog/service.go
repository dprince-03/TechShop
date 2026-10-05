// Package catalog owns products, variants, listings (seller offers), moderation, the search read
// model, search merchandising and saved items. Plan: docs/search-catalogue.md.
package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// Service is the catalogue module.
type Service struct {
	d   *kit.Deps
	syn *synonymCache
}

// New creates the catalogue module.
func New(d *kit.Deps) *Service {
	s := &Service{d: d}
	s.syn = &synonymCache{reload: s.loadSynonyms}
	if d.Hub != nil {
		d.Hub.OnChannel(realtime.ChannelSearch, s.syn.invalidate)
	}
	return s
}

// Name implements kit.Module.
func (s *Service) Name() string { return "catalog" }

func (s *Service) loadSynonyms(ctx context.Context) (map[string]string, error) {
	rows, err := s.d.Q.SearchListSynonyms(ctx)
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	for _, r := range rows {
		for _, t := range r.Terms {
			m[strings.ToLower(t)] = strings.ToLower(r.Target)
		}
	}
	return m, nil
}

// bannedPhrases auto-reject counterfeit wording (search-catalogue.md §3).
var bannedPhrases = regexp.MustCompile(`(?i)\b(replica|first copy|1st copy|clone|fake|master copy|aaa quality)\b`)

// ListingInput creates or proposes an offer on a variant.
type ListingInput struct {
	VariantID        uuid.UUID `json:"variantId"`
	Condition        string    `json:"condition" enum:"new,uk_used,refurbished,open_box"`
	Grade            string    `json:"grade,omitempty" enum:"A,B,C,"`
	Price            int64     `json:"price" minimum:"1" doc:"Kobo"`
	CompareAt        *int64    `json:"compareAt,omitempty" doc:"Previous price in kobo; must not exceed the 30-day maximum (FCCPA)"`
	FulfilledBy      string    `json:"fulfilledBy" enum:"techshop,seller"`
	SellerStock      *int32    `json:"sellerStock,omitempty" minimum:"0"`
	MinOrderQuantity *int32    `json:"minOrderQuantity,omitempty" minimum:"2"`
	HandlingDays     int16     `json:"handlingDays,omitempty" minimum:"0" maximum:"10"`
	WarrantyProvider string    `json:"warrantyProvider,omitempty" enum:"manufacturer,techshop,seller,none,"`
	WarrantyMonths   int16     `json:"warrantyMonths,omitempty" minimum:"0" maximum:"60"`
	ConditionNotes   string    `json:"conditionNotes,omitempty" maxLength:"500"`
	BatteryHealthPct *int16    `json:"batteryHealthPct,omitempty" minimum:"1" maximum:"100"`
}

// CheckResult is the outcome of automatic listing checks.
type CheckResult struct {
	Status  string   `json:"status" enum:"active,in_review,rejected"`
	Risk    int16    `json:"risk"`
	Reasons []string `json:"reasons"`
}

// autoCheck applies the automatic moderation rules. First-party listings go live directly.
func (s *Service) autoCheck(ctx context.Context, q *store.Queries, sellerID uuid.UUID, sellerType string, in ListingInput) CheckResult {
	r := CheckResult{Status: "active", Reasons: []string{}}
	if sellerType == "first_party" {
		return r
	}
	if bannedPhrases.MatchString(in.ConditionNotes) {
		return CheckResult{Status: "rejected", Risk: 99, Reasons: []string{"Banned counterfeit wording in the description"}}
	}
	if in.CompareAt != nil {
		return CheckResult{Status: "rejected", Risk: 90, Reasons: []string{"A new listing has no 30-day price history, so it can't show a “was” price (FCCPA)"}}
	}
	median, err := q.CatalogMedianPrice(ctx, store.CatalogMedianPriceParams{VariantID: in.VariantID, Condition: in.Condition})
	if err == nil && median > 0 {
		drop := float64(median-in.Price) / float64(median)
		switch {
		case drop > 0.4:
			r.Status, r.Risk = "in_review", 86
			r.Reasons = append(r.Reasons, fmt.Sprintf("Price %.0f%% below the median (₦%d) — scam signal", drop*100, median/100))
		case float64(in.Price) > float64(median)*1.6:
			r.Reasons = append(r.Reasons, "Price is more than 60% above the median (warning only)")
		}
	}
	if in.Condition != "new" && in.Condition != "open_box" && in.BatteryHealthPct == nil && in.Grade == "" {
		r.Status = "in_review"
		r.Risk = max(r.Risk, 40)
		r.Reasons = append(r.Reasons, "Used devices need a battery health or grade")
	}
	if n, err := q.CatalogSellerListingCount(ctx, sellerID); err == nil && n < 10 {
		r.Status = "in_review"
		r.Risk = max(r.Risk, 40)
		r.Reasons = append(r.Reasons, "New seller: the first 10 listings are reviewed")
	}
	return r
}

// CreateListing creates an offer, runs automatic checks and records the first price.
func (s *Service) CreateListing(ctx context.Context, tx *uow.Tx, sellerID uuid.UUID, sellerType string, in ListingInput, actor *uuid.UUID) (store.CatalogListing, CheckResult, error) {
	chk := s.autoCheck(ctx, tx.Q, sellerID, sellerType, in)
	if chk.Status == "rejected" {
		return store.CatalogListing{}, chk, httpx.Invalid("listing_rejected", strings.Join(chk.Reasons, "; "))
	}
	if in.FulfilledBy == "" {
		in.FulfilledBy = "techshop"
	}
	if in.WarrantyProvider == "" {
		in.WarrantyProvider = "manufacturer"
	}
	if in.HandlingDays == 0 {
		in.HandlingDays = 1
	}
	var publishedAt *time.Time
	if chk.Status == "active" {
		now := time.Now()
		publishedAt = &now
	}
	var grade *string
	if in.Grade != "" {
		grade = &in.Grade
	}
	risk := chk.Risk
	l, err := tx.Q.CatalogCreateListing(ctx, store.CatalogCreateListingParams{
		SellerID: sellerID, VariantID: in.VariantID, Condition: in.Condition, Grade: grade, PriceKobo: in.Price, CompareAtKobo: in.CompareAt,
		FulfilledBy: in.FulfilledBy, SellerStock: in.SellerStock, MinOrderQuantity: in.MinOrderQuantity, HandlingDays: in.HandlingDays,
		WarrantyProvider: in.WarrantyProvider, WarrantyMonths: in.WarrantyMonths, ConditionNotes: strOrNil(in.ConditionNotes),
		BatteryHealthPct: in.BatteryHealthPct, Status: chk.Status, RiskScore: &risk, PublishedAt: publishedAt,
	})
	if err != nil {
		return l, chk, httpx.DB(err, "listing")
	}
	if err := tx.Q.CatalogInsertPriceHistory(ctx, store.CatalogInsertPriceHistoryParams{ListingID: l.ID, PriceKobo: in.Price, CompareAtKobo: in.CompareAt, ChangedBy: actor}); err != nil {
		return l, chk, err
	}
	decision := "auto_passed"
	if chk.Status != "active" {
		decision = "changes_requested"
	}
	if chk.Status == "in_review" || len(chk.Reasons) > 0 {
		if err := tx.Q.CatalogInsertListingReview(ctx, store.CatalogInsertListingReviewParams{ListingID: l.ID, Decision: map[bool]string{true: "auto_passed", false: decision}[chk.Status == "active"], Reasons: chk.Reasons}); err != nil {
			return l, chk, err
		}
	}
	tx.Emit("listing", l.ID, "listing.changed", map[string]any{"listingId": l.ID, "status": l.Status})
	return l, chk, nil
}

// UpdatePrice changes a price; a compare-at ("was") price may not exceed the 30-day maximum.
func (s *Service) UpdatePrice(ctx context.Context, tx *uow.Tx, listingID uuid.UUID, price int64, compareAt *int64, actor *uuid.UUID) (store.CatalogListing, error) {
	if compareAt != nil {
		max30, err := tx.Q.CatalogMaxPrice30d(ctx, listingID)
		if err != nil {
			return store.CatalogListing{}, err
		}
		if *compareAt > max30 {
			return store.CatalogListing{}, httpx.Invalid("fake_was_price", fmt.Sprintf("The “was” price can't be above the highest price in the last 30 days (₦%d).", max30/100))
		}
	}
	l, err := tx.Q.CatalogUpdateListingPrice(ctx, store.CatalogUpdateListingPriceParams{ID: listingID, PriceKobo: price, CompareAtKobo: compareAt})
	if err != nil {
		return l, httpx.DB(err, "listing")
	}
	if err := tx.Q.CatalogInsertPriceHistory(ctx, store.CatalogInsertPriceHistoryParams{ListingID: l.ID, PriceKobo: price, CompareAtKobo: compareAt, ChangedBy: actor}); err != nil {
		return l, err
	}
	tx.Emit("listing", l.ID, "listing.changed", map[string]any{"listingId": l.ID, "price": price})
	tx.Emit("listing", l.ID, "listing.price_changed", map[string]any{"listingId": l.ID, "price": price})
	return l, nil
}

// Moderate records a moderator decision on a listing in review.
func (s *Service) Moderate(ctx context.Context, listingID uuid.UUID, decision string, reasons []string, note string) (*store.CatalogListing, error) {
	p := httpx.MustPrincipal(ctx)
	status := map[string]string{"approved": "active", "changes_requested": "changes_requested", "rejected": "rejected", "suspended": "suspended"}[decision]
	var out store.CatalogListing
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		cur, err := tx.Q.CatalogGetListing(ctx, listingID)
		if err != nil {
			return httpx.DB(err, "listing")
		}
		if decision != "suspended" && cur.Status != "in_review" {
			return httpx.Conflict("not_in_review", "This listing isn't waiting for review.")
		}
		out, err = tx.Q.CatalogSetListingStatus(ctx, store.CatalogSetListingStatusParams{ID: listingID, Status: status})
		if err != nil {
			return err
		}
		if err := tx.Q.CatalogInsertListingReview(ctx, store.CatalogInsertListingReviewParams{ListingID: listingID, ReviewerID: &p.UserID, Decision: decision, Reasons: reasons, Note: strOrNil(note)}); err != nil {
			return err
		}
		tx.Emit("listing", listingID, "listing.changed", map[string]any{"listingId": listingID, "status": status})
		tx.Emit("listing", listingID, "listing.moderated", map[string]any{"listingId": listingID, "sellerId": cur.SellerID, "decision": decision, "reasons": reasons})
		return tx.Audit("listing.moderated", "listing", listingID.String(), map[string]any{"decision": decision, "reasons": reasons})
	})
	return &out, err
}

// RefreshArgs rebuilds a product's read-model row.
type RefreshArgs struct {
	ProductID uuid.UUID `json:"productId"`
}

// Kind identifies the job.
func (RefreshArgs) Kind() string { return "catalog.refresh_summary" }

// InsertOpts deduplicates refreshes of the same product queued close together.
func (RefreshArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true, ByPeriod: 2 * time.Second}}
}

type refreshWorker struct {
	river.WorkerDefaults[RefreshArgs]
	s *Service
}

// Work rebuilds the retail row (removed if the product isn't active).
func (w *refreshWorker) Work(ctx context.Context, job *river.Job[RefreshArgs]) error {
	return w.s.Refresh(ctx, job.Args.ProductID)
}

// Refresh rebuilds one product's summary immediately.
func (s *Service) Refresh(ctx context.Context, productID uuid.UUID) error {
	p, err := s.d.Q.CatalogGetProduct(ctx, productID)
	if err != nil {
		return nil // product gone
	}
	if p.Status != "active" {
		return s.d.Q.CatalogDeleteOfferSummary(ctx, productID)
	}
	return s.d.Q.CatalogRefreshOfferSummary(ctx, store.CatalogRefreshOfferSummaryParams{ProductID: productID, Channel: "retail"})
}

// RebuildAll refreshes every product (nightly safety net and the staff "reindex" action).
func (s *Service) RebuildAll(ctx context.Context) (int, error) {
	ids, err := s.d.Q.CatalogAllProductIDs(ctx)
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		if err := s.Refresh(ctx, id); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

type rebuildArgs struct{}

func (rebuildArgs) Kind() string { return "catalog.rebuild_all" }

type rebuildWorker struct {
	river.WorkerDefaults[rebuildArgs]
	s *Service
}

func (w *rebuildWorker) Work(ctx context.Context, _ *river.Job[rebuildArgs]) error {
	if _, err := w.s.RebuildAll(ctx); err != nil {
		return err
	}
	return w.s.d.Q.SearchRebuildSuggestions(ctx)
}

// Jobs implements kit.Module.
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &refreshWorker{s: s})
	river.AddWorker(reg.Workers, &rebuildWorker{s: s})
	reg.Every(24*time.Hour, func() (river.JobArgs, *river.InsertOpts) { return rebuildArgs{}, nil })
}

// Events implements kit.Module: any listing, product or stock change refreshes the read model.
func (s *Service) Events(r *outbox.Relay) {
	byListing := func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		var p struct {
			ListingID uuid.UUID `json:"listingId"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		pid, err := tx.Q.CatalogProductIDForListing(ctx, p.ListingID)
		if err != nil {
			return nil
		}
		tx.Enqueue(RefreshArgs{ProductID: pid}, nil)
		return nil
	}
	r.On("listing.changed", byListing)
	r.On("product.changed", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		tx.Enqueue(RefreshArgs{ProductID: ev.AggregateID}, nil)
		return nil
	})
	r.On("stock.changed", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		var p struct {
			VariantID uuid.UUID `json:"variantId"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		pid, err := tx.Q.CatalogProductIDForVariant(ctx, p.VariantID)
		if err != nil {
			return nil
		}
		tx.Enqueue(RefreshArgs{ProductID: pid}, nil)
		return nil
	})
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
