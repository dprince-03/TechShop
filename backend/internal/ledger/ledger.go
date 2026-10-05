// Package ledger is the double-entry ledger (docs/payments-finance.md §4). Every posting goes
// through a typed template function; journals are idempotent by posting_key, balanced (the
// database also checks this at commit), never edited and only reversed.
package ledger

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/money"
)

// Fixed account codes (the chart in docs/payments-finance.md §4.1).
const (
	RevenueFirstParty = "revenue:first_party_sales"
	RevenueDelivery   = "revenue:delivery"
	RevenueCommission = "revenue:commission"
	LiabilityVAT      = "liability:vat"
	LiabilityRefunds  = "liability:customer_refunds"
	BankOperating     = "bank:operating"
	BankPayouts       = "bank:payouts"
	ExpenseFees       = "expense:payment_fees"
	ExpenseChargeback = "expense:chargebacks"
	ExpenseWriteOffs  = "expense:write_offs"
	// ExpensePromotions carries TechShop-funded discounts on marketplace items, so sellers are
	// still paid their full price (an addition to the §4.1 chart; logged in docs/plan.md).
	ExpensePromotions = "expense:promotions"
)

// Clearing is the provider clearing account for a payment provider.
func Clearing(provider string) string { return "clearing:" + provider }

// SellerPayable is what TechShop owes one marketplace seller.
func SellerPayable(sellerID uuid.UUID) string { return "seller_payable:" + sellerID.String() }

// StoreCredit is one customer's store credit liability.
func StoreCredit(userID uuid.UUID) string { return "liability:store_credit:" + userID.String() }

// Receivable is one business's unpaid invoices.
func Receivable(businessID uuid.UUID) string { return "receivable:b2b:" + businessID.String() }

// CashStore is the cash in one store till.
func CashStore(code string) string { return "cash:store:" + strings.ToLower(code) }

// Line is one ledger entry: positive = debit, negative = credit.
type Line struct {
	Code   string
	Amount int64
}

// Journal is one balanced accounting event.
type Journal struct {
	Key      string // posting_key: a retry returns the existing journal
	Kind     string
	RefType  string
	RefID    uuid.UUID
	Memo     string
	PostedBy *uuid.UUID
	Reverses *uuid.UUID
	Lines    []Line
}

// Service posts journals and serves the finance screens.
type Service struct{ d *kit.Deps }

// New creates the ledger service.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "ledger" }

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// ErrUnbalanced means a template produced lines that don't sum to zero (a programming error).
var ErrUnbalanced = errors.New("ledger: journal does not balance")

// Post writes a journal and its entries. Same posting_key twice = the first journal, unchanged.
func (s *Service) Post(ctx context.Context, tx *uow.Tx, j Journal) (store.FinanceLedgerJournal, bool, error) {
	lines := merge(j.Lines)
	var sum int64
	for _, l := range lines {
		sum += l.Amount
	}
	if sum != 0 || len(lines) < 2 {
		return store.FinanceLedgerJournal{}, false, fmt.Errorf("%w: %s sums to %d over %d lines", ErrUnbalanced, j.Key, sum, len(lines))
	}
	var memo *string
	if j.Memo != "" {
		memo = &j.Memo
	}
	jr, err := tx.Q.FinanceInsertJournal(ctx, store.FinanceInsertJournalParams{PostingKey: j.Key, Kind: j.Kind, ReferenceType: j.RefType, ReferenceID: j.RefID,
		ReversesJournalID: j.Reverses, Memo: memo, PostedBy: j.PostedBy})
	if errors.Is(err, pgx.ErrNoRows) {
		existing, gerr := tx.Q.FinanceGetJournalByKey(ctx, j.Key)
		return existing, false, gerr
	}
	if err != nil {
		if strings.Contains(err.Error(), "accounting period") {
			return jr, false, httpx.Conflict("period_closed", "That accounting period is closed; post the correction in the current period.")
		}
		return jr, false, err
	}
	for _, l := range lines {
		acct, err := s.ensure(ctx, tx, l.Code)
		if err != nil {
			return jr, false, err
		}
		if err := tx.Q.FinanceInsertEntry(ctx, store.FinanceInsertEntryParams{JournalID: jr.ID, AccountID: acct.ID, AmountKobo: l.Amount}); err != nil {
			return jr, false, err
		}
	}
	tx.Emit("journal", jr.ID, "ledger.posted", map[string]any{"journalId": jr.ID, "kind": j.Kind, "key": j.Key})
	return jr, true, nil
}

// merge combines lines on the same account and drops zero lines (entries can't be zero).
func merge(in []Line) []Line {
	byCode := map[string]int64{}
	order := []string{}
	for _, l := range in {
		if _, ok := byCode[l.Code]; !ok {
			order = append(order, l.Code)
		}
		byCode[l.Code] += l.Amount
	}
	out := make([]Line, 0, len(order))
	for _, c := range order {
		if byCode[c] != 0 {
			out = append(out, Line{c, byCode[c]})
		}
	}
	return out
}

