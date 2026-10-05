-- Finance: invoices, double-entry ledger (posting keys, reversals, period locks, balance check),
-- commission rules and seller payouts with dual control. Plan: docs/payments-finance.md §4, §9–§10.

-- +goose Up
create sequence finance.invoice_number_seq start 1001;

create table finance.invoices (
  id                 uuid primary key default gen_random_uuid(),
  invoice_number     text not null unique default platform.next_ref('INV-', 'finance.invoice_number_seq'),
  business_id        uuid not null references b2b.businesses (id),
  order_id           uuid not null unique references sales.orders (id),
  kind               text not null default 'tax' check (kind in ('proforma', 'tax')),
  issued_at          timestamptz not null default now(),
  due_at             timestamptz not null,
  subtotal_kobo      bigint not null check (subtotal_kobo >= 0),
  vat_kobo           bigint not null default 0 check (vat_kobo >= 0),
  total_kobo         bigint not null check (total_kobo >= 0),
  amount_paid_kobo   bigint not null default 0 check (amount_paid_kobo >= 0),
  currency           char(3) not null default 'NGN',
  status             text not null default 'issued' check (status in ('issued', 'partially_paid', 'paid', 'overdue', 'void', 'written_off')),
  pdf_file_id        uuid references platform.files (id),
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (total_kobo = subtotal_kobo + vat_kobo),
  check (amount_paid_kobo <= total_kobo),
  check (due_at >= issued_at)
);
comment on table finance.invoices is '[finance] B2B invoices (proforma or tax); on credit terms they are paid later against due_at.';
create index invoices_open_due_idx on finance.invoices (due_at) where status in ('issued', 'partially_paid', 'overdue');

alter table payments.payments
  add constraint payments_invoice_id_fkey foreign key (invoice_id) references finance.invoices (id);

create table finance.ledger_accounts (
  id          uuid primary key default gen_random_uuid(),
  code        text not null unique check (code ~ '^[a-z_]+(:[a-z0-9_-]+)+$'),
  name        text not null,
  kind        text not null check (kind in ('asset', 'liability', 'revenue', 'expense', 'equity')),
  seller_id   uuid unique references sellers.sellers (id),
  user_id     uuid unique references identity.users (id),
  created_at  timestamptz not null default now(),
  check (seller_id is null or user_id is null)
);
comment on table finance.ledger_accounts is '[finance] Chart of accounts: provider clearing, bank, revenue, VAT, refunds, one payable per seller, store credit per customer.';

create table finance.accounting_periods (
  month      date primary key check (month = date_trunc('month', month)::date),
  status     text not null default 'open' check (status in ('open', 'closing', 'closed')),
  closed_by  uuid references identity.users (id),
  closed_at  timestamptz,
  check (status <> 'closed' or (closed_by is not null and closed_at is not null))
);
comment on table finance.accounting_periods is '[finance] Month-end close. Journals cannot be posted into a closed month.';

create table finance.ledger_journals (
  id                   uuid primary key default gen_random_uuid(),
  posting_key          text not null unique,
  kind                 text not null check (kind in ('sale', 'delivery_fee', 'commission', 'payout', 'refund', 'chargeback',
                         'adjustment', 'reversal', 'settlement', 'fee', 'cogs', 'store_credit', 'write_off')),
  reference_type       text not null,
  reference_id         uuid not null,
  period               date not null default date_trunc('month', now())::date check (period = date_trunc('month', period)::date),
  reverses_journal_id  uuid unique references finance.ledger_journals (id),
  memo                 text,
  posted_by            uuid references identity.users (id),
  posted_at            timestamptz not null default now(),
  check (kind <> 'reversal' or reverses_journal_id is not null)
);
comment on table finance.ledger_journals is '[finance] One accounting event. posting_key makes posting idempotent; journals are never edited, only reversed.';
create index ledger_journals_reference_idx on finance.ledger_journals (reference_type, reference_id);

create table finance.ledger_entries (
  id           uuid primary key default gen_random_uuid(),
  journal_id   uuid not null references finance.ledger_journals (id),
  account_id   uuid not null references finance.ledger_accounts (id),
  amount_kobo  bigint not null check (amount_kobo <> 0),
  currency     char(3) not null default 'NGN',
  created_at   timestamptz not null default now()
);
comment on table finance.ledger_entries is '[finance] Double-entry lines: positive = debit, negative = credit. Each journal sums to zero (deferred check).';
create index ledger_entries_account_idx on finance.ledger_entries (account_id, created_at);

-- A journal's entries must sum to zero when the transaction commits.
-- +goose StatementBegin
create function finance.assert_journal_balanced() returns trigger
language plpgsql as $$
declare total bigint;
begin
  select coalesce(sum(amount_kobo), 0) into total from finance.ledger_entries where journal_id = new.journal_id;
  if total <> 0 then
    raise exception 'ledger journal % is not balanced (sum = %)', new.journal_id, total;
  end if;
  return null;
end $$;
-- +goose StatementEnd
create constraint trigger ledger_entries_balanced
  after insert on finance.ledger_entries
  deferrable initially deferred
  for each row execute function finance.assert_journal_balanced();

-- No postings into a closed month.
-- +goose StatementBegin
create function finance.assert_period_open() returns trigger
language plpgsql as $$
begin
  if exists (select 1 from finance.accounting_periods where month = new.period and status = 'closed') then
    raise exception 'accounting period % is closed', to_char(new.period, 'YYYY-MM');
  end if;
  return new;
end $$;
-- +goose StatementEnd
create trigger ledger_journals_period_open before insert on finance.ledger_journals
  for each row execute function finance.assert_period_open();

create trigger ledger_journals_append_only before update or delete on finance.ledger_journals
  for each row execute function platform.forbid_change();
