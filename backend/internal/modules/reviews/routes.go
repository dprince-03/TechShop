package reviews

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
)

var (
	shoppers  = []string{auth.AudMarket, auth.AudCustomerApp}
	reporters = []string{auth.AudMarket, auth.AudCustomerApp, auth.AudSeller}
)

// Routes implements kit.Module. Seller Centre's review routes live in the sellers module (they
// run in the shop-scoped transaction) and call ForSeller and Respond.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	tag := func(o httpx.Op) httpx.Op { o.Tag = "Reviews"; return o }

	httpx.Register(g, tag(httpx.Op{ID: "productReviews", Method: http.MethodGet, Path: "/api/v1/products/{slug}/reviews", Summary: "Published reviews (verified purchases first) and the rating summary"}),
		func(ctx context.Context, in *struct {
			Slug     string `path:"slug"`
			Verified string `query:"verified" enum:"true,false,"`
			Cursor   string `query:"cursor"`
			Limit    int    `query:"limit"`
		}) (*struct{ Body ProductReviews }, error) {
			var verified *bool
			if in.Verified != "" {
				v := in.Verified == "true"
				verified = &v
			}
			out, err := s.Product(ctx, in.Slug, verified, in.Cursor, in.Limit)
			return &struct{ Body ProductReviews }{out}, err
		})
	httpx.Register(g, tag(httpx.Op{ID: "createReview", Method: http.MethodPost, Path: "/api/v1/products/{slug}/reviews", Auth: shoppers, Status: 201, Summary: "Review a product (one per product)"}),
		func(ctx context.Context, in *struct {
			Slug string `path:"slug"`
			Body ReviewInput
		}) (*struct{ Body MyReview }, error) {
			out, err := s.Submit(ctx, httpx.MustPrincipal(ctx).UserID, in.Slug, in.Body)
			return &struct{ Body MyReview }{out}, err
		})
	httpx.Register(g, tag(httpx.Op{ID: "reportReview", Method: http.MethodPost, Path: "/api/v1/reviews/{id}/report", Auth: reporters, Status: 204, Summary: "Report a review that breaks the rules"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Reason string `json:"reason" enum:"abuse,personal_data,off_topic,fake,other"`
			}
		}) (*struct{}, error) {
			return nil, s.Report(ctx, httpx.MustPrincipal(ctx).UserID, in.ID, in.Body.Reason)
		})
	httpx.Register(g, tag(httpx.Op{ID: "rateSeller", Method: http.MethodPost, Path: "/api/v1/me/deliveries/{fulfilmentId}/rating", Auth: shoppers, Status: 204, Summary: "Rate the seller of a delivered part of your order"}),
		func(ctx context.Context, in *struct {
			FulfilmentID uuid.UUID `path:"fulfilmentId"`
			Body         struct {
				Rating  int16  `json:"rating" minimum:"1" maximum:"5"`
				Comment string `json:"comment,omitempty" maxLength:"500"`
			}
		}) (*struct{}, error) {
			return nil, s.RateSeller(ctx, httpx.MustPrincipal(ctx).UserID, in.FulfilmentID, in.Body.Rating, in.Body.Comment)
		})

	st := func(o httpx.Op) httpx.Op {
		o.Tag = "Staff · Reviews"
		o.Auth = []string{auth.AudStaff}
		o.Perm = "catalog.moderate"
		return o
	}
	httpx.Register(g, st(httpx.Op{ID: "staffReviewQueue", Method: http.MethodGet, Path: "/api/v1/staff/reviews/queue", Summary: "Held reviews and reviews with open reports, oldest first"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []QueueItem }, error) {
			rows, err := s.d.Q.ReviewsQueue(ctx, 100)
			out := make([]QueueItem, len(rows))
			for i, r := range rows {
				out[i] = queueItem(r)
			}
			return &struct{ Body []QueueItem }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffDecideReview", Method: http.MethodPost, Path: "/api/v1/staff/reviews/{id}/decision", Summary: "Publish, hide or reject a review (settles its reports)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Decision string `json:"decision" enum:"publish,hide,reject"`
			}
		}) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			r, err := s.Decide(ctx, in.ID, in.Body.Decision)
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = r.Status
			return out, err
		})
}
