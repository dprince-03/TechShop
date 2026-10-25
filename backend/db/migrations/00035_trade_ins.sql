-- Trade-ins (M5): catalogue match for the estimate, store drop-off and inspection, offer expiry,
-- payout details, and refurbishment jobs that aren't warranty claims.
-- Plan: docs/payments-finance.md §11, docs/plan.md 2026-10-23.

-- +goose Up
alter table aftersales.trade_ins
  add column variant_id        uuid references catalog.product_variants (id),
  add column serial_number     text check (length(serial_number) between 3 and 40),
  add column store_id          uuid references inventory.warehouses (id),
  add column received_at       timestamptz,
  add column received_by       uuid references identity.users (id),
  add column inspected_at      timestamptz,
  add column inspection_notes  text,
  add column offer_expires_at  timestamptz,
  add column decided_at        timestamptz,
  add column account_name      text,
  add column payout_reference  text,
  add column paid_by           uuid references identity.users (id),
  add column repair_job_id     uuid references aftersales.repair_jobs (id),
  add constraint trade_ins_offer_expiry_check check (status <> 'offer_sent' or offer_expires_at is not null);
comment on column aftersales.trade_ins.variant_id is 'The catalogue variant it matches (prices the estimate; needed before it can enter stock).';
create index trade_ins_queue_idx on aftersales.trade_ins (status, created_at);
create index trade_ins_offer_expiry_idx on aftersales.trade_ins (offer_expires_at) where status = 'offer_sent';
create index trade_ins_variant_id_idx on aftersales.trade_ins (variant_id);
create index trade_ins_store_id_idx on aftersales.trade_ins (store_id);
create index trade_ins_received_by_idx on aftersales.trade_ins (received_by);
create index trade_ins_paid_by_idx on aftersales.trade_ins (paid_by);
create index trade_ins_repair_job_id_idx on aftersales.trade_ins (repair_job_id);
create index trade_ins_imei_idx on aftersales.trade_ins (imei) where imei is not null;

-- Refurbishment jobs (trade-ins, returns) have no warranty claim; they say what they're for.
alter table aftersales.repair_jobs add column purpose text not null default 'warranty'
  check (purpose in ('warranty', 'refurbish'));

-- Trade-in purchases and payouts get their own journal kind.
alter table finance.ledger_journals drop constraint ledger_journals_kind_check;
alter table finance.ledger_journals add constraint ledger_journals_kind_check check (kind in ('sale', 'delivery_fee', 'commission', 'payout',
  'refund', 'chargeback', 'adjustment', 'reversal', 'settlement', 'fee', 'cogs', 'store_credit', 'write_off', 'trade_in'));

-- +goose Down
alter table finance.ledger_journals drop constraint ledger_journals_kind_check;
alter table finance.ledger_journals add constraint ledger_journals_kind_check check (kind in ('sale', 'delivery_fee', 'commission', 'payout',
  'refund', 'chargeback', 'adjustment', 'reversal', 'settlement', 'fee', 'cogs', 'store_credit', 'write_off'));
alter table aftersales.repair_jobs drop column if exists purpose;
drop index if exists aftersales.trade_ins_imei_idx;
drop index if exists aftersales.trade_ins_repair_job_id_idx;
drop index if exists aftersales.trade_ins_paid_by_idx;
drop index if exists aftersales.trade_ins_received_by_idx;
drop index if exists aftersales.trade_ins_store_id_idx;
drop index if exists aftersales.trade_ins_variant_id_idx;
drop index if exists aftersales.trade_ins_offer_expiry_idx;
drop index if exists aftersales.trade_ins_queue_idx;
alter table aftersales.trade_ins
  drop constraint if exists trade_ins_offer_expiry_check,
  drop column if exists repair_job_id,
  drop column if exists paid_by,
  drop column if exists payout_reference,
  drop column if exists account_name,
  drop column if exists decided_at,
  drop column if exists offer_expires_at,
  drop column if exists inspection_notes,
  drop column if exists inspected_at,
  drop column if exists received_by,
  drop column if exists received_at,
  drop column if exists store_id,
  drop column if exists serial_number,
  drop column if exists variant_id;
