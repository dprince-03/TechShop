package payouts

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// What Seller Centre shows. The sellers module checks shop membership before calling these.

// Balance is the seller's available, pending and held money right now.
func (s *Service) Balance(ctx context.Context, q *store.Queries, sellerID uuid.UUID) (SellerBalance, error) {
	now := time.Now()
	p, err := s.position(ctx, q, sellerID, now)
	if err != nil {
		return SellerBalance{}, err
	}
	holds := p.Holds
	if holds == nil {
		holds = []string{}
	}
	return SellerBalance{Available: p.Amount, Pending: p.Pending, OnHold: len(holds) > 0, HoldReasons: holds, Minimum: s.d.Cfg.PayoutMinimumKobo,
		NextPayoutDay: nextPayoutDay(now, s.holiday(ctx))}, nil
}

// History lists a seller's payouts, newest first.
func (s *Service) History(ctx context.Context, q *store.Queries, sellerID uuid.UUID, before *time.Time, lim int) ([]SellerPayout, error) {
	rows, err := q.PayoutsForSeller(ctx, store.PayoutsForSellerParams{SellerID: sellerID, Before: before, Lim: int32(lim)})
	out := make([]SellerPayout, len(rows))
	for i, r := range rows {
		out[i] = SellerPayout{ID: r.ID, Amount: r.AmountKobo, Status: r.Status, ScheduledFor: r.ScheduledFor, PaidAt: r.PaidAt, AccountLast4: r.AccountNumberLast4,
			FailureReason: r.FailureReason, CreatedAt: r.CreatedAt}
	}
	return out, err
}

// Statement is one payout with the orders it covers.
func (s *Service) Statement(ctx context.Context, q *store.Queries, sellerID, payoutID uuid.UUID) (SellerPayout, error) {
	p, err := q.PayoutsGet(ctx, payoutID)
	if err != nil || p.SellerID != sellerID {
		return SellerPayout{}, httpx.NotFound("Payout")
	}
	bank, err := q.PayoutsBankAccount(ctx, p.BankAccountID)
	if err != nil {
		return SellerPayout{}, err
	}
	out := SellerPayout{ID: p.ID, Amount: p.AmountKobo, Status: p.Status, ScheduledFor: p.ScheduledFor, PaidAt: p.PaidAt, AccountLast4: bank.AccountNumberLast4,
		FailureReason: p.FailureReason, CreatedAt: p.CreatedAt, Items: []PayoutItem{}}
	items, err := q.PayoutsItems(ctx, p.ID)
	for _, it := range items {
		out.Items = append(out.Items, PayoutItem{OrderNumber: it.OrderNumber, Amount: it.AmountKobo, DeliveredAt: it.DeliveredAt})
	}
	return out, err
}
