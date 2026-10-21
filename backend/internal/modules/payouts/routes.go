package payouts

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/platform/auth"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Routes implements kit.Module: finance's payout batches and holds. Seller Centre's balance and
// payout history are served by the sellers module (it owns shop membership).
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	st := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Payouts"; o.Auth = []string{auth.AudStaff}; return o }
	base := "/api/v1/staff/finance/payout-batches"
	me := func(ctx context.Context) uuid.UUID { return httpx.MustPrincipal(ctx).UserID }
	threshold := s.d.Cfg.PayoutApprovalThreshold
	s.devRoutes(r)

	httpx.Register(g, st(httpx.Op{ID: "staffPreparePayoutBatch", Method: http.MethodPost, Path: base, Perm: "payouts.prepare", Status: 201,
		Summary: "Prepare a draft batch from every seller who can be paid now"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body PayoutBatchDetail }, error) {
			b, skipped, err := s.Prepare(ctx, me(ctx))
			if err != nil {
				return nil, err
			}
			d, err := s.detail(ctx, b.ID)
			d.Skipped = skipped
			return &struct{ Body PayoutBatchDetail }{d}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffListPayoutBatches", Method: http.MethodGet, Path: base, Perm: "finance.view", Summary: "BatchPayout batches, newest first"}),
		func(ctx context.Context, in *struct {
			Status string `query:"status" enum:"draft,submitted,approved,processing,done,cancelled,"`
			Cursor string `query:"cursor"`
			Limit  int    `query:"limit"`
		}) (*struct{ Body httpx.List[PayoutBatch] }, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, httpx.BadCursor()
			}
			lim := httpx.Limit(in.Limit)
			var status *string
			if in.Status != "" {
				status = &in.Status
			}
			rows, err := s.d.Q.PayoutsListBatches(ctx, store.PayoutsListBatchesParams{Status: status, Before: before, Lim: int32(lim + 1)})
			if err != nil {
				return nil, err
			}
			views := make([]PayoutBatch, len(rows))
			for i, b := range rows {
				views[i] = batchView(b, threshold)
			}
			return &struct{ Body httpx.List[PayoutBatch] }{httpx.Paginate(s.d.Cursors, views, lim, func(v PayoutBatch) time.Time { return v.CreatedAt })}, nil
		})
	httpx.Register(g, st(httpx.Op{ID: "staffGetPayoutBatch", Method: http.MethodGet, Path: base + "/{id}", Perm: "finance.view", Summary: "A batch with its payouts"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body PayoutBatchDetail }, error) {
			d, err := s.detail(ctx, in.ID)
			return &struct{ Body PayoutBatchDetail }{d}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffExcludePayout", Method: http.MethodPost, Path: base + "/{id}/payouts/{payoutId}/exclude", Perm: "payouts.prepare",
		Summary: "Take a seller out of a draft batch (optionally holding later payouts too)"}),
		func(ctx context.Context, in *struct {
			ID       uuid.UUID `path:"id"`
			PayoutID uuid.UUID `path:"payoutId"`
			Body     struct {
				Hold bool   `json:"hold,omitempty" doc:"Also hold the seller's later payouts until finance releases it"`
				Note string `json:"note" minLength:"3" maxLength:"500" doc:"Why (kept in the audit log)"`
			}
		}) (*struct{ Body PayoutBatchDetail }, error) {
			if _, err := s.Exclude(ctx, in.ID, in.PayoutID, me(ctx), in.Body.Hold, in.Body.Note); err != nil {
				return nil, err
			}
			d, err := s.detail(ctx, in.ID)
			return &struct{ Body PayoutBatchDetail }{d}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffSubmitPayoutBatch", Method: http.MethodPost, Path: base + "/{id}/submit", Perm: "payouts.prepare", Summary: "Send a draft batch for approval"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body PayoutBatch }, error) {
			b, err := s.Submit(ctx, in.ID, me(ctx))
			return &struct{ Body PayoutBatch }{batchView(b, threshold)}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffApprovePayoutBatch", Method: http.MethodPost, Path: base + "/{id}/approve", Perm: "payouts.approve",
		Summary: "Approve a submitted batch (not its preparer; fresh MFA; two managers above the threshold)"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct {
			Body struct {
				PayoutBatch PayoutBatch `json:"batch"`
				Released    bool        `json:"released" doc:"True when the transfers have started; false when a second approver is still needed"`
			}
		}, error) {
			b, final, err := s.Approve(ctx, in.ID, me(ctx))
			out := &struct {
				Body struct {
					PayoutBatch PayoutBatch `json:"batch"`
					Released    bool        `json:"released" doc:"True when the transfers have started; false when a second approver is still needed"`
				}
			}{}
			out.Body.PayoutBatch, out.Body.Released = batchView(b, threshold), final
			return out, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffCancelPayoutBatch", Method: http.MethodPost, Path: base + "/{id}/cancel", Perm: "payouts.prepare",
		Summary: "Cancel a batch that isn't approved yet"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body PayoutBatch }, error) {
			b, err := s.Cancel(ctx, in.ID, me(ctx))
			return &struct{ Body PayoutBatch }{batchView(b, threshold)}, err
		})

	httpx.Register(g, st(httpx.Op{ID: "staffHoldSellerPayouts", Method: http.MethodPost, Path: "/api/v1/staff/finance/sellers/{sellerId}/payout-holds", Perm: "payouts.prepare",
		Status: 201, Summary: "Hold a seller's payouts"}),
		func(ctx context.Context, in *struct {
			SellerID uuid.UUID `path:"sellerId"`
			Body     struct {
				Reason string     `json:"reason" enum:"risk_case,enforcement,manual"`
				Note   string     `json:"note" minLength:"3" maxLength:"500"`
				Until  *time.Time `json:"until,omitempty" doc:"Releases itself at this time; empty = until released"`
			}
		}) (*struct{ Body store.FinancePayoutHold }, error) {
			h, err := s.Hold(ctx, in.SellerID, in.Body.Reason, in.Body.Note, in.Body.Until, me(ctx))
			return &struct{ Body store.FinancePayoutHold }{h}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffReleasePayoutHold", Method: http.MethodPost, Path: "/api/v1/staff/finance/payout-holds/{id}/release", Perm: "payouts.approve",
		Summary: "Release a payout hold"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body store.FinancePayoutHold }, error) {
			h, err := s.Release(ctx, in.ID, me(ctx))
			return &struct{ Body store.FinancePayoutHold }{h}, err
		})
}

// devRoutes ages a seller's deliveries past the return window and their bank accounts past the
// 24-hour hold (development only), so tests can pay out without waiting a week.
func (s *Service) devRoutes(r *kit.Routes) {
	if !s.d.Cfg.IsDevelopment() {
		return
	}
	r.Gin.POST("/dev/payouts/age/:seller", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("seller"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "use /dev/payouts/age/{sellerId}"})
			return
		}
		if err := s.d.Q.PayoutsDevAgeSeller(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"aged": true})
	})
}

func (s *Service) detail(ctx context.Context, batchID uuid.UUID) (PayoutBatchDetail, error) {
	b, err := s.d.Q.PayoutsGetBatch(ctx, batchID)
	if err != nil {
		return PayoutBatchDetail{}, httpx.NotFound("BatchPayout batch")
	}
	rows, err := s.d.Q.PayoutsBatchPayouts(ctx, &batchID)
	out := PayoutBatchDetail{PayoutBatch: batchView(b, s.d.Cfg.PayoutApprovalThreshold), Payouts: make([]BatchPayout, len(rows))}
	for i, r := range rows {
		out.Payouts[i] = payoutView(r)
	}
	return out, err
}
