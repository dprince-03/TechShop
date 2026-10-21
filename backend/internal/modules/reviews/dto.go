package reviews

import (
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Response types. Reviewers appear as "First L."; moderation reasons stay with staff.

// MyReview is what the reviewer gets back after submitting.
type MyReview struct {
	ID        uuid.UUID `json:"id"`
	Rating    int16     `json:"rating"`
	Verified  bool      `json:"verified"`
	Status    string    `json:"status" enum:"published,pending" doc:"pending = waiting for a moderator"`
	CreatedAt time.Time `json:"createdAt"`
}

// PublicReview is a published review on the product page.
type PublicReview struct {
	ID               uuid.UUID   `json:"id"`
	Rating           int16       `json:"rating"`
	Title            *string     `json:"title"`
	Body             *string     `json:"body"`
	Verified         bool        `json:"verified"`
	Reviewer         string      `json:"reviewer"`
	ImageFileIDs     []uuid.UUID `json:"imageFileIds"`
	SellerResponse   *string     `json:"sellerResponse"`
	SellerResponseAt *time.Time  `json:"sellerResponseAt"`
	CreatedAt        time.Time   `json:"createdAt"`
}

func publicReview(r store.ReviewsPublishedRow) PublicReview {
	imgs := r.ImageIds
	if imgs == nil {
		imgs = []uuid.UUID{}
	}
	return PublicReview{ID: r.ID, Rating: r.Rating, Title: r.Title, Body: r.Body, Verified: r.IsVerified != nil && *r.IsVerified,
		Reviewer: displayName(r.FirstName, r.LastInitial), ImageFileIDs: imgs, SellerResponse: r.SellerResponse,
		SellerResponseAt: r.SellerResponseAt, CreatedAt: r.CreatedAt}
}

// Summary is a product's rating at a glance.
type Summary struct {
	Average  float64          `json:"average" doc:"Weighted: verified purchases count 1, unverified 0.5"`
	Total    int32            `json:"total"`
	Verified int32            `json:"verified"`
	Stars    map[string]int32 `json:"stars" doc:"Count per star, keys \"1\"–\"5\""`
}

func summaryView(s store.ReviewsSummaryRow) Summary {
	return Summary{Average: s.WeightedAvg, Total: s.Total, Verified: s.Verified,
		Stars: map[string]int32{"5": s.Five, "4": s.Four, "3": s.Three, "2": s.Two, "1": s.One}}
}

// ProductReviews is the reviews section of a product page.
type ProductReviews struct {
	Summary Summary                  `json:"summary"`
	Reviews httpx.List[PublicReview] `json:"reviews"`
}

// SellerReview is a review of an item the seller shipped.
type SellerReview struct {
	ID             uuid.UUID `json:"id"`
	ProductName    string    `json:"productName"`
	ProductSlug    string    `json:"productSlug"`
	Rating         int16     `json:"rating"`
	Title          *string   `json:"title"`
	Body           *string   `json:"body"`
	Reviewer       string    `json:"reviewer"`
	SellerResponse *string   `json:"sellerResponse"`
	CreatedAt      time.Time `json:"createdAt"`
}

func sellerReview(r store.ReviewsForSellerRow) SellerReview {
	return SellerReview{ID: r.ID, ProductName: r.ProductName, ProductSlug: r.ProductSlug, Rating: r.Rating, Title: r.Title, Body: r.Body,
		Reviewer: r.FirstName, SellerResponse: r.SellerResponse, CreatedAt: r.CreatedAt}
}

// QueueItem is a review waiting for a moderator.
type QueueItem struct {
	ID          uuid.UUID `json:"id"`
	ProductName string    `json:"productName"`
	ProductSlug string    `json:"productSlug"`
	Rating      int16     `json:"rating"`
	Title       *string   `json:"title"`
	Body        *string   `json:"body"`
	Verified    bool      `json:"verified"`
	Reviewer    string    `json:"reviewer"`
	Status      string    `json:"status"`
	Reasons     []string  `json:"reasons" doc:"Why the automatic checks held it"`
	OpenReports int32     `json:"openReports"`
	CreatedAt   time.Time `json:"createdAt"`
}

func queueItem(r store.ReviewsQueueRow) QueueItem {
	reasons := r.ModerationReasons
	if reasons == nil {
		reasons = []string{}
	}
	return QueueItem{ID: r.ID, ProductName: r.ProductName, ProductSlug: r.ProductSlug, Rating: r.Rating, Title: r.Title, Body: r.Body,
		Verified: r.IsVerified != nil && *r.IsVerified, Reviewer: r.FirstName, Status: r.Status, Reasons: reasons, OpenReports: r.OpenReports, CreatedAt: r.CreatedAt}
}
