-- Reviews (M4 slice 4, trust-safety.md §6): a seller may answer a review once (they can never
-- remove it), and the automatic checks keep their reasons so moderators see why a review was
-- held back.

-- +goose Up
alter table catalog.product_reviews
  add column seller_response     text check (length(seller_response) between 2 and 1000),
  add column seller_response_at  timestamptz,
  add column moderation_reasons  text[] not null default '{}',
  add constraint product_reviews_response_check check ((seller_response is null) = (seller_response_at is null));
create index product_reviews_product_status_idx on catalog.product_reviews (product_id, status, created_at desc);

-- +goose Down
drop index catalog.product_reviews_product_status_idx;
alter table catalog.product_reviews
  drop constraint product_reviews_response_check,
  drop column moderation_reasons,
  drop column seller_response_at,
  drop column seller_response;
