-- Inventory: the single stock code path, reservations, device units, transfers, counts, costs.
-- Plan: docs/orders-fulfilment.md §3, §5, §9.

-- name: InventoryListWarehouses :many
select * from inventory.warehouses order by code;

-- name: InventoryGetWarehouse :one
select * from inventory.warehouses where id = $1;

-- name: InventoryGetWarehouseByCode :one
select * from inventory.warehouses where code = $1;

-- name: InventoryCreateWarehouse :one
insert into inventory.warehouses (code, name, kind, address, city, state_code, latitude, longitude)
values ($1, $2, $3, $4, $5, $6, $7, $8)
on conflict (code) do update set name = excluded.name, address = excluded.address
returning *;

-- name: InventoryApplyMovement :one
-- Changes on_hand (never below zero, never below reserved) and returns the new level.
insert into inventory.inventory_levels (warehouse_id, variant_id, condition, owner_seller_id, on_hand)
values (@warehouse_id, @variant_id, @condition, @owner_seller_id, greatest(@delta::int, 0))
on conflict (warehouse_id, variant_id, condition, owner_seller_id)
do update set on_hand = inventory.inventory_levels.on_hand + @delta::int, updated_at = now()
returning *;

-- name: InventoryInsertMovement :one
insert into inventory.stock_movements (warehouse_id, variant_id, condition, owner_seller_id, quantity_delta, reason, device_unit_id, reference_type, reference_id, created_by)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
returning *;

-- name: InventoryListLevels :many
select l.*, v.sku, v.name as variant_name, p.name as product_name, w.code as warehouse_code, (l.on_hand - l.reserved) as available
from inventory.inventory_levels l
join catalog.product_variants v on v.id = l.variant_id
join catalog.products p on p.id = v.product_id
join inventory.warehouses w on w.id = l.warehouse_id
where (sqlc.narg('warehouse_id')::uuid is null or l.warehouse_id = sqlc.narg('warehouse_id'))
  and (sqlc.narg('variant_id')::uuid is null or l.variant_id = sqlc.narg('variant_id'))
order by p.name, v.name
limit $1;

-- name: InventoryListMovements :many
select * from inventory.stock_movements
where (sqlc.narg('variant_id')::uuid is null or variant_id = sqlc.narg('variant_id'))
order by created_at desc limit $1;

-- name: InventoryReserveWarehouse :one
-- Race-free reservation: only succeeds if enough unreserved stock exists in one row.
update inventory.inventory_levels t
set reserved = t.reserved + @qty::int, updated_at = now()
where (t.warehouse_id, t.variant_id, t.condition, t.owner_seller_id) = (
  select il.warehouse_id, il.variant_id, il.condition, il.owner_seller_id from inventory.inventory_levels il
  where il.variant_id = @variant_id and il.condition = @condition and il.owner_seller_id = @owner_seller_id and il.on_hand - il.reserved >= @qty::int
  order by (il.warehouse_id = sqlc.narg('preferred_warehouse')::uuid) desc nulls last, il.on_hand - il.reserved desc
  limit 1
  for update)
  and t.on_hand - t.reserved >= @qty::int
returning t.warehouse_id;

-- name: InventoryReserveSellerStock :one
-- Seller-held stock: conditional decrement of the listing's stock counter.
update catalog.listings set seller_stock = seller_stock - @qty::int
where id = @listing_id and seller_stock >= @qty::int
returning seller_stock;

