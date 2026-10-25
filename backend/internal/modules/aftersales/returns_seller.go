package aftersales

// Seller returns: a shop sees returns of its items (with what it's charged) and may dispute the
// inspection within 3 days; a support lead decides. An upheld dispute gives the seller their
// money back and TechShop bears the loss. Plan: orders-fulfilment.md §7.

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/ledger"
	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// disputeWindow is how long after inspection a seller can dispute it (orders-fulfilment.md §7).
const disputeWindow = 3 * 24 * time.Hour

// SellerReturn is a return as the seller sees it: only their items, and their charge.
type SellerReturn struct {
	Return
	Charge       int64      `json:"charge" doc:"Kobo taken from your payable for these items (after commission is handed back)"`
	DisputeUntil *time.Time `json:"disputeUntil" doc:"Last moment to dispute the inspection"`
	CanDispute   bool       `json:"canDispute"`
}

func (s *Service) sellerReturn(ctx context.Context, q *store.Queries, shopID uuid.UUID, rma string) (SellerReturn, error) {
	full, err := s.view(ctx, q, rma)
	if err != nil {
		return SellerReturn{}, err
	}
	r, err := q.AftersalesGetReturn(ctx, rma)
	if err != nil {
		return SellerReturn{}, err
	}
	items, err := q.AftersalesItems(ctx, r.ID)
	if err != nil {
		return SellerReturn{}, err
	}
	mine := map[uuid.UUID]bool{}
	var shares []refundShare
	var lastInspected *time.Time
	allInspected := true
	for _, it := range items {
		if it.SellerID != shopID {
			continue
		}
		mine[it.ID] = true
		shares = append(shares, refundShare{SellerID: it.SellerID, LineTotal: it.LineTotalKobo, SellerEarning: it.SellerEarningKobo, CommissionBps: it.CommissionBps,
			LineQty: it.LineQuantity, Qty: it.Quantity, CommissionPosted: it.DeliveredAt != nil})
		if it.InspectedAt == nil {
			allInspected = false
		} else if lastInspected == nil || it.InspectedAt.After(*lastInspected) {
			lastInspected = it.InspectedAt
		}
	}
	if len(mine) == 0 {
		return SellerReturn{}, httpx.NotFound("Return")
	}
	out := SellerReturn{Return: full}
	out.Items = nil
	for _, it := range full.Items {
		if mine[it.ID] {
			out.Items = append(out.Items, it)
		}
	}
	for _, l := range refundAllocation(shares, 0) {
		if l.Code == ledger.SellerPayable(shopID) {
			out.Charge += l.Amount
		}
	}
	out.Refund = 0 // the customer's total isn't the seller's business
	if allInspected && lastInspected != nil {
		until := lastInspected.Add(disputeWindow)
		out.DisputeUntil = &until
		out.CanDispute = r.SellerDisputedAt == nil && time.Now().Before(until) && r.Status != "rejected"
	}
	return out, nil
}

// DisputeReturn challenges an inspection of the shop's items.
func (s *Service) DisputeReturn(ctx context.Context, shopID uuid.UUID, rma, reason string) (SellerReturn, error) {
	if strings.TrimSpace(reason) == "" {
		return SellerReturn{}, httpx.Invalid("reason_required", "Say what's wrong with the inspection.")
	}
	actor := httpx.MustPrincipal(ctx).UserID
	var out SellerReturn
	err := s.sellers.InShop(ctx, shopID, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		v, err := s.sellerReturn(ctx, tx.Q, shopID, rma)
		if err != nil {
			return err
		}
		if !v.CanDispute {
			if v.Dispute != nil {
				return httpx.Conflict("already_disputed", "You've already disputed this return.")
			}
			return httpx.Conflict("dispute_closed", "Disputes are open for 3 days after we inspect your items.")
		}
		r, err := tx.Q.AftersalesGetReturnForUpdate(ctx, rma)
		if err != nil {
			return err
		}
		if _, err := tx.Q.AftersalesSetDispute(ctx, store.AftersalesSetDisputeParams{ID: r.ID, SellerDisputeReason: &reason, SellerDisputedBy: &actor}); err != nil {
			return err
		}
		if err := tx.Audit("return.disputed", "return", r.ID.String(), map[string]any{"shop": shopID}); err != nil {
			return err
		}
		out, err = s.sellerReturn(ctx, tx.Q, shopID, rma)
		return err
	})
	return out, err
}

