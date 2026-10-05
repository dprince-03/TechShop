-- Fulfilment: waves and pick lists, packing, parcels, manifests, carriers, shipments, seller SLAs
-- (docs/orders-fulfilment.md §4–§5).

-- name: FulfilmentWaveCandidates :many
-- Paid TechShop-fulfilled lines whose stock was committed in this warehouse and not yet on a pick list.
select ol.id as order_line_id, ol.fulfilment_id, ol.quantity, ol.variant_id, ol.product_name, ol.sku, f.order_id
from sales.order_lines ol
join sales.fulfilments f on f.id = ol.fulfilment_id
join sales.orders o on o.id = f.order_id
join inventory.stock_reservations r on r.order_line_id = ol.id and r.status = 'committed'
where f.status = 'pending' and f.fulfilled_by = 'techshop' and o.status in ('paid', 'processing', 'partially_shipped')
  and r.warehouse_id = $1 and ol.cancelled_at is null
  and not exists (select 1 from fulfilment.pick_list_items pi where pi.order_line_id = ol.id)
order by o.placed_at, ol.id
limit $2;

-- name: FulfilmentCreatePickList :one
insert into fulfilment.pick_lists (warehouse_id, wave_at, zone) values ($1, now(), $2) returning *;

-- name: FulfilmentAddPickItem :one
insert into fulfilment.pick_list_items (pick_list_id, fulfilment_id, order_line_id, bin_id, quantity) values ($1, $2, $3, $4, $5) returning *;

-- name: FulfilmentListPickLists :many
select pl.*, (select count(*) from fulfilment.pick_list_items i where i.pick_list_id = pl.id)::int as items,
       (select count(*) from fulfilment.pick_list_items i where i.pick_list_id = pl.id and i.picked_qty = i.quantity)::int as items_done
from fulfilment.pick_lists pl
where pl.warehouse_id = $1 and (sqlc.narg('status')::text is null or pl.status = sqlc.narg('status'))
order by pl.wave_at desc limit 100;

-- name: FulfilmentGetPickList :one
select * from fulfilment.pick_lists where id = $1;

-- name: FulfilmentGetPickListForUpdate :one
select * from fulfilment.pick_lists where id = $1 for update;

-- name: FulfilmentPickItems :many
select i.*, ol.product_name, ol.variant_name, ol.sku, ol.variant_id, ol.condition, ol.seller_id, v.is_serialised, b.code as bin_code
from fulfilment.pick_list_items i
join sales.order_lines ol on ol.id = i.order_line_id
join catalog.product_variants v on v.id = ol.variant_id
left join inventory.bin_locations b on b.id = i.bin_id
where i.pick_list_id = $1 order by b.code nulls last, ol.sku;

-- name: FulfilmentSetPickListStatus :one
update fulfilment.pick_lists set status = $2, picker_id = coalesce(sqlc.narg('picker_id'), picker_id) where id = $1 returning *;

-- name: FulfilmentScanItem :one
-- Conditional: never picks more than ordered.
update fulfilment.pick_list_items set picked_qty = picked_qty + @qty::int, scanned_at = now(),
  device_unit_id = coalesce(sqlc.narg('device_unit_id'), device_unit_id)
where pick_list_id = @pick_list_id and order_line_id = @order_line_id and picked_qty + @qty::int <= quantity
returning *;

-- name: FulfilmentPickListOpenItems :one
select count(*)::int from fulfilment.pick_list_items where pick_list_id = $1 and picked_qty < quantity;

-- name: FulfilmentUnpickedForFulfilment :one
-- Lines of a TechShop fulfilment not fully picked yet (packing waits for all of them).
select count(*)::int from sales.order_lines ol
where ol.fulfilment_id = $1 and ol.cancelled_at is null
  and not exists (select 1 from fulfilment.pick_list_items i where i.order_line_id = ol.id and i.picked_qty = i.quantity);

-- name: FulfilmentExpectedWeight :one
select coalesce(sum(coalesce(v.weight_grams, 0) * ol.quantity), 0)::int
from sales.order_lines ol join catalog.product_variants v on v.id = ol.variant_id
where ol.fulfilment_id = $1 and ol.cancelled_at is null;

