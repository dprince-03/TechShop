// Package payouts pays marketplace sellers what they've earned (docs/payments-finance.md §9).
// A finance officer prepares a batch from every seller whose delivered orders are past the
// return window; a different finance manager approves it with a fresh MFA (two managers above
// the configured threshold). Each payout is a provider transfer with our payout ID as its
// reference, so a retry can never pay twice; the payout journal is posted only once the
// provider confirms the money left. Amounts are capped by the seller's live ledger balance, so
// refunds and chargebacks already booked against the seller are netted off.
package payouts

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/ledger"
	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/modules/payments"
	"github.com/dprince-03/techshop/backend/internal/modules/risk"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/money"
)

// Service is the payouts module.
type Service struct {
	d        *kit.Deps
	ledger   *ledger.Service
	payments *payments.Service
	notify   *notify.Service
	risk     *risk.Service
}

// New creates the payouts service.
func New(d *kit.Deps, led *ledger.Service, pay *payments.Service, n *notify.Service, rk *risk.Service) *Service {
	return &Service{d: d, ledger: led, payments: pay, notify: n, risk: rk}
}

// Name implements kit.Module.
func (s *Service) Name() string { return "payouts" }

// eligibleBefore: fulfilments delivered before this are past the return window.
func (s *Service) eligibleBefore(now time.Time) time.Time {
	return now.Add(-time.Duration(s.d.Cfg.ReturnWindowDays) * 24 * time.Hour)
}

// position is one seller's money right now.
type position struct {
	Eligible []store.PayoutsFulfilmentEarningsRow
	Pending  int64 // earnings not payable yet (undelivered or in the return window)
	Amount   int64 // what a payout would pay now
	Parts    []int64
	Holds    []string
	BankID   *uuid.UUID
}

// position works out what a seller can be paid now and what's stopping it.
func (s *Service) position(ctx context.Context, q *store.Queries, sellerID uuid.UUID, now time.Time) (position, error) {
	var p position
	seller, err := q.SellersGet(ctx, sellerID)
	if err != nil {
		return p, httpx.NotFound("Shop")
	}
	rows, err := q.PayoutsFulfilmentEarnings(ctx, store.PayoutsFulfilmentEarningsParams{SellerID: sellerID, EligibleBefore: s.eligibleBefore(now)})
	if err != nil {
		return p, err
	}
	var eligible []int64
	for _, r := range rows {
		if r.Eligible {
			p.Eligible = append(p.Eligible, r)
			eligible = append(eligible, r.EarningKobo)
		} else {
			p.Pending += r.EarningKobo
		}
	}
	balance, err := q.PayoutsPayableBalance(ctx, &sellerID)
	if err != nil {
		return p, err
	}
	inFlight, err := q.PayoutsInFlight(ctx, sellerID)
	if err != nil {
		return p, err
	}
	p.Amount, p.Parts = split(eligible, balance, inFlight, p.Pending)

	st := sellerState{Status: seller.Status, Available: balance - inFlight - p.Pending}
	if b, err := q.PayoutsDefaultBank(ctx, sellerID); err == nil {
		st.HasBank, st.BankAddedAt, st.BankVerified = true, b.CreatedAt, b.VerifiedAt != nil
		p.BankID = &b.ID
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return p, err
	}
	if st.CredentialMoves, err = q.PayoutsCredentialChanges(ctx, sellerID); err != nil {
		return p, err
	}
	if st.OpenRiskCases, err = q.PayoutsOpenRiskCases(ctx, sellerID); err != nil {
		return p, err
	}
	holds, err := q.PayoutsActiveHolds(ctx, sellerID)
	if err != nil {
		return p, err
	}
	for _, h := range holds {
		st.StoredHolds = append(st.StoredHolds, h.Reason)
	}
	p.Holds = holdReasons(st, now)
	return p, nil
}

func (s *Service) holiday(ctx context.Context) func(time.Time) bool {
	return func(d time.Time) bool {
		h, err := s.d.Q.PayoutsIsHoliday(ctx, d)
		return err == nil && h
	}
}

