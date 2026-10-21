package aftersales

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

var shoppers = []string{auth.AudMarket, auth.AudCustomerApp}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	me := func(o httpx.Op) httpx.Op { o.Tag = "Returns"; o.Auth = shoppers; return o }

	httpx.Register(g, me(httpx.Op{ID: "requestReturn", Method: http.MethodPost, Path: "/api/v1/me/orders/{number}/returns", Status: 201, Summary: "Return items from a delivered order (7 days, faulty / damaged / wrong / not as described)"}),
		func(ctx context.Context, in *struct {
			Number string `path:"number"`
			Body   ReturnInput
		}) (*struct{ Body Return }, error) {
			out, err := s.RequestReturn(ctx, httpx.MustPrincipal(ctx).UserID, in.Number, in.Body)
			return &struct{ Body Return }{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myReturns", Method: http.MethodGet, Path: "/api/v1/me/returns", Summary: "Your returns, newest first"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []ReturnRow }, error) {
			rows, err := s.d.Q.AftersalesReturnsForUser(ctx, httpx.MustPrincipal(ctx).UserID)
			out := make([]ReturnRow, len(rows))
			for i, r := range rows {
				out[i] = ReturnRow{RMA: r.RmaNumber, OrderNumber: r.OrderNumber, Reason: r.Reason, Status: r.Status, AutoApproved: r.AutoApproved, CreatedAt: r.CreatedAt}
			}
			return &struct{ Body []ReturnRow }{out}, err
		})
	httpx.Register(g, me(httpx.Op{ID: "myReturn", Method: http.MethodGet, Path: "/api/v1/me/returns/{rma}", Summary: "One of your returns"}),
		func(ctx context.Context, in *struct {
			RMA string `path:"rma"`
		}) (*struct{ Body Return }, error) {
			r, err := s.d.Q.AftersalesGetReturn(ctx, in.RMA)
			if err != nil || r.RequestedBy != httpx.MustPrincipal(ctx).UserID {
				return nil, httpx.NotFound("Return")
			}
			out, err := s.view(ctx, s.d.Q, in.RMA)
			return &struct{ Body Return }{out}, err
		})

	st := func(o httpx.Op) httpx.Op {
		o.Tag = "Staff · Returns"
		o.Auth = []string{auth.AudStaff}
		o.Perm = "returns.manage"
		return o
	}
	httpx.Register(g, st(httpx.Op{ID: "staffListReturns", Method: http.MethodGet, Path: "/api/v1/staff/returns", Summary: "Returns, newest first"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"requested,approved,rejected,awaiting_pickup,received,inspected,refunded,replaced,closed,"`
			Cursor string `query:"cursor"`
			Limit  int    `query:"limit"`
		}) (*struct{ Body httpx.List[ReturnRow] }, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, httpx.BadCursor()
			}
			lim := httpx.Limit(in.Limit)
			rows, err := s.d.Q.AftersalesStaffList(ctx, store.AftersalesStaffListParams{Status: optional(in.Status), Before: before, Lim: int32(lim + 1)})
			if err != nil {
				return nil, err
			}
			views := make([]ReturnRow, len(rows))
			for i, r := range rows {
				views[i] = ReturnRow{RMA: r.RmaNumber, OrderNumber: r.OrderNumber, Reason: r.Reason, Status: r.Status, AutoApproved: r.AutoApproved, CreatedAt: r.CreatedAt}
			}
			return &struct{ Body httpx.List[ReturnRow] }{httpx.Paginate(s.d.Cursors, views, lim, func(v ReturnRow) time.Time { return v.CreatedAt })}, nil
		})
	httpx.Register(g, st(httpx.Op{ID: "staffGetReturn", Method: http.MethodGet, Path: "/api/v1/staff/returns/{rma}", Summary: "A return with its items and inspections"}),
		func(ctx context.Context, in *struct {
			RMA string `path:"rma"`
		}) (*struct{ Body Return }, error) {
			out, err := s.view(ctx, s.d.Q, in.RMA)
			return &struct{ Body Return }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffDecideReturn", Method: http.MethodPost, Path: "/api/v1/staff/returns/{rma}/decision", Summary: "Approve or reject a return waiting for support"}),
		func(ctx context.Context, in *struct {
			RMA  string `path:"rma"`
			Body struct {
				Decision string `json:"decision" enum:"approve,reject"`
				Reason   string `json:"reason,omitempty" maxLength:"500" doc:"Required to reject; the customer sees it"`
			}
		}) (*struct{ Body Return }, error) {
			out, err := s.Decide(ctx, in.RMA, in.Body.Decision, in.Body.Reason)
			return &struct{ Body Return }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffReceiveReturn", Method: http.MethodPost, Path: "/api/v1/staff/returns/{rma}/receive", Summary: "The item arrived (warehouse or store)"}),
		func(ctx context.Context, in *struct {
			RMA string `path:"rma"`
		}) (*struct{ Body Return }, error) {
			out, err := s.Receive(ctx, in.RMA)
			return &struct{ Body Return }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffInspectReturnItem", Method: http.MethodPost, Path: "/api/v1/staff/returns/{rma}/items/{itemId}/inspection", Summary: "Grade an item and decide where it goes (IMEI must match the unit sold)"}),
		func(ctx context.Context, in *struct {
			RMA    string    `path:"rma"`
			ItemID uuid.UUID `path:"itemId"`
			Body   InspectInput
		}) (*struct{ Body Return }, error) {
			out, err := s.Inspect(ctx, in.RMA, in.ItemID, in.Body)
			return &struct{ Body Return }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffRefundReturn", Method: http.MethodPost, Path: "/api/v1/staff/returns/{rma}/refund", Summary: "Raise the refund for an inspected return (finance approves it)"}),
		func(ctx context.Context, in *struct {
			RMA string `path:"rma"`
		}) (*struct{ Body Return }, error) {
			out, err := s.Refund(ctx, in.RMA)
			return &struct{ Body Return }{out}, err
		})
}
