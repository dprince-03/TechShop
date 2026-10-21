-- Seller payouts (M4, docs/payments-finance.md §9): what each marketplace line earns the seller
-- is snapshotted at payment capture, batches need a second approver above the configured
-- threshold, and payouts keep the provider transfer code apart from our own reference.

-- +goose Up
alter table sales.order_lines add column seller_earning_kobo bigint check (seller_earning_kobo >= 0);
comment on column sales.order_lines.seller_earning_kobo is 'Marketplace lines: what the seller is owed before commission (net price plus any TechShop-funded discount), set at payment capture.';

alter table finance.payout_batches
  add column submitted_at         timestamptz,
  add column second_approved_by   uuid references identity.users (id),
  add column second_approved_at   timestamptz,
  add constraint payout_batches_second_approver_check check (
    second_approved_by is null or (second_approved_by <> prepared_by and second_approved_by <> approved_by and second_approved_at is not null));
create index payout_batches_second_approved_by_idx on finance.payout_batches (second_approved_by);

alter table finance.payouts
  add column transfer_code text,
  add column attempts      integer not null default 0 check (attempts >= 0);
comment on column finance.payouts.provider_reference is 'Our transfer reference sent to the provider (the payout id), so retries never pay twice.';
comment on column finance.payouts.transfer_code is 'The provider''s own code for the transfer.';
create index payouts_processing_idx on finance.payouts (updated_at) where status = 'processing';

-- +goose Down
drop index finance.payouts_processing_idx;
alter table finance.payouts drop column attempts, drop column transfer_code;
drop index finance.payout_batches_second_approved_by_idx;
alter table finance.payout_batches drop constraint payout_batches_second_approver_check,
  drop column second_approved_at, drop column second_approved_by, drop column submitted_at;
alter table sales.order_lines drop column seller_earning_kobo;