// SkippedSeller is a seller left out of a batch, and why.
type SkippedSeller struct {
	SellerID uuid.UUID `json:"sellerId"`
	Name     string    `json:"name"`
	Reasons  []string  `json:"reasons" doc:"Hold reasons, or below_minimum"`
	Amount   int64     `json:"amount" doc:"Kobo: what would have been paid"`
}

// Prepare builds a draft batch: one payout per seller who can be paid now, at least the
// minimum. Each seller is locked while their payout is worked out, so two batches being
// prepared at once can't both take the same earnings.
func (s *Service) Prepare(ctx context.Context, actor uuid.UUID) (store.FinancePayoutBatch, []SkippedSeller, error) {
	var batch store.FinancePayoutBatch
	skipped := []SkippedSeller{}
	now := time.Now()
	day := nextPayoutDay(now, s.holiday(ctx))
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		sellers, err := tx.Q.PayoutsCandidateSellers(ctx)
		if err != nil {
			return err
		}
		if batch, err = tx.Q.PayoutsCreateBatch(ctx, actor); err != nil {
			return err
		}
		for _, sl := range sellers {
			if _, err := tx.PgTx.Exec(ctx, "select pg_advisory_xact_lock(hashtext('payout:' || $1))", sl.ID.String()); err != nil {
				return err
			}
			p, err := s.position(ctx, tx.Q, sl.ID, now)
			if err != nil {
				return err
			}
			if len(p.Holds) > 0 || p.Amount < s.d.Cfg.PayoutMinimumKobo || p.BankID == nil {
				reasons := p.Holds
				if len(reasons) == 0 {
					reasons = []string{"below_minimum"}
				}
				if p.Amount > 0 || len(p.Holds) > 0 {
					skipped = append(skipped, SkippedSeller{SellerID: sl.ID, Name: sl.DisplayName, Reasons: reasons, Amount: p.Amount})
				}
				continue
			}
			held, err := s.payoutRisk(ctx, tx, sl.ID, sl.DisplayName, p.Amount)
			if err != nil {
				return err
			}
			if held {
				skipped = append(skipped, SkippedSeller{SellerID: sl.ID, Name: sl.DisplayName, Reasons: []string{"risk_review"}, Amount: p.Amount})
				continue
			}
			po, err := tx.Q.PayoutsCreatePayout(ctx, store.PayoutsCreatePayoutParams{BatchID: &batch.ID, SellerID: sl.ID, BankAccountID: *p.BankID,
				AmountKobo: p.Amount, ScheduledFor: day})
			if err != nil {
				return err
			}
			for i, f := range p.Eligible {
				if p.Parts[i] <= 0 {
					continue
				}
				if err := tx.Q.PayoutsAddItem(ctx, store.PayoutsAddItemParams{PayoutID: po.ID, FulfilmentID: f.ID, AmountKobo: p.Parts[i]}); err != nil {
					return httpx.DB(err, "payout item")
				}
			}
		}
		if batch, err = tx.Q.PayoutsRefreshBatchTotals(ctx, batch.ID); err != nil {
			return err
		}
		return tx.Audit("payout.batch_prepared", "payout_batch", batch.ID.String(), map[string]any{"payouts": batch.PayoutCount, "totalKobo": batch.TotalKobo, "skipped": len(skipped)})
	})
	return batch, skipped, err
}

// payoutRisk runs the payout checkpoint for one seller's payout. A hold opens a seller case,
// and the open case keeps them out of later batches until an analyst clears it.
func (s *Service) payoutRisk(ctx context.Context, tx *uow.Tx, sellerID uuid.UUID, name string, amount int64) (bool, error) {
	f, err := tx.Q.PayoutsRiskFeatures(ctx, sellerID)
	if err != nil {
		return false, err
	}
	vsAvg, cbRate := 0.0, 0.0
	if f.AvgPaid30d > 0 {
		vsAvg = float64(amount) / f.AvgPaid30d
	}
	if f.Orders90d > 0 {
		cbRate = 100 * float64(f.Chargebacks90d) / float64(f.Orders90d)
	}
	dec, err := s.risk.Evaluate(ctx, tx, risk.Check{Checkpoint: "payout", SubjectType: "seller", SubjectID: sellerID, MoneyKobo: amount, CaseKind: "seller",
		CaseTitle: "Payout to " + name, Features: risk.Features{"payout_vs_avg": vsAvg, "chargeback_rate_pct": cbRate, "open_cases": 0}})
	return dec.Held(), err
}

