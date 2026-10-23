-- Trade-ins: catalogue estimate, drop-off, inspection, offer, payout. Plan: docs/plan.md 2026-10-23.

-- name: TradeInDevices :many
-- Catalogue variants we can price, with today's lowest new price on a live listing.
select v.id as variant_id, p.name as product_name, v.name as variant_name, coalesce(b.name, '') as brand,
       min(l.price_kobo)::bigint as new_price_kobo
from catalog.product_variants v
join catalog.products p on p.id = v.product_id
join catalog.categories c on c.id = p.category_id
left join catalog.categories pc on pc.id = c.parent_id
left join catalog.brands b on b.id = p.brand_id
join catalog.listings l on l.variant_id = v.id and l.status = 'active' and l.condition = 'new'
where p.status = 'active' and (c.slug = any(@slugs::text[]) or pc.slug = any(@slugs::text[]))
group by v.id, p.name, v.name, b.name
order by b.name, p.name, v.name
limit 300;

-- name: TradeInVariantPrice :one
-- One variant's lowest new price (for the estimate) and its category.
select v.id as variant_id, p.name as product_name, v.name as variant_name, coalesce(b.name, '') as brand, c.slug as category_slug,
       coalesce(pc.slug, '') as parent_category_slug, v.is_serialised,
       (select min(l.price_kobo) from catalog.listings l where l.variant_id = v.id and l.status = 'active' and l.condition = 'new')::bigint as new_price_kobo
from catalog.product_variants v
join catalog.products p on p.id = v.product_id
join catalog.categories c on c.id = p.category_id
left join catalog.categories pc on pc.id = c.parent_id
left join catalog.brands b on b.id = p.brand_id
where v.id = $1;

-- name: TradeInCreate :one
insert into aftersales.trade_ins (user_id, device_type, brand, model, storage, declared_condition, imei, variant_id, estimate_kobo, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
returning *;

-- name: TradeInsForUser :many
select * from aftersales.trade_ins where user_id = $1 order by created_at desc limit 100;

-- name: TradeIn :one
select t.*, st.name as store_name, u.first_name, u.last_name
from aftersales.trade_ins t
left join inventory.warehouses st on st.id = t.store_id
join identity.users u on u.id = t.user_id
where t.reference = $1;

-- name: TradeInForUpdate :one
select * from aftersales.trade_ins where reference = $1 for update;

-- name: TradeInStaffList :many
select t.*, st.name as store_name, u.first_name, u.last_name
from aftersales.trade_ins t
left join inventory.warehouses st on st.id = t.store_id
join identity.users u on u.id = t.user_id
where (sqlc.narg('status')::text is null or t.status = sqlc.narg('status')
       or (sqlc.narg('status') = 'active' and t.status in ('submitted', 'estimated', 'awaiting_inspection', 'inspected', 'offer_sent', 'accepted'))
       or (sqlc.narg('status') = 'payout_due' and t.status = 'accepted' and t.payout_method = 'bank'))
  and (sqlc.narg('q')::text is null or t.reference ilike sqlc.narg('q') || '%' or t.imei = sqlc.narg('q'))
order by (t.status in ('submitted', 'awaiting_inspection', 'accepted')) desc, t.created_at
limit 200;

-- name: TradeInSummary :one
select
  count(*) filter (where status = 'submitted')::int as needs_estimate,
  count(*) filter (where status = 'awaiting_inspection')::int as awaiting_inspection,
  count(*) filter (where status = 'offer_sent')::int as offers_out,
  count(*) filter (where status = 'accepted' and payout_method = 'bank')::int as payouts_due,
  count(*) filter (where status = 'paid' and paid_at >= date_trunc('month', now()))::int as paid_this_month,
  coalesce(avg(offer_kobo) filter (where status = 'paid' and paid_at > now() - interval '90 days'), 0)::bigint as avg_paid_kobo
from aftersales.trade_ins;

-- name: TradeInRecent :one
-- Trade-ins by the same person in the last 7 days (a velocity feature).
select count(*)::int from aftersales.trade_ins where user_id = $1 and created_at > now() - interval '7 days';

-- name: TradeInAccountAgeDays :one
select (extract(epoch from now() - created_at) / 86400)::float8 from identity.users where id = $1;

-- name: TradeInSetEstimate :one
update aftersales.trade_ins set estimate_kobo = $2, variant_id = coalesce(sqlc.narg('variant_id'), variant_id), status = 'estimated'
where id = $1 returning *;

-- name: TradeInReceive :one
update aftersales.trade_ins
set status = 'awaiting_inspection', store_id = $2, received_by = $3, received_at = now(),
    imei = coalesce(sqlc.narg('imei'), imei), serial_number = coalesce(sqlc.narg('serial'), serial_number)
where id = $1 returning *;

-- name: TradeInBlock :one
update aftersales.trade_ins set status = 'blocked', inspection_notes = $2,
  imei = coalesce(sqlc.narg('imei'), imei), serial_number = coalesce(sqlc.narg('serial'), serial_number)
where id = $1 returning *;

-- name: TradeInOffer :one
update aftersales.trade_ins
set status = 'offer_sent', inspected_condition = $2, offer_kobo = $3, inspection_notes = $4, inspected_by = $5, inspected_at = now(),
    offer_expires_at = $6, variant_id = coalesce(sqlc.narg('variant_id'), variant_id)
where id = $1 returning *;

-- name: TradeInDecide :one
update aftersales.trade_ins set status = $2, decided_at = now(), payout_method = $3, bank_code = $4,
  account_number_encrypted = $5, account_number_last4 = $6, encryption_key_id = $7, account_name = $8
where id = $1 returning *;

-- name: TradeInSetStatus :one
update aftersales.trade_ins set status = $2 where id = $1 returning *;

-- name: TradeInPaid :one
update aftersales.trade_ins set status = 'paid', paid_at = now(), paid_by = $2, payout_reference = $3 where id = $1 returning *;

-- name: TradeInLinkUnit :exec
update aftersales.trade_ins set device_unit_id = $2, repair_job_id = $3 where id = $1;

-- name: TradeInExpireOffers :many
-- Offers not answered within 7 days lapse: the device waits at the store for collection.
update aftersales.trade_ins set status = 'declined', decided_at = now(),
  inspection_notes = coalesce(inspection_notes || E'\n', '') || 'Offer expired without an answer.'
where status = 'offer_sent' and offer_expires_at < now()
returning id, reference, user_id;

-- name: TradeInFirstPartySeller :one
select id from sellers.sellers where type = 'first_party' order by created_at limit 1;

-- name: TradeInCreateUnit :one
insert into inventory.device_units (variant_id, condition, imei, serial_number, warehouse_id, owner_seller_id, cost_kobo, status, acquired_via)
values ($1, 'refurbished', $2, $3, $4, $5, $6, 'in_repair', 'trade_in')
returning *;

-- name: TradeInCreateRefurb :one
insert into aftersales.repair_jobs (device_unit_id, purpose, diagnosis) values ($1, 'refurbish', $2) returning *;
