-- Catalogue queries (docs/search-catalogue.md). Search itself is the documented dynamic
-- query in internal/catalog/search_query.go.

-- name: CatalogListCategories :many
select * from catalog.categories where is_active order by parent_id nulls first, position, name;

-- name: CatalogGetCategoryBySlug :one
select * from catalog.categories where slug = $1;

-- name: CatalogCreateCategory :one
insert into catalog.categories (parent_id, slug, name, description, kind, repurchase_days, position)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (slug) do update set name = excluded.name, description = excluded.description, parent_id = excluded.parent_id
returning *;

-- name: CatalogListBrands :many
select * from catalog.brands order by name;

-- name: CatalogUpsertBrand :one
insert into catalog.brands (slug, name) values ($1, $2)
on conflict (slug) do update set name = excluded.name
returning *;

-- name: CatalogCreateProduct :one
insert into catalog.products (category_id, brand_id, slug, name, model, description, status, created_by)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: CatalogUpdateProduct :one
update catalog.products
set name = coalesce(sqlc.narg('name'), name), description = coalesce(sqlc.narg('description'), description),
    status = coalesce(sqlc.narg('status'), status), brand_id = coalesce(sqlc.narg('brand_id'), brand_id),
    category_id = coalesce(sqlc.narg('category_id'), category_id)
where id = $1
returning *;

-- name: CatalogGetProductBySlug :one
select p.*, c.slug as category_slug, c.name as category_name, b.name as brand_name
from catalog.products p
join catalog.categories c on c.id = p.category_id
left join catalog.brands b on b.id = p.brand_id
where p.slug = $1;

-- name: CatalogGetProduct :one
select * from catalog.products where id = $1;

-- name: CatalogGetProductBySlugForUpdate :one
select * from catalog.products where slug = $1;

-- name: CatalogListVariants :many
select * from catalog.product_variants where product_id = $1 and is_active order by name;

-- name: CatalogCreateVariant :one
insert into catalog.product_variants (product_id, sku, name, axis_values, gtin, weight_grams, is_serialised)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (sku) do update set name = excluded.name
returning *;

-- name: CatalogGetVariant :one
select v.*, p.name as product_name, p.slug as product_slug, p.category_id
from catalog.product_variants v join catalog.products p on p.id = v.product_id
where v.id = $1;

-- name: CatalogListImages :many
select i.*, f.storage_key, f.visibility from catalog.product_images i join platform.files f on f.id = i.file_id
where i.product_id = $1 order by i.position;

-- name: CatalogAddImage :one
insert into catalog.product_images (product_id, variant_id, file_id, alt, position) values ($1, $2, $3, $4, $5) returning *;

-- name: CatalogListAttributeValues :many
select d.key, d.label, d.unit, d.is_filterable, v.value_text, v.value_number, v.value_bool
from catalog.product_attribute_values v
join catalog.attribute_definitions d on d.id = v.attribute_id
where v.product_id = $1 order by d.position;

-- name: CatalogUpsertAttributeDefinition :one
insert into catalog.attribute_definitions (category_id, key, label, data_type, unit, options, is_required, is_filterable, is_variant_axis, position)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
on conflict (category_id, key) do update set label = excluded.label, is_filterable = excluded.is_filterable
returning *;

-- name: CatalogListAttributeDefinitions :many
select * from catalog.attribute_definitions where category_id = $1 order by position;

-- name: CatalogSetAttributeValue :exec
insert into catalog.product_attribute_values (product_id, attribute_id, value_text, value_number, value_bool)
values ($1, $2, $3, $4, $5)
on conflict (product_id, attribute_id) do update set value_text = excluded.value_text, value_number = excluded.value_number, value_bool = excluded.value_bool;

-- name: CatalogListOffers :many
-- Every sellable offer for a product (buy box candidates).
select l.*, v.name as variant_name, v.sku, s.display_name as seller_name, s.type as seller_type, s.state_code as seller_state,
       s.rating_avg as seller_rating, s.slug as seller_slug
from catalog.listings l
join catalog.product_variants v on v.id = l.variant_id
join sellers.sellers s on s.id = l.seller_id
where v.product_id = $1 and l.status = 'active' and s.status = 'active'
order by l.price_kobo;