// Exclude takes one seller out of a draft batch; with hold, they stay out of later batches
// until finance releases the hold.
func (s *Service) Exclude(ctx context.Context, batchID, payoutID, actor uuid.UUID, hold bool, note string) (store.FinancePayoutBatch, error) {
	var b store.FinancePayoutBatch
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		if b, err = tx.Q.PayoutsGetBatchForUpdate(ctx, batchID); err != nil {
			return httpx.NotFound("BatchPayout batch")
		}
		if b.Status != "draft" {
			return httpx.Conflict("batch_not_draft", "Only a draft batch can be changed.")
		}
		po, err := tx.Q.PayoutsGetForUpdate(ctx, payoutID)
		if err != nil || po.BatchID == nil || *po.BatchID != batchID {
			return httpx.NotFound("Payout")
		}
		if err := tx.Q.PayoutsDelete(ctx, po.ID); err != nil {
			return err
		}
		if hold {
			var n *string
			if note != "" {
				n = &note
			}
			if _, err := tx.Q.PayoutsCreateHold(ctx, store.PayoutsCreateHoldParams{SellerID: po.SellerID, Reason: "manual", Note: n, CreatedBy: &actor}); err != nil {
				return err
			}
		}
		if b, err = tx.Q.PayoutsRefreshBatchTotals(ctx, batchID); err != nil {
			return err
		}
		return tx.Audit("payout.excluded", "payout_batch", batchID.String(), map[string]any{"seller": po.SellerID, "amountKobo": po.AmountKobo, "hold": hold, "note": note})
	})
	return b, err
}

// Submit sends a draft batch for approval.
func (s *Service) Submit(ctx context.Context, batchID, actor uuid.UUID) (store.FinancePayoutBatch, error) {
	var b store.FinancePayoutBatch
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		cur, err := tx.Q.PayoutsGetBatchForUpdate(ctx, batchID)
		if err != nil {
			return httpx.NotFound("BatchPayout batch")
		}
		if cur.PayoutCount == 0 {
			return httpx.Conflict("batch_empty", "This batch has no payouts to submit.")
		}
		if b, err = tx.Q.PayoutsSubmitBatch(ctx, batchID); err != nil {
			return httpx.Conflict("batch_not_draft", "Only a draft batch can be submitted.")
		}
		return tx.Audit("payout.batch_submitted", "payout_batch", batchID.String(), map[string]any{"by": actor, "totalKobo": b.TotalKobo})
	})
	return b, err
}

// Approve is the second pair of eyes (never the preparer). Above the threshold two different
// managers approve; the first is recorded and the batch waits for the second. Once approved
// the transfers start in the background.
func (s *Service) Approve(ctx context.Context, batchID, approver uuid.UUID) (store.FinancePayoutBatch, bool, error) {
	var b store.FinancePayoutBatch
	final := false
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		if b, err = tx.Q.PayoutsGetBatchForUpdate(ctx, batchID); err != nil {
			return httpx.NotFound("BatchPayout batch")
		}
		if b.Status != "submitted" {
			return httpx.Conflict("batch_not_submitted", "This batch isn't waiting for approval.")
		}
		if b.PreparedBy == approver {
			return httpx.E(403, "separation_of_duties", "You prepared this batch, so someone else must approve it.")
		}
		if b.SecondApprovedBy != nil && *b.SecondApprovedBy == approver {
			return httpx.E(403, "separation_of_duties", "You've already approved this batch; a second manager must approve it too.")
		}
		if b.TotalKobo > s.d.Cfg.PayoutApprovalThreshold && b.SecondApprovedBy == nil {
			if b, err = tx.Q.PayoutsSecondApprove(ctx, store.PayoutsSecondApproveParams{ID: batchID, Approver: &approver}); err != nil {
				return err
			}
			return tx.Audit("payout.batch_first_approval", "payout_batch", batchID.String(), map[string]any{"totalKobo": b.TotalKobo,
				"threshold": money.Format(s.d.Cfg.PayoutApprovalThreshold)})
		}
		if b, err = tx.Q.PayoutsApproveBatch(ctx, store.PayoutsApproveBatchParams{ID: batchID, Approver: &approver}); err != nil {
			return httpx.DB(err, "payout batch")
		}
		final = true
		tx.Enqueue(processBatchArgs{BatchID: batchID}, finance())
		return tx.Audit("payout.batch_approved", "payout_batch", batchID.String(), map[string]any{"totalKobo": b.TotalKobo, "payouts": b.PayoutCount})
	})
	return b, final, err
}

