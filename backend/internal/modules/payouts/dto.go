package payouts

import (
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
)

// PayoutBatch is a payout run as finance sees it.
type PayoutBatch struct {
	ID                uuid.UUID  `json:"id"`
	Status            string     `json:"status" enum:"draft,submitted,approved,processing,done,cancelled"`
	PreparedBy        uuid.UUID  `json:"preparedBy"`
	SubmittedAt       *time.Time `json:"submittedAt"`
	FirstApprovedBy   *uuid.UUID `json:"firstApprovedBy" doc:"Above the threshold: the first of two approvers"`
	ApprovedBy        *uuid.UUID `json:"approvedBy"`
	ApprovedAt        *time.Time `json:"approvedAt"`
	Total             int64      `json:"total" doc:"Kobo"`
	Count             int32      `json:"count"`
	NeedsTwoApprovers bool       `json:"needsTwoApprovers"`
	CreatedAt         time.Time  `json:"createdAt"`
}

func batchView(b store.FinancePayoutBatch, threshold int64) PayoutBatch {
	return PayoutBatch{ID: b.ID, Status: b.Status, PreparedBy: b.PreparedBy, SubmittedAt: b.SubmittedAt, FirstApprovedBy: b.SecondApprovedBy, ApprovedBy: b.ApprovedBy,
		ApprovedAt: b.ApprovedAt, Total: b.TotalKobo, Count: b.PayoutCount, NeedsTwoApprovers: b.TotalKobo > threshold, CreatedAt: b.CreatedAt}
}

// BatchPayout is one seller's transfer in a batch. Only the last 4 digits of the account show.
type BatchPayout struct {
	ID            uuid.UUID  `json:"id"`
	SellerID      uuid.UUID  `json:"sellerId"`
	Shop          string     `json:"shop"`
	Amount        int64      `json:"amount" doc:"Kobo"`
	Status        string     `json:"status" enum:"scheduled,processing,paid,failed,on_hold"`
	ScheduledFor  time.Time  `json:"scheduledFor" format:"date"`
	BankCode      string     `json:"bankCode"`
	AccountName   string     `json:"accountName"`
	AccountLast4  string     `json:"accountLast4"`
	PaidAt        *time.Time `json:"paidAt"`
	FailureReason *string    `json:"failureReason"`
}

func payoutView(p store.PayoutsBatchPayoutsRow) BatchPayout {
	return BatchPayout{ID: p.ID, SellerID: p.SellerID, Shop: p.DisplayName, Amount: p.AmountKobo, Status: p.Status, ScheduledFor: p.ScheduledFor,
		BankCode: p.BankCode, AccountName: p.AccountName, AccountLast4: p.AccountNumberLast4, PaidAt: p.PaidAt, FailureReason: p.FailureReason}
}

// PayoutBatchDetail is a batch with its payouts (and, right after preparing, who was left out).
type PayoutBatchDetail struct {
	PayoutBatch PayoutBatch     `json:"batch"`
	Payouts     []BatchPayout   `json:"payouts"`
	Skipped     []SkippedSeller `json:"skipped,omitempty" doc:"Only when the batch was just prepared"`
}

// SellerBalance is a seller's money as Seller Centre shows it.
type SellerBalance struct {
	Available     int64     `json:"available" doc:"Kobo: payable in the next payout run"`
	Pending       int64     `json:"pending" doc:"Kobo: earned, waiting for delivery or the return window"`
	OnHold        bool      `json:"onHold"`
	HoldReasons   []string  `json:"holdReasons"`
	Minimum       int64     `json:"minimum" doc:"Kobo: the smallest payout we send"`
	NextPayoutDay time.Time `json:"nextPayoutDay" format:"date" doc:"Payouts run on Tuesdays and Fridays, business days only"`
}

// SellerPayout is one payout in a seller's history.
type SellerPayout struct {
	ID            uuid.UUID    `json:"id"`
	Amount        int64        `json:"amount" doc:"Kobo"`
	Status        string       `json:"status" enum:"scheduled,processing,paid,failed,on_hold"`
	ScheduledFor  time.Time    `json:"scheduledFor" format:"date"`
	PaidAt        *time.Time   `json:"paidAt"`
	AccountLast4  string       `json:"accountLast4"`
	FailureReason *string      `json:"failureReason"`
	CreatedAt     time.Time    `json:"createdAt"`
	Items         []PayoutItem `json:"items,omitempty"`
}

// PayoutItem is one delivered order a payout covers.
type PayoutItem struct {
	OrderNumber string     `json:"orderNumber"`
	Amount      int64      `json:"amount" doc:"Kobo: after commission"`
	DeliveredAt *time.Time `json:"deliveredAt"`
}
