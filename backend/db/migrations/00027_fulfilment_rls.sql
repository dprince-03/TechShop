-- Row-level security on seller fulfilments (M4 slice 3). Seller Centre requests (app.seller_id
-- set) may only change their own fulfilments. Reads stay open: the order status is derived from
-- every fulfilment of the order, so hiding other sellers' parts would compute the wrong status.
-- Seller read queries always filter by seller_id themselves.

-- +goose Up
alter table sales.fulfilments enable row level security;
create policy read_all on sales.fulfilments for select using (true);
create policy seller_insert on sales.fulfilments for insert
  with check (coalesce(current_setting('app.seller_id', true), '') = ''
              or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);
create policy seller_update on sales.fulfilments for update
  using (coalesce(current_setting('app.seller_id', true), '') = ''
         or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid)
  with check (coalesce(current_setting('app.seller_id', true), '') = ''
              or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);
create policy seller_delete on sales.fulfilments for delete
  using (coalesce(current_setting('app.seller_id', true), '') = ''
         or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);

-- +goose Down
drop policy seller_delete on sales.fulfilments;
drop policy seller_update on sales.fulfilments;
drop policy seller_insert on sales.fulfilments;
drop policy read_all on sales.fulfilments;
alter table sales.fulfilments disable row level security;
