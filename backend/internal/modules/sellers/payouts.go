package sellers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/modules/payouts"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Seller Centre › Payouts (payments-finance.md §9): the balance (available, pending, on hold)
// and payout history with statements. Owners and finance members only.

func (s *Service) payoutRoutes(r *kit.Routes) {
	g := r.Public
	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre · Payouts"; o.Auth = sellerAud; return o }
	base := "/api/v1/seller/shops/{shopId}"

	httpx.Register(g, sel(httpx.Op{ID: "sellerBalance", Method: http.MethodGet, Path: base + "/balance", Summary: "Money available, pending and on hold"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
		}) (*struct{ Body payouts.SellerBalance }, error) {
			var b payouts.SellerBalance
			err := s.inShop(ctx, in.ShopID, moneyRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				var err error
				b, err = s.payouts.Balance(ctx, tx.Q, in.ShopID)
				return err
			})
			return &struct{ Body payouts.SellerBalance }{b}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerPayouts", Method: http.MethodGet, Path: base + "/payouts", Summary: "Payouts, newest first"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			Cursor string    `query:"cursor"`
			Limit  int       `query:"limit"`
		}) (*struct {
			Body httpx.List[payouts.SellerPayout]
		}, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, httpx.BadCursor()
			}
			lim := httpx.Limit(in.Limit)
			var rows []payouts.SellerPayout
			err = s.inShop(ctx, in.ShopID, moneyRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				rows, err = s.payouts.History(ctx, tx.Q, in.ShopID, before, lim+1)
				return err
			})
			if err != nil {
				return nil, err
			}
			return &struct {
				Body httpx.List[payouts.SellerPayout]
			}{httpx.Paginate(s.d.Cursors, rows, lim, func(v payouts.SellerPayout) time.Time { return v.CreatedAt })}, nil
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerPayout", Method: http.MethodGet, Path: base + "/payouts/{payoutId}", Summary: "A payout statement: the orders it covers"}),
		func(ctx context.Context, in *struct {
			ShopID   uuid.UUID `path:"shopId"`
			PayoutID uuid.UUID `path:"payoutId"`
		}) (*struct{ Body payouts.SellerPayout }, error) {
			var p payouts.SellerPayout
			err := s.inShop(ctx, in.ShopID, moneyRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				var err error
				p, err = s.payouts.Statement(ctx, tx.Q, in.ShopID, in.PayoutID)
				return err
			})
			return &struct{ Body payouts.SellerPayout }{p}, err
		})
}
