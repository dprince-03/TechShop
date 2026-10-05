package logistics

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// housekeepingArgs closes forgotten shifts (12-hour timeout stops location sharing) and
// deletes raw location points older than 90 days.
type housekeepingArgs struct{}

func (housekeepingArgs) Kind() string { return "logistics.housekeeping" }

type housekeepingWorker struct {
	river.WorkerDefaults[housekeepingArgs]
	s *Service
}

func (w *housekeepingWorker) Work(ctx context.Context, _ *river.Job[housekeepingArgs]) error {
	stale, err := w.s.d.Q.LogisticsStaleShifts(ctx)
	if err != nil {
		return err
	}
	for _, sh := range stale {
		err := w.s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
			if _, err := tx.Q.LogisticsEndShift(ctx, sh.RiderID); err != nil {
				return nil // ended meanwhile
			}
			return tx.Q.LogisticsSetRiderStatus(ctx, store.LogisticsSetRiderStatusParams{UserID: sh.RiderID, Status: "off_shift"})
		})
		if err != nil {
			return err
		}
	}
	n, err := w.s.d.Q.LogisticsPurgePings(ctx)
	if n > 0 {
		w.s.d.Logger.Info("purged old rider location points", "rows", n)
	}
	return err
}

// Jobs implements kit.Module.
func (s *Service) Jobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &housekeepingWorker{s: s})
	reg.Every(15*time.Minute, func() (river.JobArgs, *river.InsertOpts) { return housekeepingArgs{}, nil })
}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}