-- name: FulfilmentCreateParcel :one
insert into fulfilment.parcels (fulfilment_id, label_code, box_code, weight_grams) values ($1, $2, $3, $4) returning *;

-- name: FulfilmentCreatePackRecord :one
insert into fulfilment.pack_records (fulfilment_id, parcel_id, weight_grams, expected_weight_grams, packed_by) values ($1, $2, $3, $4, $5) returning *;

-- name: FulfilmentParcelsForFulfilment :many
select * from fulfilment.parcels where fulfilment_id = $1 order by created_at;

-- name: FulfilmentParcelByLabel :one
select * from fulfilment.parcels where label_code = $1;

-- name: FulfilmentGetParcel :one
select * from fulfilment.parcels where id = $1;

-- name: FulfilmentMarkParcelScanned :exec
update fulfilment.parcels set scanned_at = now() where id = $1;

-- name: FulfilmentCreateManifest :one
insert into fulfilment.manifests (warehouse_id, rider_id, carrier_id, created_by) values ($1, $2, $3, $4) returning *;

-- name: FulfilmentAddManifestParcel :exec
insert into fulfilment.manifest_parcels (manifest_id, parcel_id, scanned_at) values ($1, $2, now());

-- name: FulfilmentGetManifest :one
select * from fulfilment.manifests where id = $1;

-- name: FulfilmentGetManifestForUpdate :one
select * from fulfilment.manifests where id = $1 for update;

-- name: FulfilmentManifestParcels :many
select p.*, mp.scanned_at as manifest_scanned_at from fulfilment.manifest_parcels mp join fulfilment.parcels p on p.id = mp.parcel_id
where mp.manifest_id = $1 order by p.label_code;

-- name: FulfilmentSignManifest :one
update fulfilment.manifests set signed_by_name = $2, signed_at = now() where id = $1 and signed_at is null returning *;

-- name: FulfilmentListManifests :many
select * from fulfilment.manifests where warehouse_id = $1 order by created_at desc limit 100;

-- name: FulfilmentListCarriers :many
select * from fulfilment.carriers order by name;

-- name: FulfilmentUpsertCarrier :one
insert into fulfilment.carriers (code, name, is_active) values ($1, $2, $3)
on conflict (code) do update set name = excluded.name, is_active = excluded.is_active returning *;

-- name: FulfilmentGetCarrierByCode :one
select * from fulfilment.carriers where code = $1;

-- name: FulfilmentCreateShipment :one
insert into fulfilment.shipments (fulfilment_id, carrier_id, carrier_name, tracking_number, label_file_id) values ($1, $2, $3, $4, $5) returning *;

-- name: FulfilmentShipmentByTracking :one
select * from fulfilment.shipments where carrier_id = $1 and tracking_number = $2;

-- name: FulfilmentShipmentsForFulfilment :many
select * from fulfilment.shipments where fulfilment_id = $1 order by created_at;

-- name: FulfilmentSetShipmentStatus :one
update fulfilment.shipments set status = $2 where id = $1 returning *;

-- name: FulfilmentInsertShipmentEvent :one
insert into fulfilment.shipment_events (shipment_id, carrier_event_id, status, location, occurred_at) values ($1, $2, $3, $4, $5)
on conflict (shipment_id, carrier_event_id) do nothing returning *;

-- name: FulfilmentShipmentEvents :many
select * from fulfilment.shipment_events where shipment_id = $1 order by occurred_at;

-- name: FulfilmentInsertSLAEvent :exec
insert into fulfilment.seller_sla_events (seller_id, fulfilment_id, kind) values ($1, $2, $3) on conflict do nothing;

-- name: FulfilmentSellerSLAStats :one
select count(*) filter (where kind = 'late_accept')::int as late_accept, count(*) filter (where kind = 'late_pack')::int as late_pack,
       count(*) filter (where kind = 'seller_cancel')::int as seller_cancel, count(*) filter (where kind = 'late_handover')::int as late_handover
from fulfilment.seller_sla_events where seller_id = $1 and created_at > now() - interval '90 days';
