-- Seller hand-over at a TechShop hub (M4, docs/plan.md 2026-10-21 decision 1): a packed seller
-- parcel must be dropped off at a warehouse or hub within 24 hours (Sundays skipped), or the
-- SLA job cancels it as a late hand-over.

-- +goose Up
alter table sales.fulfilments add column handover_by timestamptz;
comment on column sales.fulfilments.handover_by is 'Seller-shipped only: drop the packed parcel at a hub before this.';
create index fulfilments_handover_due_idx on sales.fulfilments (handover_by) where status = 'packed' and fulfilled_by = 'seller';

-- +goose Down
drop index sales.fulfilments_handover_due_idx;
alter table sales.fulfilments drop column handover_by;
