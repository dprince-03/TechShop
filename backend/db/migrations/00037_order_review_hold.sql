-- Orders held by the checkout risk check (M5): paid, but not released to the warehouse or the
-- seller until a risk analyst clears the case. Plan: docs/trust-safety.md §3, docs/plan.md 2026-10-23.

-- +goose Up
alter table sales.orders
  add column review_hold_at  timestamptz,
  add column review_case_id  uuid references risk.risk_cases (id);
comment on column sales.orders.review_hold_at is 'Set while the order waits for a risk review; cleared when the case is cleared.';
create index orders_review_hold_idx on sales.orders (review_hold_at) where review_hold_at is not null;
create index orders_review_case_id_idx on sales.orders (review_case_id);

-- +goose Down
drop index if exists sales.orders_review_case_id_idx;
drop index if exists sales.orders_review_hold_idx;
alter table sales.orders drop column if exists review_case_id, drop column if exists review_hold_at;
