// Package reviews handles product reviews, reports of reviews and customer ratings of sellers
// (M4 slice 4). Plan: trust-safety.md §6. Rules:
//   - a review is a "verified purchase" only when tied to the reviewer's own delivered order line;
//     unverified reviews are allowed, listed after verified ones and count half in ratings;
//   - suspicious reviews (rules.go) are held for moderation; the others publish straight away;
//   - sellers can answer a review of an item they shipped once, and report it, never remove it;
//   - customers rate a seller once per delivered fulfilment.
package reviews

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/outbox"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Service is the reviews module.
type Service struct{ d *kit.Deps }

// New creates the module.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "reviews" }

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// ReviewInput is a customer's review.
type ReviewInput struct {
	Rating      int16       `json:"rating" minimum:"1" maximum:"5"`
	Title       string      `json:"title,omitempty" maxLength:"120"`
	Body        string      `json:"body,omitempty" maxLength:"3000"`
	OrderLineID *uuid.UUID  `json:"orderLineId,omitempty" doc:"Your delivered order line for this product: marks the review as a verified purchase"`
	ImageIDs    []uuid.UUID `json:"imageFileIds,omitempty" maxItems:"5" doc:"Files uploaded with purpose review_image"`
}

// LineReviewInput is a review of a received item (the order line identifies the product).
type LineReviewInput struct {
	Rating   int16       `json:"rating" minimum:"1" maximum:"5"`
	Title    string      `json:"title,omitempty" maxLength:"120"`
	Body     string      `json:"body,omitempty" maxLength:"3000"`
	ImageIDs []uuid.UUID `json:"imageFileIds,omitempty" maxItems:"5"`
}

// Submit records a review. It publishes at once unless a rule holds it for moderation.
func (s *Service) Submit(ctx context.Context, userID uuid.UUID, slug string, in ReviewInput) (MyReview, error) {
	var out MyReview
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		p, err := tx.Q.CatalogGetProductBySlug(ctx, slug)
		if err != nil || p.Status != "active" {
			return httpx.NotFound("Product")
		}
		if in.OrderLineID != nil {
			if _, err := tx.Q.ReviewsDeliveredLine(ctx, store.ReviewsDeliveredLineParams{LineID: *in.OrderLineID, UserID: &userID, ProductID: p.ID}); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return httpx.Invalid("not_your_delivery", "That order item isn't a delivered purchase of this product on your account.")
				}
				return err
			}
		}
		if err := checkImages(ctx, tx.Q, userID, in.ImageIDs); err != nil {
			return err
		}
		var sig Signals
		recent, err := tx.Q.ReviewsByUserLast24h(ctx, userID)
		if err != nil {
			return err
		}
		sig.ReviewsLast24h = int(recent)
		if body := normalise(in.Body); len(body) >= 20 { // short texts ("great phone") repeat innocently
			if sig.SameTextByOthers, err = tx.Q.ReviewsSameTextByOthers(ctx, store.ReviewsSameTextByOthersParams{UserID: userID, Normalised: body}); err != nil {
				return err
			}
		}
		reasons := holdReasons(in.Title, in.Body, sig)
		status := "published"
		if len(reasons) > 0 {
			status = "pending"
		} else {
			reasons = []string{}
		}
		r, err := tx.Q.ReviewsInsert(ctx, store.ReviewsInsertParams{ProductID: p.ID, OrderLineID: in.OrderLineID, UserID: userID, Rating: in.Rating,
			Title: optional(in.Title), Body: optional(in.Body), Status: status, ModerationReasons: reasons})
		if err != nil {
			if isUnique(err) {
				return httpx.Conflict("already_reviewed", "You've already reviewed this product.")
			}
			return httpx.DB(err, "review")
		}
		for i, f := range in.ImageIDs {
			if err := tx.Q.ReviewsInsertImage(ctx, store.ReviewsInsertImageParams{ReviewID: r.ID, FileID: f, Position: int16(i)}); err != nil {
				return err
			}
		}
		if status == "published" {
			tx.Emit("product", p.ID, "product.changed", map[string]any{"productId": p.ID, "reason": "review"})
		}
		out = MyReview{ID: r.ID, Rating: r.Rating, Verified: r.IsVerified != nil && *r.IsVerified, Status: r.Status, CreatedAt: r.CreatedAt}
		return nil
	})
	return out, err
}

