-- Warranty claims and repairs (M5). Warranty terms are copied onto each order line at checkout
-- so a later listing change can't shorten a warranty already sold.
-- Plan: docs/orders-fulfilment.md §7, docs/plan.md 2026-10-23.

-- +goose Up
alter table sales.order_lines
  add column warranty_provider text check (warranty_provider in ('manufacturer', 'techshop', 'seller', 'none')),
  add column warranty_months   smallint check (warranty_months between 0 and 60);
update sales.order_lines ol set warranty_provider = l.warranty_provider, warranty_months = l.warranty_months
from catalog.listings l where l.id = ol.listing_id and ol.warranty_months is null;

alter table aftersales.warranty_claims drop constraint warranty_claims_status_check;
alter table aftersales.warranty_claims
  add constraint warranty_claims_status_check check (status in ('new', 'received', 'in_repair', 'ready', 'collected', 'rejected', 'closed')),
  add column warranty_provider   text check (warranty_provider in ('manufacturer', 'techshop', 'seller')),
  add column warranty_ends_on    date,
  add column store_id            uuid references inventory.warehouses (id),
  add column received_at         timestamptz,
  add column received_by         uuid references identity.users (id),
  add column rejection_reason    text,
  add column resolution_note     text,
  add column replacement_unit_id uuid references inventory.device_units (id),
  add column refund_id           uuid references payments.refunds (id),
  add column closed_at           timestamptz,
  add constraint warranty_claims_rejected_check check (status <> 'rejected' or rejection_reason is not null),
  add constraint warranty_claims_resolved_check check (status not in ('collected', 'closed') or resolution is not null);
comment on column aftersales.warranty_claims.status is
  'new (waiting for drop-off) → received → in_repair → ready (for collection) → collected; or rejected; closed = refunded, device kept.';

-- One open claim per item: a second claim waits until the first is finished.
create unique index warranty_claims_one_open on aftersales.warranty_claims (order_line_id, imei_or_serial)
  where status in ('new', 'received', 'in_repair', 'ready');
create index warranty_claims_queue_idx on aftersales.warranty_claims (status, created_at);
create index warranty_claims_store_id_idx on aftersales.warranty_claims (store_id);
create index warranty_claims_received_by_idx on aftersales.warranty_claims (received_by);
create index warranty_claims_replacement_unit_id_idx on aftersales.warranty_claims (replacement_unit_id);
create index warranty_claims_refund_id_idx on aftersales.warranty_claims (refund_id);
create index warranty_claims_imei_idx on aftersales.warranty_claims (imei_or_serial);

alter table aftersales.repair_jobs add column notes text;
create index repair_jobs_status_idx on aftersales.repair_jobs (status, created_at);

-- A replacement unit handed over at the counter leaves stock with its own reason.
alter table inventory.stock_movements drop constraint stock_movements_reason_check;
alter table inventory.stock_movements add constraint stock_movements_reason_check check (reason in ('purchase_receipt', 'sale', 'return',
  'transfer_in', 'transfer_out', 'adjustment', 'trade_in', 'write_off', 'count_variance', 'warranty_replacement'));

-- +goose Down
alter table inventory.stock_movements drop constraint stock_movements_reason_check;
alter table inventory.stock_movements add constraint stock_movements_reason_check check (reason in ('purchase_receipt', 'sale', 'return',
  'transfer_in', 'transfer_out', 'adjustment', 'trade_in', 'write_off', 'count_variance'));
drop index if exists aftersales.repair_jobs_status_idx;
alter table aftersales.repair_jobs drop column if exists notes;
drop index if exists aftersales.warranty_claims_imei_idx;
drop index if exists aftersales.warranty_claims_refund_id_idx;
drop index if exists aftersales.warranty_claims_replacement_unit_id_idx;
drop index if exists aftersales.warranty_claims_received_by_idx;
drop index if exists aftersales.warranty_claims_store_id_idx;
drop index if exists aftersales.warranty_claims_queue_idx;
drop index if exists aftersales.warranty_claims_one_open;
alter table aftersales.warranty_claims
  drop constraint if exists warranty_claims_resolved_check,
  drop constraint if exists warranty_claims_rejected_check,
  drop column if exists closed_at,
  drop column if exists refund_id,
  drop column if exists replacement_unit_id,
  drop column if exists resolution_note,
  drop column if exists rejection_reason,
  drop column if exists received_by,
  drop column if exists received_at,
  drop column if exists store_id,
  drop column if exists warranty_ends_on,
  drop column if exists warranty_provider;
alter table aftersales.warranty_claims drop constraint warranty_claims_status_check;
alter table aftersales.warranty_claims add constraint warranty_claims_status_check
  check (status in ('new', 'received', 'in_repair', 'ready', 'collected', 'rejected'));
alter table sales.order_lines drop column if exists warranty_months, drop column if exists warranty_provider;
