-- Warranty claims and repair jobs. Plan: orders-fulfilment.md §7, docs/plan.md 2026-10-23.

-- name: WarrantyCoveredLines :many
-- The caller's delivered items still under warranty, with any open claim.
select ol.id, o.order_number, ol.product_name, ol.variant_name, ol.quantity, ol.warranty_provider, ol.warranty_months,
       f.delivered_at, (f.delivered_at + make_interval(months => ol.warranty_months::int))::date as ends_on,
       du.imei, du.serial_number,
       coalesce((select wc.claim_number from aftersales.warranty_claims wc where wc.order_line_id = ol.id
          and wc.status in ('new', 'received', 'in_repair', 'ready') order by wc.created_at desc limit 1), '')::text as open_claim
from sales.order_lines ol
join sales.orders o on o.id = ol.order_id
join sales.fulfilments f on f.id = ol.fulfilment_id
left join inventory.device_units du on du.id = ol.device_unit_id
where o.customer_user_id = $1 and f.status = 'delivered' and ol.cancelled_at is null
  and coalesce(ol.warranty_months, 0) > 0 and coalesce(ol.warranty_provider, 'none') <> 'none'
  and f.delivered_at + make_interval(months => ol.warranty_months::int) >= now()
order by f.delivered_at desc
limit 100;

-- name: WarrantyLine :one
-- One order line with what a claim needs.
select ol.id, ol.order_id, o.order_number, o.customer_user_id, ol.variant_id, ol.seller_id, ol.product_name, ol.variant_name, ol.quantity,
       ol.line_total_kobo, ol.device_unit_id, ol.warranty_provider, ol.warranty_months, ol.cancelled_at,
       f.status as fulfilment_status, f.delivered_at, du.imei, du.serial_number
from sales.order_lines ol
join sales.orders o on o.id = ol.order_id
join sales.fulfilments f on f.id = ol.fulfilment_id
left join inventory.device_units du on du.id = ol.device_unit_id
where ol.id = $1;

-- name: WarrantySoldUnit :one
-- The order line that sold a device, found by its IMEI or serial (walk-in claims).
select ol.id from inventory.device_units du
join sales.order_lines ol on ol.device_unit_id = du.id
where du.imei = @serial::text or upper(du.serial_number) = upper(@serial::text)
order by ol.created_at desc limit 1;

-- name: WarrantyCreateClaim :one
insert into aftersales.warranty_claims (customer_user_id, order_line_id, device_unit_id, imei_or_serial, fault_description, warranty_provider, warranty_ends_on)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: WarrantyClaimsForUser :many
select wc.*, ol.product_name, ol.variant_name, o.order_number
from aftersales.warranty_claims wc
left join sales.order_lines ol on ol.id = wc.order_line_id
left join sales.orders o on o.id = ol.order_id
where wc.customer_user_id = $1
order by wc.created_at desc limit 100;

-- name: WarrantyClaim :one
select wc.*, ol.product_name, ol.variant_name, ol.variant_id, ol.line_total_kobo, ol.quantity as line_quantity, ol.order_id, o.order_number,
       st.name as store_name, u.first_name as customer_first_name, u.last_name as customer_last_name
from aftersales.warranty_claims wc
left join sales.order_lines ol on ol.id = wc.order_line_id
left join sales.orders o on o.id = ol.order_id
left join inventory.warehouses st on st.id = wc.store_id
join identity.users u on u.id = wc.customer_user_id
where wc.claim_number = $1;

-- name: WarrantyClaimForUpdate :one
select * from aftersales.warranty_claims where claim_number = $1 for update;

-- name: WarrantyStaffList :many
select wc.*, ol.product_name, ol.variant_name, o.order_number, st.name as store_name,
       u.first_name as customer_first_name, u.last_name as customer_last_name
from aftersales.warranty_claims wc
left join sales.order_lines ol on ol.id = wc.order_line_id
left join sales.orders o on o.id = ol.order_id
left join inventory.warehouses st on st.id = wc.store_id
join identity.users u on u.id = wc.customer_user_id
where (sqlc.narg('status')::text is null or wc.status = sqlc.narg('status')
       or (sqlc.narg('status') = 'active' and wc.status in ('new', 'received', 'in_repair', 'ready')))
  and (sqlc.narg('q')::text is null or wc.claim_number ilike sqlc.narg('q') || '%' or wc.imei_or_serial = sqlc.narg('q') or o.order_number ilike sqlc.narg('q') || '%')