create trigger ledger_entries_append_only before update or delete on finance.ledger_entries
  for each row execute function platform.forbid_change();

create table finance.commission_rules (
  id           uuid primary key default gen_random_uuid(),
  category_id  uuid references catalog.categories (id),
  seller_id    uuid references sellers.sellers (id),
  rate_bps     integer not null check (rate_bps between 0 and 10000),
  valid_from   date not null default current_date,
  valid_to     date,
  created_by   uuid not null references identity.users (id),
  created_at   timestamptz not null default now(),
  check (valid_to is null or valid_to > valid_from)
);
comment on table finance.commission_rules is '[finance] Marketplace commission (basis points) by category and/or seller; the rate is snapshotted on order lines.';

create table finance.payout_batches (
  id           uuid primary key default gen_random_uuid(),
  status       text not null default 'draft' check (status in ('draft', 'submitted', 'approved', 'processing', 'done', 'cancelled')),
  prepared_by  uuid not null references identity.users (id),
  approved_by  uuid references identity.users (id),
  approved_at  timestamptz,
  total_kobo   bigint not null default 0 check (total_kobo >= 0),
  payout_count integer not null default 0 check (payout_count >= 0),
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now(),
  -- Dual control: whoever prepares a batch can't approve it.
  check (approved_by is distinct from prepared_by or approved_by is null),
  check (status not in ('approved', 'processing', 'done') or (approved_by is not null and approved_at is not null))
);
comment on table finance.payout_batches is '[finance] A payout run: prepared by one finance person, approved by a different one (with MFA step-up).';

create table finance.payouts (
  id                  uuid primary key default gen_random_uuid(),
  batch_id            uuid references finance.payout_batches (id),
  seller_id           uuid not null references sellers.sellers (id),
  bank_account_id     uuid not null references sellers.seller_bank_accounts (id),
  amount_kobo         bigint not null check (amount_kobo > 0),
  currency            char(3) not null default 'NGN',
  status              text not null default 'scheduled' check (status in ('scheduled', 'processing', 'paid', 'failed', 'on_hold')),
  scheduled_for       date not null,
  paid_at             timestamptz,
  provider_reference  text unique,
  failure_reason      text,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  check (status <> 'paid' or paid_at is not null),
  check (status <> 'failed' or failure_reason is not null)
);
comment on table finance.payouts is '[finance] Transfers of earnings to a seller''s bank account (journal posted only on success).';

create table finance.payout_items (
  payout_id      uuid not null references finance.payouts (id) on delete cascade,
  fulfilment_id  uuid not null unique references sales.fulfilments (id),
  amount_kobo    bigint not null check (amount_kobo > 0),
  primary key (payout_id, fulfilment_id)
);
comment on table finance.payout_items is '[finance] Which delivered fulfilments a payout covers; each fulfilment is paid out once.';

create table finance.payout_holds (
  id          uuid primary key default gen_random_uuid(),
  seller_id   uuid not null references sellers.sellers (id),
  reason      text not null check (reason in ('kyc_incomplete', 'risk_case', 'bank_account_changed', 'credentials_changed',
                'negative_balance', 'enforcement', 'manual')),
  note        text,
  until       timestamptz,
  created_by  uuid references identity.users (id),
  released_at timestamptz,
  created_at  timestamptz not null default now()
);
comment on table finance.payout_holds is '[finance] Reasons a seller is not paid out right now (e.g. 24 hours after a bank or credential change).';
create index payout_holds_active_idx on finance.payout_holds (seller_id) where released_at is null;

alter table payments.payment_events
  add constraint payment_events_payout_id_fkey foreign key (payout_id) references finance.payouts (id);

-- Indexes on foreign-key columns (every FK is indexed).
create index accounting_periods_closed_by_idx on finance.accounting_periods (closed_by);
create index commission_rules_category_id_idx on finance.commission_rules (category_id);
create index commission_rules_created_by_idx on finance.commission_rules (created_by);
create index commission_rules_seller_id_idx on finance.commission_rules (seller_id);
create index invoices_business_id_idx on finance.invoices (business_id);
create index invoices_pdf_file_id_idx on finance.invoices (pdf_file_id);
create index ledger_entries_journal_id_idx on finance.ledger_entries (journal_id);
create index ledger_journals_posted_by_idx on finance.ledger_journals (posted_by);
create index payout_batches_approved_by_idx on finance.payout_batches (approved_by);
create index payout_batches_prepared_by_idx on finance.payout_batches (prepared_by);
create index payout_holds_created_by_idx on finance.payout_holds (created_by);
create index payouts_bank_account_id_idx on finance.payouts (bank_account_id);
create index payouts_batch_id_idx on finance.payouts (batch_id);
create index payouts_seller_id_idx on finance.payouts (seller_id);

-- Keep updated_at current.
create trigger invoices_set_updated_at before update on finance.invoices
  for each row execute function platform.set_updated_at();
create trigger payout_batches_set_updated_at before update on finance.payout_batches
  for each row execute function platform.set_updated_at();
create trigger payouts_set_updated_at before update on finance.payouts
  for each row execute function platform.set_updated_at();

-- +goose Down
alter table payments.payment_events drop constraint if exists payment_events_payout_id_fkey;
alter table payments.payments drop constraint if exists payments_invoice_id_fkey;
drop table if exists finance.payout_holds, finance.payout_items, finance.payouts, finance.payout_batches,
  finance.commission_rules, finance.ledger_entries, finance.ledger_journals, finance.accounting_periods,
  finance.ledger_accounts, finance.invoices cascade;
drop function if exists finance.assert_period_open();
drop function if exists finance.assert_journal_balanced();
drop sequence if exists finance.invoice_number_seq;