-- name: CatalogGetListing :one
select l.*, v.product_id, v.name as variant_name, v.sku, v.weight_grams, v.is_serialised, p.name as product_name, p.slug as product_slug, p.category_id,
       s.type as seller_type, s.status as seller_status, s.display_name as seller_name, s.state_code as seller_state
from catalog.listings l
join catalog.product_variants v on v.id = l.variant_id
join catalog.products p on p.id = v.product_id
join sellers.sellers s on s.id = l.seller_id
where l.id = $1;

-- name: CatalogCreateListing :one
insert into catalog.listings (seller_id, variant_id, condition, grade, price_kobo, compare_at_kobo, fulfilled_by, seller_stock,
  min_order_quantity, handling_days, warranty_provider, warranty_months, condition_notes, battery_health_pct, status, risk_score, published_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
returning *;

-- name: CatalogUpdateListingPrice :one
update catalog.listings set price_kobo = $2, compare_at_kobo = $3 where id = $1 returning *;

-- name: CatalogSetListingStock :one
update catalog.listings set seller_stock = $2 where id = $1 returning *;

-- name: CatalogSetListingStatus :one
update catalog.listings
set status = $2, risk_score = coalesce(sqlc.narg('risk_score'), risk_score),
    published_at = case when $2 = 'active' then coalesce(published_at, now()) else published_at end
where id = $1
returning *;

-- name: CatalogListSellerListings :many
select l.*, v.name as variant_name, v.sku, p.name as product_name, p.slug as product_slug
from catalog.listings l
join catalog.product_variants v on v.id = l.variant_id
join catalog.products p on p.id = v.product_id
where l.seller_id = $1
order by l.created_at desc
limit $2;

-- name: CatalogInsertPriceHistory :exec
insert into catalog.listing_price_history (listing_id, price_kobo, compare_at_kobo, changed_by) values ($1, $2, $3, $4);

-- name: CatalogMaxPrice30d :one
-- The highest price this listing sold at in the last 30 days (FCCPA guard for "was" prices).
select coalesce(max(price_kobo), 0)::bigint from catalog.listing_price_history
where listing_id = $1 and changed_at >= now() - interval '30 days';

-- name: CatalogMedianPrice :one
-- Median active price for a variant and condition (price-outlier check).
select coalesce(percentile_cont(0.5) within group (order by price_kobo), 0)::bigint
from catalog.listings where variant_id = $1 and condition = $2 and status = 'active';

-- name: CatalogInsertListingReview :exec
insert into catalog.listing_reviews (listing_id, reviewer_id, decision, reasons, note) values ($1, $2, $3, $4, $5);

-- name: CatalogListListingReviews :many
select * from catalog.listing_reviews where listing_id = $1 order by created_at desc;

-- name: CatalogModerationQueue :many
select l.*, v.name as variant_name, p.name as product_name, p.slug as product_slug, s.display_name as seller_name,
       (select r.reasons from catalog.listing_reviews r where r.listing_id = l.id order by r.created_at desc limit 1)::text[] as reasons
from catalog.listings l
join catalog.product_variants v on v.id = l.variant_id
join catalog.products p on p.id = v.product_id
join sellers.sellers s on s.id = l.seller_id
where l.status = 'in_review'
order by l.risk_score desc nulls last, l.created_at
limit $1;

-- name: CatalogSellerListingCount :one
select count(*)::int from catalog.listings where seller_id = $1;

-- name: CatalogCreateSuggestion :one
insert into catalog.product_suggestions (seller_id, product_id, kind, payload) values ($1, $2, $3, $4) returning *;

-- name: CatalogListSuggestions :many
select * from catalog.product_suggestions where status = 'pending' order by created_at limit $1;

-- name: CatalogDecideSuggestion :one
update catalog.product_suggestions set status = $2, reviewed_by = $3, reviewed_at = now() where id = $1 and status = 'pending' returning *;

-- name: CatalogDeleteSaved :execrows
delete from catalog.saved_items where user_id = $1 and product_id = $2;

-- name: CatalogInsertSaved :exec
insert into catalog.saved_items (user_id, product_id) values ($1, $2) on conflict do nothing;

-- name: CatalogListSaved :many
select p.id, p.slug, p.name, s.created_at as saved_at
from catalog.saved_items s join catalog.products p on p.id = s.product_id
where s.user_id = $1 order by s.created_at desc;

-- name: CatalogAddStockAlert :exec
insert into catalog.stock_alerts (user_id, product_id) values ($1, $2) on conflict do nothing;

-- name: CatalogDueStockAlerts :many
select a.user_id, a.product_id, p.name as product_name
from catalog.stock_alerts a
join catalog.products p on p.id = a.product_id
join catalog.product_offer_summary s on s.product_id = a.product_id and s.channel = 'retail'
where a.notified_at is null and s.in_stock
limit 200;

-- name: CatalogMarkStockAlertNotified :exec
update catalog.stock_alerts set notified_at = now() where user_id = $1 and product_id = $2;

-- name: CatalogListCompatible :many
select p.id, p.slug, p.name from catalog.product_compatibility c
join catalog.products p on p.id = c.accessory_product_id
where c.device_product_id = $1 and p.status = 'active'
limit 12;

-- name: CatalogAddCompatibility :exec
insert into catalog.product_compatibility (accessory_product_id, device_product_id, source, verified) values ($1, $2, $3, $4);

-- Read model ------------------------------------------------------------------

-- name: CatalogRefreshOfferSummary :exec
-- Rebuilds the search/browse row for one product and channel from listings, stock, ratings and
-- promotions. Called by the outbox subscriber within seconds of a change (and nightly).
with recursive up as (
  select c.id, c.parent_id, c.name from catalog.categories c join catalog.products p on p.category_id = c.id where p.id = @product_id
  union all
  select c.id, c.parent_id, c.name from catalog.categories c join up on up.parent_id = c.id
), offers as (
  select l.*, s.type as seller_type,
         coalesce((select sum(il.on_hand - il.reserved) from inventory.inventory_levels il
                   where il.variant_id = l.variant_id and il.condition = l.condition and il.owner_seller_id = l.seller_id), 0) as wh_stock
  from catalog.listings l
  join catalog.product_variants v on v.id = l.variant_id
  join sellers.sellers s on s.id = l.seller_id
  where v.product_id = @product_id and l.status = 'active' and s.status = 'active'
), stock as (
  select o.*, case when o.fulfilled_by = 'seller' then coalesce(o.seller_stock, 0) else o.wh_stock end as available
  from offers o
), promo as (
  select min(plp.price_kobo) as promo_price from marketing.promotion_listing_prices plp
  join marketing.promotions pr on pr.id = plp.promotion_id
  where pr.status = 'live' and now() between pr.starts_at and pr.ends_at and plp.listing_id in (select id from offers)
), best as (
  select id from stock order by (available > 0) desc, price_kobo limit 1
), cats as (
  select array_agg(id) as ids, string_agg(name, ' ') as names from up
), attrs as (
  select coalesce(jsonb_object_agg(d.key, coalesce(to_jsonb(v.value_number), to_jsonb(v.value_text), to_jsonb(v.value_bool))) filter (where d.is_filterable), '{}'::jsonb) as filterable,
         string_agg(coalesce(v.value_text, v.value_number::text, ''), ' ') as all_text
  from catalog.product_attribute_values v join catalog.attribute_definitions d on d.id = v.attribute_id
  where v.product_id = @product_id
), ratings as (
  select avg(rating)::numeric(3, 2) as avg, count(*)::int as cnt from catalog.product_reviews where product_id = @product_id and status = 'published'
), sales as (
  select coalesce(sum(ol.quantity), 0)::int as qty from sales.order_lines ol
  join sales.orders o on o.id = ol.order_id
  join catalog.product_variants v on v.id = ol.variant_id
  where v.product_id = @product_id and o.status not in ('pending_payment', 'cancelled') and o.placed_at >= now() - interval '30 days'
)
insert into catalog.product_offer_summary (product_id, channel, slug, title, category_ids, brand_id, best_listing_id, min_price_kobo, max_price_kobo,
  compare_at_kobo, promo_price_kobo, discount_pct, offer_count, conditions, seller_types, in_stock, stock_band, rating_avg, rating_count, sales_30d,
  attrs, facet_keys, search_vector, refreshed_at)
select p.id, @channel::text, p.slug, p.name, coalesce(cats.ids, '{}'), p.brand_id, (select id from best),
  (select min(price_kobo) from stock), (select max(price_kobo) from stock),
  (select max(compare_at_kobo) from stock where id = (select id from best)),
  (select promo_price from promo),
  (select case when max(compare_at_kobo) > 0 then round(100 - 100.0 * min(price_kobo) / max(compare_at_kobo))::smallint end from stock where compare_at_kobo is not null),
  (select count(*) from stock)::int,
  coalesce((select array_agg(distinct condition) from stock), '{}'),
  coalesce((select array_agg(distinct seller_type) from stock), '{}'),
  coalesce((select bool_or(available > 0) from stock), false),
  case when coalesce((select sum(available) from stock), 0) <= 0 then 'out' when (select sum(available) from stock) <= 3 then 'low' else 'in' end,
  ratings.avg, coalesce(ratings.cnt, 0), sales.qty,
  attrs.filterable,
  coalesce((select array_agg(distinct k) from jsonb_object_keys(attrs.filterable) k), '{}'),
  setweight(to_tsvector('simple', coalesce(p.name, '') || ' ' || coalesce(p.model, '') || ' ' || coalesce(b.name, '')), 'A') ||
  setweight(to_tsvector('simple', coalesce(cats.names, '')), 'B') ||
  setweight(to_tsvector('simple', coalesce(attrs.all_text, '') || ' ' || coalesce(p.description, '')), 'C'),
  now()
from catalog.products p
left join catalog.brands b on b.id = p.brand_id
cross join cats cross join attrs cross join ratings cross join sales
where p.id = @product_id and p.status = 'active'
on conflict (product_id, channel) do update set
  slug = excluded.slug, title = excluded.title, category_ids = excluded.category_ids, brand_id = excluded.brand_id,
  best_listing_id = excluded.best_listing_id, min_price_kobo = excluded.min_price_kobo, max_price_kobo = excluded.max_price_kobo,
  compare_at_kobo = excluded.compare_at_kobo, promo_price_kobo = excluded.promo_price_kobo, discount_pct = excluded.discount_pct,
  offer_count = excluded.offer_count, conditions = excluded.conditions, seller_types = excluded.seller_types, in_stock = excluded.in_stock,
  stock_band = excluded.stock_band, rating_avg = excluded.rating_avg, rating_count = excluded.rating_count, sales_30d = excluded.sales_30d,
  attrs = excluded.attrs, facet_keys = excluded.facet_keys, search_vector = excluded.search_vector, refreshed_at = now();

-- name: CatalogDeleteOfferSummary :exec
delete from catalog.product_offer_summary where product_id = $1;

-- name: CatalogAllProductIDs :many
select id from catalog.products where status = 'active';

-- name: CatalogProductIDForListing :one
select v.product_id from catalog.listings l join catalog.product_variants v on v.id = l.variant_id where l.id = $1;

-- name: CatalogProductIDForVariant :one
select product_id from catalog.product_variants where id = $1;

-- name: CatalogSummaryByProduct :one
select * from catalog.product_offer_summary where product_id = $1 and channel = $2;

-- name: CatalogSuggestProducts :many
select slug, title, min_price_kobo from catalog.product_offer_summary
where channel = 'retail' and (title ilike @prefix || '%' or title ilike '% ' || @prefix || '%')
order by sales_30d desc limit 4;

-- name: CatalogSuggestQueries :many
select suggestion from search.search_suggestions where prefix = $1 order by weight desc limit 4;

-- name: CatalogSuggestCategories :many
select slug, name from catalog.categories where is_active and name ilike $1 || '%' order by position limit 3;

-- name: CatalogInsertTier :exec
insert into catalog.price_tiers (listing_id, min_quantity, unit_price_kobo) values ($1, $2, $3)
on conflict (listing_id, min_quantity) do update set unit_price_kobo = excluded.unit_price_kobo;

-- name: CatalogListTiers :many
select * from catalog.price_tiers where listing_id = $1 order by min_quantity;
