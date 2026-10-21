package payouts

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/outbox"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
)

func finance() *river.InsertOpts { return &river.InsertOpts{Queue: jobs.QueueFinance} }

// processBatchArgs starts the transfers of an approved batch.
type processBatchArgs struct {
	BatchID uuid.UUID `json:"batchId"`
}

// Kind identifies the job.
func (processBatchArgs) Kind() string { return "payouts.process_batch" }

type processBatchWorker struct {
	river.WorkerDefaults[processBatchArgs]
	s *Service
}

func (w *processBatchWorker) Work(ctx context.Context, j *river.Job[processBatchArgs]) error {
	return w.s.ProcessBatch(ctx, j.Args.BatchID)
}

// recheckArgs asks the provider about one payout (after a transfer webhook).
type recheckArgs struct {
	PayoutID uuid.UUID `json:"payoutId"`
}

// Kind identifies the job.
func (recheckArgs) Kind() string { return "payouts.recheck" }

type recheckWorker struct {
	river.WorkerDefaults[recheckArgs]
	s *Service
}

func (w *recheckWorker) Work(ctx context.Context, j *river.Job[recheckArgs]) error {
	return w.s.Recheck(ctx, j.Args.PayoutID)
}

type sweepArgs struct{}

// Kind identifies the job.
func (sweepArgs) Kind() string { return "payouts.sweep" }

type sweepWorker struct {
	river.WorkerDefaults[sweepArgs]
	s *Service
}

func (w *sweepWorker) Work(ctx context.Context, _ *river.Job[sweepArgs]) error { return w.s.Sweep(ctx) }

// Jobs implements kit.Module: batch processing, webhook re-checks and a 10-minute sweep of
// payouts stuck in processing.
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &processBatchWorker{s: s})
	river.AddWorker(reg.Workers, &recheckWorker{s: s})
	river.AddWorker(reg.Workers, &sweepWorker{s: s})
	reg.Every(10*time.Minute, func() (river.JobArgs, *river.InsertOpts) { return sweepArgs{}, finance() })
}

// Events implements kit.Module: a provider transfer webhook (relayed by payments) prompts a
// re-check of that payout; the webhook body itself is never trusted.
func (s *Service) Events(r *outbox.Relay) {
	r.On("transfer.updated", func(ctx context.Context, tx *uow.Tx, ev outbox.Event) error {
		var p struct {
			PayoutID uuid.UUID `json:"payoutId"`
		}
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return err
		}
		tx.Enqueue(recheckArgs{PayoutID: p.PayoutID}, finance())
		return nil
	})
}
