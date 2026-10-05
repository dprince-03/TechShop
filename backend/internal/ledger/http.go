package ledger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// JournalView is a journal with its entries.
type JournalView struct {
	store.FinanceLedgerJournal
	Entries []store.FinanceJournalEntriesRow `json:"entries"`
}

// AccountLedger is one account's entries plus its live balance.
type AccountLedger struct {
	httpx.List[store.FinanceAccountLedgerRow]
	BalanceKobo int64 `json:"balanceKobo"`
}

// Routes implements kit.Module (staff finance: ledger, accounts, periods).
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	op := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Finance"; o.Auth = []string{auth.AudStaff}; return o }

	httpx.Register(g, op(httpx.Op{ID: "financeDashboard", Method: http.MethodGet, Path: "/api/v1/staff/finance/dashboard", Perm: "finance.view", Summary: "Cash position, owed to sellers, VAT, revenue"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body store.FinanceDashboardRow }, error) {
			d, err := s.d.Q.FinanceDashboard(ctx)
			return &struct{ Body store.FinanceDashboardRow }{d}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "financeTrialBalance", Method: http.MethodGet, Path: "/api/v1/staff/finance/trial-balance", Perm: "finance.view", Summary: "Trial balance (must be zero) and account balances"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body struct {
				Total    int64                             `json:"totalKobo"`
				Balanced bool                              `json:"balanced"`
				Entries  int64                             `json:"entries"`
				Accounts []store.FinanceAccountBalancesRow `json:"accounts"`
			}
		}, error) {
			out := &struct {
				Body struct {
					Total    int64                             `json:"totalKobo"`
					Balanced bool                              `json:"balanced"`
					Entries  int64                             `json:"entries"`
					Accounts []store.FinanceAccountBalancesRow `json:"accounts"`
				}
			}{}
			tb, err := s.d.Q.FinanceTrialBalance(ctx)
			if err != nil {
				return nil, err
			}
			out.Body.Total, out.Body.Entries, out.Body.Balanced = tb.Total, tb.Entries, tb.Total == 0
			out.Body.Accounts, err = s.d.Q.FinanceAccountBalances(ctx)
			return out, err
		})
	httpx.Register(g, op(httpx.Op{ID: "financeAccountLedger", Method: http.MethodGet, Path: "/api/v1/staff/finance/accounts/{code}/ledger", Perm: "finance.view", Summary: "General ledger for one account"}),
		func(ctx context.Context, in *struct {
			Code   string `path:"code"`
			Cursor string `query:"cursor"`
			Limit  int    `query:"limit"`
		}) (*struct{ Body AccountLedger }, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, err
			}
			lim := httpx.Limit(in.Limit)
			rows, err := s.d.Q.FinanceAccountLedger(ctx, store.FinanceAccountLedgerParams{Code: in.Code, Before: before, Limit: int32(lim + 1)})
			if err != nil {
				return nil, err
			}
			out := AccountLedger{List: httpx.Paginate(s.d.Cursors, rows, lim, func(r store.FinanceAccountLedgerRow) time.Time { return r.CreatedAt })}
			out.BalanceKobo, err = s.d.Q.FinanceAccountBalance(ctx, in.Code)
			return &struct{ Body AccountLedger }{out}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "financeListJournals", Method: http.MethodGet, Path: "/api/v1/staff/finance/journals", Perm: "finance.view", Summary: "Journals, newest first"}),
		func(ctx context.Context, in *struct {
			Kind   string `query:"kind"`
			Cursor string `query:"cursor"`
			Limit  int    `query:"limit"`
		}) (*struct {
			Body httpx.List[store.FinanceLedgerJournal]
		}, error) {
			before, err := s.d.Cursors.Before(in.Cursor)
			if err != nil {
				return nil, err
			}
			lim := httpx.Limit(in.Limit)
			p := store.FinanceListJournalsParams{Limit: int32(lim + 1), Before: before}
			if in.Kind != "" {
				p.Kind = &in.Kind
			}
			rows, err := s.d.Q.FinanceListJournals(ctx, p)
			if err != nil {
				return nil, err
			}
			return &struct {
				Body httpx.List[store.FinanceLedgerJournal]
			}{httpx.Paginate(s.d.Cursors, rows, lim, func(r store.FinanceLedgerJournal) time.Time { return r.PostedAt })}, nil
		})
	httpx.Register(g, op(httpx.Op{ID: "financeGetJournal", Method: http.MethodGet, Path: "/api/v1/staff/finance/journals/{id}", Perm: "finance.view", Summary: "One journal with its entries"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body JournalView }, error) {
			v, err := s.View(ctx, s.d.Q, in.ID)
			return &struct{ Body JournalView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "financeReverseJournal", Method: http.MethodPost, Path: "/api/v1/staff/finance/journals/{id}/reverse", Perm: "ledger.adjust", StepUp: true, Status: 201,
		Summary: "Reverse a journal (posts the negation; journals are never edited)"}),
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Memo string `json:"memo" minLength:"5" maxLength:"500"`
			}
		}) (*struct{ Body JournalView }, error) {
			actor := httpx.MustPrincipal(ctx).UserID
			var id uuid.UUID
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				jr, err := s.Reverse(ctx, tx, in.ID, in.Body.Memo, &actor)
				if err != nil {
					return err
				}
				id = jr.ID
				return tx.Audit("ledger.reversed", "journal", in.ID.String(), map[string]any{"reversal": jr.ID, "memo": in.Body.Memo})
			})
			if err != nil {
				return nil, err
			}
			v, err := s.View(ctx, s.d.Q, id)
			return &struct{ Body JournalView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "financePostAdjustment", Method: http.MethodPost, Path: "/api/v1/staff/finance/adjustments", Perm: "ledger.adjust", StepUp: true, Status: 201,
		Summary: "Post a manual adjustment (memo required, must balance)"}),
		func(ctx context.Context, in *struct {
			Key  string `header:"Idempotency-Key" required:"true"`
			Body struct {
				Memo  string `json:"memo" minLength:"5" maxLength:"500"`
				Lines []struct {
					Account    string `json:"account" minLength:"3"`
					AmountKobo int64  `json:"amountKobo"`
				} `json:"lines" minItems:"2" maxItems:"20"`
			}
		}) (*struct{ Body JournalView }, error) {
			actor := httpx.MustPrincipal(ctx).UserID
			lines := make([]Line, len(in.Body.Lines))
			for i, l := range in.Body.Lines {
				lines[i] = Line{l.Account, l.AmountKobo}
			}
			var id uuid.UUID
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				jr, err := s.Adjust(ctx, tx, in.Key, in.Body.Memo, lines, actor)
				if err != nil {
					return err
				}
				id = jr.ID
				return tx.Audit("ledger.adjusted", "journal", jr.ID.String(), in.Body)
			})
			if err != nil {
				return nil, err
			}
			v, err := s.View(ctx, s.d.Q, id)
			return &struct{ Body JournalView }{v}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "financeListPeriods", Method: http.MethodGet, Path: "/api/v1/staff/finance/periods", Perm: "finance.view", Summary: "Accounting periods"}),
		func(ctx context.Context, _ *struct{}) (*struct {
			Body []store.FinanceAccountingPeriod
		}, error) {
			rows, err := s.d.Q.FinanceListPeriods(ctx)
			return &struct {
				Body []store.FinanceAccountingPeriod
			}{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "financeClosePeriod", Method: http.MethodPost, Path: "/api/v1/staff/finance/periods/{month}/close", Perm: "finance.period_close", StepUp: true,
		Summary: "Close a month: no more postings into it"}),
		func(ctx context.Context, in *struct {
			Month string `path:"month" pattern:"^[0-9]{4}-[0-9]{2}$"`
		}) (*struct{ Body store.FinanceAccountingPeriod }, error) {
			m, err := time.Parse("2006-01", in.Month)
			if err != nil {
				return nil, httpx.Invalid("bad_month", "Use YYYY-MM.")
			}
			if !m.AddDate(0, 1, 0).Before(time.Now()) {
				return nil, httpx.Invalid("period_not_over", "A month can only be closed after it ends.")
			}
			actor := httpx.MustPrincipal(ctx).UserID
			var out store.FinanceAccountingPeriod
			err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.FinanceClosePeriod(ctx, store.FinanceClosePeriodParams{Month: m, ClosedBy: &actor})
				if err != nil {
					return httpx.Conflict("already_closed", "That period is already closed.")
				}
				return tx.Audit("finance.period_closed", "period", in.Month, nil)
			})
			return &struct{ Body store.FinanceAccountingPeriod }{out}, err
		})
}

// View loads a journal and its entries.
func (s *Service) View(ctx context.Context, q *store.Queries, id uuid.UUID) (JournalView, error) {
	j, err := q.FinanceGetJournal(ctx, id)
	if err != nil {
		return JournalView{}, httpx.NotFound("Journal")
	}
	es, err := q.FinanceJournalEntries(ctx, id)
	return JournalView{j, es}, err
}

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
