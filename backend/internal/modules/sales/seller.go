package sales

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// Seller-shipped fulfilment (orders-fulfilment.md §2.1, §4 "Seller SLAs"): sellers accept within
// 24 hours and pack within their listings' handling time. A breach cancels the fulfilment,
// raises the refund request, records an SLA event and tells the seller; three breaches in 30
// days send the seller's live listings back to moderation (trust-safety.md §7).

// slaStrikeLimit is the number of breaches in 30 days that triggers listing review.
const slaStrikeLimit = 3

// handoverWindow is how long a seller has to drop a packed parcel at a TechShop hub (Jumia
// Nigeria's vendor rule: drop off within 24 hours of ready-to-ship; docs/plan.md 2026-10-21).
const handoverWindow = 24 * time.Hour

// wat is Nigeria's time zone (UTC+1, no daylight saving), fixed so no tzdata is needed.
var wat = time.FixedZone("WAT", 60*60)

// handoverDeadline is packedAt plus the window; a deadline falling on a Sunday (WAT), when
// hubs don't receive, moves to the same time on Monday.
func handoverDeadline(packedAt time.Time) time.Time {
	d := packedAt.Add(handoverWindow)
	if d.In(wat).Weekday() == time.Sunday {
		d = d.Add(24 * time.Hour)
	}
	return d
}

// PackSellerFulfilment marks an accepted seller fulfilment packed and starts the hand-over
// deadline.
func (s *Service) PackSellerFulfilment(ctx context.Context, tx *uow.Tx, fulfilmentID, actor uuid.UUID, note string) (store.SalesFulfilment, error) {
	out, err := s.Move(ctx, tx, fulfilmentID, "packed", &actor, note)
	if err != nil {
		return out, err
	}
	by := handoverDeadline(time.Now())
	out.HandoverBy = &by
	return out, tx.Q.SalesSetHandoverBy(ctx, store.SalesSetHandoverByParams{ID: fulfilmentID, HandoverBy: &by})
}

// AcceptFulfilment records the seller's acceptance and sets the packing deadline from the
// longest handling time among the fulfilment's listings.
func (s *Service) AcceptFulfilment(ctx context.Context, tx *uow.Tx, f store.SalesFulfilment, actor uuid.UUID) (store.SalesFulfilment, error) {
	out, err := s.Move(ctx, tx, f.ID, "accepted", &actor, "")
	if err != nil {
		return out, err
	}
	days, err := tx.Q.SalesFulfilmentHandlingDays(ctx, f.ID)
	if err != nil {
		return out, err
	}
	packBy := time.Now().Add(time.Duration(max(days, 1)) * 24 * time.Hour)
	out.PackBy = &packBy
	return out, tx.Q.SalesSetPackBy(ctx, store.SalesSetPackByParams{ID: f.ID, PackBy: &packBy})
}

// CancelFulfilment cancels a seller-shipped fulfilment that hasn't been packed: every open line
// is cancelled and its stock released, and one refund request covers the lines and the delivery
// fee (the payments module listens for order.line_cancelled). slaKind records why it counts
// against the seller ("seller_cancel", "late_accept", "late_pack", "late_handover"). A packed
// parcel can only be cancelled for a missed hand-over: it never reached TechShop.
func (s *Service) CancelFulfilment(ctx context.Context, tx *uow.Tx, fulfilmentID uuid.UUID, actor *uuid.UUID, reason, slaKind string) error {
	f, err := tx.Q.SalesGetFulfilmentForUpdate(ctx, fulfilmentID)
	if err != nil {
		return httpx.NotFound("Fulfilment")
	}
	lateHandover := f.Status == "packed" && f.FulfilledBy == "seller" && slaKind == "late_handover"
	if f.Status != "pending" && f.Status != "accepted" && !lateHandover {
		return httpx.Conflict("too_late_to_cancel", "Only orders that haven't been packed can be cancelled.")
	}
	o, err := tx.Q.SalesGetOrderForUpdate(ctx, f.OrderID)
	if err != nil {
		return err
	}
	lines, err := tx.Q.SalesFulfilmentLines(ctx, f.ID)
	if err != nil {
		return err
	}
	refund := f.DeliveryFeeKobo
	for _, l := range lines {
		if l.CancelledAt != nil {
			continue
		}
		cancelled, err := tx.Q.SalesCancelLine(ctx, store.SalesCancelLineParams{ID: l.ID, OrderID: o.ID})
		if err != nil {
			return err
		}
		if err := s.restockLine(ctx, tx, cancelled, actor); err != nil {
			return err
		}
		refund += cancelled.LineTotalKobo
	}
	if _, err := s.Move(ctx, tx, f.ID, "cancelled", actor, reason); err != nil {
		return err
	}
	if slaKind != "" {
		if err := tx.Q.FulfilmentInsertSLAEvent(ctx, store.FulfilmentInsertSLAEventParams{SellerID: f.SellerID, FulfilmentID: f.ID, Kind: slaKind}); err != nil {
			return err
		}
	}
	tx.Emit("order", o.ID, "order.line_cancelled", map[string]any{"orderId": o.ID, "orderNumber": o.OrderNumber, "fulfilmentId": f.ID, "refundKobo": refund, "reason": reason})
	if o.CustomerUserID != nil {
		if err := s.notify.Send(ctx, tx, notify.Msg{UserID: o.CustomerUserID, Channel: "push", Category: notify.CatOrders, Template: "order_cancelled",
			Data: map[string]any{"order": o.OrderNumber, "reason": reason}, Key: "fulfilment:" + f.ID.String() + ":cancelled"}); err != nil {
			return err
		}
	}
	return nil
}

