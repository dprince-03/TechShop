-- After-sales: returns (RMA) with inspection and grading, warranty claims, repairs and trade-ins.
-- Plan: docs/orders-fulfilment.md §7.

-- +goose Up
create sequence aftersales.rma_number_seq start 1001;
create sequence aftersales.claim_number_seq start 1001;
create sequence aftersales.trade_in_ref_seq start 1001;

create table aftersales.return_requests (
  id            uuid primary key default gen_random_uuid(),
  rma_number    text not null unique default platform.next_ref('RMA-', 'aftersales.rma_number_seq'),
  order_id      uuid not null references sales.orders (id),
  requested_by  uuid not null references identity.users (id),
  reason        text not null check (reason in ('faulty', 'damaged', 'not_as_described', 'wrong_item', 'change_of_mind')),
  details       text,
  return_method text check (return_method in ('pickup', 'store_dropoff', 'seller_return')),
  status        text not null default 'requested' check (status in ('requested', 'approved', 'rejected', 'awaiting_pickup',
                  'received', 'inspected', 'refunded', 'replaced', 'closed')),
  auto_approved boolean not null default false,
  resolution    text check (resolution in ('refund', 'replace', 'repair')),
  decided_by    uuid references identity.users (id),
  rejection_reason text,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  check (status <> 'rejected' or rejection_reason is not null)
);
comment on table aftersales.return_requests is '[aftersales] Customer return requests (RMA) and their outcome. Faulty items within 7 days may be auto-approved.';

alter table payments.refunds
  add constraint refunds_return_request_id_fkey foreign key (return_request_id) references aftersales.return_requests (id);
alter table logistics.delivery_jobs
  add constraint delivery_jobs_return_request_id_fkey foreign key (return_request_id) references aftersales.return_requests (id);

create table aftersales.return_items (
  id                  uuid primary key default gen_random_uuid(),
  return_request_id   uuid not null references aftersales.return_requests (id) on delete cascade,
  order_line_id       uuid not null references sales.order_lines (id),
  quantity            integer not null check (quantity > 0),
  unique (return_request_id, order_line_id)
);
comment on table aftersales.return_items is '[aftersales] Which order lines (and how many) are being returned.';

create table aftersales.return_inspections (
  id              uuid primary key default gen_random_uuid(),
  return_item_id  uuid not null references aftersales.return_items (id) on delete cascade,
  device_unit_id  uuid references inventory.device_units (id),
  scanned_serial  text,
  imei_match      boolean,
  grade           text not null check (grade in ('A', 'B', 'C', 'faulty')),
  disposition     text not null check (disposition in ('restock_new', 'restock_open_box', 'repair', 'return_to_seller',
                    'return_to_supplier', 'write_off')),
  notes           text,
  inspected_by    uuid not null references identity.users (id),
  inspected_at    timestamptz not null default now(),
  check (device_unit_id is null or imei_match is not null)
);
comment on table aftersales.return_inspections is '[aftersales] Inspection of returned items: serial must match the unit sold (swap fraud), grade and disposition.';

create table aftersales.warranty_claims (
  id                 uuid primary key default gen_random_uuid(),
  claim_number       text not null unique default platform.next_ref('WR-', 'aftersales.claim_number_seq'),
  customer_user_id   uuid not null references identity.users (id),
  order_line_id      uuid references sales.order_lines (id),
  device_unit_id     uuid references inventory.device_units (id),
  imei_or_serial     text not null,
  fault_description  text not null,
  status             text not null default 'new' check (status in ('new', 'received', 'in_repair', 'ready', 'collected', 'rejected')),
  resolution         text check (resolution in ('repaired', 'replaced', 'refunded', 'not_covered')),
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now()
);
comment on table aftersales.warranty_claims is '[aftersales] Warranty claims, identified by IMEI/serial.';

create table aftersales.repair_jobs (
  id                 uuid primary key default gen_random_uuid(),
  warranty_claim_id  uuid references aftersales.warranty_claims (id),
  device_unit_id     uuid references inventory.device_units (id),
  technician_id      uuid references identity.staff_members (user_id),
  status             text not null default 'queued' check (status in ('queued', 'diagnosing', 'awaiting_parts', 'repairing', 'done', 'unrepairable')),
  diagnosis          text,
  chargeable         boolean not null default false,
  cost_kobo          bigint check (cost_kobo >= 0),
  started_at         timestamptz,
  finished_at        timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (warranty_claim_id is not null or device_unit_id is not null),
  check (not chargeable or cost_kobo is not null)
);
comment on table aftersales.repair_jobs is '[aftersales] Workshop jobs for warranty claims, paid repairs and refurbishment of returns.';