// Cancel drops a batch that hasn't been approved; its earnings go back to the next run.
func (s *Service) Cancel(ctx context.Context, batchID, actor uuid.UUID) (store.FinancePayoutBatch, error) {
	var b store.FinancePayoutBatch
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		if b, err = tx.Q.PayoutsGetBatchForUpdate(ctx, batchID); err != nil {
			return httpx.NotFound("BatchPayout batch")
		}
		if b.Status != "draft" && b.Status != "submitted" {
			return httpx.Conflict("batch_started", "An approved batch can't be cancelled; its transfers have started.")
		}
		pos, err := tx.Q.PayoutsBatchPayouts(ctx, &batchID)
		if err != nil {
			return err
		}
		for _, p := range pos {
			if err := tx.Q.PayoutsDelete(ctx, p.ID); err != nil {
				return err
			}
		}
		if _, err = tx.Q.PayoutsRefreshBatchTotals(ctx, batchID); err != nil {
			return err
		}
		if b, err = tx.Q.PayoutsSetBatchStatus(ctx, store.PayoutsSetBatchStatusParams{ID: batchID, Status: "cancelled"}); err != nil {
			return err
		}
		return tx.Audit("payout.batch_cancelled", "payout_batch", batchID.String(), map[string]any{"by": actor})
	})
	return b, err
}

// ProcessBatch starts every transfer in an approved batch. Each payout is handled on its own,
// so one bank's failure doesn't hold up the rest.
func (s *Service) ProcessBatch(ctx context.Context, batchID uuid.UUID) error {
	b, err := s.d.Q.PayoutsGetBatch(ctx, batchID)
	if err != nil {
		return err
	}
	if b.Status == "approved" {
		if _, err := s.d.Q.PayoutsSetBatchStatus(ctx, store.PayoutsSetBatchStatusParams{ID: batchID, Status: "processing"}); err != nil {
			return err
		}
	} else if b.Status != "processing" {
		return nil
	}
	pos, err := s.d.Q.PayoutsBatchPayouts(ctx, &batchID)
	if err != nil {
		return err
	}
	var firstErr error
	for _, p := range pos {
		if p.Status != "scheduled" {
			continue
		}
		if err := s.pay(ctx, p.ID); err != nil {
			s.d.Logger.Error("payout transfer failed to start", "payout", p.ID, "error", err)
			if firstErr == nil {
				firstErr = err // River retries the batch; started payouts are skipped
			}
		}
	}
	if err := s.closeBatchIfDone(ctx, batchID); err != nil {
		return err
	}
	return firstErr
}

// pay starts one transfer. The payout is marked processing with its reference before the
// provider is called, so a crash in between is picked up by the sweep, which re-checks the
// reference (and re-sends it, safely, if the provider never saw it).
func (s *Service) pay(ctx context.Context, payoutID uuid.UUID) error {
	po, err := s.d.Q.PayoutsGet(ctx, payoutID)
	if err != nil {
		return err
	}
	recipient, err := s.recipient(ctx, po)
	if err != nil {
		return err
	}
	ref := po.ID.String()
	if _, err := s.d.Q.PayoutsSetProcessing(ctx, store.PayoutsSetProcessingParams{ID: po.ID, Reference: &ref}); err != nil {
		return err
	}
	res, err := s.payments.Transfers().Initiate(ctx, payments.TransferRequest{Reference: ref, AmountKobo: po.AmountKobo, Recipient: recipient,
		Reason: "TechShop payout " + ref[:8]})
	if err != nil {
		return err // stays processing; the sweep verifies it
	}
	return s.apply(ctx, po.ID, res)
}

