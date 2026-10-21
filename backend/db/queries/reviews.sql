-- Product reviews, reports and seller ratings (M4 slice 4). Plan: trust-safety.md §6.

-- name: ReviewsDeliveredLine :one
-- The caller's delivered, not-cancelled order line for this product (a verified purchase).
select ol.id, f.seller_id
from sales.order_lines ol
join sales.orders o on o.id = ol.order_id
join sales.fulfilments f on f.id = ol.fulfilment_id
join catalog.product_variants v on v.id = ol.variant_id
where ol.id = @line_id and o.customer_user_id = @user_id and v.product_id = @product_id
  and f.status = 'delivered' and ol.cancelled_at is null;

-- name: ReviewsByUserLast24h :one
select count(*)::int from catalog.product_reviews where user_id = $1 and created_at > now() - interval '24 hours';

-- name: ReviewsSameTextByOthers :one
-- Another account already posted this text (normalised): a sign of copied or bought reviews.
select exists (
  select 1 from catalog.product_reviews
  where user_id <> @user_id and body is not null and lower(regexp_replace(trim(body), '\s+', ' ', 'g')) = @normalised::text
);

-- name: ReviewsInsert :one
insert into catalog.product_reviews (product_id, order_line_id, user_id, rating, title, body, status, moderation_reasons)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: ReviewsInsertImage :exec
insert into catalog.review_images (review_id, file_id, position) values ($1, $2, $3);

-- name: ReviewsPublished :many
select r.id, r.rating, r.title, r.body, r.is_verified, r.seller_response, r.seller_response_at, r.created_at,
       u.first_name, left(u.last_name, 1) as last_initial,
       coalesce((select array_agg(i.file_id order by i.position) from catalog.review_images i where i.review_id = r.id), '{}')::uuid[] as image_ids
from catalog.product_reviews r join identity.users u on u.id = r.user_id
where r.product_id = @product_id and r.status = 'published'
  and (sqlc.narg('verified')::boolean is null or r.is_verified = sqlc.narg('verified'))
  and (sqlc.narg('before')::timestamptz is null or r.created_at < sqlc.narg('before'))
order by r.is_verified desc, r.created_at desc limit @lim;

-- name: ReviewsSummary :one
-- Star counts and the weighted average (verified reviews count 1, unverified 0.5).
select count(*)::int as total,
       count(*) filter (where is_verified)::int as verified,
       count(*) filter (where rating = 5)::int as five, count(*) filter (where rating = 4)::int as four,
       count(*) filter (where rating = 3)::int as three, count(*) filter (where rating = 2)::int as two,
       count(*) filter (where rating = 1)::int as one,
       coalesce(round(sum(rating * case when is_verified then 1.0 else 0.5 end) / nullif(sum(case when is_verified then 1.0 else 0.5 end), 0), 2), 0)::float8 as weighted_avg
from catalog.product_reviews where product_id = $1 and status = 'published';

-- name: ReviewsGet :one
select * from catalog.product_reviews where id = $1;

-- name: ReviewsGetForUpdate :one
select * from catalog.product_reviews where id = $1 for update;

-- name: ReviewsSetStatus :one
update catalog.product_reviews set status = $2 where id = $1 returning *;

-- name: ReviewsQueue :many
-- Staff moderation: held reviews, oldest first, and published ones with open reports.
select r.*, p.name as product_name, p.slug as product_slug, u.first_name,
       (select count(*) from risk.review_reports rr where rr.review_id = r.id and rr.status = 'open')::int as open_reports
from catalog.product_reviews r
join catalog.products p on p.id = r.product_id
join identity.users u on u.id = r.user_id
where r.status = 'pending' or exists (select 1 from risk.review_reports rr where rr.review_id = r.id and rr.status = 'open')
order by r.created_at limit $1;

-- name: ReviewsInsertReport :one
insert into risk.review_reports (review_id, reporter_id, reason) values ($1, $2, $3) returning *;

-- name: ReviewsReportedBy :one
select exists (select 1 from risk.review_reports where review_id = $1 and reporter_id = $2 and status = 'open');

-- name: ReviewsDecideReports :exec
update risk.review_reports set status = $2, decided_by = $3 where review_id = $1 and status = 'open';

-- name: ReviewsForSeller :many
-- Published reviews of items this seller shipped (the reviews it may answer).
select r.id, r.product_id, r.rating, r.title, r.body, r.is_verified, r.seller_response, r.seller_response_at, r.created_at,
       p.name as product_name, p.slug as product_slug, u.first_name
from catalog.product_reviews r
join sales.order_lines ol on ol.id = r.order_line_id
join sales.fulfilments f on f.id = ol.fulfilment_id
join catalog.products p on p.id = r.product_id
join identity.users u on u.id = r.user_id
where f.seller_id = @seller_id and r.status = 'published'
  and (sqlc.narg('before')::timestamptz is null or r.created_at < sqlc.narg('before'))
order by r.created_at desc limit @lim;

-- name: ReviewsSellerOf :one
-- The seller that shipped the reviewed item (null for unverified reviews).
select f.seller_id from catalog.product_reviews r
join sales.order_lines ol on ol.id = r.order_line_id
join sales.fulfilments f on f.id = ol.fulfilment_id
where r.id = $1;

-- name: ReviewsSetSellerResponse :execrows
update catalog.product_reviews set seller_response = $2, seller_response_at = now()
where id = $1 and seller_response is null;

-- name: RatingsDeliveredFulfilment :one
-- A delivered fulfilment of the caller's own order (the only thing a customer may rate).
select f.id, f.seller_id from sales.fulfilments f join sales.orders o on o.id = f.order_id
where f.id = @fulfilment_id and o.customer_user_id = @user_id and f.status = 'delivered';

-- name: RatingsInsert :one
insert into catalog.seller_ratings (seller_id, fulfilment_id, user_id, rating, comment) values ($1, $2, $3, $4, $5)
returning *;

-- name: RatingsRecomputeSeller :exec
update sellers.sellers s
set rating_avg = (select round(avg(rating), 2) from catalog.seller_ratings where seller_id = s.id),
    rating_count = (select count(*) from catalog.seller_ratings where seller_id = s.id)
where s.id = $1;

-- name: ReviewsLineProductSlug :one
-- The product slug of one of the caller's order lines (to review it from the order page).
select p.slug from sales.order_lines ol
join sales.orders o on o.id = ol.order_id
join catalog.product_variants v on v.id = ol.variant_id
join catalog.products p on p.id = v.product_id
where ol.id = @line_id and o.customer_user_id = @user_id;