-- name: InventoryInsertReservation :one
insert into inventory.stock_reservations (order_id, order_line_id, warehouse_id, listing_id, variant_id, condition, owner_seller_id, quantity, expires_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *;

-- name: InventoryOrderReservations :many
select * from inventory.stock_reservations where order_id = $1 and status = 'reserved' for update;

-- name: InventorySetReservationStatus :exec
update inventory.stock_reservations set status = $2 where id = $1;

-- name: InventoryReleaseWarehouseReservation :exec
update inventory.inventory_levels set reserved = reserved - @qty::int, updated_at = now()
where warehouse_id = @warehouse_id and variant_id = @variant_id and condition = @condition and owner_seller_id = @owner_seller_id;

-- name: InventoryReleaseSellerStock :exec
update catalog.listings set seller_stock = seller_stock + @qty::int where id = @listing_id;

-- name: InventoryCommitWarehouseReservation :exec
-- On payment: the reserved units leave stock (on_hand and reserved both drop).
update inventory.inventory_levels set reserved = reserved - @qty::int, on_hand = on_hand - @qty::int, updated_at = now()
where warehouse_id = @warehouse_id and variant_id = @variant_id and condition = @condition and owner_seller_id = @owner_seller_id;

-- name: InventoryExpiredReservationOrders :many
select distinct r.order_id from inventory.stock_reservations r
where r.status = 'reserved' and r.expires_at < now() limit 100;

-- name: InventoryIsBlocklisted :one
select exists (select 1 from risk.device_blocklist where imei = $1) as blocked;

-- name: InventoryCreateDeviceUnit :one
insert into inventory.device_units (variant_id, condition, grade, imei, serial_number, warehouse_id, owner_seller_id, cost_kobo, status, acquired_via)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
returning *;

-- name: InventoryGetDeviceUnitBySerial :one
select * from inventory.device_units where imei = $1 or serial_number = $1 limit 1;

-- name: InventoryGetDeviceUnit :one
select * from inventory.device_units where id = $1;

-- name: InventorySetDeviceUnitStatus :one
update inventory.device_units set status = $2, warehouse_id = coalesce(sqlc.narg('warehouse_id'), warehouse_id) where id = $1 returning *;

-- name: InventoryListDeviceUnits :many
select * from inventory.device_units
where (sqlc.narg('variant_id')::uuid is null or variant_id = sqlc.narg('variant_id'))
  and (sqlc.narg('status')::text is null or status = sqlc.narg('status'))
order by created_at desc limit $1;

-- name: InventoryCreateTransfer :one
insert into inventory.stock_transfers (from_warehouse_id, to_warehouse_id, created_by) values ($1, $2, $3) returning *;

-- name: InventoryAddTransferLine :one
insert into inventory.stock_transfer_lines (transfer_id, variant_id, condition, owner_seller_id, quantity) values ($1, $2, $3, $4, $5) returning *;

-- name: InventoryGetTransfer :one
select * from inventory.stock_transfers where id = $1 for update;

-- name: InventoryTransferLines :many
select * from inventory.stock_transfer_lines where transfer_id = $1;

-- name: InventorySetTransferStatus :one
update inventory.stock_transfers set status = $2,
  shipped_at = case when $2 = 'in_transit' then now() else shipped_at end,
  received_at = case when $2 = 'received' then now() else received_at end
where id = $1 returning *;

-- name: InventoryIntegrityDrift :many
-- Nightly check: levels whose on_hand differs from the sum of their movements.
select l.warehouse_id, l.variant_id, l.condition, l.owner_seller_id, l.on_hand, coalesce(m.total, 0)::int as movements
from inventory.inventory_levels l
left join (select warehouse_id, variant_id, condition, owner_seller_id, sum(quantity_delta) as total
           from inventory.stock_movements group by 1, 2, 3, 4) m
  using (warehouse_id, variant_id, condition, owner_seller_id)
where l.on_hand <> coalesce(m.total, 0);

-- name: InventoryCreateCount :one
insert into inventory.inventory_counts (warehouse_id, counted_by) values ($1, $2) returning *;

-- name: InventoryAddCountLine :one
insert into inventory.inventory_count_lines (count_id, variant_id, condition, owner_seller_id, expected_qty, counted_qty)
values ($1, $2, $3, $4,
  coalesce((select on_hand from inventory.inventory_levels where warehouse_id = (select warehouse_id from inventory.inventory_counts where id = $1)
            and variant_id = $2 and condition = $3 and owner_seller_id = $4), 0), $5)
on conflict (count_id, variant_id, condition, owner_seller_id) do update set counted_qty = excluded.counted_qty
returning *;

-- name: InventoryGetCount :one
select * from inventory.inventory_counts where id = $1 for update;

-- name: InventoryCountLines :many
select * from inventory.inventory_count_lines where count_id = $1;

-- name: InventorySetCountStatus :one
update inventory.inventory_counts set status = $2, approved_by = sqlc.narg('approved_by'), approved_at = case when $2 = 'approved' then now() else approved_at end
where id = $1 returning *;

-- name: InventoryUpsertCost :exec
insert into inventory.inventory_costs (variant_id, condition, warehouse_id, owner_seller_id, avg_cost_kobo, quantity_basis)
values ($1, $2, $3, $4, $5, $6)
on conflict (variant_id, condition, warehouse_id, owner_seller_id) do update set
  avg_cost_kobo = (inventory.inventory_costs.avg_cost_kobo * inventory.inventory_costs.quantity_basis + excluded.avg_cost_kobo * excluded.quantity_basis)
                  / greatest(inventory.inventory_costs.quantity_basis + excluded.quantity_basis, 1),
  quantity_basis = inventory.inventory_costs.quantity_basis + excluded.quantity_basis, updated_at = now();
