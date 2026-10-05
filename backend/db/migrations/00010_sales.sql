-- Sales: carts, B2B quotes, orders split into per-seller fulfilments, order lines, status history,
-- stock reservations, and reviews/ratings (which need order lines).
-- Plans: docs/orders-fulfilment.md §2–§3, docs/backend.md §3.1.

-- +goose Up
create sequence sales.order_number_seq start 10001;
create sequence b2b.quote_number_seq start 1001;

create table sales.carts (
  id                  uuid primary key default gen_random_uuid(),
  user_id             uuid references identity.users (id) on delete cascade,
  session_token_hash  bytea unique,
  business_id         uuid references b2b.businesses (id),
  status              text not null default 'active' check (status in ('active', 'converted', 'merged', 'abandoned')),
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  check (user_id is not null or session_token_hash is not null)
);
comment on table sales.carts is '[sales] Shopping carts for signed-in users or guests (by session token hash).';
create unique index carts_one_active_per_user on sales.carts (user_id) where status = 'active' and user_id is not null and business_id is null;

create table sales.cart_items (
  id          uuid primary key default gen_random_uuid(),
  cart_id     uuid not null references sales.carts (id) on delete cascade,
  listing_id  uuid not null references catalog.listings (id),
  quantity    integer not null check (quantity > 0),
  added_at    timestamptz not null default now(),
  unique (cart_id, listing_id)
);
comment on table sales.cart_items is '[sales] Listings in a cart. Prices are read live; they are snapshotted only on order lines.';

create table b2b.quotes (
  id                 uuid primary key default gen_random_uuid(),
  quote_number       text not null unique default platform.next_ref('QT-', 'b2b.quote_number_seq'),
  business_id        uuid references b2b.businesses (id),
  requested_by       uuid references identity.users (id),
  contact_name       text,
  contact_email      citext,
  contact_phone      text check (contact_phone ~ '^\+234[0-9]{10}$'),
  company_name       text,
  assigned_to        uuid references identity.staff_members (user_id),
  status             text not null default 'requested' check (status in ('requested', 'drafting', 'sent', 'accepted', 'declined', 'expired')),
  delivery_state     text references identity.nigerian_states (code),
  needed_by          date,
  needs_vat_invoice  boolean not null default false,
  wants_credit       boolean not null default false,
  notes              text,
  total_kobo         bigint check (total_kobo >= 0),
  currency           char(3) not null default 'NGN',
  valid_until        date,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (status <> 'sent' or (total_kobo is not null and valid_until is not null)),
  check (requested_by is not null or contact_email is not null or contact_phone is not null)
);
comment on table b2b.quotes is '[b2b] B2B quote requests (from signed-in buyers or the public form) and the priced quotes sent back.';

create table b2b.quote_lines (
  id               uuid primary key default gen_random_uuid(),
  quote_id         uuid not null references b2b.quotes (id) on delete cascade,
  listing_id       uuid references catalog.listings (id),
  description      text not null,
  quantity         integer not null check (quantity > 0),
  unit_price_kobo  bigint check (unit_price_kobo > 0)
);
comment on table b2b.quote_lines is '[b2b] Requested items (listed or free text) and their quoted unit prices.';

create table sales.orders (
  id                          uuid primary key default gen_random_uuid(),
  order_number                text not null unique default platform.next_ref('TS-', 'sales.order_number_seq'),
  channel                     text not null check (channel in ('market', 'wholesale', 'app', 'pos')),
  customer_user_id            uuid references identity.users (id),
  business_id                 uuid references b2b.businesses (id),
  quote_id                    uuid unique references b2b.quotes (id),
  pos_shift_id                uuid references pos.pos_shifts (id),
  status                      text not null default 'pending_payment' check (status in ('pending_payment', 'paid', 'processing',
                                'partially_shipped', 'shipped', 'delivered', 'cancelled', 'refunded', 'partially_refunded')),
  subtotal_kobo               bigint not null check (subtotal_kobo >= 0),
  delivery_fee_kobo           bigint not null default 0 check (delivery_fee_kobo >= 0),
  discount_kobo               bigint not null default 0 check (discount_kobo >= 0),
  vat_kobo                    bigint not null default 0 check (vat_kobo >= 0),
  total_kobo                  bigint not null check (total_kobo >= 0),
  currency                    char(3) not null default 'NGN',
  ship_to                     jsonb,
  contact_phone               text,
  price_hash                  text,
  payment_due_at              timestamptz,
  attributed_notification_id  uuid,
  placed_at                   timestamptz not null default now(),
  cancelled_at                timestamptz,
  created_at                  timestamptz not null default now(),
  updated_at                  timestamptz not null default now(),
  check (total_kobo = subtotal_kobo + delivery_fee_kobo - discount_kobo + vat_kobo),
  check (channel <> 'pos' or pos_shift_id is not null),
  check (channel = 'pos' or ship_to is not null),
  check (status <> 'cancelled' or cancelled_at is not null)
);
comment on table sales.orders is '[sales] One checkout. Totals must add up; ship_to snapshots the address; status is derived from fulfilments.';
create index orders_customer_placed_idx on sales.orders (customer_user_id, placed_at desc);
create index orders_unpaid_due_idx on sales.orders (payment_due_at) where status = 'pending_payment';

