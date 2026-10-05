-- Payments: charges, provider webhooks, refunds (dual control), transfer accounts, disputes,
-- provider health and daily reconciliation. Plan: docs/payments-finance.md.

-- +goose Up
create table payments.payments (
  id                  uuid primary key default gen_random_uuid(),
  order_id            uuid references sales.orders (id),
  invoice_id          uuid,
  provider            text not null check (provider in ('paystack', 'opay', 'moniepoint', 'bank_transfer', 'credit', 'cash', 'store_credit', 'pos_terminal')),
  channel             text check (channel in ('card', 'bank_transfer', 'ussd', 'wallet', 'cash', 'credit')),
  provider_reference  text,
  provider_status     text,
  checkout_url        text,
  idempotency_key     text not null unique,
  amount_kobo         bigint not null check (amount_kobo > 0),
  fees_kobo           bigint check (fees_kobo >= 0),
  currency            char(3) not null default 'NGN',
  status              text not null default 'initiated' check (status in ('initiated', 'pending', 'pending_review', 'succeeded',
                        'failed', 'cancelled', 'reversed')),
  expires_at          timestamptz,
  paid_at             timestamptz,
  failure_reason      text,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  unique (provider, provider_reference),
  check ((order_id is null) <> (invoice_id is null)),
  check (status <> 'succeeded' or paid_at is not null)
);
comment on table payments.payments is '[payments] A charge for an order or invoice. Always re-verified with the provider; card data is never stored.';
create index payments_pending_idx on payments.payments (created_at) where status in ('pending', 'pending_review');

create table payments.payment_events (
  id                 uuid primary key default gen_random_uuid(),
  provider           text not null check (provider in ('paystack', 'opay', 'moniepoint')),
  provider_event_id  text not null,
  event_type         text not null,
  payment_id         uuid references payments.payments (id),
  refund_id          uuid,
  payout_id          uuid,
  payload            jsonb not null,
  signature_valid    boolean not null,
  received_at        timestamptz not null default now(),
  processed_at       timestamptz,
  unique (provider, provider_event_id)
);
comment on table payments.payment_events is '[payments] Raw provider webhooks, stored once per provider event so retries are ignored.';

create table payments.refunds (
  id                  uuid primary key default gen_random_uuid(),
  payment_id          uuid not null references payments.payments (id),
  order_id            uuid not null references sales.orders (id),
  return_request_id   uuid,
  amount_kobo         bigint not null check (amount_kobo > 0),
  currency            char(3) not null default 'NGN',
  method              text not null default 'original' check (method in ('original', 'bank_transfer', 'store_credit')),
  reason              text not null,
  status              text not null default 'requested' check (status in ('requested', 'approved', 'processing', 'succeeded', 'failed', 'rejected')),
  requested_by        uuid references identity.users (id),
  approved_by         uuid references identity.users (id),
  provider_reference  text,
  failure_reason      text,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  check (status not in ('approved', 'processing', 'succeeded') or approved_by is not null),
  -- Separation of duties: whoever asks for a refund can't approve it.
  check (approved_by is distinct from requested_by or approved_by is null),
  check (status <> 'failed' or failure_reason is not null)
);
comment on table payments.refunds is '[payments] Money returned to a customer. Requester and approver must be different people; requested_by is null for automatic refunds.';

alter table payments.payment_events
  add constraint payment_events_refund_id_fkey foreign key (refund_id) references payments.refunds (id);

create table payments.virtual_accounts (
  id              uuid primary key default gen_random_uuid(),
  provider        text not null check (provider in ('paystack', 'moniepoint')),
  account_number  text not null check (account_number ~ '^[0-9]{10}$'),
  bank_name       text not null,
  owner_type      text not null check (owner_type in ('order', 'business')),
  order_id        uuid references sales.orders (id),
  business_id     uuid references b2b.businesses (id),
  expected_kobo   bigint check (expected_kobo > 0),
  received_kobo   bigint not null default 0 check (received_kobo >= 0),
  expires_at      timestamptz,
  status          text not null default 'active' check (status in ('active', 'paid', 'expired', 'closed')),
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  unique (provider, account_number),
  check ((owner_type = 'order' and order_id is not null and business_id is null and expected_kobo is not null and expires_at is not null)
      or (owner_type = 'business' and business_id is not null and order_id is null))
);
comment on table payments.virtual_accounts is '[payments] Pay-by-transfer accounts: one-time per order (exact amount, expires) or permanent per business.';

create table payments.disputes (
  id                   uuid primary key default gen_random_uuid(),
  payment_id           uuid not null references payments.payments (id),
  provider_dispute_id  text not null,
  provider             text not null check (provider in ('paystack', 'opay', 'moniepoint')),
  reason               text not null,
  amount_kobo          bigint not null check (amount_kobo > 0),
  due_by               timestamptz not null,
  status               text not null default 'open' check (status in ('open', 'evidence_submitted', 'won', 'lost', 'accepted')),
  decided_by           uuid references identity.users (id),
  created_at           timestamptz not null default now(),
  updated_at           timestamptz not null default now(),
  unique (provider, provider_dispute_id)
);
comment on table payments.disputes is '[payments] Chargebacks from card schemes, with an evidence deadline.';

create table payments.dispute_evidence (
  id          uuid primary key default gen_random_uuid(),
  dispute_id  uuid not null references payments.disputes (id) on delete cascade,
  kind        text not null check (kind in ('delivery_proof', 'otp_record', 'invoice', 'device_serial', 'sign_in_history', 'messages', 'other')),
  file_id     uuid references platform.files (id),
  summary     text,
  created_at  timestamptz not null default now(),
  check (file_id is not null or summary is not null)
);
comment on table payments.dispute_evidence is '[payments] Evidence gathered automatically for a dispute.';