// ensure creates an account on first use; its kind and owner come from the code.
func (s *Service) ensure(ctx context.Context, tx *uow.Tx, code string) (store.FinanceLedgerAccount, error) {
	var kind string
	name := code
	var sellerID, userID *uuid.UUID
	parts := strings.Split(code, ":")
	switch parts[0] {
	case "clearing", "bank", "receivable", "cash":
		kind = "asset"
	case "revenue":
		kind = "revenue"
	case "expense":
		kind = "expense"
	case "equity":
		kind = "equity"
	case "seller_payable":
		kind = "liability"
		if id, err := uuid.Parse(parts[len(parts)-1]); err == nil {
			sellerID = &id
			name = "Payable to seller " + id.String()[:8]
		}
	case "liability":
		kind = "liability"
		if len(parts) == 3 && parts[1] == "store_credit" {
			if id, err := uuid.Parse(parts[2]); err == nil {
				userID = &id
				name = "Store credit " + id.String()[:8]
			}
		}
	default:
		return store.FinanceLedgerAccount{}, fmt.Errorf("ledger: unknown account prefix in %q", code)
	}
	return tx.Q.FinanceEnsureAccount(ctx, store.FinanceEnsureAccountParams{Code: code, Name: name, Kind: kind, SellerID: sellerID, UserID: userID})
}

// ---- Templates (docs/database/flows.md "Journal templates"). VAT treatment is a PLACEHOLDER
// until finance confirms it (VAT_BPS, owner decision #2).

// SaleLine is one paid order line, as the ledger needs it.
type SaleLine struct {
	SellerID       uuid.UUID
	FirstParty     bool
	Gross          int64 // unit price × quantity
	Discount       int64 // total discount on the line
	TechShopFunded int64 // the part of Discount TechShop pays (the seller is still paid for it)
	VAT            int64 // VAT inside the line (first-party only)
}

// Sale posts "payment captured": clearing ← paid; revenue/VAT for TechShop stock, seller
// payables for marketplace items, delivery revenue, TechShop-funded discounts as an expense,
// and any overpayment as a refund liability.
func (s *Service) Sale(ctx context.Context, tx *uow.Tx, paymentID, orderID uuid.UUID, provider string, paid, deliveryFee, overpaid int64, lines []SaleLine) error {
	ls := []Line{{Clearing(provider), paid}}
	for _, l := range lines {
		net := l.Gross - l.Discount
		switch {
		case l.FirstParty:
			ls = append(ls, Line{RevenueFirstParty, -(net - l.VAT)}, Line{LiabilityVAT, -l.VAT})
		default:
			ls = append(ls, Line{SellerPayable(l.SellerID), -(net + l.TechShopFunded)}, Line{ExpensePromotions, l.TechShopFunded})
		}
	}
	ls = append(ls, Line{RevenueDelivery, -deliveryFee})
	if overpaid > 0 {
		ls = append(ls, Line{LiabilityRefunds, -overpaid})
	}
	_, _, err := s.Post(ctx, tx, Journal{Key: "payment:" + paymentID.String() + ":captured", Kind: "sale", RefType: "order", RefID: orderID, Lines: ls})
	return err
}

// Commission posts the marketplace commission when a fulfilment is delivered.
func (s *Service) Commission(ctx context.Context, tx *uow.Tx, fulfilmentID, sellerID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return nil
	}
	_, _, err := s.Post(ctx, tx, Journal{Key: "fulfilment:" + fulfilmentID.String() + ":commission", Kind: "commission", RefType: "fulfilment", RefID: fulfilmentID,
		Lines: []Line{{SellerPayable(sellerID), amount}, {RevenueCommission, -amount}}})
	return err
}

