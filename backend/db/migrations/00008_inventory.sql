-- Inventory (locations, stock levels, serialised units, movements, transfers, counts, costs)
-- and purchasing (suppliers, purchase orders, goods receipts).
-- Plan: docs/orders-fulfilment.md §3, §5, §9. Stock reservations are created with orders (00010).

-- +goose Up
create table inventory.warehouses (
  id          uuid primary key default gen_random_uuid(),
  code        text not null unique,
  name        text not null,
  kind        text not null check (kind in ('warehouse', 'store', 'hub')),
  address     text not null,
  city        text not null,
  state_code  text not null references identity.nigerian_states (code),
  latitude    numeric(9, 6),
  longitude   numeric(9, 6),
  is_active   boolean not null default true,
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now()
);
comment on table inventory.warehouses is '[inventory] Physical locations: warehouses, retail stores (POS) and dispatch hubs.';

create table inventory.bin_locations (
  id            uuid primary key default gen_random_uuid(),
  warehouse_id  uuid not null references inventory.warehouses (id) on delete cascade,
  code          text not null,
  zone          text,
  unique (warehouse_id, code)
);
comment on table inventory.bin_locations is '[inventory] Shelf/bin codes inside a warehouse (optional), used on pick lists.';

-- Vendor stock held in TechShop warehouses is kept apart by owner_seller_id (TechShop's own
-- stock uses the first-party seller id).
create table inventory.inventory_levels (
  warehouse_id     uuid not null references inventory.warehouses (id),
  variant_id       uuid not null references catalog.product_variants (id),
  condition        text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  owner_seller_id  uuid not null references sellers.sellers (id),
  on_hand          integer not null default 0 check (on_hand >= 0),
  reserved         integer not null default 0 check (reserved >= 0),
  reorder_point    integer not null default 0 check (reorder_point >= 0),
  bin_id           uuid references inventory.bin_locations (id),
  updated_at       timestamptz not null default now(),
  primary key (warehouse_id, variant_id, condition, owner_seller_id),
  check (reserved <= on_hand)
);
comment on table inventory.inventory_levels is '[inventory] Current stock per location × variant × condition × owner. Changed only together with a stock movement.';

create table inventory.device_units (
  id               uuid primary key default gen_random_uuid(),
  variant_id       uuid not null references catalog.product_variants (id),
  condition        text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  grade            text check (grade in ('A', 'B', 'C')),
  imei             text unique check (imei ~ '^[0-9]{15}$'),
  serial_number    text,
  warehouse_id     uuid references inventory.warehouses (id),
  owner_seller_id  uuid not null references sellers.sellers (id),
  cost_kobo        bigint check (cost_kobo >= 0),
  status           text not null default 'in_stock' check (status in ('in_stock', 'reserved', 'sold', 'in_repair', 'returned',
                     'quarantined', 'written_off')),
  acquired_via     text not null check (acquired_via in ('purchase', 'trade_in', 'return', 'consignment')),
  created_at       timestamptz not null default now(),
  updated_at       timestamptz not null default now(),
  check (imei is not null or serial_number is not null)
);
comment on table inventory.device_units is '[inventory] Individually tracked devices (IMEI/serial): picking, warranty, returns, theft checks, trade-ins.';
create unique index device_units_serial_idx on inventory.device_units (variant_id, serial_number) where serial_number is not null;

create table inventory.stock_transfers (
  id                 uuid primary key default gen_random_uuid(),
  from_warehouse_id  uuid not null references inventory.warehouses (id),
  to_warehouse_id    uuid not null references inventory.warehouses (id),
  status             text not null default 'draft' check (status in ('draft', 'in_transit', 'received', 'cancelled')),
  created_by         uuid not null references identity.users (id),
  shipped_at         timestamptz,
  received_at        timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (from_warehouse_id <> to_warehouse_id),
  check (status <> 'received' or received_at is not null)
);
comment on table inventory.stock_transfers is '[inventory] Moving stock between locations.';

create table inventory.stock_transfer_lines (
  id                 uuid primary key default gen_random_uuid(),
  transfer_id        uuid not null references inventory.stock_transfers (id) on delete cascade,
  variant_id         uuid not null references catalog.product_variants (id),
  condition          text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  owner_seller_id    uuid not null references sellers.sellers (id),
  quantity           integer not null check (quantity > 0),
  received_quantity  integer check (received_quantity >= 0)
);
comment on table inventory.stock_transfer_lines is '[inventory] Items in a stock transfer.';

