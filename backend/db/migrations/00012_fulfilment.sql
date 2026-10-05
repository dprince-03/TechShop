-- Warehouse work and shipping: parcels, pick lists, packing, manifests, third-party carriers,
-- and seller SLA events. Plan: docs/orders-fulfilment.md §4–§5.

-- +goose Up
create table fulfilment.parcels (
  id              uuid primary key default gen_random_uuid(),
  fulfilment_id   uuid not null references sales.fulfilments (id) on delete cascade,
  label_code      text not null unique,
  box_code        text,
  weight_grams    integer check (weight_grams > 0),
  scanned_at      timestamptz,
  created_at      timestamptz not null default now()
);
comment on table fulfilment.parcels is '[fulfilment] Physical parcels with a printed label code, scanned at handover and pickup.';

create table fulfilment.pick_lists (
  id            uuid primary key default gen_random_uuid(),
  warehouse_id  uuid not null references inventory.warehouses (id),
  wave_at       timestamptz not null,
  zone          text,
  status        text not null default 'open' check (status in ('open', 'picking', 'done', 'cancelled')),
  picker_id     uuid references identity.staff_members (user_id),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table fulfilment.pick_lists is '[fulfilment] Wave pick lists (every 30 minutes) per warehouse and zone.';

create table fulfilment.pick_list_items (
  id              uuid primary key default gen_random_uuid(),
  pick_list_id    uuid not null references fulfilment.pick_lists (id) on delete cascade,
  fulfilment_id   uuid not null references sales.fulfilments (id),
  order_line_id   uuid not null references sales.order_lines (id),
  bin_id          uuid references inventory.bin_locations (id),
  quantity        integer not null check (quantity > 0),
  picked_qty      integer not null default 0 check (picked_qty >= 0),
  device_unit_id  uuid references inventory.device_units (id),
  scanned_at      timestamptz,
  unique (pick_list_id, order_line_id),
  check (picked_qty <= quantity)
);
comment on table fulfilment.pick_list_items is '[fulfilment] Lines to pick; scanning the serial binds the exact device unit to the order line.';

create table fulfilment.pack_records (
  id                     uuid primary key default gen_random_uuid(),
  fulfilment_id          uuid not null references sales.fulfilments (id),
  parcel_id              uuid not null unique references fulfilment.parcels (id) on delete cascade,
  weight_grams           integer not null check (weight_grams > 0),
  expected_weight_grams  integer check (expected_weight_grams > 0),
  flagged                boolean generated always as (
                           expected_weight_grams is not null and abs(weight_grams - expected_weight_grams) * 10 > expected_weight_grams
                         ) stored,
  packed_by              uuid not null references identity.users (id),
  packed_at              timestamptz not null default now()
);
comment on table fulfilment.pack_records is '[fulfilment] Packing: weighed parcel; more than 10% off the expected weight is flagged.';

create table fulfilment.carriers (
  id          uuid primary key default gen_random_uuid(),
  code        text not null unique,
  name        text not null,
  is_active   boolean not null default true,
  created_at  timestamptz not null default now()
);
comment on table fulfilment.carriers is '[fulfilment] Third-party carriers for interstate delivery (behind carrier.Port).';

create table fulfilment.manifests (
  id                 uuid primary key default gen_random_uuid(),
  warehouse_id       uuid not null references inventory.warehouses (id),
  rider_id           uuid references logistics.riders (user_id),
  carrier_id         uuid references fulfilment.carriers (id),
  signed_by_name     text,
  signed_at          timestamptz,
  created_by         uuid not null references identity.users (id),
  created_at         timestamptz not null default now(),
  check ((rider_id is null) <> (carrier_id is null))
);
comment on table fulfilment.manifests is '[fulfilment] Handover batches to a rider or carrier; counts must match before signing.';

create table fulfilment.manifest_parcels (
  manifest_id  uuid not null references fulfilment.manifests (id) on delete cascade,
  parcel_id    uuid not null unique references fulfilment.parcels (id),
  scanned_at   timestamptz,
  primary key (manifest_id, parcel_id)
);
comment on table fulfilment.manifest_parcels is '[fulfilment] Parcels on a manifest and when each was scanned.';

create table fulfilment.shipments (
  id               uuid primary key default gen_random_uuid(),
  fulfilment_id    uuid not null references sales.fulfilments (id),
  carrier_id       uuid references fulfilment.carriers (id),
  carrier_name     text,
  tracking_number  text not null,
  label_file_id    uuid references platform.files (id),
  status           text not null default 'booked' check (status in ('booked', 'in_transit', 'delivered', 'exception', 'lost', 'returned')),
  created_at       timestamptz not null default now(),
  updated_at       timestamptz not null default now(),
  unique (carrier_id, tracking_number),
  check (carrier_id is not null or carrier_name is not null)
);
comment on table fulfilment.shipments is '[fulfilment] Carrier shipments (TechShop-booked or seller''s own carrier with tracking).';

create table fulfilment.shipment_events (
  id                 uuid primary key default gen_random_uuid(),
  shipment_id        uuid not null references fulfilment.shipments (id) on delete cascade,
  carrier_event_id   text,
  status             text not null,
  location           text,
  occurred_at        timestamptz not null,
  received_at        timestamptz not null default now(),
  unique (shipment_id, carrier_event_id)
);
comment on table fulfilment.shipment_events is '[fulfilment] Tracking events from carrier webhooks or polling (deduplicated).';

create table fulfilment.seller_sla_events (
  id             uuid primary key default gen_random_uuid(),
  seller_id      uuid not null references sellers.sellers (id),
  fulfilment_id  uuid not null references sales.fulfilments (id),
  kind           text not null check (kind in ('late_accept', 'late_pack', 'seller_cancel', 'late_handover')),
  created_at     timestamptz not null default now(),
  unique (fulfilment_id, kind)
);
comment on table fulfilment.seller_sla_events is '[fulfilment] Seller SLA breaches (feed seller performance and enforcement).';
create index seller_sla_events_seller_idx on fulfilment.seller_sla_events (seller_id, created_at desc);

-- Indexes on foreign-key columns (every FK is indexed).
create index manifests_carrier_id_idx on fulfilment.manifests (carrier_id);
create index manifests_created_by_idx on fulfilment.manifests (created_by);
create index manifests_rider_id_idx on fulfilment.manifests (rider_id);
create index manifests_warehouse_id_idx on fulfilment.manifests (warehouse_id);
create index pack_records_fulfilment_id_idx on fulfilment.pack_records (fulfilment_id);
create index pack_records_packed_by_idx on fulfilment.pack_records (packed_by);
create index parcels_fulfilment_id_idx on fulfilment.parcels (fulfilment_id);
create index pick_list_items_bin_id_idx on fulfilment.pick_list_items (bin_id);
create index pick_list_items_device_unit_id_idx on fulfilment.pick_list_items (device_unit_id);
create index pick_list_items_fulfilment_id_idx on fulfilment.pick_list_items (fulfilment_id);
create index pick_list_items_order_line_id_idx on fulfilment.pick_list_items (order_line_id);
create index pick_lists_picker_id_idx on fulfilment.pick_lists (picker_id);
create index pick_lists_warehouse_id_idx on fulfilment.pick_lists (warehouse_id);
create index shipments_fulfilment_id_idx on fulfilment.shipments (fulfilment_id);
create index shipments_label_file_id_idx on fulfilment.shipments (label_file_id);

-- Keep updated_at current.
create trigger pick_lists_set_updated_at before update on fulfilment.pick_lists
  for each row execute function platform.set_updated_at();
create trigger shipments_set_updated_at before update on fulfilment.shipments
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists fulfilment.seller_sla_events, fulfilment.shipment_events, fulfilment.shipments,
  fulfilment.manifest_parcels, fulfilment.manifests, fulfilment.carriers, fulfilment.pack_records,
  fulfilment.pick_list_items, fulfilment.pick_lists, fulfilment.parcels cascade;