// SubmitForLine reviews the product of one of the caller's delivered order lines (a verified
// purchase), so order pages needn't know product slugs.
func (s *Service) SubmitForLine(ctx context.Context, userID, lineID uuid.UUID, in ReviewInput) (MyReview, error) {
	slug, err := s.d.Q.ReviewsLineProductSlug(ctx, store.ReviewsLineProductSlugParams{LineID: lineID, UserID: &userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return MyReview{}, httpx.NotFound("Order item")
	}
	if err != nil {
		return MyReview{}, err
	}
	in.OrderLineID = &lineID
	return s.Submit(ctx, userID, slug, in)
}

func checkImages(ctx context.Context, q *store.Queries, userID uuid.UUID, ids []uuid.UUID) error {
	seen := map[uuid.UUID]bool{}
	for _, id := range ids {
		f, err := q.PlatformGetFile(ctx, id)
		if seen[id] || err != nil || f.UploadedBy == nil || *f.UploadedBy != userID || f.Purpose != "review_image" || f.Status != "ready" {
			return httpx.Invalid("invalid_image", "One of the photos isn't a finished review upload of yours.")
		}
		seen[id] = true
	}
	return nil
}

// Product returns a product's published reviews (verified first) and its rating summary.
func (s *Service) Product(ctx context.Context, slug string, verified *bool, cursor string, limit int) (ProductReviews, error) {
	p, err := s.d.Q.CatalogGetProductBySlug(ctx, slug)
	if err != nil || p.Status != "active" {
		return ProductReviews{}, httpx.NotFound("Product")
	}
	before, err := s.d.Cursors.Before(cursor)
	if err != nil {
		return ProductReviews{}, httpx.BadCursor()
	}
	lim := httpx.Limit(limit)
	rows, err := s.d.Q.ReviewsPublished(ctx, store.ReviewsPublishedParams{ProductID: p.ID, Verified: verified, Before: before, Lim: int32(lim + 1)})
	if err != nil {
		return ProductReviews{}, err
	}
	sum, err := s.d.Q.ReviewsSummary(ctx, p.ID)
	if err != nil {
		return ProductReviews{}, err
	}
	views := make([]PublicReview, len(rows))
	for i, r := range rows {
		views[i] = publicReview(r)
	}
	return ProductReviews{Summary: summaryView(sum), Reviews: httpx.Paginate(s.d.Cursors, views, lim, func(v PublicReview) time.Time { return v.CreatedAt })}, nil
}

// Report flags a published review for moderators. One open report per person per review.
func (s *Service) Report(ctx context.Context, userID, reviewID uuid.UUID, reason string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.ReviewsGet(ctx, reviewID)
		if err != nil || r.Status != "published" {
			return httpx.NotFound("Review")
		}
		if r.UserID == userID {
			return httpx.Invalid("own_review", "You can't report your own review.")
		}
		dup, err := tx.Q.ReviewsReportedBy(ctx, store.ReviewsReportedByParams{ReviewID: reviewID, ReporterID: &userID})
		if err != nil {
			return err
		}
		if dup {
			return httpx.Conflict("already_reported", "You've already reported this review. A moderator will look at it.")
		}
		_, err = tx.Q.ReviewsInsertReport(ctx, store.ReviewsInsertReportParams{ReviewID: reviewID, ReporterID: &userID, Reason: reason})
		return err
	})
}

