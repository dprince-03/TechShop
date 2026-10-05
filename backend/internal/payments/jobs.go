package payments

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

func critical() *river.InsertOpts { return &river.InsertOpts{Queue: jobs.QueueCritical} }
func finance() *river.InsertOpts  { return &river.InsertOpts{Queue: jobs.QueueFinance} }

// ProcessEventArgs processes one stored webhook.
type ProcessEventArgs struct {
	EventID uuid.UUID `json:"eventId"`
}

// Kind identifies the job.
func (ProcessEventArgs) Kind() string { return "payments.process_event" }

type processEventWorker struct {
	river.WorkerDefaults[ProcessEventArgs]
	s *Service
}

func (w *processEventWorker) Work(ctx context.Context, j *river.Job[ProcessEventArgs]) error {
	return w.s.ProcessEvent(ctx, j.Args.EventID)
}

// ExecuteRefundArgs pays out one approved refund.
type ExecuteRefundArgs struct {
	RefundID uuid.UUID `json:"refundId"`
}

// Kind identifies the job.
func (ExecuteRefundArgs) Kind() string { return "payments.execute_refund" }

type executeRefundWorker struct {
	river.WorkerDefaults[ExecuteRefundArgs]
	s *Service
}

func (w *executeRefundWorker) Work(ctx context.Context, j *river.Job[ExecuteRefundArgs]) error {
	return w.s.Execute(ctx, j.Args.RefundID)
}

type expiryArgs struct{}

func (expiryArgs) Kind() string { return "payments.expire_orders" }

type expiryWorker struct {
	river.WorkerDefaults[expiryArgs]
	s *Service
}

func (w *expiryWorker) Work(ctx context.Context, _ *river.Job[expiryArgs]) error {
	_, err := w.s.ExpireOrders(ctx)
	return err
}

type sweepArgs struct{}

func (sweepArgs) Kind() string { return "payments.sweep" }

type sweepWorker struct {
	river.WorkerDefaults[sweepArgs]
	s *Service
}

func (w *sweepWorker) Work(ctx context.Context, _ *river.Job[sweepArgs]) error { return w.s.Sweep(ctx) }

// Jobs implements kit.Module: webhook processing, refunds, expiry (every minute) and the
// pending-payment sweep (every 5 minutes).
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &processEventWorker{s: s})
	river.AddWorker(reg.Workers, &executeRefundWorker{s: s})
	river.AddWorker(reg.Workers, &expiryWorker{s: s})
	river.AddWorker(reg.Workers, &sweepWorker{s: s})
	reg.Every(time.Minute, func() (river.JobArgs, *river.InsertOpts) { return expiryArgs{}, critical() })
	reg.Every(5*time.Minute, func() (river.JobArgs, *river.InsertOpts) { return sweepArgs{}, critical() })
}

// Events implements kit.Module: cancelling a paid order (or a line) raises an automatic refund
// request; a finance officer approves it (dual control holds for system requests too).
func (s *Service) Events(r *outbox.Relay) {
	r.On("order.cancelled", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		var p struct {
			OrderID uuid.UUID `json:"orderId"`
			Paid    bool      `json:"paid"`
			Reason  string    `json:"reason"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil || !p.Paid {
			return err
		}
		return s.systemRefund(ctx, tx, p.OrderID, 0, "Order cancelled: "+p.Reason)
	})
	r.On("order.line_cancelled", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		var p struct {
			OrderID    uuid.UUID `json:"orderId"`
			RefundKobo int64     `json:"refundKobo"`
			Reason     string    `json:"reason"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return err
		}
		return s.systemRefund(ctx, tx, p.OrderID, p.RefundKobo, "Item cancelled: "+p.Reason)
	})
}

// systemRefund requests a refund of amount (0 = everything still refundable).
func (s *Service) systemRefund(ctx context.Context, tx *uow.Tx, orderID uuid.UUID, amount int64, reason string) error {
	pay, err := tx.Q.PaymentsSucceededForOrder(ctx, &orderID)
	if err != nil {
		return nil // nothing captured, nothing to refund
	}
	already, err := tx.Q.PaymentsRefundedTotal(ctx, pay.ID)
	if err != nil {
		return err
	}
	left := pay.AmountKobo - already
	if amount == 0 || amount > left {
		amount = left
	}
	if amount <= 0 {
		return nil
	}
	_, err = s.RequestRefund(ctx, tx, orderID, amount, reason, "original", nil)
	return err
}