create table aftersales.trade_ins (
  id                        uuid primary key default gen_random_uuid(),
  reference                 text not null unique default platform.next_ref('TI-', 'aftersales.trade_in_ref_seq'),
  user_id                   uuid not null references identity.users (id),
  device_type               text not null check (device_type in ('phone', 'laptop', 'tablet', 'console', 'smartwatch')),
  brand                     text not null,
  model                     text not null,
  storage                   text,
  declared_condition        text not null check (declared_condition in ('like_new', 'good', 'fair', 'faulty')),
  imei                      text check (imei ~ '^[0-9]{15}$'),
  estimate_kobo             bigint check (estimate_kobo >= 0),
  inspected_condition       text check (inspected_condition in ('like_new', 'good', 'fair', 'faulty')),
  offer_kobo                bigint check (offer_kobo >= 0),
  payout_method             text check (payout_method in ('bank', 'store_credit')),
  bank_code                 text,
  account_number_encrypted  bytea,
  account_number_last4      text check (length(account_number_last4) = 4),
  encryption_key_id         uuid references platform.encryption_keys (id),
  status                    text not null default 'submitted' check (status in ('submitted', 'estimated', 'awaiting_inspection',
                              'inspected', 'offer_sent', 'accepted', 'declined', 'paid', 'cancelled', 'blocked')),
  inspected_by              uuid references identity.staff_members (user_id),
  device_unit_id            uuid unique references inventory.device_units (id),
  paid_at                   timestamptz,
  created_at                timestamptz not null default now(),
  updated_at                timestamptz not null default now(),
  check (status not in ('offer_sent', 'accepted', 'paid') or offer_kobo is not null),
  check (payout_method is distinct from 'bank' or status not in ('paid') or account_number_encrypted is not null),
  check (status <> 'paid' or paid_at is not null)
);
comment on table aftersales.trade_ins is '[aftersales] Trade-ins from estimate to inspection, offer and payout (bank or store credit); accepted devices become device units.';

-- Indexes on foreign-key columns (every FK is indexed).
create index repair_jobs_device_unit_id_idx on aftersales.repair_jobs (device_unit_id);
create index repair_jobs_technician_id_idx on aftersales.repair_jobs (technician_id);
create index repair_jobs_warranty_claim_id_idx on aftersales.repair_jobs (warranty_claim_id);
create index return_inspections_device_unit_id_idx on aftersales.return_inspections (device_unit_id);
create index return_inspections_inspected_by_idx on aftersales.return_inspections (inspected_by);
create index return_inspections_return_item_id_idx on aftersales.return_inspections (return_item_id);
create index return_items_order_line_id_idx on aftersales.return_items (order_line_id);
create index return_requests_decided_by_idx on aftersales.return_requests (decided_by);
create index return_requests_order_id_idx on aftersales.return_requests (order_id);
create index return_requests_requested_by_idx on aftersales.return_requests (requested_by);
create index trade_ins_encryption_key_id_idx on aftersales.trade_ins (encryption_key_id);
create index trade_ins_inspected_by_idx on aftersales.trade_ins (inspected_by);
create index trade_ins_user_id_idx on aftersales.trade_ins (user_id);
create index warranty_claims_customer_user_id_idx on aftersales.warranty_claims (customer_user_id);
create index warranty_claims_device_unit_id_idx on aftersales.warranty_claims (device_unit_id);
create index warranty_claims_order_line_id_idx on aftersales.warranty_claims (order_line_id);

-- Keep updated_at current.
create trigger repair_jobs_set_updated_at before update on aftersales.repair_jobs
  for each row execute function platform.set_updated_at();
create trigger return_requests_set_updated_at before update on aftersales.return_requests
  for each row execute function platform.set_updated_at();
create trigger trade_ins_set_updated_at before update on aftersales.trade_ins
  for each row execute function platform.set_updated_at();
create trigger warranty_claims_set_updated_at before update on aftersales.warranty_claims
  for each row execute function platform.set_updated_at();

-- +goose Down
alter table logistics.delivery_jobs drop constraint if exists delivery_jobs_return_request_id_fkey;
alter table payments.refunds drop constraint if exists refunds_return_request_id_fkey;
drop table if exists aftersales.trade_ins, aftersales.repair_jobs, aftersales.warranty_claims,
  aftersales.return_inspections, aftersales.return_items, aftersales.return_requests cascade;
drop sequence if exists aftersales.trade_in_ref_seq, aftersales.claim_number_seq, aftersales.rma_number_seq;