// Decide is a moderator's call on a held or reported review. Hiding or rejecting upholds the
// open reports; publishing dismisses them.
func (s *Service) Decide(ctx context.Context, reviewID uuid.UUID, decision string) (store.CatalogProductReview, error) {
	moderator := httpx.MustPrincipal(ctx).UserID
	status := map[string]string{"publish": "published", "hide": "hidden", "reject": "rejected"}[decision]
	var out store.CatalogProductReview
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.ReviewsGetForUpdate(ctx, reviewID)
		if err != nil {
			return httpx.NotFound("Review")
		}
		if r.UserID == moderator {
			return httpx.E(403, "separation_of_duties", "You can't moderate your own review.")
		}
		if out, err = tx.Q.ReviewsSetStatus(ctx, store.ReviewsSetStatusParams{ID: reviewID, Status: status}); err != nil {
			return err
		}
		reports := "upheld"
		if status == "published" {
			reports = "dismissed"
		}
		if err := tx.Q.ReviewsDecideReports(ctx, store.ReviewsDecideReportsParams{ReviewID: reviewID, Status: reports, DecidedBy: &moderator}); err != nil {
			return err
		}
		if r.Status == "published" || status == "published" { // the visible rating changes
			tx.Emit("product", r.ProductID, "product.changed", map[string]any{"productId": r.ProductID, "reason": "review"})
		}
		return tx.Audit("review.moderated", "review", reviewID.String(), map[string]string{"from": r.Status, "to": status})
	})
	return out, err
}

// ForSeller lists published reviews of items the seller shipped (inside the seller's scoped
// transaction).
func (s *Service) ForSeller(ctx context.Context, tx *uow.Tx, sellerID uuid.UUID, cursor string, limit int) (httpx.List[SellerReview], error) {
	before, err := s.d.Cursors.Before(cursor)
	if err != nil {
		return httpx.List[SellerReview]{}, httpx.BadCursor()
	}
	lim := httpx.Limit(limit)
	rows, err := tx.Q.ReviewsForSeller(ctx, store.ReviewsForSellerParams{SellerID: sellerID, Before: before, Lim: int32(lim + 1)})
	if err != nil {
		return httpx.List[SellerReview]{}, err
	}
	views := make([]SellerReview, len(rows))
	for i, r := range rows {
		views[i] = sellerReview(r)
	}
	return httpx.Paginate(s.d.Cursors, views, lim, func(v SellerReview) time.Time { return v.CreatedAt }), nil
}

// Respond records the seller's one public answer to a review of an item it shipped.
func (s *Service) Respond(ctx context.Context, tx *uow.Tx, sellerID, reviewID uuid.UUID, text string) error {
	owner, err := tx.Q.ReviewsSellerOf(ctx, reviewID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && owner != sellerID) {
		return httpx.NotFound("Review")
	}
	if err != nil {
		return err
	}
	r, err := tx.Q.ReviewsGet(ctx, reviewID)
	if err != nil || r.Status != "published" {
		return httpx.NotFound("Review")
	}
	n, err := tx.Q.ReviewsSetSellerResponse(ctx, store.ReviewsSetSellerResponseParams{ID: reviewID, SellerResponse: &text})
	if err != nil {
		return httpx.DB(err, "response")
	}
	if n == 0 {
		return httpx.Conflict("already_responded", "You've already answered this review.")
	}
	return tx.Audit("review.seller_responded", "review", reviewID.String(), map[string]any{"seller": sellerID})
}

// RateSeller records a customer's rating of the seller for one delivered fulfilment.
func (s *Service) RateSeller(ctx context.Context, userID, fulfilmentID uuid.UUID, rating int16, comment string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		f, err := tx.Q.RatingsDeliveredFulfilment(ctx, store.RatingsDeliveredFulfilmentParams{FulfilmentID: fulfilmentID, UserID: &userID})
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound("Delivered order")
		}
		if err != nil {
			return err
		}
		if _, err := tx.Q.RatingsInsert(ctx, store.RatingsInsertParams{SellerID: f.SellerID, FulfilmentID: f.ID, UserID: userID, Rating: rating, Comment: optional(comment)}); err != nil {
			if isUnique(err) {
				return httpx.Conflict("already_rated", "You've already rated this delivery.")
			}
			return httpx.DB(err, "rating")
		}
		return tx.Q.RatingsRecomputeSeller(ctx, f.SellerID)
	})
}

func optional(s string) *string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return &s
}

// isUnique reports a unique-constraint violation (a second review or rating).
func isUnique(err error) bool {
	var p interface{ SQLState() string }
	return errors.As(err, &p) && p.SQLState() == "23505"
}