// recipient returns the provider's recipient code for the payout's bank account, creating it
// on first use (the account number is decrypted only for that call).
func (s *Service) recipient(ctx context.Context, po store.FinancePayout) (string, error) {
	bank, err := s.d.Q.PayoutsBankAccount(ctx, po.BankAccountID)
	if err != nil {
		return "", err
	}
	if bank.ProviderRecipientCode != nil && *bank.ProviderRecipientCode != "" {
		return *bank.ProviderRecipientCode, nil
	}
	number, err := s.d.Sealer.Open(bank.AccountNumberEncrypted, "sellers.seller_bank_accounts:"+po.SellerID.String())
	if err != nil {
		return "", err
	}
	name := bank.AccountName
	if bank.ResolvedAccountName != nil && *bank.ResolvedAccountName != "" {
		name = *bank.ResolvedAccountName
	}
	code, err := s.payments.Transfers().CreateRecipient(ctx, name, string(number), bank.BankCode)
	if err != nil {
		return "", err
	}
	return code, s.d.Q.PayoutsSetRecipientCode(ctx, store.PayoutsSetRecipientCodeParams{ID: bank.ID, ProviderRecipientCode: &code})
}

// apply records the provider's answer. Success posts the payout journal and tells the seller;
// failure gives the fulfilments back for the next run; a reversal after success undoes the
// journal. Repeating an answer changes nothing.
func (s *Service) apply(ctx context.Context, payoutID uuid.UUID, res payments.TransferResult) error {
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		po, err := tx.Q.PayoutsGetForUpdate(ctx, payoutID)
		if err != nil {
			return err
		}
		var code *string
		if res.Code != "" {
			code = &res.Code
		}
		switch {
		case res.Status == "success" && po.Status == "processing":
			if _, err := tx.Q.PayoutsMarkPaid(ctx, store.PayoutsMarkPaidParams{ID: po.ID, TransferCode: code}); err != nil {
				return err
			}
			if err := s.ledger.PayoutPaid(ctx, tx, po.ID, po.SellerID, po.AmountKobo); err != nil {
				return err
			}
			tx.Emit("payout", po.ID, "payout.paid", map[string]any{"payoutId": po.ID, "sellerId": po.SellerID, "amountKobo": po.AmountKobo})
			return s.tellSeller(ctx, tx, po, "payout_sent", "")
		case (res.Status == "failed" && (po.Status == "processing" || po.Status == "scheduled")) || (res.Status == "reversed" && po.Status != "failed"):
			reason := res.Reason
			if reason == "" {
				reason = "The bank couldn't complete the transfer."
			}
			if po.Status == "paid" {
				if err := s.ledger.PayoutReversed(ctx, tx, po.ID); err != nil {
					return err
				}
			}
			if _, err := tx.Q.PayoutsMarkFailed(ctx, store.PayoutsMarkFailedParams{ID: po.ID, Reason: &reason}); err != nil {
				return err
			}
			if err := tx.Q.PayoutsReleaseItems(ctx, po.ID); err != nil {
				return err
			}
			if err := tx.Audit("payout.failed", "payout", po.ID.String(), map[string]any{"reason": reason, "status": res.Status}); err != nil {
				return err
			}
			return s.tellSeller(ctx, tx, po, "payout_failed", reason)
		}
		return nil // pending, or an answer already applied
	})
	if err != nil {
		return err
	}
	po, err := s.d.Q.PayoutsGet(ctx, payoutID)
	if err != nil || po.BatchID == nil {
		return err
	}
	return s.closeBatchIfDone(ctx, *po.BatchID)
}

func (s *Service) closeBatchIfDone(ctx context.Context, batchID uuid.UUID) error {
	open, err := s.d.Q.PayoutsBatchOpenCount(ctx, &batchID)
	if err != nil || open > 0 {
		return err
	}
	b, err := s.d.Q.PayoutsGetBatch(ctx, batchID)
	if err != nil || b.Status != "processing" {
		return err
	}
	_, err = s.d.Q.PayoutsSetBatchStatus(ctx, store.PayoutsSetBatchStatusParams{ID: batchID, Status: "done"})
	return err
}

