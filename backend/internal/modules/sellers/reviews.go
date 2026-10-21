package sellers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/modules/reviews"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Seller Centre reviews (M4 slice 4): the shop sees reviews of items it shipped and may answer
// each once. Reporting a review uses the shared POST /reviews/{id}/report.

func (s *Service) reviewRoutes(r *kit.Routes) {
	g := r.Public
	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre · Reviews"; o.Auth = sellerAud; return o }

	httpx.Register(g, sel(httpx.Op{ID: "sellerListReviews", Method: http.MethodGet, Path: "/api/v1/seller/shops/{shopId}/reviews", Summary: "Reviews of items you shipped, newest first"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Cursor string    `query:"cursor"`
			Limit  int       `query:"limit"`
		}) (*struct {
			Body httpx.List[reviews.SellerReview]
		}, error) {
			var out httpx.List[reviews.SellerReview]
			err := s.inShop(ctx, in.ShopID, anyRole, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				var err error
				out, err = s.reviews.ForSeller(ctx, tx, in.ShopID, in.Cursor, in.Limit)
				return err
			})
			return &struct {
				Body httpx.List[reviews.SellerReview]
			}{out}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerRespondToReview", Method: http.MethodPost, Path: "/api/v1/seller/shops/{shopId}/reviews/{reviewId}/response", Status: 204, Summary: "Answer a review publicly (once)"}),
		func(ctx context.Context, in *struct {
			ShopID   uuid.UUID `path:"shopId"`
			ReviewID uuid.UUID `path:"reviewId"`
			Body     struct {
				Text string `json:"text" minLength:"2" maxLength:"1000"`
			}
		}) (*struct{}, error) {
			return nil, s.inShop(ctx, in.ShopID, []string{"owner", "manager"}, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				return s.reviews.Respond(ctx, tx, in.ShopID, in.ReviewID, in.Body.Text)
			})
		})
}