// DecideDispute settles a seller's dispute. Upheld: the seller is credited what the refund took
// from them and TechShop bears it as a write-off. The refund must be raised first, so the amount
// is known.
func (s *Service) DecideDispute(ctx context.Context, rma, decision, note string) (Return, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	if strings.TrimSpace(note) == "" {
		return Return{}, httpx.Invalid("note_required", "Explain the decision; the seller sees it.")
	}
	var out Return
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.AftersalesGetReturnForUpdate(ctx, rma)
		if err != nil {
			return httpx.NotFound("Return")
		}
		if r.SellerDisputedAt == nil || r.DisputeDecision != nil {
			return httpx.Conflict("no_open_dispute", "This return has no open dispute.")
		}
		if decision == "upheld" {
			if err := s.creditSellers(ctx, tx, r); err != nil {
				return err
			}
		}
		if _, err := tx.Q.AftersalesDecideDispute(ctx, store.AftersalesDecideDisputeParams{ID: r.ID, DisputeDecision: &decision, DisputeNote: &note, DisputeDecidedBy: &actor}); err != nil {
			return err
		}
		if err := tx.Audit("return.dispute_"+decision, "return", r.ID.String(), nil); err != nil {
			return err
		}
		out, err = s.view(ctx, tx.Q, rma)
		return err
	})
	return out, err
}

// creditSellers gives back what the return's refund charged each seller.
func (s *Service) creditSellers(ctx context.Context, tx *uow.Tx, r store.AftersalesReturnRequest) error {
	ref, err := tx.Q.AftersalesReturnRefundRow(ctx, &r.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.Conflict("refund_first", "Raise the refund before upholding the dispute, so the seller's charge is known.")
	}
	if err != nil {
		return err
	}
	items, err := tx.Q.AftersalesItems(ctx, r.ID)
	if err != nil {
		return err
	}
	_, alloc := returnRefund(r.Reason, items)
	var ls []ledger.Line
	var total int64
	for _, l := range alloc {
		if strings.HasPrefix(l.Code, "seller_payable:") && l.Amount != 0 {
			ls = append(ls, ledger.Line{Code: l.Code, Amount: -l.Amount})
			total += l.Amount
		}
	}
	if total <= 0 {
		return nil
	}
	ls = append(ls, ledger.Line{Code: ledger.ExpenseWriteOffs, Amount: total})
	_, _, err = s.ledger.Post(ctx, tx, ledger.Journal{Key: "return:" + r.ID.String() + ":dispute_upheld", Kind: "adjustment", RefType: "refund", RefID: ref.ID,
		Memo: "Return " + r.RmaNumber + ": seller's dispute upheld", PostedBy: &httpx.MustPrincipal(ctx).UserID, Lines: ls})
	return err
}

func (s *Service) sellerReturnRoutes(r *kit.Routes) {
	g := r.Public
	sel := func(o httpx.Op) httpx.Op { o.Tag = "Seller Centre"; o.Auth = []string{auth.AudSeller}; return o }
	httpx.Register(g, sel(httpx.Op{ID: "sellerReturns", Method: http.MethodGet, Path: "/api/v1/seller/shops/{shopId}/returns", Summary: "Returns of your items, newest first"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
		}) (*struct{ Body []ReturnRow }, error) {
			var out []ReturnRow
			err := s.sellers.InShop(ctx, in.ShopID, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				rows, err := tx.Q.AftersalesSellerReturns(ctx, in.ShopID)
				out = make([]ReturnRow, len(rows))
				for i, r := range rows {
					out[i] = ReturnRow{RMA: r.RmaNumber, OrderNumber: r.OrderNumber, Reason: r.Reason, Status: r.Status, AutoApproved: r.AutoApproved, CreatedAt: r.CreatedAt}
				}
				return err
			})
			return &struct{ Body []ReturnRow }{out}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerReturn", Method: http.MethodGet, Path: "/api/v1/seller/shops/{shopId}/returns/{rma}", Summary: "A return of your items, with your charge"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			RMA    string    `path:"rma"`
		}) (*struct{ Body SellerReturn }, error) {
			var out SellerReturn
			err := s.sellers.InShop(ctx, in.ShopID, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
				var err error
				out, err = s.sellerReturn(ctx, tx.Q, in.ShopID, in.RMA)
				return err
			})
			return &struct{ Body SellerReturn }{out}, err
		})
	httpx.Register(g, sel(httpx.Op{ID: "sellerDisputeReturn", Method: http.MethodPost, Path: "/api/v1/seller/shops/{shopId}/returns/{rma}/dispute", Summary: "Dispute our inspection (within 3 days)"}),
		func(ctx context.Context, in *struct {
			ShopID uuid.UUID `path:"shopId"`
			RMA    string    `path:"rma"`
			Body   struct {
				Reason string `json:"reason" maxLength:"1000"`
			}
		}) (*struct{ Body SellerReturn }, error) {
			out, err := s.DisputeReturn(ctx, in.ShopID, in.RMA, in.Body.Reason)
			return &struct{ Body SellerReturn }{out}, err
		})
}