create table inventory.stock_movements (
  id               uuid primary key default gen_random_uuid(),
  warehouse_id     uuid not null references inventory.warehouses (id),
  variant_id       uuid not null references catalog.product_variants (id),
  condition        text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  owner_seller_id  uuid not null references sellers.sellers (id),
  quantity_delta   integer not null check (quantity_delta <> 0),
  reason           text not null check (reason in ('purchase_receipt', 'sale', 'return', 'transfer_in', 'transfer_out',
                     'adjustment', 'trade_in', 'write_off', 'count_variance')),
  device_unit_id   uuid references inventory.device_units (id),
  reference_type   text,
  reference_id     uuid,
  created_by       uuid references identity.users (id),
  created_at       timestamptz not null default now()
);
comment on table inventory.stock_movements is '[inventory] Append-only stock ledger; every change to inventory_levels has a movement (nightly check: sum = on hand).';
create trigger stock_movements_append_only before update or delete on inventory.stock_movements
  for each row execute function platform.forbid_change();

create table inventory.inventory_counts (
  id            uuid primary key default gen_random_uuid(),
  warehouse_id  uuid not null references inventory.warehouses (id),
  status        text not null default 'open' check (status in ('open', 'submitted', 'approved', 'cancelled')),
  counted_by    uuid not null references identity.users (id),
  approved_by   uuid references identity.users (id),
  approved_at   timestamptz,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  check (approved_by is distinct from counted_by),
  check (status <> 'approved' or approved_by is not null)
);
comment on table inventory.inventory_counts is '[inventory] Cycle counts (blind). Variances above a threshold need a different person to approve.';

create table inventory.inventory_count_lines (
  id                uuid primary key default gen_random_uuid(),
  count_id          uuid not null references inventory.inventory_counts (id) on delete cascade,
  variant_id        uuid not null references catalog.product_variants (id),
  condition         text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  owner_seller_id   uuid not null references sellers.sellers (id),
  bin_id            uuid references inventory.bin_locations (id),
  expected_qty      integer not null check (expected_qty >= 0),
  counted_qty       integer check (counted_qty >= 0),
  variance          integer generated always as (counted_qty - expected_qty) stored,
  unique (count_id, variant_id, condition, owner_seller_id)
);
comment on table inventory.inventory_count_lines is '[inventory] Expected vs counted per item in a cycle count.';

create table inventory.inventory_costs (
  variant_id         uuid not null references catalog.product_variants (id),
  condition          text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  warehouse_id       uuid not null references inventory.warehouses (id),
  owner_seller_id    uuid not null references sellers.sellers (id),
  avg_cost_kobo      bigint not null check (avg_cost_kobo >= 0),
  quantity_basis     integer not null check (quantity_basis >= 0),
  updated_at         timestamptz not null default now(),
  primary key (variant_id, condition, warehouse_id, owner_seller_id)
);
comment on table inventory.inventory_costs is '[inventory] Weighted average cost for non-serialised stock (COGS once finance confirms the method).';

create sequence purchasing.po_number_seq start 2001;

