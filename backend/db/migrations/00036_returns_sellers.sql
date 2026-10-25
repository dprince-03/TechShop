-- Seller returns and item-exact refunds (M5).
-- Plan: docs/orders-fulfilment.md §7, docs/plan.md 2026-10-23.

-- +goose Up
-- How a refund is charged, worked out when it's raised: the ledger lines to debit (revenue and
-- VAT for TechShop items, the seller's payable for theirs, delivery). Without it, approval falls
-- back to reversing the whole sale pro rata.
alter table payments.refunds add column allocation jsonb;
comment on column payments.refunds.allocation is
  'Ledger debits for this refund as [{"code": account, "amount": kobo}], summing to the amount. Null = pro rata over the sale.';

-- A seller may dispute the inspection of their returned item within 3 days; support decides.
alter table aftersales.return_requests
  add column seller_dispute_reason text,
  add column seller_disputed_at    timestamptz,
  add column seller_disputed_by    uuid references identity.users (id),
  add column dispute_decision      text check (dispute_decision in ('upheld', 'rejected')),
  add column dispute_note          text,
  add column dispute_decided_by    uuid references identity.users (id),
  add column dispute_decided_at    timestamptz,
  add constraint return_requests_dispute_check check (seller_disputed_at is null or seller_dispute_reason is not null),
  add constraint return_requests_dispute_decided_check check (dispute_decision is null or (seller_disputed_at is not null and dispute_decided_by is not null));
create index return_requests_seller_disputed_by_idx on aftersales.return_requests (seller_disputed_by);
create index return_requests_dispute_decided_by_idx on aftersales.return_requests (dispute_decided_by);
create index return_requests_disputed_idx on aftersales.return_requests (seller_disputed_at) where seller_disputed_at is not null and dispute_decision is null;

-- +goose Down
drop index if exists aftersales.return_requests_disputed_idx;
drop index if exists aftersales.return_requests_dispute_decided_by_idx;
drop index if exists aftersales.return_requests_seller_disputed_by_idx;
alter table aftersales.return_requests
  drop constraint if exists return_requests_dispute_decided_check,
  drop constraint if exists return_requests_dispute_check,
  drop column if exists dispute_decided_at,
  drop column if exists dispute_decided_by,
  drop column if exists dispute_note,
  drop column if exists dispute_decision,
  drop column if exists seller_disputed_by,
  drop column if exists seller_disputed_at,
  drop column if exists seller_dispute_reason;
alter table payments.refunds drop column if exists allocation;
