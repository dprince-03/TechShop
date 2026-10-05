-- Finance: chart of accounts, double-entry journals, periods (docs/payments-finance.md §4).

-- name: FinanceEnsureAccount :one
-- Get-or-create an account by code (seller payables and store credit accounts are created on first use).
insert into finance.ledger_accounts (code, name, kind, seller_id, user_id) values ($1, $2, $3, $4, $5)
on conflict (code) do update set code = excluded.code
returning *;

-- name: FinanceGetAccountByCode :one
select * from finance.ledger_accounts where code = $1;

-- name: FinanceGetJournalByKey :one
select * from finance.ledger_journals where posting_key = $1;

-- name: FinanceInsertJournal :one
-- Idempotent by posting_key: returns no row when the key was already posted.
insert into finance.ledger_journals (posting_key, kind, reference_type, reference_id, reverses_journal_id, memo, posted_by)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (posting_key) do nothing
returning *;

-- name: FinanceInsertEntry :exec
insert into finance.ledger_entries (journal_id, account_id, amount_kobo) values ($1, $2, $3);

-- name: FinanceJournalEntries :many
select e.*, a.code as account_code, a.name as account_name, a.kind as account_kind
from finance.ledger_entries e join finance.ledger_accounts a on a.id = e.account_id
where e.journal_id = $1 order by e.amount_kobo desc;

-- name: FinanceJournalsForReference :many
select * from finance.ledger_journals where reference_type = $1 and reference_id = $2 order by posted_at;

-- name: FinanceGetJournal :one
select * from finance.ledger_journals where id = $1;

-- name: FinanceListJournals :many
select * from finance.ledger_journals
where (sqlc.narg('kind')::text is null or kind = sqlc.narg('kind'))
  and (sqlc.narg('before')::timestamptz is null or posted_at < sqlc.narg('before'))
order by posted_at desc limit $1;

-- name: FinanceAccountBalances :many
select a.id, a.code, a.name, a.kind, coalesce(sum(e.amount_kobo), 0)::bigint as balance_kobo
from finance.ledger_accounts a left join finance.ledger_entries e on e.account_id = a.id
group by a.id order by a.code;

-- name: FinanceAccountBalance :one
select coalesce(sum(e.amount_kobo), 0)::bigint from finance.ledger_entries e join finance.ledger_accounts a on a.id = e.account_id where a.code = $1;

-- name: FinanceAccountLedger :many
select e.id, e.amount_kobo, e.created_at, j.id as journal_id, j.kind, j.posting_key, j.reference_type, j.reference_id, j.memo
from finance.ledger_entries e join finance.ledger_journals j on j.id = e.journal_id
join finance.ledger_accounts a on a.id = e.account_id
where a.code = $1 and (sqlc.narg('before')::timestamptz is null or e.created_at < sqlc.narg('before'))
order by e.created_at desc limit $2;

-- name: FinanceTrialBalance :one
-- Must always be zero: every journal balances.
select coalesce(sum(amount_kobo), 0)::bigint as total, count(*)::bigint as entries from finance.ledger_entries;

-- name: FinanceSucceededPaymentsWithoutSale :many
-- Nightly check: every succeeded order payment has exactly one sale journal.
select p.id, p.provider, p.provider_reference, p.amount_kobo from payments.payments p
where p.status = 'succeeded' and p.order_id is not null
  and not exists (select 1 from finance.ledger_journals j where j.posting_key = 'payment:' || p.id::text || ':captured')
limit 100;

-- name: FinanceGetPeriod :one
select * from finance.accounting_periods where month = $1;

-- name: FinanceClosePeriod :one
insert into finance.accounting_periods (month, status, closed_by, closed_at) values ($1, 'closed', $2, now())
on conflict (month) do update set status = 'closed', closed_by = excluded.closed_by, closed_at = now()
where finance.accounting_periods.status <> 'closed'
returning *;

-- name: FinanceListPeriods :many
select * from finance.accounting_periods order by month desc limit 24;

-- name: FinanceDashboard :one
select
  coalesce(sum(e.amount_kobo) filter (where a.code like 'clearing:%'), 0)::bigint as clearing_kobo,
  coalesce(sum(e.amount_kobo) filter (where a.code like 'bank:%'), 0)::bigint as bank_kobo,
  coalesce(-sum(e.amount_kobo) filter (where a.code like 'seller_payable:%'), 0)::bigint as owed_to_sellers_kobo,
  coalesce(-sum(e.amount_kobo) filter (where a.code = 'liability:vat'), 0)::bigint as vat_kobo,
  coalesce(-sum(e.amount_kobo) filter (where a.kind = 'revenue'), 0)::bigint as revenue_kobo,
  coalesce(-sum(e.amount_kobo) filter (where a.code = 'liability:customer_refunds'), 0)::bigint as refunds_owed_kobo,
  coalesce(sum(e.amount_kobo) filter (where a.kind = 'expense'), 0)::bigint as expenses_kobo
from finance.ledger_entries e join finance.ledger_accounts a on a.id = e.account_id;
