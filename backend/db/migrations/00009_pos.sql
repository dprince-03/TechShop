-- Point of sale: tills in physical stores and cashier shifts. POS sales are orders with channel 'pos'.

-- +goose Up
create table pos.pos_terminals (
  id            uuid primary key default gen_random_uuid(),
  warehouse_id  uuid not null references inventory.warehouses (id),
  label         text not null,
  status        text not null default 'active' check (status in ('active', 'retired')),
  created_at    timestamptz not null default now(),
  unique (warehouse_id, label)
);
comment on table pos.pos_terminals is '[pos] Tills in physical stores (warehouses of kind store).';

create table pos.pos_shifts (
  id                   uuid primary key default gen_random_uuid(),
  terminal_id          uuid not null references pos.pos_terminals (id),
  cashier_id           uuid not null references identity.staff_members (user_id),
  opened_at            timestamptz not null default now(),
  closed_at            timestamptz,
  opening_float_kobo   bigint not null default 0 check (opening_float_kobo >= 0),
  expected_cash_kobo   bigint,
  counted_cash_kobo    bigint,
  variance_kobo        bigint generated always as (counted_cash_kobo - expected_cash_kobo) stored,
  check (closed_at is null or closed_at > opened_at),
  check (closed_at is null or (expected_cash_kobo is not null and counted_cash_kobo is not null))
);
comment on table pos.pos_shifts is '[pos] A cashier''s shift on a till, with cash reconciliation (variance flagged).';
create unique index pos_shifts_one_open_per_terminal on pos.pos_shifts (terminal_id) where closed_at is null;

-- Indexes on foreign-key columns (every FK is indexed).
create index pos_shifts_cashier_id_idx on pos.pos_shifts (cashier_id);

-- +goose Down
drop table if exists pos.pos_shifts, pos.pos_terminals cascade;
