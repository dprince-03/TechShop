-- Sales: carts, orders, fulfilments, lines, status history (docs/orders-fulfilment.md §2).

-- name: SalesGetUserCart :one
select * from sales.carts where user_id = $1 and status = 'active' and business_id is null;

-- name: SalesGetGuestCart :one
select * from sales.carts where session_token_hash = $1 and status = 'active';

-- name: SalesCreateCart :one
insert into sales.carts (user_id, session_token_hash) values ($1, $2) returning *;

-- name: SalesSetCartStatus :exec
update sales.carts set status = $2 where id = $1;

-- name: SalesTouchCart :exec
update sales.carts set updated_at = now() where id = $1;

-- name: SalesCartLines :many
-- Cart lines with live prices and everything pricing and checkout need.
select ci.id, ci.cart_id, ci.listing_id, ci.quantity, ci.added_at,
       l.price_kobo, l.compare_at_kobo, l.condition, l.status as listing_status, l.fulfilled_by, l.seller_stock, l.min_order_quantity,
       l.seller_id, l.variant_id, v.name as variant_name, v.sku, v.weight_grams, p.id as product_id, p.name as product_name, p.slug as product_slug,
       p.category_id, s.display_name as seller_name, s.type as seller_type, s.state_code as seller_state, s.status as seller_status
from sales.cart_items ci
join catalog.listings l on l.id = ci.listing_id
join catalog.product_variants v on v.id = l.variant_id
join catalog.products p on p.id = v.product_id
join sellers.sellers s on s.id = l.seller_id
where ci.cart_id = $1
order by ci.listing_id;

-- name: SalesUpsertCartItem :one
insert into sales.cart_items (cart_id, listing_id, quantity) values ($1, $2, $3)
on conflict (cart_id, listing_id) do update set quantity = case when @replace::bool then excluded.quantity else sales.cart_items.quantity + excluded.quantity end
returning *;

-- name: SalesSetCartItemQuantity :execrows
update sales.cart_items set quantity = $3 where id = $1 and cart_id = $2;

-- name: SalesDeleteCartItem :execrows
delete from sales.cart_items where id = $1 and cart_id = $2;

-- name: SalesClearCart :exec
delete from sales.cart_items where cart_id = $1;

-- name: SalesMergeCartItems :exec
insert into sales.cart_items (cart_id, listing_id, quantity)
select @target::uuid, listing_id, quantity from sales.cart_items where cart_id = @source::uuid
on conflict (cart_id, listing_id) do update set quantity = greatest(sales.cart_items.quantity, excluded.quantity);

-- name: SalesTierPrice :one
-- Best wholesale tier unit price for a quantity (null when no tier applies).
select unit_price_kobo from catalog.price_tiers where listing_id = $1 and min_quantity <= $2 order by min_quantity desc limit 1;

-- name: SalesLiveFlashPrice :one
select plp.promotion_id, plp.price_kobo, plp.stock_limit, plp.claimed, plp.per_user_limit
from marketing.promotion_listing_prices plp
join marketing.promotions pr on pr.id = plp.promotion_id
where plp.listing_id = $1 and pr.status = 'live' and now() between pr.starts_at and pr.ends_at
order by plp.price_kobo limit 1;

-- name: SalesCommissionBps :one
-- Seller-specific rule, then category rule, then the seller override, then the configured default.
select coalesce(
  (select cs.rate_bps from finance.commission_rules cs where cs.seller_id = @seller_id::uuid and cs.category_id is null
     and current_date >= cs.valid_from and (cs.valid_to is null or current_date < cs.valid_to) order by cs.valid_from desc limit 1),
  (select cc.rate_bps from finance.commission_rules cc where cc.category_id = @category_id::uuid and cc.seller_id is null
     and current_date >= cc.valid_from and (cc.valid_to is null or current_date < cc.valid_to) order by cc.valid_from desc limit 1),
  (select so.commission_override_bps from sellers.sellers so where so.id = @seller_id::uuid),
  @default_bps::int
)::int as bps;