// tellSellerNewOrder lets the shop's owner know a paid order is waiting for acceptance.
func (s *Service) tellSellerNewOrder(ctx context.Context, tx *uow.Tx, fulfilmentID, sellerID uuid.UUID, orderNumber string) error {
	seller, err := tx.Q.SellersGet(ctx, sellerID)
	if err != nil || seller.OwnerUserID == nil {
		return err
	}
	owner, err := tx.Q.IdentityGetUser(ctx, *seller.OwnerUserID)
	if err != nil {
		return err
	}
	data := map[string]any{"order": orderNumber}
	if err := s.notify.Send(ctx, tx, notify.Msg{UserID: &owner.ID, Channel: "push", Category: notify.CatSeller, Template: "seller_new_order",
		Data: data, Key: "fulfilment:" + fulfilmentID.String() + ":new:push"}); err != nil {
		return err
	}
	if owner.Email == nil {
		return nil
	}
	return s.notify.Send(ctx, tx, notify.Msg{UserID: &owner.ID, To: *owner.Email, Channel: "email", Category: notify.CatSeller, Template: "seller_new_order",
		Data: data, Key: "fulfilment:" + fulfilmentID.String() + ":new:email"})
}

// ---- SLA enforcement (every 5 minutes)

type slaArgs struct{}

// Kind identifies the job.
func (slaArgs) Kind() string { return "sales.seller_sla" }

type slaWorker struct {
	river.WorkerDefaults[slaArgs]
	s *Service
}

func (w *slaWorker) Work(ctx context.Context, _ *river.Job[slaArgs]) error {
	return w.s.EnforceSellerSLAs(ctx)
}

// EnforceSellerSLAs cancels seller fulfilments that missed the accept, pack or hand-over deadline. Each
// breach is its own transaction, so one failure doesn't block the rest.
func (s *Service) EnforceSellerSLAs(ctx context.Context) error {
	late, err := s.d.Q.SalesPendingAcceptBreaches(ctx)
	if err != nil {
		return err
	}
	for _, f := range late {
		if err := s.breach(ctx, f.ID, f.SellerID, "late_accept", "The seller didn't accept the order in time."); err != nil {
			s.d.Logger.Error("seller SLA cancel failed", "fulfilment", f.ID, "error", err)
		}
	}
	unpacked, err := s.d.Q.SalesPendingPackBreaches(ctx)
	if err != nil {
		return err
	}
	for _, f := range unpacked {
		if err := s.breach(ctx, f.ID, f.SellerID, "late_pack", "The seller didn't pack the order in time."); err != nil {
			s.d.Logger.Error("seller SLA cancel failed", "fulfilment", f.ID, "error", err)
		}
	}
	undelivered, err := s.d.Q.SalesPendingHandoverBreaches(ctx)
	if err != nil {
		return err
	}
	for _, f := range undelivered {
		if err := s.breach(ctx, f.ID, f.SellerID, "late_handover", "The seller didn't bring the parcel to a TechShop hub in time."); err != nil {
			s.d.Logger.Error("seller SLA cancel failed", "fulfilment", f.ID, "error", err)
		}
	}
	return nil
}

func (s *Service) breach(ctx context.Context, fulfilmentID, sellerID uuid.UUID, kind, reason string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if err := s.CancelFulfilment(ctx, tx, fulfilmentID, nil, reason, kind); err != nil {
			return err
		}
		if err := tx.Audit("seller.sla_breach", "fulfilment", fulfilmentID.String(), map[string]string{"kind": kind}); err != nil {
			return err
		}
		strikes, err := tx.Q.FulfilmentSLABreaches30d(ctx, sellerID)
		if err != nil || strikes < slaStrikeLimit {
			return err
		}
		ids, err := tx.Q.CatalogSendSellerListingsToReview(ctx, sellerID)
		for _, id := range ids {
			tx.Emit("listing", id, "listing.changed", map[string]any{"listingId": id, "status": "in_review"})
		}
		if len(ids) > 0 {
			return tx.Audit("seller.listings_to_review", "seller", sellerID.String(), map[string]any{"reason": "sla_breaches_30d", "strikes": strikes, "listings": len(ids)})
		}
		return err
	})
}

func (s *Service) registerSellerJobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &slaWorker{s: s})
	reg.Every(5*time.Minute, func() (river.JobArgs, *river.InsertOpts) {
		return slaArgs{}, &river.InsertOpts{Queue: jobs.QueueCritical}
	})
}

// devSLARoutes lets tests trigger an SLA breach now (development only): the deadline moves into
// the past and the enforcement runs immediately.
func (s *Service) devSLARoutes(r *kit.Routes) {
	if !s.d.Cfg.IsDevelopment() {
		return
	}
	r.Gin.POST("/dev/sla/:fulfilment/:which", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("fulfilment"))
		which := c.Param("which")
		if err != nil || (which != "accept" && which != "pack" && which != "handover") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "use /dev/sla/{fulfilmentId}/accept|pack|handover"})
			return
		}
		if err := s.d.Q.SalesDevExpireDeadline(c.Request.Context(), store.SalesDevExpireDeadlineParams{ID: id, Which: which}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := s.EnforceSellerSLAs(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"enforced": true})
	})
}