// RefundApprove moves an approved refund into liability:customer_refunds, reversing the
// order's sale credits pro rata (revenue, VAT, seller payables, delivery) so the right party
// bears it. TechShop-funded promotion expense is reversed in the same proportion.
func (s *Service) RefundApprove(ctx context.Context, tx *uow.Tx, refundID, orderID, paymentID uuid.UUID, amount int64, actor *uuid.UUID) error {
	sale, err := tx.Q.FinanceGetJournalByKey(ctx, "payment:"+paymentID.String()+":captured")
	if err != nil {
		return fmt.Errorf("ledger: sale journal for payment %s: %w", paymentID, err)
	}
	entries, err := tx.Q.FinanceJournalEntries(ctx, sale.ID)
	if err != nil {
		return err
	}
	var paid int64
	var credits []store.FinanceJournalEntriesRow
	var promo int64
	for _, e := range entries {
		switch {
		case strings.HasPrefix(e.AccountCode, "clearing:"):
			paid += e.AmountKobo
		case e.AccountCode == ExpensePromotions:
			promo += e.AmountKobo
		case e.AmountKobo < 0 && e.AccountCode != LiabilityRefunds:
			credits = append(credits, e)
		}
	}
	if paid <= 0 {
		return fmt.Errorf("ledger: sale journal %s has no clearing debit", sale.ID)
	}
	// Reverse the credits for amount × (credits+promo)/paid: the promo share is added back.
	promoBack := amount * promo / paid
	weights := make([]int64, len(credits))
	for i, c := range credits {
		weights[i] = -c.AmountKobo
	}
	parts := money.Allocate(amount+promoBack, weights)
	ls := []Line{{LiabilityRefunds, -amount}}
	if promoBack > 0 {
		ls = append(ls, Line{ExpensePromotions, -promoBack})
	}
	for i, c := range credits {
		ls = append(ls, Line{c.AccountCode, parts[i]})
	}
	_, _, err = s.Post(ctx, tx, Journal{Key: "refund:" + refundID.String() + ":approved", Kind: "refund", RefType: "refund", RefID: refundID, PostedBy: actor,
		Memo: "Refund approved for order " + orderID.String(), Lines: ls})
	return err
}

// RefundPaid clears the refund liability when the provider confirms the money went back.
func (s *Service) RefundPaid(ctx context.Context, tx *uow.Tx, refundID uuid.UUID, provider string, amount int64, toStoreCredit *uuid.UUID) error {
	target := Clearing(provider)
	if toStoreCredit != nil {
		target = StoreCredit(*toStoreCredit)
	}
	_, _, err := s.Post(ctx, tx, Journal{Key: "refund:" + refundID.String() + ":paid", Kind: "refund", RefType: "refund", RefID: refundID,
		Lines: []Line{{LiabilityRefunds, amount}, {target, -amount}}})
	return err
}

// Reverse posts the negation of a journal (journals are never edited).
func (s *Service) Reverse(ctx context.Context, tx *uow.Tx, journalID uuid.UUID, memo string, actor *uuid.UUID) (store.FinanceLedgerJournal, error) {
	orig, err := tx.Q.FinanceGetJournal(ctx, journalID)
	if err != nil {
		return orig, httpx.NotFound("Journal")
	}
	if orig.Kind == "reversal" {
		return orig, httpx.Invalid("cannot_reverse_reversal", "A reversal can't itself be reversed; post a corrected journal instead.")
	}
	entries, err := tx.Q.FinanceJournalEntries(ctx, journalID)
	if err != nil {
		return orig, err
	}
	ls := make([]Line, len(entries))
	for i, e := range entries {
		ls[i] = Line{e.AccountCode, -e.AmountKobo}
	}
	jr, created, err := s.Post(ctx, tx, Journal{Key: "reversal:" + journalID.String(), Kind: "reversal", RefType: orig.ReferenceType, RefID: orig.ReferenceID,
		Reverses: &journalID, Memo: memo, PostedBy: actor, Lines: ls})
	if err != nil {
		return jr, httpx.DB(err, "reversal")
	}
	if !created {
		return jr, httpx.Conflict("already_reversed", "This journal was already reversed.")
	}
	return jr, nil
}

// Adjust posts a free-form adjustment (ledger.adjust ★ — memo and step-up required).
func (s *Service) Adjust(ctx context.Context, tx *uow.Tx, key, memo string, lines []Line, actor uuid.UUID) (store.FinanceLedgerJournal, error) {
	for _, l := range lines {
		if !validCode(l.Code) {
			return store.FinanceLedgerJournal{}, httpx.Invalid("bad_account", "Unknown account code "+l.Code+".")
		}
	}
	id := uuid.New()
	jr, created, err := s.Post(ctx, tx, Journal{Key: "adjustment:" + key, Kind: "adjustment", RefType: "adjustment", RefID: id, Memo: memo, PostedBy: &actor, Lines: lines})
	if errors.Is(err, ErrUnbalanced) {
		return jr, httpx.Invalid("unbalanced", "Debits and credits must sum to zero.")
	}
	if err == nil && !created {
		return jr, httpx.Conflict("already_posted", "An adjustment with this key was already posted.")
	}
	return jr, err
}

func validCode(code string) bool {
	known := []string{"clearing:", "bank:", "receivable:", "cash:", "revenue:", "expense:", "equity:", "seller_payable:", "liability:"}
	for _, k := range known {
		if strings.HasPrefix(code, k) && len(code) > len(k) {
			return true
		}
	}
	return false
}

// Jobs implements kit.Module: the nightly integrity checks (trial balance, sale journals).
func (s *Service) Jobs(reg *jobs.Registry) { s.registerJobs(reg) }