// tellSeller emails the shop owner about a payout.
func (s *Service) tellSeller(ctx context.Context, tx *uow.Tx, po store.FinancePayout, template, reason string) error {
	seller, err := tx.Q.SellersGet(ctx, po.SellerID)
	if err != nil || seller.OwnerUserID == nil {
		return err
	}
	owner, err := tx.Q.IdentityGetUser(ctx, *seller.OwnerUserID)
	if err != nil || owner.Email == nil {
		return err
	}
	bank, err := tx.Q.PayoutsBankAccount(ctx, po.BankAccountID)
	if err != nil {
		return err
	}
	return s.notify.Send(ctx, tx, notify.Msg{UserID: &owner.ID, To: *owner.Email, Channel: "email", Category: notify.CatSeller, Template: template,
		Data: map[string]any{"amount": money.Format(po.AmountKobo), "last4": bank.AccountNumberLast4, "reason": reason}, Key: "payout:" + po.ID.String() + ":" + template})
}

// Recheck asks the provider about a processing payout (webhook prompt or the sweep). A
// reference the provider has never seen is sent again — the same reference can't pay twice.
func (s *Service) Recheck(ctx context.Context, payoutID uuid.UUID) error {
	po, err := s.d.Q.PayoutsGet(ctx, payoutID)
	if err != nil {
		return err
	}
	if po.Status != "processing" && po.Status != "paid" {
		return nil
	}
	res, err := s.payments.Transfers().VerifyTransfer(ctx, po.ID.String())
	if err != nil {
		if po.Status != "processing" {
			return err
		}
		recipient, rerr := s.recipient(ctx, po)
		if rerr != nil {
			return rerr
		}
		if res, err = s.payments.Transfers().Initiate(ctx, payments.TransferRequest{Reference: po.ID.String(), AmountKobo: po.AmountKobo,
			Recipient: recipient, Reason: "TechShop payout " + po.ID.String()[:8]}); err != nil {
			return err
		}
	}
	return s.apply(ctx, po.ID, res)
}

// Sweep re-checks payouts stuck in processing (missed webhooks, crashes mid-transfer).
func (s *Service) Sweep(ctx context.Context) error {
	stale, err := s.d.Q.PayoutsStaleProcessing(ctx)
	if err != nil {
		return err
	}
	for _, p := range stale {
		if err := s.Recheck(ctx, p.ID); err != nil {
			s.d.Logger.Warn("payout recheck failed", "payout", p.ID, "error", err)
		}
	}
	return nil
}

// Hold stops a seller's payouts until it's released (or until the given time).
func (s *Service) Hold(ctx context.Context, sellerID uuid.UUID, reason, note string, until *time.Time, actor uuid.UUID) (store.FinancePayoutHold, error) {
	var h store.FinancePayoutHold
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if _, err := tx.Q.SellersGet(ctx, sellerID); err != nil {
			return httpx.NotFound("Shop")
		}
		var err error
		if h, err = tx.Q.PayoutsCreateHold(ctx, store.PayoutsCreateHoldParams{SellerID: sellerID, Reason: reason, Note: &note, Until: until, CreatedBy: &actor}); err != nil {
			return err
		}
		return tx.Audit("payout.hold_placed", "seller", sellerID.String(), map[string]any{"hold": h.ID, "reason": reason, "note": note})
	})
	return h, err
}

// Release lifts a hold.
func (s *Service) Release(ctx context.Context, holdID, actor uuid.UUID) (store.FinancePayoutHold, error) {
	var h store.FinancePayoutHold
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var err error
		if h, err = tx.Q.PayoutsReleaseHold(ctx, holdID); err != nil {
			return httpx.Conflict("hold_not_active", "This hold doesn't exist or was already released.")
		}
		return tx.Audit("payout.hold_released", "seller", h.SellerID.String(), map[string]any{"hold": h.ID, "by": actor})
	})
	return h, err
}
