-- Seller listings (M4 slice 2).
-- 1. Automatic checks that send a listing to review record their reasons as an 'auto_flagged'
--    review with no reviewer (only people need a reviewer_id). Before this, flagging a seller's
--    listing broke the reviewer check and the insert failed.
-- 2. Row-level security on listings for Seller Centre requests (app.seller_id set): a seller
--    writes only its own listings and reads its own plus every *active* listing (those are
--    public on the market, and the price-outlier check needs other sellers' prices).

-- +goose Up
alter table catalog.listing_reviews drop constraint listing_reviews_decision_check;
alter table catalog.listing_reviews add constraint listing_reviews_decision_check
  check (decision in ('auto_passed', 'auto_flagged', 'approved', 'changes_requested', 'rejected', 'suspended'));
alter table catalog.listing_reviews drop constraint listing_reviews_check;
alter table catalog.listing_reviews add constraint listing_reviews_reviewer_check
  check (decision in ('auto_passed', 'auto_flagged') or reviewer_id is not null);

alter table catalog.listings enable row level security;
create policy seller_read on catalog.listings for select
  using (coalesce(current_setting('app.seller_id', true), '') = ''
         or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid
         or status = 'active');
create policy seller_insert on catalog.listings for insert
  with check (coalesce(current_setting('app.seller_id', true), '') = ''
              or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);
create policy seller_update on catalog.listings for update
  using (coalesce(current_setting('app.seller_id', true), '') = ''
         or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid)
  with check (coalesce(current_setting('app.seller_id', true), '') = ''
              or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);
create policy seller_delete on catalog.listings for delete
  using (coalesce(current_setting('app.seller_id', true), '') = ''
         or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);

-- +goose Down
drop policy seller_delete on catalog.listings;
drop policy seller_update on catalog.listings;
drop policy seller_insert on catalog.listings;
drop policy seller_read on catalog.listings;
alter table catalog.listings disable row level security;
delete from catalog.listing_reviews where decision = 'auto_flagged';
alter table catalog.listing_reviews drop constraint listing_reviews_reviewer_check;
alter table catalog.listing_reviews add constraint listing_reviews_check
  check (decision = 'auto_passed' or reviewer_id is not null);
alter table catalog.listing_reviews drop constraint listing_reviews_decision_check;
alter table catalog.listing_reviews add constraint listing_reviews_decision_check
  check (decision in ('auto_passed', 'approved', 'changes_requested', 'rejected', 'suspended'));
