package ledger

import (
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
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
