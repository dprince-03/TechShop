-- After-sales: returns (RMA). Plan: orders-fulfilment.md §7.

-- name: AftersalesReturnableLines :many
-- The order's lines with what a return needs: delivery state, who shipped it, the unit sold and
-- how many are already in an open (not rejected) return.
select ol.id, ol.order_id, ol.fulfilment_id, ol.listing_id, ol.variant_id, ol.product_name, ol.variant_name, ol.condition,
       ol.quantity, ol.unit_price_kobo, ol.line_total_kobo, ol.device_unit_id, ol.cancelled_at,
       f.status as fulfilment_status, f.fulfilled_by, f.delivered_at, f.seller_id,
       coalesce((select sum(ri.quantity) from aftersales.return_items ri
                 join aftersales.return_requests rr on rr.id = ri.return_request_id
                 where ri.order_line_id = ol.id and rr.status <> 'rejected'), 0)::int as already_returned
from sales.order_lines ol join sales.fulfilments f on f.id = ol.fulfilment_id
where ol.order_id = $1
order by ol.created_at;

-- name: AftersalesCreateReturn :one
insert into aftersales.return_requests (order_id, requested_by, reason, details, return_method, status, auto_approved)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: AftersalesAddItem :exec
insert into aftersales.return_items (return_request_id, order_line_id, quantity) values ($1, $2, $3);

-- name: AftersalesReturnsForUser :many
select r.*, o.order_number from aftersales.return_requests r join sales.orders o on o.id = r.order_id
where r.requested_by = $1 order by r.created_at desc limit 50;

-- name: AftersalesGetReturn :one
select r.*, o.order_number, o.customer_user_id from aftersales.return_requests r join sales.orders o on o.id = r.order_id
where r.rma_number = $1;

-- name: AftersalesGetReturnForUpdate :one
select * from aftersales.return_requests where rma_number = $1 for update;

-- name: AftersalesItems :many
select ri.id, ri.order_line_id, ri.quantity, ol.product_name, ol.variant_name, ol.condition, ol.variant_id, ol.listing_id,
       ol.unit_price_kobo, ol.line_total_kobo, ol.quantity as line_quantity, ol.device_unit_id, f.seller_id,
       ins.grade, ins.disposition, ins.imei_match, ins.inspected_at
from aftersales.return_items ri
join sales.order_lines ol on ol.id = ri.order_line_id
join sales.fulfilments f on f.id = ol.fulfilment_id
left join aftersales.return_inspections ins on ins.return_item_id = ri.id
where ri.return_request_id = $1
order by ol.created_at;

-- name: AftersalesStaffList :many
select r.*, o.order_number from aftersales.return_requests r join sales.orders o on o.id = r.order_id
where (sqlc.narg('status')::text is null or r.status = sqlc.narg('status'))
  and (sqlc.narg('before')::timestamptz is null or r.created_at < sqlc.narg('before'))
order by r.created_at desc limit @lim;

-- name: AftersalesDecide :one
update aftersales.return_requests set status = $2, decided_by = $3, rejection_reason = $4 where id = $1 returning *;

-- name: AftersalesSetStatus :one
update aftersales.return_requests set status = $2, resolution = coalesce(sqlc.narg('resolution'), resolution) where id = $1 returning *;

-- name: AftersalesInsertInspection :one
insert into aftersales.return_inspections (return_item_id, device_unit_id, scanned_serial, imei_match, grade, disposition, notes, inspected_by)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: AftersalesStanding :one
-- Returns vs delivered orders in the last 90 days, and open risk cases: "good standing" for auto-approval.
select
  (select count(*) from aftersales.return_requests where requested_by = @user_id and created_at > now() - interval '90 days')::int as returns_90d,
  (select count(*) from sales.orders where customer_user_id = @user_id and status = 'delivered' and placed_at > now() - interval '90 days')::int as orders_90d,
  (select count(*) from risk.risk_cases where subject_type = 'user' and subject_id = @user_id and status in ('open', 'in_review', 'escalated'))::int as open_cases;

-- name: AftersalesOpenRiskCase :one
insert into risk.risk_cases (kind, subject_type, subject_id, reason, score, money_at_risk_kobo)
values ($1, $2, $3, $4, $5, $6)
returning id;

-- name: AftersalesSetUnitCondition :exec
update inventory.device_units set condition = $2, status = $3, warehouse_id = $4 where id = $1;