create table purchasing.suppliers (
  id            uuid primary key default gen_random_uuid(),
  name          text not null,
  rc_number     text,
  contact_name  text,
  email         citext,
  phone         text,
  status        text not null default 'active' check (status in ('active', 'on_hold', 'inactive')),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table purchasing.suppliers is '[purchasing] Companies TechShop buys stock from.';

create table purchasing.purchase_orders (
  id            uuid primary key default gen_random_uuid(),
  po_number     text not null unique default platform.next_ref('PO-', 'purchasing.po_number_seq'),
  supplier_id   uuid not null references purchasing.suppliers (id),
  warehouse_id  uuid not null references inventory.warehouses (id),
  status        text not null default 'draft' check (status in ('draft', 'awaiting_approval', 'sent', 'partially_received', 'received', 'cancelled')),
  expected_on   date,
  total_kobo    bigint not null default 0 check (total_kobo >= 0),
  currency      char(3) not null default 'NGN',
  created_by    uuid not null references identity.users (id),
  approved_by   uuid references identity.users (id),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  check (approved_by is distinct from created_by)
);
comment on table purchasing.purchase_orders is '[purchasing] Orders placed with suppliers, delivered to a warehouse. Large POs need finance approval.';

create table purchasing.purchase_order_lines (
  id                 uuid primary key default gen_random_uuid(),
  purchase_order_id  uuid not null references purchasing.purchase_orders (id) on delete cascade,
  variant_id         uuid not null references catalog.product_variants (id),
  condition          text not null default 'new' check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  quantity_ordered   integer not null check (quantity_ordered > 0),
  quantity_received  integer not null default 0 check (quantity_received >= 0),
  unit_cost_kobo     bigint not null check (unit_cost_kobo > 0),
  check (quantity_received <= quantity_ordered * 1.05)
);
comment on table purchasing.purchase_order_lines is '[purchasing] Items, quantities and costs on a purchase order (over-receipt above 5% needs approval).';

create table purchasing.goods_receipts (
  id                 uuid primary key default gen_random_uuid(),
  purchase_order_id  uuid not null references purchasing.purchase_orders (id),
  received_by        uuid not null references identity.users (id),
  received_at        timestamptz not null default now(),
  notes              text
);
comment on table purchasing.goods_receipts is '[purchasing] A delivery received against a PO; posts stock movements (purchase_receipt) and device units.';

-- Indexes on foreign-key columns (every FK is indexed).
create index device_units_owner_seller_id_idx on inventory.device_units (owner_seller_id);
create index device_units_warehouse_id_idx on inventory.device_units (warehouse_id);
create index inventory_costs_owner_seller_id_idx on inventory.inventory_costs (owner_seller_id);
create index inventory_costs_warehouse_id_idx on inventory.inventory_costs (warehouse_id);
create index inventory_count_lines_bin_id_idx on inventory.inventory_count_lines (bin_id);
create index inventory_count_lines_owner_seller_id_idx on inventory.inventory_count_lines (owner_seller_id);
create index inventory_count_lines_variant_id_idx on inventory.inventory_count_lines (variant_id);
create index inventory_counts_approved_by_idx on inventory.inventory_counts (approved_by);
create index inventory_counts_counted_by_idx on inventory.inventory_counts (counted_by);
create index inventory_counts_warehouse_id_idx on inventory.inventory_counts (warehouse_id);
create index inventory_levels_bin_id_idx on inventory.inventory_levels (bin_id);
create index inventory_levels_owner_seller_id_idx on inventory.inventory_levels (owner_seller_id);
create index inventory_levels_variant_id_idx on inventory.inventory_levels (variant_id);
create index stock_movements_created_by_idx on inventory.stock_movements (created_by);
create index stock_movements_device_unit_id_idx on inventory.stock_movements (device_unit_id);
create index stock_movements_owner_seller_id_idx on inventory.stock_movements (owner_seller_id);
create index stock_movements_variant_id_idx on inventory.stock_movements (variant_id);
create index stock_movements_warehouse_id_idx on inventory.stock_movements (warehouse_id);
create index stock_transfer_lines_owner_seller_id_idx on inventory.stock_transfer_lines (owner_seller_id);
create index stock_transfer_lines_transfer_id_idx on inventory.stock_transfer_lines (transfer_id);
create index stock_transfer_lines_variant_id_idx on inventory.stock_transfer_lines (variant_id);
create index stock_transfers_created_by_idx on inventory.stock_transfers (created_by);
create index stock_transfers_from_warehouse_id_idx on inventory.stock_transfers (from_warehouse_id);
create index stock_transfers_to_warehouse_id_idx on inventory.stock_transfers (to_warehouse_id);
create index warehouses_state_code_idx on inventory.warehouses (state_code);
create index goods_receipts_purchase_order_id_idx on purchasing.goods_receipts (purchase_order_id);
create index goods_receipts_received_by_idx on purchasing.goods_receipts (received_by);
create index purchase_order_lines_purchase_order_id_idx on purchasing.purchase_order_lines (purchase_order_id);
create index purchase_order_lines_variant_id_idx on purchasing.purchase_order_lines (variant_id);
create index purchase_orders_approved_by_idx on purchasing.purchase_orders (approved_by);
create index purchase_orders_created_by_idx on purchasing.purchase_orders (created_by);
create index purchase_orders_supplier_id_idx on purchasing.purchase_orders (supplier_id);
create index purchase_orders_warehouse_id_idx on purchasing.purchase_orders (warehouse_id);

-- Keep updated_at current.
create trigger device_units_set_updated_at before update on inventory.device_units
  for each row execute function platform.set_updated_at();
create trigger inventory_costs_set_updated_at before update on inventory.inventory_costs
  for each row execute function platform.set_updated_at();
create trigger inventory_counts_set_updated_at before update on inventory.inventory_counts
  for each row execute function platform.set_updated_at();
create trigger inventory_levels_set_updated_at before update on inventory.inventory_levels
  for each row execute function platform.set_updated_at();
create trigger stock_transfers_set_updated_at before update on inventory.stock_transfers
  for each row execute function platform.set_updated_at();
create trigger warehouses_set_updated_at before update on inventory.warehouses
  for each row execute function platform.set_updated_at();
create trigger purchase_orders_set_updated_at before update on purchasing.purchase_orders
  for each row execute function platform.set_updated_at();
create trigger suppliers_set_updated_at before update on purchasing.suppliers
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists purchasing.goods_receipts, purchasing.purchase_order_lines,
  purchasing.purchase_orders, purchasing.suppliers, inventory.inventory_costs,
  inventory.inventory_count_lines, inventory.inventory_counts, inventory.stock_movements,
  inventory.stock_transfer_lines, inventory.stock_transfers, inventory.device_units,
  inventory.inventory_levels, inventory.bin_locations, inventory.warehouses cascade;
drop sequence if exists purchasing.po_number_seq;
