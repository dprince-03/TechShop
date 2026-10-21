package payouts

import (
	"time"

	"github.com/dprince-03/techshop/backend/pkg/money"
)

// Payout rules (docs/payments-finance.md §9). Pure, so they're unit-tested without a database.

// wat is Nigeria's time zone (UTC+1, no daylight saving).
var wat = time.FixedZone("WAT", 60*60)

// bankChangeHold is how long a new payout account waits before it can be paid (account
// takeover protection, identity-access.md §6).
const bankChangeHold = 24 * time.Hour

// nextPayoutDay is the first Tuesday or Friday on or after now (WAT) that isn't a public
// holiday: the date a payout prepared now is scheduled for.
func nextPayoutDay(now time.Time, holiday func(time.Time) bool) time.Time {
	d := now.In(wat)
	day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, wat)
	for i := 0; i < 30; i++ {
		if wd := day.Weekday(); (wd == time.Tuesday || wd == time.Friday) && !holiday(day) {
			return day
		}
		day = day.AddDate(0, 0, 1)
	}
	return day
}

// Hold reasons beyond the stored finance.payout_holds rows, worked out live.
const (
	HoldKYC         = "kyc_incomplete"
	HoldEnforcement = "enforcement"
	HoldBankChanged = "bank_account_changed"
	HoldCredentials = "credentials_changed"
	HoldRiskCase    = "risk_case"
	HoldNegative    = "negative_balance"
	HoldNoBank      = "no_bank_account"
)

// sellerState is everything that decides whether a seller can be paid now.
type sellerState struct {
	Status          string // sellers.sellers.status
	HasBank         bool
	BankAddedAt     time.Time
	BankVerified    bool
	CredentialMoves int32 // password/phone/email/MFA changes in the last 24 hours
	OpenRiskCases   int32
	StoredHolds     []string
	Available       int64
}

// holdReasons lists why a seller can't be paid now (empty = no hold).
func holdReasons(st sellerState, now time.Time) []string {
	var out []string
	switch st.Status {
	case "active":
	case "suspended", "closed":
		out = append(out, HoldEnforcement)
	default:
		out = append(out, HoldKYC)
	}
	switch {
	case !st.HasBank || !st.BankVerified:
		out = append(out, HoldNoBank)
	case now.Sub(st.BankAddedAt) < bankChangeHold:
		out = append(out, HoldBankChanged)
	}
	if st.CredentialMoves > 0 {
		out = append(out, HoldCredentials)
	}
	if st.OpenRiskCases > 0 {
		out = append(out, HoldRiskCase)
	}
	if st.Available < 0 {
		out = append(out, HoldNegative)
	}
	return append(out, st.StoredHolds...)
}

// split is the amount a seller can be paid and how it spreads over the eligible fulfilments.
// What's owed per the ledger (balance) less money already in other payouts (inFlight) less
// earnings that aren't payable yet (pending) caps the payout, so refunds and chargebacks
// already booked against the seller are netted off.
func split(eligible []int64, balance, inFlight, pending int64) (amount int64, parts []int64) {
	var sum int64
	for _, e := range eligible {
		if e > 0 {
			sum += e
		}
	}
	available := balance - inFlight - pending
	amount = min(sum, available)
	if amount <= 0 {
		return 0, make([]int64, len(eligible))
	}
	weights := make([]int64, len(eligible))
	for i, e := range eligible {
		weights[i] = max(e, 0)
	}
	return amount, money.Allocate(amount, weights)
}