create table payments.provider_health (
  provider         text primary key check (provider in ('paystack', 'opay', 'moniepoint')),
  breaker_state    text not null default 'closed' check (breaker_state in ('closed', 'open', 'half_open')),
  failures         integer not null default 0 check (failures >= 0),
  last_failure_at  timestamptz,
  updated_at       timestamptz not null default now()
);
comment on table payments.provider_health is '[payments] Shared circuit-breaker state, so checkout hides a provider that is failing right now.';

create table payments.settlement_reports (
  id           uuid primary key default gen_random_uuid(),
  provider     text not null check (provider in ('paystack', 'opay', 'moniepoint')),
  report_date  date not null,
  batch_ref    text not null,
  gross_kobo   bigint not null,
  fees_kobo    bigint not null check (fees_kobo >= 0),
  net_kobo     bigint not null,
  imported_at  timestamptz not null default now(),
  unique (provider, batch_ref),
  check (net_kobo = gross_kobo - fees_kobo)
);
comment on table payments.settlement_reports is '[payments] Provider settlement batches imported for daily reconciliation.';

create table payments.settlement_lines (
  id                  uuid primary key default gen_random_uuid(),
  report_id           uuid not null references payments.settlement_reports (id) on delete cascade,
  provider_reference  text not null,
  amount_kobo         bigint not null,
  fee_kobo            bigint not null default 0,
  payment_id          uuid references payments.payments (id),
  match_status        text not null default 'unmatched' check (match_status in ('unmatched', 'matched', 'exception')),
  unique (report_id, provider_reference)
);
comment on table payments.settlement_lines is '[payments] One line per settled transaction; matched to our payments by reference and amount.';

create table payments.bank_statement_lines (
  id            uuid primary key default gen_random_uuid(),
  bank_account  text not null,
  value_date    date not null,
  amount_kobo   bigint not null,
  narration     text,
  reference     text,
  match_status  text not null default 'unmatched' check (match_status in ('unmatched', 'matched', 'exception')),
  matched_type  text check (matched_type in ('settlement', 'payout', 'b2b_payment', 'other')),
  matched_id    uuid,
  imported_at   timestamptz not null default now(),
  check (match_status <> 'matched' or (matched_type is not null))
);
comment on table payments.bank_statement_lines is '[payments] Imported bank statement lines, matched to settlements and payouts.';

create table payments.reconciliation_exceptions (
  id             uuid primary key default gen_random_uuid(),
  kind           text not null check (kind in ('missing_in_provider', 'missing_in_ours', 'amount_mismatch', 'fee_mismatch',
                   'duplicate', 'unsettled_batch')),
  provider       text,
  reference      text,
  amount_kobo    bigint,
  detail         text not null,
  status         text not null default 'open' check (status in ('open', 'resolved', 'escalated')),
  owner_id       uuid references identity.users (id),
  resolution     text,
  resolved_by    uuid references identity.users (id),
  resolved_at    timestamptz,
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),
  check (status = 'open' or (resolution is not null and resolved_by is not null and resolved_at is not null))
);
comment on table payments.reconciliation_exceptions is '[payments] Anything that did not reconcile, with an owner and a 48-hour target.';
create index reconciliation_exceptions_open_idx on payments.reconciliation_exceptions (created_at) where status = 'open';

-- Indexes on foreign-key columns (every FK is indexed).
create index dispute_evidence_dispute_id_idx on payments.dispute_evidence (dispute_id);
create index dispute_evidence_file_id_idx on payments.dispute_evidence (file_id);
create index disputes_decided_by_idx on payments.disputes (decided_by);
create index disputes_payment_id_idx on payments.disputes (payment_id);
create index payment_events_payment_id_idx on payments.payment_events (payment_id);
create index payment_events_payout_id_idx on payments.payment_events (payout_id);
create index payment_events_refund_id_idx on payments.payment_events (refund_id);
create index payments_invoice_id_idx on payments.payments (invoice_id);
create index payments_order_id_idx on payments.payments (order_id);
create index reconciliation_exceptions_owner_id_idx on payments.reconciliation_exceptions (owner_id);
create index reconciliation_exceptions_resolved_by_idx on payments.reconciliation_exceptions (resolved_by);
create index refunds_approved_by_idx on payments.refunds (approved_by);
create index refunds_order_id_idx on payments.refunds (order_id);
create index refunds_payment_id_idx on payments.refunds (payment_id);
create index refunds_requested_by_idx on payments.refunds (requested_by);
create index refunds_return_request_id_idx on payments.refunds (return_request_id);
create index settlement_lines_payment_id_idx on payments.settlement_lines (payment_id);
create index virtual_accounts_business_id_idx on payments.virtual_accounts (business_id);
create index virtual_accounts_order_id_idx on payments.virtual_accounts (order_id);

-- Keep updated_at current.
create trigger disputes_set_updated_at before update on payments.disputes
  for each row execute function platform.set_updated_at();
create trigger payments_set_updated_at before update on payments.payments
  for each row execute function platform.set_updated_at();
create trigger provider_health_set_updated_at before update on payments.provider_health
  for each row execute function platform.set_updated_at();
create trigger reconciliation_exceptions_set_updated_at before update on payments.reconciliation_exceptions
  for each row execute function platform.set_updated_at();
create trigger refunds_set_updated_at before update on payments.refunds
  for each row execute function platform.set_updated_at();
create trigger virtual_accounts_set_updated_at before update on payments.virtual_accounts
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists payments.reconciliation_exceptions, payments.bank_statement_lines,
  payments.settlement_lines, payments.settlement_reports, payments.provider_health,
  payments.dispute_evidence, payments.disputes, payments.virtual_accounts, payments.refunds,
  payments.payment_events, payments.payments cascade;