order by (wc.status in ('new', 'received', 'in_repair', 'ready')) desc, wc.created_at
limit 200;

-- name: WarrantySummary :one
select
  count(*) filter (where status in ('new', 'received', 'in_repair', 'ready'))::int as open,
  count(*) filter (where status = 'new')::int as awaiting_dropoff,
  count(*) filter (where status = 'in_repair')::int as in_repair,
  count(*) filter (where status = 'ready')::int as ready,
  count(*) filter (where resolution = 'replaced' and updated_at >= date_trunc('month', now()))::int as replaced_this_month,
  coalesce(extract(epoch from avg(updated_at - received_at) filter (where status in ('ready', 'collected', 'closed')
    and received_at > now() - interval '30 days')) / 86400, 0)::float8 as avg_turnaround_days
from aftersales.warranty_claims;

-- name: WarrantyHistory :many
-- Earlier claims on the same device (staff see a device's repair history).
select wc.claim_number, wc.status, wc.resolution, wc.fault_description, wc.created_at
from aftersales.warranty_claims wc
where wc.imei_or_serial = $1 and wc.id <> $2
order by wc.created_at desc limit 20;

-- name: WarrantyReceive :one
update aftersales.warranty_claims set status = 'received', store_id = $2, received_by = $3, received_at = now()
where id = $1 returning *;

-- name: WarrantyReject :one
update aftersales.warranty_claims set status = 'rejected', resolution = 'not_covered', rejection_reason = $2
where id = $1 returning *;

-- name: WarrantySetStatus :one
update aftersales.warranty_claims set status = $2 where id = $1 returning *;

-- name: WarrantyReady :one
update aftersales.warranty_claims set status = 'ready', resolution = $2, resolution_note = $3, replacement_unit_id = $4
where id = $1 returning *;

-- name: WarrantyCollected :one
update aftersales.warranty_claims set status = 'collected' where id = $1 and status = 'ready' returning *;

-- name: WarrantyClose :one
update aftersales.warranty_claims set status = 'closed', resolution = 'refunded', resolution_note = $2, refund_id = $3, closed_at = now()
where id = $1 returning *;

-- name: WarrantyCreateRepair :one
insert into aftersales.repair_jobs (warranty_claim_id, device_unit_id, technician_id, diagnosis)
values ($1, $2, $3, $4)
returning *;

-- name: WarrantyRepairs :many
select r.*, u.first_name as technician_first_name, u.last_name as technician_last_name
from aftersales.repair_jobs r left join identity.users u on u.id = r.technician_id
where r.warranty_claim_id = $1 order by r.created_at;

-- name: WarrantyRepairForUpdate :one
select r.*, wc.claim_number from aftersales.repair_jobs r join aftersales.warranty_claims wc on wc.id = r.warranty_claim_id
where r.id = $1 for update of r;

-- name: WarrantyUpdateRepair :one
update aftersales.repair_jobs
set status = @status::text,
    diagnosis = coalesce(sqlc.narg('diagnosis'), diagnosis),
    notes = coalesce(sqlc.narg('notes'), notes),
    technician_id = coalesce(sqlc.narg('technician_id'), technician_id),
    started_at = case when @status::text in ('diagnosing', 'awaiting_parts', 'repairing') then coalesce(started_at, now()) else started_at end,
    finished_at = case when @status::text in ('done', 'unrepairable') then now() else null end
where id = @id
returning *;

-- name: WarrantyOpenRepairs :one
select count(*)::int from aftersales.repair_jobs where warranty_claim_id = $1 and status not in ('done', 'unrepairable');

-- name: WarrantyReplacementUnit :one
-- An in-stock unit of the same variant to swap in, found by its IMEI or serial.
select * from inventory.device_units
where (imei = @serial::text or upper(serial_number) = upper(@serial::text)) and variant_id = @variant_id and status = 'in_stock'
for update;

-- name: WarrantySetUnit :exec
update inventory.device_units set status = $2, warehouse_id = $3 where id = $1;

-- name: WarrantyStores :many
-- Where customers can bring a device: TechShop stores and warehouses (not hubs).
select id, name, address, city, state_code from inventory.warehouses where kind in ('store', 'warehouse') and is_active order by kind desc, name;