-- name: SalesCreateOrder :one
insert into sales.orders (channel, customer_user_id, business_id, quote_id, pos_shift_id, status, subtotal_kobo, delivery_fee_kobo, discount_kobo,
  vat_kobo, total_kobo, ship_to, contact_phone, price_hash, payment_due_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
returning *;

-- name: SalesCreateFulfilment :one
insert into sales.fulfilments (order_id, seller_id, fulfilled_by, method, warehouse_id, collection_store_id, delivery_fee_kobo)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: SalesCreateLine :one
insert into sales.order_lines (order_id, fulfilment_id, listing_id, seller_id, variant_id, product_name, variant_name, sku, condition, quantity,
  unit_price_kobo, discount_kobo, line_total_kobo, vat_kobo, commission_bps)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
returning *;

-- name: SalesInsertHistory :exec
insert into sales.order_status_history (order_id, fulfilment_id, from_status, to_status, actor_user_id, note) values ($1, $2, $3, $4, $5, $6);

-- name: SalesGetOrder :one
select * from sales.orders where id = $1;

-- name: SalesGetOrderForUpdate :one
select * from sales.orders where id = $1 for update;

-- name: SalesGetOrderByNumber :one
select * from sales.orders where order_number = $1;

-- name: SalesSetOrderStatus :one
update sales.orders set status = $2, cancelled_at = case when $2 = 'cancelled' then now() else cancelled_at end where id = $1 returning *;

-- name: SalesSetAttribution :exec
update sales.orders set attributed_notification_id = $2 where id = $1 and attributed_notification_id is null;

-- name: SalesOrderFulfilments :many
select f.*, s.display_name as seller_name, s.type as seller_type from sales.fulfilments f join sellers.sellers s on s.id = f.seller_id
where f.order_id = $1 order by f.created_at;

-- name: SalesOrderLines :many
select * from sales.order_lines where order_id = $1 order by created_at;

-- name: SalesOrderHistory :many
select * from sales.order_status_history where order_id = $1 order by created_at;

-- name: SalesListCustomerOrders :many
select * from sales.orders where customer_user_id = $1
  and (sqlc.narg('before')::timestamptz is null or placed_at < sqlc.narg('before'))
order by placed_at desc limit $2;

-- name: SalesStaffListOrders :many
select o.*, u.first_name, u.last_name from sales.orders o left join identity.users u on u.id = o.customer_user_id
where (sqlc.narg('status')::text is null or o.status = sqlc.narg('status'))
  and (sqlc.narg('q')::text is null or o.order_number ilike '%' || sqlc.narg('q') || '%' or (u.first_name || ' ' || u.last_name) ilike '%' || sqlc.narg('q') || '%')
  and (sqlc.narg('before')::timestamptz is null or o.placed_at < sqlc.narg('before'))
order by o.placed_at desc limit $1;

-- name: SalesGetFulfilment :one
select * from sales.fulfilments where id = $1;

-- name: SalesGetFulfilmentForUpdate :one
select * from sales.fulfilments where id = $1 for update;

-- name: SalesSetFulfilmentStatus :one
update sales.fulfilments
set status = $2, delivered_at = case when $2 = 'delivered' then now() else delivered_at end,
    cancelled_reason = coalesce(sqlc.narg('reason'), cancelled_reason),
    accept_by = coalesce(sqlc.narg('accept_by'), accept_by), pack_by = coalesce(sqlc.narg('pack_by'), pack_by)
where id = $1 returning *;

-- name: SalesCancelLine :one
update sales.order_lines set cancelled_at = now() where id = $1 and order_id = $2 and cancelled_at is null returning *;

-- name: SalesUnpaidExpiredOrders :many
select * from sales.orders where status = 'pending_payment' and payment_due_at < now() order by payment_due_at limit 50;

-- name: SalesSellerFulfilments :many
select f.*, o.order_number, o.ship_to, o.placed_at from sales.fulfilments f join sales.orders o on o.id = f.order_id
where f.seller_id = $1 and (sqlc.narg('status')::text is null or f.status = sqlc.narg('status'))
order by f.created_at desc limit $2;

-- name: SalesFulfilmentLines :many
select * from sales.order_lines where fulfilment_id = $1 order by created_at;

-- name: SalesPendingAcceptBreaches :many
select f.*, o.order_number from sales.fulfilments f join sales.orders o on o.id = f.order_id
where f.status = 'pending' and f.fulfilled_by = 'seller' and f.accept_by < now() limit 50;

-- name: SalesListingAvailable :one
-- Units that can still be sold for a listing: seller-held stock, or unreserved warehouse stock.
select case when l.fulfilled_by = 'seller' then coalesce(l.seller_stock, 0)
  else coalesce((select sum(il.on_hand - il.reserved) from inventory.inventory_levels il
                 where il.variant_id = l.variant_id and il.condition = l.condition and il.owner_seller_id = l.seller_id), 0) end::int as available
from catalog.listings l where l.id = $1;

-- name: SalesLineReservations :many
select * from inventory.stock_reservations where order_line_id = $1;

-- name: SalesCustomerOrderByNumber :one
select * from sales.orders where order_number = $1 and customer_user_id = $2;

-- name: SalesPaidOrdersCount :one
select count(*)::int from sales.orders where customer_user_id = $1 and status not in ('pending_payment', 'cancelled');

-- name: SalesSellerFundedAt :one
-- Was this listing's discount seller-funded when the order was placed? (Seller-funded
-- promotions always target the seller's own listings and never carry coupons.)
select exists (
  select 1 from marketing.promotions p join marketing.promotion_targets t on t.promotion_id = p.id
  where p.funded_by = 'seller' and t.listing_id = $1 and p.starts_at <= $2 and p.ends_at > $2 and p.status in ('live', 'ended')
)::bool;

-- name: SalesBindDeviceUnit :one
-- The exact device sold (IMEI/serial), bound when the unit is picked.
update sales.order_lines set device_unit_id = $2 where id = $1 and device_unit_id is null and quantity = 1 returning *;

-- name: SalesGetLine :one
select * from sales.order_lines where id = $1;

-- name: SalesCommissionForFulfilment :one
-- Commission owed on a delivered fulfilment: Σ line total × the line's snapshot rate.
select coalesce(sum((ol.line_total_kobo * ol.commission_bps + 5000) / 10000), 0)::bigint
from sales.order_lines ol where ol.fulfilment_id = $1 and ol.cancelled_at is null;

-- name: SalesSetFulfilmentWarehouse :exec
update sales.fulfilments set warehouse_id = $2 where id = $1;

-- name: SalesSetFulfilmentMethod :exec
update sales.fulfilments set method = $2, collection_store_id = $3 where id = $1;
