package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
)

// ---- Nightly integrity checks (§4.2): trial balance = 0; every succeeded payment has a sale journal.

type checkArgs struct{}

func (checkArgs) Kind() string { return "ledger.integrity_check" }

type checkWorker struct {
	river.WorkerDefaults[checkArgs]
	s *Service
}

func (w *checkWorker) Work(ctx context.Context, _ *river.Job[checkArgs]) error {
	_, err := w.s.Check(ctx)
	return err
}

// Check runs the integrity checks and raises reconciliation exceptions for problems found.
func (s *Service) Check(ctx context.Context) ([]string, error) {
	var problems []string
	tb, err := s.d.Q.FinanceTrialBalance(ctx)
	if err != nil {
		return nil, err
	}
	if tb.Total != 0 {
		problems = append(problems, fmt.Sprintf("trial balance is %d kobo, not zero", tb.Total))
	}
	missing, err := s.d.Q.FinanceSucceededPaymentsWithoutSale(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range missing {
		problems = append(problems, "payment "+m.ID.String()+" has no sale journal")
	}
	if len(problems) == 0 {
		return nil, nil
	}
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		for _, p := range problems {
			if _, err := tx.Q.PaymentsCreateException(ctx, store.PaymentsCreateExceptionParams{Kind: "missing_in_ours", Detail: "Ledger check: " + p}); err != nil {
				return err
			}
		}
		return nil
	})
	s.d.Logger.Error("ledger integrity check failed", "problems", problems)
	return problems, err
}

func (s *Service) registerJobs(reg *jobs.Registry) {
	river.AddWorker(reg.Workers, &checkWorker{s: s})
	reg.Every(24*time.Hour, func() (river.JobArgs, *river.InsertOpts) {
		return checkArgs{}, &river.InsertOpts{Queue: jobs.QueueFinance}
	})
}
