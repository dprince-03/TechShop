package marketing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// attributionWindow: an order placed within 7 days of clicking a marketing message is credited to it.
const attributionWindow = 7 * 24 * time.Hour

func marketingQueue() *river.InsertOpts { return &river.InsertOpts{Queue: jobs.QueueMarketing} }

// SendCampaignArgs sends a campaign to its pending recipients in batches.
type SendCampaignArgs struct {
	CampaignID uuid.UUID `json:"campaignId"`
}

// Kind identifies the job.
func (SendCampaignArgs) Kind() string { return "marketing.send_campaign" }

type sendCampaignWorker struct {
	river.WorkerDefaults[SendCampaignArgs]
	s *Service
}

func (w *sendCampaignWorker) Work(ctx context.Context, j *river.Job[SendCampaignArgs]) error {
	for {
		more, err := w.s.SendBatch(ctx, j.Args.CampaignID)
		if err != nil || !more {
			return err
		}
	}
}

// tickArgs runs the per-minute marketing chores: promotion schedule, due campaigns, journeys.
type tickArgs struct{}

func (tickArgs) Kind() string { return "marketing.tick" }

type tickWorker struct {
	river.WorkerDefaults[tickArgs]
	s *Service
}

func (w *tickWorker) Work(ctx context.Context, _ *river.Job[tickArgs]) error {
	if _, err := w.s.d.Q.MarketingAdvancePromotions(ctx); err != nil {
		return err
	}
	if err := w.s.StartDue(ctx); err != nil {
		return err
	}
	return w.s.RunDue(ctx)
}

type abandonedArgs struct{}

func (abandonedArgs) Kind() string { return "marketing.abandoned_carts" }

type abandonedWorker struct {
	river.WorkerDefaults[abandonedArgs]
	s *Service
}

// Work enrols people with idle carts in the abandoned-cart journey (reminders only go out with
// marketing consent — the notification policy checks it per message).
func (w *abandonedWorker) Work(ctx context.Context, _ *river.Job[abandonedArgs]) error {
	carts, err := w.s.d.Q.MarketingAbandonedCarts(ctx)
	if err != nil {
		return err
	}
	return w.s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		for _, c := range carts {
			if c.UserID == nil {
				continue
			}
			if err := w.s.Enroll(ctx, tx, "cart_abandoned", *c.UserID, map[string]any{"cartId": c.CartID, "items": c.Items, "valueKobo": c.ValueKobo}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Jobs implements kit.Module.
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &sendCampaignWorker{s: s})
	river.AddWorker(reg.Workers, &tickWorker{s: s})
	river.AddWorker(reg.Workers, &abandonedWorker{s: s})
	reg.Every(time.Minute, func() (river.JobArgs, *river.InsertOpts) { return tickArgs{}, marketingQueue() })
	reg.Every(time.Hour, func() (river.JobArgs, *river.InsertOpts) { return abandonedArgs{}, marketingQueue() })
}

// Events implements kit.Module: journey triggers and order attribution.
func (s *Service) Events(r *outbox.Relay) {
	userOf := func(ev outbox.Event) (uuid.UUID, bool) {
		var p struct {
			UserID *uuid.UUID `json:"userId"`
		}
		if json.Unmarshal(ev.Payload, &p) != nil || p.UserID == nil {
			return uuid.Nil, false
		}
		return *p.UserID, true
	}
	r.On("user.signed_up", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		if u, ok := userOf(ev); ok {
			return s.Enroll(ctx, tx, "signed_up", u, nil)
		}
		return nil
	})
	r.On("order.paid", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		u, ok := userOf(ev)
		if !ok {
			return nil
		}
		var p struct {
			OrderID uuid.UUID `json:"orderId"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		since := time.Now().Add(-attributionWindow)
		if n, err := tx.Q.MarketingLastClick(ctx, store.MarketingLastClickParams{UserID: &u, ClickedAt: &since}); err == nil {
			if err := tx.Q.SalesSetAttribution(ctx, store.SalesSetAttributionParams{ID: p.OrderID, AttributedNotificationID: &n}); err != nil {
				return err
			}
		}
		return s.Enroll(ctx, tx, "order_paid", u, map[string]any{"orderId": p.OrderID})
	})
	r.On("order.delivered", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		if u, ok := userOf(ev); ok {
			return s.Enroll(ctx, tx, "order_delivered", u, nil)
		}
		return nil
	})
}