create table sales.fulfilments (
  id                    uuid primary key default gen_random_uuid(),
  order_id              uuid not null references sales.orders (id) on delete cascade,
  seller_id             uuid not null references sellers.sellers (id),
  fulfilled_by          text not null check (fulfilled_by in ('techshop', 'seller')),
  method                text not null default 'delivery' check (method in ('delivery', 'carrier', 'collection')),
  warehouse_id          uuid references inventory.warehouses (id),
  collection_store_id   uuid references inventory.warehouses (id),
  status                text not null default 'pending' check (status in ('pending', 'accepted', 'picking', 'packed', 'handed_over',
                          'in_transit', 'delivered', 'failed', 'returned', 'cancelled')),
  delivery_fee_kobo     bigint not null default 0 check (delivery_fee_kobo >= 0),
  accept_by             timestamptz,
  pack_by               timestamptz,
  delivered_at          timestamptz,
  cancelled_reason      text,
  created_at            timestamptz not null default now(),
  updated_at            timestamptz not null default now(),
  unique (order_id, seller_id),
  check (status <> 'delivered' or delivered_at is not null),
  check (method <> 'collection' or collection_store_id is not null),
  check (status <> 'cancelled' or cancelled_reason is not null)
);
comment on table sales.fulfilments is '[sales] The part of an order handled by one seller. Moves on its own; drives delivery, earnings and payouts.';
create index fulfilments_seller_status_idx on sales.fulfilments (seller_id, status);
create index fulfilments_sla_idx on sales.fulfilments (accept_by) where status = 'pending';

create table sales.order_lines (
  id               uuid primary key default gen_random_uuid(),
  order_id         uuid not null references sales.orders (id) on delete cascade,
  fulfilment_id    uuid not null references sales.fulfilments (id) on delete cascade,
  listing_id       uuid not null references catalog.listings (id),
  seller_id        uuid not null references sellers.sellers (id),
  variant_id       uuid not null references catalog.product_variants (id),
  product_name     text not null,
  variant_name     text not null,
  sku              text not null,
  condition        text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  quantity         integer not null check (quantity > 0),
  unit_price_kobo  bigint not null check (unit_price_kobo > 0),
  discount_kobo    bigint not null default 0 check (discount_kobo >= 0),
  line_total_kobo  bigint not null check (line_total_kobo >= 0),
  vat_kobo         bigint not null default 0 check (vat_kobo >= 0),
  commission_bps   integer not null check (commission_bps between 0 and 10000),
  device_unit_id   uuid unique references inventory.device_units (id),
  cancelled_at     timestamptz,
  created_at       timestamptz not null default now(),
  check (line_total_kobo = unit_price_kobo * quantity - discount_kobo),
  check (vat_kobo <= line_total_kobo),
  check (device_unit_id is null or quantity = 1)
);
comment on table sales.order_lines is '[sales] Snapshot of what was bought (name, price, VAT, condition, commission), optionally the exact device unit.';

create table sales.order_status_history (
  id             uuid primary key default gen_random_uuid(),
  order_id       uuid not null references sales.orders (id) on delete cascade,
  fulfilment_id  uuid references sales.fulfilments (id) on delete cascade,
  from_status    text,
  to_status      text not null,
  actor_user_id  uuid references identity.users (id),
  note           text,
  created_at     timestamptz not null default now()
);
comment on table sales.order_status_history is '[sales] Every order and fulfilment status change and who made it (shown on tracking).';

-- Exactly what was reserved, so expiry releases exactly that (docs/orders-fulfilment.md §3).
create table inventory.stock_reservations (
  id               uuid primary key default gen_random_uuid(),
  order_id         uuid not null references sales.orders (id) on delete cascade,
  order_line_id    uuid not null references sales.order_lines (id) on delete cascade,
  warehouse_id     uuid references inventory.warehouses (id),
  listing_id       uuid not null references catalog.listings (id),
  variant_id       uuid not null references catalog.product_variants (id),
  condition        text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  owner_seller_id  uuid not null references sellers.sellers (id),
  quantity         integer not null check (quantity > 0),
  status           text not null default 'reserved' check (status in ('reserved', 'committed', 'released')),
  expires_at       timestamptz not null,
  created_at       timestamptz not null default now(),
  updated_at       timestamptz not null default now()
);
comment on table inventory.stock_reservations is '[inventory] Stock held for an unpaid order (warehouse level or seller-held listing stock); committed on payment, released on expiry.';
create index stock_reservations_expiry_idx on inventory.stock_reservations (expires_at) where status = 'reserved';

create table catalog.product_reviews (
  id             uuid primary key default gen_random_uuid(),
  product_id     uuid not null references catalog.products (id) on delete cascade,
  order_line_id  uuid unique references sales.order_lines (id),
  user_id        uuid not null references identity.users (id),
  rating         smallint not null check (rating between 1 and 5),
  title          text,
  body           text,
  is_verified    boolean generated always as (order_line_id is not null) stored,
  status         text not null default 'pending' check (status in ('pending', 'published', 'hidden', 'rejected')),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now()
);
comment on table catalog.product_reviews is '[catalog] Product reviews. Verified purchase when tied to the order line bought; unverified ones are shown separately and weigh less.';
create unique index product_reviews_one_per_user_product on catalog.product_reviews (user_id, product_id);

create table catalog.review_images (
  review_id  uuid not null references catalog.product_reviews (id) on delete cascade,
  file_id    uuid not null references platform.files (id),
  position   smallint not null default 0,
  primary key (review_id, file_id)
);
comment on table catalog.review_images is '[catalog] Photos attached to a review.';

create table catalog.seller_ratings (
  id             uuid primary key default gen_random_uuid(),
  seller_id      uuid not null references sellers.sellers (id),
  fulfilment_id  uuid not null unique references sales.fulfilments (id),
  user_id        uuid not null references identity.users (id),
  rating         smallint not null check (rating between 1 and 5),
  comment        text,
  created_at     timestamptz not null default now()
);
comment on table catalog.seller_ratings is '[catalog] Customer rating of a seller per delivered fulfilment; aggregated onto sellers.rating_avg.';

-- Indexes on foreign-key columns (every FK is indexed).
create index quote_lines_listing_id_idx on b2b.quote_lines (listing_id);
create index quote_lines_quote_id_idx on b2b.quote_lines (quote_id);
create index quotes_assigned_to_idx on b2b.quotes (assigned_to);
create index quotes_business_id_idx on b2b.quotes (business_id);
create index quotes_delivery_state_idx on b2b.quotes (delivery_state);
create index quotes_requested_by_idx on b2b.quotes (requested_by);
create index product_reviews_product_id_idx on catalog.product_reviews (product_id);
create index review_images_file_id_idx on catalog.review_images (file_id);
create index seller_ratings_seller_id_idx on catalog.seller_ratings (seller_id);
create index seller_ratings_user_id_idx on catalog.seller_ratings (user_id);
create index stock_reservations_listing_id_idx on inventory.stock_reservations (listing_id);
create index stock_reservations_order_id_idx on inventory.stock_reservations (order_id);
create index stock_reservations_order_line_id_idx on inventory.stock_reservations (order_line_id);
create index stock_reservations_owner_seller_id_idx on inventory.stock_reservations (owner_seller_id);
create index stock_reservations_variant_id_idx on inventory.stock_reservations (variant_id);
create index stock_reservations_warehouse_id_idx on inventory.stock_reservations (warehouse_id);
create index cart_items_listing_id_idx on sales.cart_items (listing_id);
create index carts_business_id_idx on sales.carts (business_id);
create index fulfilments_collection_store_id_idx on sales.fulfilments (collection_store_id);
create index fulfilments_warehouse_id_idx on sales.fulfilments (warehouse_id);
create index order_lines_fulfilment_id_idx on sales.order_lines (fulfilment_id);
create index order_lines_listing_id_idx on sales.order_lines (listing_id);
create index order_lines_order_id_idx on sales.order_lines (order_id);
create index order_lines_seller_id_idx on sales.order_lines (seller_id);
create index order_lines_variant_id_idx on sales.order_lines (variant_id);
create index order_status_history_actor_user_id_idx on sales.order_status_history (actor_user_id);
create index order_status_history_fulfilment_id_idx on sales.order_status_history (fulfilment_id);
create index order_status_history_order_id_idx on sales.order_status_history (order_id);
create index orders_attributed_notification_id_idx on sales.orders (attributed_notification_id);
create index orders_business_id_idx on sales.orders (business_id);
create index orders_pos_shift_id_idx on sales.orders (pos_shift_id);

-- Keep updated_at current.
create trigger quotes_set_updated_at before update on b2b.quotes
  for each row execute function platform.set_updated_at();
create trigger product_reviews_set_updated_at before update on catalog.product_reviews
  for each row execute function platform.set_updated_at();
create trigger stock_reservations_set_updated_at before update on inventory.stock_reservations
  for each row execute function platform.set_updated_at();
create trigger carts_set_updated_at before update on sales.carts
  for each row execute function platform.set_updated_at();
create trigger fulfilments_set_updated_at before update on sales.fulfilments
  for each row execute function platform.set_updated_at();
create trigger orders_set_updated_at before update on sales.orders
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists catalog.seller_ratings, catalog.review_images, catalog.product_reviews,
  inventory.stock_reservations, sales.order_status_history, sales.order_lines, sales.fulfilments,
  sales.orders, b2b.quote_lines, b2b.quotes, sales.cart_items, sales.carts cascade;
drop sequence if exists b2b.quote_number_seq, sales.order_number_seq;
