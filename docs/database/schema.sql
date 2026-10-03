-- =============================================================================
-- TechShop — proposed database schema (PostgreSQL 17)
--
-- DESIGN DOCUMENT, NOT A MIGRATION. When backend work starts this file is split
-- into goose migrations in backend/db/migrations/ (see docs/database.md §9).
--
-- Conventions (docs/database.md §2):
--   * uuid primary keys (gen_random_uuid); human references from sequences
--   * money = bigint kobo + currency char(3) default 'NGN'
--   * statuses = text + CHECK constraints
--   * created_at / updated_at timestamptz; updated_at maintained by trigger
--   * every table has COMMENT ON TABLE '[domain] purpose' — diagrams are
--     generated from these by docs/database/tools/generate.sh
--   * every foreign key gets an index (created automatically at the end)
-- =============================================================================

create extension if not exists citext;
create extension if not exists pg_trgm;

-- Keeps updated_at current on every table that has the column (attached at the end).
create function set_updated_at() returns trigger
language plpgsql as $$
begin
  new.updated_at := now();
  return new;
end $$;

-- Human-readable reference generator, e.g. ref('TS-', 'order_number_seq') → 'TS-10001'.
create function next_ref(prefix text, seq regclass) returns text
language sql volatile as $$ select prefix || nextval(seq)::text $$;

create sequence order_number_seq start 10001;
create sequence quote_number_seq start 1001;
create sequence invoice_number_seq start 1001;
create sequence po_number_seq start 2001;
create sequence rma_number_seq start 1001;
create sequence ticket_number_seq start 5001;
create sequence claim_number_seq start 1001;
create sequence trade_in_ref_seq start 1001;
create sequence car_ref_seq start 1001;

-- =============================================================================
-- 1. IDENTITY & ACCESS
-- =============================================================================

create table users (
  id                 uuid primary key default gen_random_uuid(),
  email              citext unique,
  phone              text unique check (phone ~ '^\+234[0-9]{10}$'),
  password_hash      text,
  first_name         text not null,
  last_name          text not null,
  status             text not null default 'active' check (status in ('active', 'suspended', 'deleted')),
  email_verified_at  timestamptz,
  phone_verified_at  timestamptz,
  marketing_opt_in   boolean not null default false,
  last_login_at      timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (email is not null or phone is not null)
);
comment on table users is '[identity] Every person who signs in: customers, seller staff, business buyers and TechShop staff. Phone stored in E.164 (+234…).';

create table user_sessions (
  id                  uuid primary key default gen_random_uuid(),
  user_id             uuid not null references users (id) on delete cascade,
  refresh_token_hash  bytea not null unique,
  user_agent          text,
  ip                  inet,
  expires_at          timestamptz not null,
  revoked_at          timestamptz,
  created_at          timestamptz not null default now()
);
comment on table user_sessions is '[identity] Refresh-token sessions per device; tokens stored only as hashes.';

create table verification_codes (
  id           uuid primary key default gen_random_uuid(),
  user_id      uuid references users (id) on delete cascade,
  channel      text not null check (channel in ('sms', 'email')),
  destination  text not null,
  purpose      text not null check (purpose in ('signup', 'login', 'password_reset', 'phone_change', 'delivery_otp')),
  code_hash    bytea not null,
  attempts     smallint not null default 0 check (attempts between 0 and 10),
  expires_at   timestamptz not null,
  consumed_at  timestamptz,
  created_at   timestamptz not null default now()
);
comment on table verification_codes is '[identity] One-time codes sent by SMS or email (sign-up, login, password reset); hashed, rate-limited by attempts.';

create table nigerian_states (
  code  text primary key,
  name  text not null unique
);
comment on table nigerian_states is '[identity] Lookup: the 36 states and the FCT, used by addresses, zones and listings.';

create table roles (
  id           uuid primary key default gen_random_uuid(),
  key          text not null unique,
  name         text not null,
  description  text,
  is_system    boolean not null default false,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now()
);
comment on table roles is '[identity] Staff roles (e.g. finance, dispatch). Proposed defaults in docs/database/access.md; owner to decide.';

create table permissions (
  id           uuid primary key default gen_random_uuid(),
  key          text not null unique check (key ~ '^[a-z_-]+\.[a-z_]+$'),
  module       text not null,
  description  text
);
comment on table permissions is '[identity] Fine-grained actions per staff module, e.g. orders.refund, finance.payouts.';

create table role_permissions (
  role_id        uuid not null references roles (id) on delete cascade,
  permission_id  uuid not null references permissions (id) on delete cascade,
  primary key (role_id, permission_id)
);
comment on table role_permissions is '[identity] Which permissions each role grants.';

create table staff_members (
  user_id      uuid primary key references users (id),
  employee_no  text not null unique,
  department   text not null,
  job_title    text not null,
  status       text not null default 'active' check (status in ('active', 'on_leave', 'exited')),
  hired_on     date,
  exited_on    date,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now(),
  check (exited_on is null or hired_on is null or exited_on >= hired_on)
);
comment on table staff_members is '[identity] TechShop employees (HR & staff module). A staff member is a user with an employee record.';

create table staff_roles (
  user_id     uuid not null references staff_members (user_id) on delete cascade,
  role_id     uuid not null references roles (id) on delete cascade,
  granted_by  uuid references users (id),
  granted_at  timestamptz not null default now(),
  primary key (user_id, role_id)
);
comment on table staff_roles is '[identity] Roles assigned to each staff member.';

-- =============================================================================
-- 2. SELLERS & BUSINESSES (+ shared files)
-- =============================================================================

create table files (
  id            uuid primary key default gen_random_uuid(),
  storage_key   text not null unique,
  content_type  text not null,
  size_bytes    bigint not null check (size_bytes > 0),
  checksum      text,
  visibility    text not null default 'private' check (visibility in ('public', 'private')),
  uploaded_by   uuid references users (id),
  created_at    timestamptz not null default now()
);
comment on table files is '[sellers] Uploaded files (product photos, KYC documents, proofs of delivery, invoices). Bytes live in object storage.';

create table sellers (
  id                     uuid primary key default gen_random_uuid(),
  type                   text not null check (type in ('first_party', 'business', 'individual')),
  display_name           text not null,
  slug                   text not null unique,
  owner_user_id          uuid references users (id),
  status                 text not null default 'pending_kyc' check (status in ('pending_kyc', 'active', 'suspended', 'closed')),
  state_code             text references nigerian_states (code),
  city                   text,
  rating_avg             numeric(3, 2) check (rating_avg between 1 and 5),
  rating_count           integer not null default 0 check (rating_count >= 0),
  commission_override_bps integer check (commission_override_bps between 0 and 10000),
  created_at             timestamptz not null default now(),
  updated_at             timestamptz not null default now(),
  check (type = 'first_party' or owner_user_id is not null)
);
comment on table sellers is '[sellers] Everyone who sells: TechShop itself (first_party), marketplace businesses and individuals.';
create unique index sellers_one_first_party on sellers ((true)) where type = 'first_party';

create table seller_members (
  seller_id   uuid not null references sellers (id) on delete cascade,
  user_id     uuid not null references users (id) on delete cascade,
  role        text not null check (role in ('owner', 'manager', 'staff')),
  created_at  timestamptz not null default now(),
  primary key (seller_id, user_id)
);
comment on table seller_members is '[sellers] Users who can act for a seller in the Seller Centre.';

create table seller_kyc_submissions (
  id                   uuid primary key default gen_random_uuid(),
  seller_id            uuid not null references sellers (id) on delete cascade,
  status               text not null default 'submitted' check (status in ('submitted', 'in_review', 'approved', 'rejected', 'needs_more_info')),
  id_type              text not null check (id_type in ('nin_slip', 'passport', 'drivers_licence', 'voters_card')),
  id_number_encrypted  bytea not null,
  id_number_last4      text not null check (length(id_number_last4) = 4),
  cac_rc_number        text,
  tin                  text,
  reviewed_by          uuid references users (id),
  reviewed_at          timestamptz,
  rejection_reason     text,
  created_at           timestamptz not null default now(),
  updated_at           timestamptz not null default now(),
  check (status not in ('approved', 'rejected') or (reviewed_by is not null and reviewed_at is not null))
);
comment on table seller_kyc_submissions is '[sellers] Identity / business verification applications. ID numbers encrypted by the API (NDPA).';

create table kyc_documents (
  id             uuid primary key default gen_random_uuid(),
  submission_id  uuid not null references seller_kyc_submissions (id) on delete cascade,
  kind           text not null check (kind in ('id_front', 'id_back', 'cac_certificate', 'utility_bill', 'selfie')),
  file_id        uuid not null references files (id),
  created_at     timestamptz not null default now()
);
comment on table kyc_documents is '[sellers] Documents uploaded with a KYC submission.';

create table seller_bank_accounts (
  id                        uuid primary key default gen_random_uuid(),
  seller_id                 uuid not null references sellers (id) on delete cascade,
  bank_code                 text not null,
  account_name              text not null,
  account_number_encrypted  bytea not null,
  account_number_last4      text not null check (length(account_number_last4) = 4),
  verified_at               timestamptz,
  is_default                boolean not null default false,
  created_at                timestamptz not null default now(),
  updated_at                timestamptz not null default now()
);
comment on table seller_bank_accounts is '[sellers] Payout bank accounts (NUBAN encrypted, last 4 shown).';
create unique index seller_bank_accounts_one_default on seller_bank_accounts (seller_id) where is_default;

create table businesses (
  id                  uuid primary key default gen_random_uuid(),
  legal_name          text not null,
  rc_number           text unique,
  tin                 text,
  type                text not null check (type in ('office_sme', 'school', 'reseller', 'government', 'ngo', 'healthcare', 'other')),
  size_band           text check (size_band in ('1-10', '11-50', '51-200', '201-1000', '1000+')),
  status              text not null default 'pending_verification' check (status in ('pending_verification', 'active', 'suspended')),
  account_manager_id  uuid references staff_members (user_id),
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now()
);
comment on table businesses is '[sellers] B2B buyer organisations (wholesale site): trade pricing, VAT invoices, credit.';

create table business_members (
  business_id  uuid not null references businesses (id) on delete cascade,
  user_id      uuid not null references users (id) on delete cascade,
  role         text not null check (role in ('owner', 'buyer', 'finance')),
  created_at   timestamptz not null default now(),
  primary key (business_id, user_id)
);
comment on table business_members is '[sellers] Users who buy or pay on behalf of a business.';

create table credit_accounts (
  id           uuid primary key default gen_random_uuid(),
  business_id  uuid not null unique references businesses (id) on delete cascade,
  limit_kobo   bigint not null default 0 check (limit_kobo >= 0),
  currency     char(3) not null default 'NGN',
  term_days    integer not null default 30 check (term_days between 1 and 180),
  status       text not null default 'applied' check (status in ('applied', 'approved', 'suspended', 'closed')),
  approved_by  uuid references users (id),
  approved_at  timestamptz,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now()
);
comment on table credit_accounts is '[sellers] Pay-on-invoice credit for approved businesses (limit and payment term).';

create table addresses (
  id              uuid primary key default gen_random_uuid(),
  user_id         uuid references users (id) on delete cascade,
  business_id     uuid references businesses (id) on delete cascade,
  label           text,
  recipient_name  text not null,
  phone           text not null check (phone ~ '^\+234[0-9]{10}$'),
  line1           text not null,
  line2           text,
  landmark        text,
  city            text not null,
  lga             text,
  state_code      text not null references nigerian_states (code),
  latitude        numeric(9, 6),
  longitude       numeric(9, 6),
  is_default      boolean not null default false,
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  check ((user_id is null) <> (business_id is null))
);
comment on table addresses is '[identity] Saved delivery / business addresses. Owned by exactly one user or one business.';
create unique index addresses_one_default_per_user on addresses (user_id) where is_default and user_id is not null;

-- =============================================================================
-- 3. CATALOGUE
-- =============================================================================

create table categories (
  id           uuid primary key default gen_random_uuid(),
  parent_id    uuid references categories (id),
  slug         text not null unique,
  name         text not null,
  description  text,
  kind         text not null default 'product' check (kind in ('product', 'vehicle')),
  position     integer not null default 0,
  is_active    boolean not null default true,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now(),
  check (parent_id is distinct from id)
);
comment on table categories is '[catalogue] Category tree (Phones → Android…). kind = vehicle routes cars to the car tables.';

create table brands (
  id            uuid primary key default gen_random_uuid(),
  slug          text not null unique,
  name          text not null,
  logo_file_id  uuid references files (id),
  created_at    timestamptz not null default now()
);
comment on table brands is '[catalogue] Manufacturers / brands.';

create table attribute_definitions (
  id               uuid primary key default gen_random_uuid(),
  category_id      uuid not null references categories (id) on delete cascade,
  key              text not null,
  label            text not null,
  data_type        text not null check (data_type in ('text', 'number', 'boolean', 'enum')),
  unit             text,
  options          jsonb,
  is_filterable    boolean not null default false,
  is_variant_axis  boolean not null default false,
  position         integer not null default 0,
  unique (category_id, key),
  check (data_type <> 'enum' or jsonb_typeof(options) = 'array')
);
comment on table attribute_definitions is '[catalogue] Spec fields per category (RAM, storage, GPU…): drive spec tables, filters and variant axes.';

create table products (
  id             uuid primary key default gen_random_uuid(),
  category_id    uuid not null references categories (id),
  brand_id       uuid references brands (id),
  slug           text not null unique,
  name           text not null,
  description    text,
  status         text not null default 'draft' check (status in ('draft', 'in_review', 'active', 'archived')),
  search_vector  tsvector generated always as (
                   setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
                   setweight(to_tsvector('simple', coalesce(description, '')), 'B')
                 ) stored,
  created_by     uuid references users (id),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now()
);
comment on table products is '[catalogue] Catalogue entries (model level). Sold through listings on its variants.';
create index products_search_idx on products using gin (search_vector);
create index products_name_trgm_idx on products using gin (name gin_trgm_ops);

create table product_attribute_values (
  product_id    uuid not null references products (id) on delete cascade,
  attribute_id  uuid not null references attribute_definitions (id) on delete cascade,
  value_text    text,
  value_number  numeric,
  value_bool    boolean,
  primary key (product_id, attribute_id),
  check (num_nonnulls(value_text, value_number, value_bool) = 1)
);
comment on table product_attribute_values is '[catalogue] Spec values for a product (one typed value per attribute).';

create table product_variants (
  id            uuid primary key default gen_random_uuid(),
  product_id    uuid not null references products (id) on delete cascade,
  sku           text not null unique,
  name          text not null,
  axis_values   jsonb not null default '{}'::jsonb,
  gtin          text,
  weight_grams  integer check (weight_grams > 0),
  is_active     boolean not null default true,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table product_variants is '[catalogue] Purchasable configurations of a product, e.g. 256GB · Titanium.';

create table product_images (
  id          uuid primary key default gen_random_uuid(),
  product_id  uuid not null references products (id) on delete cascade,
  variant_id  uuid references product_variants (id) on delete cascade,
  file_id     uuid not null references files (id),
  alt         text not null,
  position    integer not null default 0
);
comment on table product_images is '[catalogue] Product photos (optionally per variant); alt text required.';

create table listings (
  id                  uuid primary key default gen_random_uuid(),
  seller_id           uuid not null references sellers (id),
  variant_id          uuid not null references product_variants (id),
  condition           text not null check (condition in ('new', 'uk_used', 'refurbished')),
  price_kobo          bigint not null check (price_kobo > 0),
  compare_at_kobo     bigint,
  currency            char(3) not null default 'NGN',
  fulfilled_by        text not null default 'techshop' check (fulfilled_by in ('techshop', 'seller')),
  seller_stock        integer check (seller_stock >= 0),
  min_order_quantity  integer check (min_order_quantity > 1),
  warranty_provider   text not null default 'manufacturer' check (warranty_provider in ('manufacturer', 'techshop', 'seller', 'none')),
  warranty_months     smallint not null default 0 check (warranty_months between 0 and 60),
  condition_notes     text,
  battery_health_pct  smallint check (battery_health_pct between 1 and 100),
  status              text not null default 'draft' check (status in ('draft', 'in_review', 'active', 'paused', 'rejected')),
  published_at        timestamptz,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  unique (seller_id, variant_id, condition),
  check (compare_at_kobo is null or compare_at_kobo > price_kobo),
  check (condition = 'new' or condition_notes is not null),
  check (fulfilled_by = 'techshop' or seller_stock is not null)
);
comment on table listings is '[catalogue] A seller''s offer on a variant: condition, price, stock, warranty. TechShop stock is a first_party listing.';
create index listings_variant_status_idx on listings (variant_id, status);

create table price_tiers (
  id               uuid primary key default gen_random_uuid(),
  listing_id       uuid not null references listings (id) on delete cascade,
  min_quantity     integer not null check (min_quantity > 1),
  unit_price_kobo  bigint not null check (unit_price_kobo > 0),
  unique (listing_id, min_quantity)
);
comment on table price_tiers is '[catalogue] Wholesale quantity breaks: unit price from min_quantity upward.';

create table listing_price_history (
  id               uuid primary key default gen_random_uuid(),
  listing_id       uuid not null references listings (id) on delete cascade,
  price_kobo       bigint not null check (price_kobo > 0),
  compare_at_kobo  bigint,
  changed_by       uuid references users (id),
  changed_at       timestamptz not null default now()
);
comment on table listing_price_history is '[catalogue] Every price change. Proves discounts are genuine (FCCPA) and feeds analytics.';

create table saved_items (
  user_id     uuid not null references users (id) on delete cascade,
  product_id  uuid not null references products (id) on delete cascade,
  created_at  timestamptz not null default now(),
  primary key (user_id, product_id)
);
comment on table saved_items is '[catalogue] Customers'' saved products (the heart button).';

-- =============================================================================
-- 4. INVENTORY
-- =============================================================================

create table warehouses (
  id          uuid primary key default gen_random_uuid(),
  code        text not null unique,
  name        text not null,
  kind        text not null check (kind in ('warehouse', 'store', 'hub')),
  address     text not null,
  city        text not null,
  state_code  text not null references nigerian_states (code),
  is_active   boolean not null default true,
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now()
);
comment on table warehouses is '[inventory] Physical locations: warehouses, retail stores (POS) and dispatch hubs.';

create table inventory_levels (
  warehouse_id   uuid not null references warehouses (id),
  variant_id     uuid not null references product_variants (id),
  condition      text not null check (condition in ('new', 'uk_used', 'refurbished')),
  on_hand        integer not null default 0 check (on_hand >= 0),
  reserved       integer not null default 0 check (reserved >= 0),
  reorder_point  integer not null default 0 check (reorder_point >= 0),
  updated_at     timestamptz not null default now(),
  primary key (warehouse_id, variant_id, condition),
  check (reserved <= on_hand)
);
comment on table inventory_levels is '[inventory] Current stock per location × variant × condition (derived from stock_movements).';

create table device_units (
  id             uuid primary key default gen_random_uuid(),
  variant_id     uuid not null references product_variants (id),
  condition      text not null check (condition in ('new', 'uk_used', 'refurbished')),
  imei           text unique check (imei ~ '^[0-9]{15}$'),
  serial_number  text,
  warehouse_id   uuid references warehouses (id),
  status         text not null default 'in_stock' check (status in ('in_stock', 'reserved', 'sold', 'in_repair', 'returned', 'written_off')),
  acquired_via   text not null check (acquired_via in ('purchase', 'trade_in', 'return')),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),
  check (imei is not null or serial_number is not null)
);
comment on table device_units is '[inventory] Individually tracked devices (IMEI/serial): warranty, returns, theft checks, trade-ins.';

create table stock_transfers (
  id                 uuid primary key default gen_random_uuid(),
  from_warehouse_id  uuid not null references warehouses (id),
  to_warehouse_id    uuid not null references warehouses (id),
  status             text not null default 'draft' check (status in ('draft', 'in_transit', 'received', 'cancelled')),
  created_by         uuid not null references users (id),
  shipped_at         timestamptz,
  received_at        timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (from_warehouse_id <> to_warehouse_id)
);
comment on table stock_transfers is '[inventory] Moving stock between locations.';

create table stock_transfer_lines (
  id           uuid primary key default gen_random_uuid(),
  transfer_id  uuid not null references stock_transfers (id) on delete cascade,
  variant_id   uuid not null references product_variants (id),
  condition    text not null check (condition in ('new', 'uk_used', 'refurbished')),
  quantity     integer not null check (quantity > 0)
);
comment on table stock_transfer_lines is '[inventory] Items in a stock transfer.';

-- =============================================================================
-- 5. PURCHASING
-- =============================================================================

create table suppliers (
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
comment on table suppliers is '[purchasing] Companies TechShop buys stock from.';

create table purchase_orders (
  id            uuid primary key default gen_random_uuid(),
  po_number     text not null unique default next_ref('PO-', 'po_number_seq'),
  supplier_id   uuid not null references suppliers (id),
  warehouse_id  uuid not null references warehouses (id),
  status        text not null default 'draft' check (status in ('draft', 'sent', 'partially_received', 'received', 'cancelled')),
  expected_on   date,
  total_kobo    bigint not null default 0 check (total_kobo >= 0),
  currency      char(3) not null default 'NGN',
  created_by    uuid not null references users (id),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table purchase_orders is '[purchasing] Orders placed with suppliers, delivered to a warehouse.';

create table purchase_order_lines (
  id                 uuid primary key default gen_random_uuid(),
  purchase_order_id  uuid not null references purchase_orders (id) on delete cascade,
  variant_id         uuid not null references product_variants (id),
  condition          text not null default 'new' check (condition in ('new', 'uk_used', 'refurbished')),
  quantity_ordered   integer not null check (quantity_ordered > 0),
  quantity_received  integer not null default 0 check (quantity_received >= 0),
  unit_cost_kobo     bigint not null check (unit_cost_kobo > 0),
  check (quantity_received <= quantity_ordered)
);
comment on table purchase_order_lines is '[purchasing] Items, quantities and costs on a purchase order.';

create table goods_receipts (
  id                 uuid primary key default gen_random_uuid(),
  purchase_order_id  uuid not null references purchase_orders (id),
  received_by        uuid not null references users (id),
  received_at        timestamptz not null default now(),
  notes              text
);
comment on table goods_receipts is '[purchasing] A delivery received against a PO; posts stock_movements (reason purchase_receipt).';

create table stock_movements (
  id              uuid primary key default gen_random_uuid(),
  warehouse_id    uuid not null references warehouses (id),
  variant_id      uuid not null references product_variants (id),
  condition       text not null check (condition in ('new', 'uk_used', 'refurbished')),
  quantity_delta  integer not null check (quantity_delta <> 0),
  reason          text not null check (reason in ('purchase_receipt', 'sale', 'return', 'transfer_in', 'transfer_out', 'adjustment', 'trade_in', 'write_off')),
  reference_type  text,
  reference_id    uuid,
  created_by      uuid references users (id),
  created_at      timestamptz not null default now()
);
comment on table stock_movements is '[inventory] Append-only stock ledger; every change to inventory_levels has a movement.';

-- =============================================================================
-- 6. POINT OF SALE
-- =============================================================================

create table pos_terminals (
  id            uuid primary key default gen_random_uuid(),
  warehouse_id  uuid not null references warehouses (id),
  label         text not null,
  status        text not null default 'active' check (status in ('active', 'retired')),
  created_at    timestamptz not null default now(),
  unique (warehouse_id, label)
);
comment on table pos_terminals is '[pos] Tills in physical stores (warehouses of kind store).';

create table pos_shifts (
  id                   uuid primary key default gen_random_uuid(),
  terminal_id          uuid not null references pos_terminals (id),
  cashier_id           uuid not null references staff_members (user_id),
  opened_at            timestamptz not null default now(),
  closed_at            timestamptz,
  opening_float_kobo   bigint not null default 0 check (opening_float_kobo >= 0),
  expected_cash_kobo   bigint,
  counted_cash_kobo    bigint,
  check (closed_at is null or closed_at > opened_at)
);
comment on table pos_shifts is '[pos] A cashier''s shift on a till, with cash reconciliation. POS sales are orders with channel pos.';

-- =============================================================================
-- 7. SALES
-- =============================================================================

create table carts (
  id                  uuid primary key default gen_random_uuid(),
  user_id             uuid references users (id) on delete cascade,
  session_token_hash  bytea unique,
  business_id         uuid references businesses (id),
  status              text not null default 'active' check (status in ('active', 'converted', 'abandoned')),
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  check (user_id is not null or session_token_hash is not null)
);
comment on table carts is '[sales] Shopping carts for signed-in users or guests (by session).';

create table cart_items (
  id          uuid primary key default gen_random_uuid(),
  cart_id     uuid not null references carts (id) on delete cascade,
  listing_id  uuid not null references listings (id),
  quantity    integer not null check (quantity > 0),
  added_at    timestamptz not null default now(),
  unique (cart_id, listing_id)
);
comment on table cart_items is '[sales] Listings in a cart. Prices are read live; they are snapshotted only on order lines.';

create table quotes (
  id                 uuid primary key default gen_random_uuid(),
  quote_number       text not null unique default next_ref('QT-', 'quote_number_seq'),
  business_id        uuid references businesses (id),
  requested_by       uuid not null references users (id),
  assigned_to        uuid references staff_members (user_id),
  status             text not null default 'requested' check (status in ('requested', 'drafting', 'sent', 'accepted', 'declined', 'expired')),
  delivery_state     text references nigerian_states (code),
  needed_by          date,
  needs_vat_invoice  boolean not null default false,
  wants_credit       boolean not null default false,
  notes              text,
  total_kobo         bigint check (total_kobo >= 0),
  currency           char(3) not null default 'NGN',
  valid_until        date,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (status <> 'sent' or (total_kobo is not null and valid_until is not null))
);
comment on table quotes is '[sales] B2B quote requests and the priced quotes sent back.';

create table quote_lines (
  id               uuid primary key default gen_random_uuid(),
  quote_id         uuid not null references quotes (id) on delete cascade,
  listing_id       uuid references listings (id),
  description      text not null,
  quantity         integer not null check (quantity > 0),
  unit_price_kobo  bigint check (unit_price_kobo > 0)
);
comment on table quote_lines is '[sales] Requested items (listed or free text) and their quoted unit prices.';

create table orders (
  id                 uuid primary key default gen_random_uuid(),
  order_number       text not null unique default next_ref('TS-', 'order_number_seq'),
  channel            text not null check (channel in ('market', 'wholesale', 'app', 'pos')),
  customer_user_id   uuid references users (id),
  business_id        uuid references businesses (id),
  quote_id           uuid unique references quotes (id),
  pos_shift_id       uuid references pos_shifts (id),
  status             text not null default 'pending_payment' check (status in ('pending_payment', 'paid', 'processing', 'partially_shipped', 'shipped', 'delivered', 'cancelled', 'refunded', 'partially_refunded')),
  subtotal_kobo      bigint not null check (subtotal_kobo >= 0),
  delivery_fee_kobo  bigint not null default 0 check (delivery_fee_kobo >= 0),
  discount_kobo      bigint not null default 0 check (discount_kobo >= 0),
  vat_kobo           bigint not null default 0 check (vat_kobo >= 0),
  total_kobo         bigint not null check (total_kobo >= 0),
  currency           char(3) not null default 'NGN',
  ship_to            jsonb,
  contact_phone      text,
  placed_at          timestamptz not null default now(),
  cancelled_at       timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (total_kobo = subtotal_kobo + delivery_fee_kobo - discount_kobo + vat_kobo),
  check (channel <> 'pos' or pos_shift_id is not null),
  check (channel = 'pos' or ship_to is not null)
);
comment on table orders is '[sales] One checkout. Totals must add up; ship_to is a snapshot of the address used.';
create index orders_customer_placed_idx on orders (customer_user_id, placed_at desc);

create table fulfilments (
  id                 uuid primary key default gen_random_uuid(),
  order_id           uuid not null references orders (id) on delete cascade,
  seller_id          uuid not null references sellers (id),
  fulfilled_by       text not null check (fulfilled_by in ('techshop', 'seller')),
  warehouse_id       uuid references warehouses (id),
  status             text not null default 'pending' check (status in ('pending', 'packed', 'handed_over', 'in_transit', 'delivered', 'failed', 'returned', 'cancelled')),
  delivery_fee_kobo  bigint not null default 0 check (delivery_fee_kobo >= 0),
  delivered_at       timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  unique (order_id, seller_id),
  check (status <> 'delivered' or delivered_at is not null)
);
comment on table fulfilments is '[sales] The part of an order shipped by one seller. Drives delivery, seller earnings and payouts.';
create index fulfilments_seller_status_idx on fulfilments (seller_id, status);

create table order_lines (
  id               uuid primary key default gen_random_uuid(),
  order_id         uuid not null references orders (id) on delete cascade,
  fulfilment_id    uuid not null references fulfilments (id) on delete cascade,
  listing_id       uuid not null references listings (id),
  seller_id        uuid not null references sellers (id),
  variant_id       uuid not null references product_variants (id),
  product_name     text not null,
  variant_name     text not null,
  sku              text not null,
  condition        text not null check (condition in ('new', 'uk_used', 'refurbished')),
  quantity         integer not null check (quantity > 0),
  unit_price_kobo  bigint not null check (unit_price_kobo > 0),
  discount_kobo    bigint not null default 0 check (discount_kobo >= 0),
  line_total_kobo  bigint not null check (line_total_kobo >= 0),
  commission_bps   integer not null check (commission_bps between 0 and 10000),
  device_unit_id   uuid unique references device_units (id),
  created_at       timestamptz not null default now(),
  check (line_total_kobo = unit_price_kobo * quantity - discount_kobo),
  check (device_unit_id is null or quantity = 1)
);
comment on table order_lines is '[sales] Snapshot of what was bought (name, price, condition, commission), optionally the exact device unit.';

create table order_status_history (
  id             uuid primary key default gen_random_uuid(),
  order_id       uuid not null references orders (id) on delete cascade,
  from_status    text,
  to_status      text not null,
  actor_user_id  uuid references users (id),
  note           text,
  created_at     timestamptz not null default now()
);
comment on table order_status_history is '[sales] Every order status change and who made it (shown on order tracking).';

create table invoices (
  id                 uuid primary key default gen_random_uuid(),
  invoice_number     text not null unique default next_ref('INV-', 'invoice_number_seq'),
  business_id        uuid not null references businesses (id),
  order_id           uuid not null unique references orders (id),
  issued_at          timestamptz not null default now(),
  due_at             timestamptz not null,
  subtotal_kobo      bigint not null check (subtotal_kobo >= 0),
  vat_kobo           bigint not null default 0 check (vat_kobo >= 0),
  total_kobo         bigint not null check (total_kobo >= 0),
  amount_paid_kobo   bigint not null default 0 check (amount_paid_kobo >= 0),
  currency           char(3) not null default 'NGN',
  status             text not null default 'issued' check (status in ('issued', 'partially_paid', 'paid', 'overdue', 'void')),
  pdf_file_id        uuid references files (id),
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (total_kobo = subtotal_kobo + vat_kobo),
  check (amount_paid_kobo <= total_kobo),
  check (due_at >= issued_at)
);
comment on table invoices is '[sales] B2B VAT invoices; on credit terms they are paid later against due_at.';

create table product_reviews (
  id             uuid primary key default gen_random_uuid(),
  product_id     uuid not null references products (id) on delete cascade,
  order_line_id  uuid not null unique references order_lines (id),
  user_id        uuid not null references users (id),
  rating         smallint not null check (rating between 1 and 5),
  title          text,
  body           text,
  status         text not null default 'pending' check (status in ('pending', 'published', 'rejected')),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now()
);
comment on table product_reviews is '[sales] Verified-purchase reviews only: each review is tied to the order line that was bought.';

create table seller_ratings (
  id             uuid primary key default gen_random_uuid(),
  seller_id      uuid not null references sellers (id),
  fulfilment_id  uuid not null unique references fulfilments (id),
  user_id        uuid not null references users (id),
  rating         smallint not null check (rating between 1 and 5),
  comment        text,
  created_at     timestamptz not null default now()
);
comment on table seller_ratings is '[sales] Customer rating of a seller per delivered fulfilment; aggregated onto sellers.rating_avg.';

-- =============================================================================
-- 8. PAYMENTS & MONEY
-- =============================================================================

create table payments (
  id                  uuid primary key default gen_random_uuid(),
  order_id            uuid references orders (id),
  invoice_id          uuid references invoices (id),
  provider            text not null check (provider in ('paystack', 'opay', 'moniepoint', 'bank_transfer', 'credit', 'cash')),
  provider_reference  text,
  idempotency_key     text not null unique,
  amount_kobo         bigint not null check (amount_kobo > 0),
  currency            char(3) not null default 'NGN',
  status              text not null default 'initiated' check (status in ('initiated', 'pending', 'succeeded', 'failed', 'cancelled', 'reversed')),
  paid_at             timestamptz,
  failure_reason      text,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  unique (provider, provider_reference),
  check ((order_id is null) <> (invoice_id is null)),
  check (status <> 'succeeded' or paid_at is not null)
);
comment on table payments is '[payments] A charge for an order or an invoice through a provider. Card data is never stored.';

create table payment_events (
  id                 uuid primary key default gen_random_uuid(),
  provider           text not null check (provider in ('paystack', 'opay', 'moniepoint')),
  provider_event_id  text not null,
  event_type         text not null,
  payment_id         uuid references payments (id),
  payload            jsonb not null,
  signature_valid    boolean not null,
  received_at        timestamptz not null default now(),
  processed_at       timestamptz,
  unique (provider, provider_event_id)
);
comment on table payment_events is '[payments] Raw provider webhooks; unique per provider event so retries are processed once.';

create table refunds (
  id                  uuid primary key default gen_random_uuid(),
  payment_id          uuid not null references payments (id),
  order_id            uuid not null references orders (id),
  return_request_id   uuid,
  amount_kobo         bigint not null check (amount_kobo > 0),
  currency            char(3) not null default 'NGN',
  reason              text not null,
  status              text not null default 'requested' check (status in ('requested', 'approved', 'processing', 'succeeded', 'failed', 'rejected')),
  approved_by         uuid references users (id),
  provider_reference  text,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  check (status not in ('approved', 'processing', 'succeeded') or approved_by is not null)
);
comment on table refunds is '[payments] Money returned to a customer against a payment (full or partial).';

create table ledger_accounts (
  id          uuid primary key default gen_random_uuid(),
  code        text not null unique,
  name        text not null,
  kind        text not null check (kind in ('asset', 'liability', 'revenue', 'expense')),
  seller_id   uuid unique references sellers (id),
  created_at  timestamptz not null default now()
);
comment on table ledger_accounts is '[payments] Chart of accounts: provider clearing, seller payables (one per seller), commission, delivery, VAT, refunds.';

create table ledger_journals (
  id              uuid primary key default gen_random_uuid(),
  kind            text not null check (kind in ('sale', 'delivery_fee', 'commission', 'payout', 'refund', 'chargeback', 'adjustment')),
  reference_type  text not null,
  reference_id    uuid not null,
  memo            text,
  posted_by       uuid references users (id),
  posted_at       timestamptz not null default now()
);
comment on table ledger_journals is '[payments] One accounting event (e.g. a sale). Its entries must sum to zero.';

create table ledger_entries (
  id           uuid primary key default gen_random_uuid(),
  journal_id   uuid not null references ledger_journals (id),
  account_id   uuid not null references ledger_accounts (id),
  amount_kobo  bigint not null check (amount_kobo <> 0),
  currency     char(3) not null default 'NGN',
  created_at   timestamptz not null default now()
);
comment on table ledger_entries is '[payments] Double-entry lines: positive = debit, negative = credit. Balanced per journal (deferred trigger).';

create table commission_rules (
  id           uuid primary key default gen_random_uuid(),
  category_id  uuid references categories (id),
  seller_id    uuid references sellers (id),
  rate_bps     integer not null check (rate_bps between 0 and 10000),
  valid_from   date not null default current_date,
  valid_to     date,
  created_by   uuid not null references users (id),
  created_at   timestamptz not null default now(),
  check (valid_to is null or valid_to > valid_from)
);
comment on table commission_rules is '[payments] Marketplace commission (basis points) by category and/or seller; the rate is snapshotted on order lines.';

create table payouts (
  id                  uuid primary key default gen_random_uuid(),
  seller_id           uuid not null references sellers (id),
  bank_account_id     uuid not null references seller_bank_accounts (id),
  amount_kobo         bigint not null check (amount_kobo > 0),
  currency            char(3) not null default 'NGN',
  status              text not null default 'scheduled' check (status in ('scheduled', 'processing', 'paid', 'failed', 'on_hold')),
  scheduled_for       date not null,
  paid_at             timestamptz,
  provider_reference  text,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  check (status <> 'paid' or paid_at is not null)
);
comment on table payouts is '[payments] Transfers of earnings to a seller''s bank account.';

create table payout_items (
  payout_id      uuid not null references payouts (id) on delete cascade,
  fulfilment_id  uuid not null unique references fulfilments (id),
  amount_kobo    bigint not null check (amount_kobo > 0),
  primary key (payout_id, fulfilment_id)
);
comment on table payout_items is '[payments] Which delivered fulfilments a payout covers; each fulfilment is paid out once.';

-- =============================================================================
-- 9. LOGISTICS
-- =============================================================================

create table delivery_zones (
  id          uuid primary key default gen_random_uuid(),
  name        text not null unique,
  state_code  text not null references nigerian_states (code),
  lgas        text[] not null default '{}',
  is_active   boolean not null default true,
  created_at  timestamptz not null default now()
);
comment on table delivery_zones is '[logistics] Delivery areas (state + LGAs) used for fees, ETAs and rider assignment.';

create table delivery_rates (
  id               uuid primary key default gen_random_uuid(),
  zone_id          uuid not null references delivery_zones (id) on delete cascade,
  fulfilled_by     text not null check (fulfilled_by in ('techshop', 'seller')),
  max_weight_grams integer not null check (max_weight_grams > 0),
  fee_kobo         bigint not null check (fee_kobo >= 0),
  eta_min_days     smallint not null check (eta_min_days >= 0),
  eta_max_days     smallint not null,
  valid_from       date not null default current_date,
  check (eta_max_days >= eta_min_days)
);
comment on table delivery_rates is '[logistics] Delivery fee and ETA per zone and weight band.';

create table riders (
  user_id            uuid primary key references staff_members (user_id),
  vehicle_type       text not null check (vehicle_type in ('bike', 'car', 'van')),
  plate_number       text unique,
  home_warehouse_id  uuid not null references warehouses (id),
  status             text not null default 'off_shift' check (status in ('off_shift', 'available', 'on_delivery')),
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now()
);
comment on table riders is '[logistics] Dispatch riders (staff using the logistics app).';

create table delivery_jobs (
  id              uuid primary key default gen_random_uuid(),
  fulfilment_id   uuid not null references fulfilments (id),
  rider_id        uuid references riders (user_id),
  zone_id         uuid not null references delivery_zones (id),
  status          text not null default 'unassigned' check (status in ('unassigned', 'assigned', 'picked_up', 'en_route', 'delivered', 'failed', 'returned')),
  window_start    timestamptz,
  window_end      timestamptz,
  attempts        smallint not null default 0 check (attempts between 0 and 5),
  assigned_by     uuid references users (id),
  assigned_at     timestamptz,
  delivered_at    timestamptz,
  recipient_name  text,
  otp_hash        bytea,
  proof_file_id   uuid references files (id),
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  check (window_end is null or window_end > window_start),
  check (status = 'unassigned' or rider_id is not null),
  check (status <> 'delivered' or (delivered_at is not null and (proof_file_id is not null or otp_hash is not null)))
);
comment on table delivery_jobs is '[logistics] A delivery run for a fulfilment. Delivered requires OTP or photo proof.';
create index delivery_jobs_rider_status_idx on delivery_jobs (rider_id, status);

create table delivery_events (
  id               uuid primary key default gen_random_uuid(),
  delivery_job_id  uuid not null references delivery_jobs (id) on delete cascade,
  kind             text not null check (kind in ('assigned', 'picked_up', 'location', 'attempt_failed', 'delivered', 'returned', 'note')),
  latitude         numeric(9, 6),
  longitude        numeric(9, 6),
  note             text,
  created_by       uuid references users (id),
  created_at       timestamptz not null default now()
);
comment on table delivery_events is '[logistics] Append-only tracking trail for a delivery job (shown to customers and dispatch).';

-- =============================================================================
-- 10. AFTER-SALES
-- =============================================================================

create table return_requests (
  id            uuid primary key default gen_random_uuid(),
  rma_number    text not null unique default next_ref('RMA-', 'rma_number_seq'),
  order_id      uuid not null references orders (id),
  requested_by  uuid not null references users (id),
  reason        text not null check (reason in ('faulty', 'damaged', 'not_as_described', 'wrong_item', 'change_of_mind')),
  details       text,
  status        text not null default 'requested' check (status in ('requested', 'approved', 'rejected', 'awaiting_pickup', 'received', 'inspected', 'refunded', 'closed')),
  resolution    text check (resolution in ('refund', 'replace', 'repair')),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table return_requests is '[after-sales] Customer return requests (RMA) and their outcome.';

alter table refunds
  add constraint refunds_return_request_id_fkey foreign key (return_request_id) references return_requests (id);

create table return_items (
  id                  uuid primary key default gen_random_uuid(),
  return_request_id   uuid not null references return_requests (id) on delete cascade,
  order_line_id       uuid not null references order_lines (id),
  quantity            integer not null check (quantity > 0),
  condition_received  text,
  unique (return_request_id, order_line_id)
);
comment on table return_items is '[after-sales] Which order lines (and how many) are being returned.';

create table warranty_claims (
  id                 uuid primary key default gen_random_uuid(),
  claim_number       text not null unique default next_ref('WR-', 'claim_number_seq'),
  customer_user_id   uuid not null references users (id),
  order_line_id      uuid references order_lines (id),
  device_unit_id     uuid references device_units (id),
  imei_or_serial     text not null,
  fault_description  text not null,
  status             text not null default 'new' check (status in ('new', 'received', 'in_repair', 'ready', 'collected', 'rejected')),
  resolution         text check (resolution in ('repaired', 'replaced', 'refunded', 'not_covered')),
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now()
);
comment on table warranty_claims is '[after-sales] Warranty claims, identified by IMEI/serial.';

create table repair_jobs (
  id                 uuid primary key default gen_random_uuid(),
  warranty_claim_id  uuid references warranty_claims (id),
  device_unit_id     uuid references device_units (id),
  technician_id      uuid references staff_members (user_id),
  status             text not null default 'queued' check (status in ('queued', 'diagnosing', 'awaiting_parts', 'repairing', 'done', 'unrepairable')),
  diagnosis          text,
  chargeable         boolean not null default false,
  cost_kobo          bigint check (cost_kobo >= 0),
  started_at         timestamptz,
  finished_at        timestamptz,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now(),
  check (warranty_claim_id is not null or device_unit_id is not null),
  check (not chargeable or cost_kobo is not null)
);
comment on table repair_jobs is '[after-sales] Workshop jobs for warranty claims, paid repairs and refurbishment.';

create table trade_ins (
  id                   uuid primary key default gen_random_uuid(),
  reference            text not null unique default next_ref('TI-', 'trade_in_ref_seq'),
  user_id              uuid not null references users (id),
  device_type          text not null check (device_type in ('phone', 'laptop', 'tablet', 'console', 'smartwatch')),
  brand                text not null,
  model                text not null,
  storage              text,
  declared_condition   text not null check (declared_condition in ('like_new', 'good', 'fair', 'faulty')),
  imei                 text check (imei ~ '^[0-9]{15}$'),
  estimate_kobo        bigint check (estimate_kobo >= 0),
  inspected_condition  text check (inspected_condition in ('like_new', 'good', 'fair', 'faulty')),
  offer_kobo           bigint check (offer_kobo >= 0),
  payout_method        text check (payout_method in ('bank', 'store_credit')),
  status               text not null default 'submitted' check (status in ('submitted', 'estimated', 'awaiting_inspection', 'inspected', 'offer_sent', 'accepted', 'declined', 'paid', 'cancelled')),
  inspected_by         uuid references staff_members (user_id),
  device_unit_id       uuid unique references device_units (id),
  created_at           timestamptz not null default now(),
  updated_at           timestamptz not null default now(),
  check (status not in ('offer_sent', 'accepted', 'paid') or offer_kobo is not null)
);
comment on table trade_ins is '[after-sales] Trade-in requests from estimate to inspection, offer and payout; accepted devices become device_units.';

-- =============================================================================
-- 11. CARS
-- =============================================================================

create table car_listings (
  id              uuid primary key default gen_random_uuid(),
  reference       text not null unique default next_ref('CAR-', 'car_ref_seq'),
  seller_id       uuid not null references sellers (id),
  make            text not null,
  model           text not null,
  year            smallint not null check (year between 1950 and 2100),
  trim            text,
  body_type       text,
  transmission    text not null check (transmission in ('automatic', 'manual')),
  fuel_type       text check (fuel_type in ('petrol', 'diesel', 'hybrid', 'electric', 'cng')),
  mileage_km      integer not null check (mileage_km >= 0),
  condition       text not null check (condition in ('brand_new', 'foreign_used', 'nigerian_used')),
  vin             text unique check (vin ~ '^[A-HJ-NPR-Z0-9]{17}$'),
  colour          text,
  state_code      text not null references nigerian_states (code),
  city            text not null,
  price_kobo      bigint not null check (price_kobo > 0),
  currency        char(3) not null default 'NGN',
  status          text not null default 'draft' check (status in ('draft', 'in_review', 'active', 'reserved', 'sold', 'withdrawn')),
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  check (condition <> 'brand_new' or mileage_km < 500)
);
comment on table car_listings is '[cars] Cars for sale (separate from products): viewing and inspection flow, never add-to-cart.';

create table car_images (
  id              uuid primary key default gen_random_uuid(),
  car_listing_id  uuid not null references car_listings (id) on delete cascade,
  file_id         uuid not null references files (id),
  alt             text not null,
  position        integer not null default 0
);
comment on table car_images is '[cars] Photos of a car listing.';

create table car_inspections (
  id              uuid primary key default gen_random_uuid(),
  car_listing_id  uuid not null references car_listings (id) on delete cascade,
  inspector_id    uuid references staff_members (user_id),
  status          text not null default 'scheduled' check (status in ('scheduled', 'passed', 'passed_with_notes', 'failed')),
  checklist       jsonb,
  report_file_id  uuid references files (id),
  inspected_at    timestamptz,
  created_at      timestamptz not null default now(),
  check (status = 'scheduled' or inspected_at is not null)
);
comment on table car_inspections is '[cars] Inspection reports (engine, brakes, body, electricals, documents/VIN).';

create table car_documents (
  id              uuid primary key default gen_random_uuid(),
  car_listing_id  uuid not null references car_listings (id) on delete cascade,
  kind            text not null check (kind in ('customs_papers', 'proof_of_ownership', 'vehicle_licence', 'roadworthiness', 'insurance')),
  file_id         uuid not null references files (id),
  verified_by     uuid references users (id),
  verified_at     timestamptz
);
comment on table car_documents is '[cars] Ownership and import documents, verified by staff.';

create table viewing_bookings (
  id                uuid primary key default gen_random_uuid(),
  car_listing_id    uuid not null references car_listings (id),
  customer_user_id  uuid references users (id),
  name              text not null,
  phone             text not null check (phone ~ '^\+234[0-9]{10}$'),
  preferred_date    date not null,
  slot              text not null check (slot in ('morning', 'afternoon', 'late_afternoon')),
  notes             text,
  status            text not null default 'requested' check (status in ('requested', 'confirmed', 'completed', 'no_show', 'cancelled')),
  created_at        timestamptz not null default now(),
  updated_at        timestamptz not null default now()
);
comment on table viewing_bookings is '[cars] Requests to see a car (the Book a viewing form).';

create table financing_enquiries (
  id                   uuid primary key default gen_random_uuid(),
  car_listing_id       uuid not null references car_listings (id),
  customer_user_id     uuid references users (id),
  name                 text not null,
  phone                text not null check (phone ~ '^\+234[0-9]{10}$'),
  email                citext,
  monthly_income_band  text,
  down_payment_kobo    bigint check (down_payment_kobo >= 0),
  status               text not null default 'new' check (status in ('new', 'contacted', 'referred', 'closed')),
  created_at           timestamptz not null default now(),
  updated_at           timestamptz not null default now()
);
comment on table financing_enquiries is '[cars] Car financing enquiries, referred to finance partners.';

-- =============================================================================
-- 12. MARKETING
-- =============================================================================

create table promotions (
  id          uuid primary key default gen_random_uuid(),
  name        text not null,
  kind        text not null check (kind in ('percentage', 'fixed_amount', 'free_delivery', 'flash_price')),
  value       integer check (value > 0),
  starts_at   timestamptz not null,
  ends_at     timestamptz not null,
  status      text not null default 'draft' check (status in ('draft', 'scheduled', 'live', 'ended', 'cancelled')),
  created_by  uuid not null references users (id),
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now(),
  check (ends_at > starts_at),
  check (kind in ('free_delivery', 'flash_price') or value is not null),
  check (kind <> 'percentage' or value <= 90)
);
comment on table promotions is '[marketing] Campaigns and deals with real start/end times (the countdown reads ends_at; no fake timers).';

create table promotion_targets (
  id            uuid primary key default gen_random_uuid(),
  promotion_id  uuid not null references promotions (id) on delete cascade,
  category_id   uuid references categories (id),
  listing_id    uuid references listings (id),
  check ((category_id is null) <> (listing_id is null))
);
comment on table promotion_targets is '[marketing] What a promotion applies to (a category or a listing); none = sitewide.';

create table promotion_listing_prices (
  promotion_id  uuid not null references promotions (id) on delete cascade,
  listing_id    uuid not null references listings (id),
  price_kobo    bigint not null check (price_kobo > 0),
  stock_limit   integer check (stock_limit > 0),
  primary key (promotion_id, listing_id)
);
comment on table promotion_listing_prices is '[marketing] Flash-deal prices per listing, optionally limited to a quantity.';

create table coupons (
  id               uuid primary key default gen_random_uuid(),
  code             citext not null unique,
  promotion_id     uuid not null references promotions (id) on delete cascade,
  max_redemptions  integer check (max_redemptions > 0),
  per_user_limit   integer not null default 1 check (per_user_limit > 0),
  min_order_kobo   bigint check (min_order_kobo >= 0),
  created_at       timestamptz not null default now()
);
comment on table coupons is '[marketing] Codes customers enter at checkout to apply a promotion.';

create table coupon_redemptions (
  id             uuid primary key default gen_random_uuid(),
  coupon_id      uuid not null references coupons (id),
  order_id       uuid not null unique references orders (id),
  user_id        uuid references users (id),
  discount_kobo  bigint not null check (discount_kobo > 0),
  redeemed_at    timestamptz not null default now()
);
comment on table coupon_redemptions is '[marketing] Each coupon use (one per order) for limits and reporting.';

-- =============================================================================
-- 13. SUPPORT & RISK
-- =============================================================================

create table support_tickets (
  id                uuid primary key default gen_random_uuid(),
  ticket_number     text not null unique default next_ref('TKT-', 'ticket_number_seq'),
  customer_user_id  uuid references users (id),
  seller_id         uuid references sellers (id),
  order_id          uuid references orders (id),
  channel           text not null check (channel in ('web', 'app', 'whatsapp', 'email', 'phone')),
  topic             text not null,
  priority          text not null default 'normal' check (priority in ('low', 'normal', 'high', 'urgent')),
  status            text not null default 'open' check (status in ('open', 'pending_customer', 'resolved', 'closed')),
  assigned_to       uuid references staff_members (user_id),
  first_response_at timestamptz,
  resolved_at       timestamptz,
  created_at        timestamptz not null default now(),
  updated_at        timestamptz not null default now(),
  check (customer_user_id is not null or seller_id is not null)
);
comment on table support_tickets is '[support] Support conversations with customers or sellers, across channels.';

create table ticket_messages (
  id              uuid primary key default gen_random_uuid(),
  ticket_id       uuid not null references support_tickets (id) on delete cascade,
  author_user_id  uuid references users (id),
  author_kind     text not null check (author_kind in ('customer', 'seller', 'staff', 'system')),
  body            text not null,
  is_internal     boolean not null default false,
  created_at      timestamptz not null default now()
);
comment on table ticket_messages is '[support] Messages in a ticket; internal notes are hidden from customers.';

create table message_attachments (
  message_id  uuid not null references ticket_messages (id) on delete cascade,
  file_id     uuid not null references files (id),
  primary key (message_id, file_id)
);
comment on table message_attachments is '[support] Files attached to ticket messages.';

create table risk_cases (
  id            uuid primary key default gen_random_uuid(),
  kind          text not null check (kind in ('order', 'payment', 'device', 'seller', 'account')),
  subject_type  text not null,
  subject_id    uuid not null,
  reason        text not null,
  score         smallint check (score between 0 and 100),
  status        text not null default 'open' check (status in ('open', 'cleared', 'blocked')),
  assigned_to   uuid references staff_members (user_id),
  resolved_by   uuid references users (id),
  resolved_at   timestamptz,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  check (status = 'open' or (resolved_by is not null and resolved_at is not null))
);
comment on table risk_cases is '[support] Fraud and risk reviews of orders, payments, devices, sellers and accounts.';

create table device_blocklist (
  id          uuid primary key default gen_random_uuid(),
  imei        text not null unique check (imei ~ '^[0-9]{15}$'),
  reason      text not null check (reason in ('reported_stolen', 'fraud', 'counterfeit')),
  source      text not null,
  created_by  uuid references users (id),
  created_at  timestamptz not null default now()
);
comment on table device_blocklist is '[support] IMEIs that may not be sold, traded in or repaired (stolen, fraud, counterfeit).';

-- =============================================================================
-- 14. CONTENT
-- =============================================================================

create table cms_pages (
  id            uuid primary key default gen_random_uuid(),
  site          text not null check (site in ('corporate', 'market', 'wholesale', 'seller')),
  slug          text not null,
  title         text not null,
  body          jsonb not null default '[]'::jsonb,
  status        text not null default 'draft' check (status in ('draft', 'in_review', 'published')),
  published_at  timestamptz,
  updated_by    uuid references users (id),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  unique (site, slug),
  check (status <> 'published' or published_at is not null)
);
comment on table cms_pages is '[content] Editable pages on each site (about, legal, credit terms…), as content blocks.';

create table banners (
  id             uuid primary key default gen_random_uuid(),
  site           text not null check (site in ('corporate', 'market', 'wholesale', 'seller')),
  placement      text not null,
  title          text not null,
  image_file_id  uuid references files (id),
  link_url       text not null,
  starts_at      timestamptz,
  ends_at        timestamptz,
  position       integer not null default 0,
  status         text not null default 'draft' check (status in ('draft', 'live', 'ended')),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),
  check (ends_at is null or starts_at is null or ends_at > starts_at)
);
comment on table banners is '[content] Hero and promo banners per site and placement.';

create table help_articles (
  id            uuid primary key default gen_random_uuid(),
  site          text not null check (site in ('market', 'wholesale', 'seller')),
  topic         text not null,
  slug          text not null,
  title         text not null,
  body          jsonb not null default '[]'::jsonb,
  status        text not null default 'draft' check (status in ('draft', 'in_review', 'published')),
  published_at  timestamptz,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  unique (site, slug)
);
comment on table help_articles is '[content] Help centre articles and FAQs per site.';

-- =============================================================================
-- 15. PLATFORM
-- =============================================================================

create table audit_log (
  id             bigint generated always as identity primary key,
  actor_user_id  uuid references users (id),
  action         text not null,
  entity_type    text not null,
  entity_id      text not null,
  changes        jsonb,
  ip             inet,
  created_at     timestamptz not null default now()
);
comment on table audit_log is '[platform] Append-only record of who changed what (staff and seller actions, sensitive reads).';

create table notifications (
  id          uuid primary key default gen_random_uuid(),
  user_id     uuid not null references users (id) on delete cascade,
  channel     text not null check (channel in ('sms', 'email', 'push', 'in_app')),
  template    text not null,
  payload     jsonb not null default '{}'::jsonb,
  status      text not null default 'queued' check (status in ('queued', 'sent', 'failed', 'read')),
  sent_at     timestamptz,
  read_at     timestamptz,
  created_at  timestamptz not null default now()
);
comment on table notifications is '[platform] SMS, email, push and in-app notifications per user (staff bell, order updates).';
create index notifications_unread_idx on notifications (user_id, created_at desc) where status <> 'read' and channel = 'in_app';

create table outbox_events (
  id              bigint generated always as identity primary key,
  aggregate_type  text not null,
  aggregate_id    uuid not null,
  event_type      text not null,
  payload         jsonb not null,
  created_at      timestamptz not null default now(),
  published_at    timestamptz
);
comment on table outbox_events is '[platform] Transactional outbox: events written with the data change, then delivered (SMS, email, webhooks).';
create index outbox_events_unpublished_idx on outbox_events (id) where published_at is null;

create table idempotency_keys (
  key            text primary key,
  user_id        uuid references users (id),
  request_hash   bytea not null,
  response_code  smallint,
  response_body  jsonb,
  created_at     timestamptz not null default now()
);
comment on table idempotency_keys is '[platform] Remembers responses to retried POSTs (checkout, payments) so they run once.';

create table feature_flags (
  key          text primary key,
  description  text not null,
  enabled      boolean not null default false,
  rules        jsonb,
  updated_by   uuid references users (id),
  updated_at   timestamptz not null default now()
);
comment on table feature_flags is '[platform] Runtime feature switches (IT tools module).';

-- =============================================================================
-- CROSS-CUTTING: triggers, ledger balance, foreign-key indexes
-- =============================================================================

-- updated_at trigger on every table that has the column.
do $$
declare t record;
begin
  for t in
    select table_name from information_schema.columns
    where table_schema = 'public' and column_name = 'updated_at'
  loop
    execute format(
      'create trigger %I before update on %I for each row execute function set_updated_at()',
      t.table_name || '_set_updated_at', t.table_name);
  end loop;
end $$;

-- A journal's entries must sum to zero when the transaction commits.
create function assert_journal_balanced() returns trigger
language plpgsql as $$
declare total bigint;
begin
  select coalesce(sum(amount_kobo), 0) into total from ledger_entries where journal_id = new.journal_id;
  if total <> 0 then
    raise exception 'ledger journal % is not balanced (sum = %)', new.journal_id, total;
  end if;
  return null;
end $$;

create constraint trigger ledger_entries_balanced
  after insert or update on ledger_entries
  deferrable initially deferred
  for each row execute function assert_journal_balanced();

-- Ledger and audit rows are never edited or deleted.
create function forbid_change() returns trigger
language plpgsql as $$
begin
  raise exception '% is append-only', tg_table_name;
end $$;

create trigger ledger_entries_append_only before update or delete on ledger_entries
  for each row execute function forbid_change();
create trigger audit_log_append_only before update or delete on audit_log
  for each row execute function forbid_change();
create trigger stock_movements_append_only before update or delete on stock_movements
  for each row execute function forbid_change();

-- Index every single-column foreign key that isn't already the leading column of an index.
do $$
declare fk record;
begin
  for fk in
    select c.conrelid::regclass as tbl, a.attname as col
    from pg_constraint c
    join pg_attribute a on a.attrelid = c.conrelid and a.attnum = c.conkey[1]
    where c.contype = 'f' and array_length(c.conkey, 1) = 1
      and not exists (
        select 1 from pg_index i
        where i.indrelid = c.conrelid and i.indkey[0] = c.conkey[1]
      )
  loop
    execute format('create index %I on %s (%I)', fk.tbl::text || '_' || fk.col || '_idx', fk.tbl, fk.col);
  end loop;
end $$;

-- =============================================================================
-- SEED / REFERENCE DATA
-- =============================================================================

insert into nigerian_states (code, name) values
  ('AB', 'Abia'), ('AD', 'Adamawa'), ('AK', 'Akwa Ibom'), ('AN', 'Anambra'), ('BA', 'Bauchi'),
  ('BY', 'Bayelsa'), ('BE', 'Benue'), ('BO', 'Borno'), ('CR', 'Cross River'), ('DE', 'Delta'),
  ('EB', 'Ebonyi'), ('ED', 'Edo'), ('EK', 'Ekiti'), ('EN', 'Enugu'), ('FC', 'FCT (Abuja)'),
  ('GO', 'Gombe'), ('IM', 'Imo'), ('JI', 'Jigawa'), ('KD', 'Kaduna'), ('KN', 'Kano'),
  ('KT', 'Katsina'), ('KE', 'Kebbi'), ('KO', 'Kogi'), ('KW', 'Kwara'), ('LA', 'Lagos'),
  ('NA', 'Nasarawa'), ('NI', 'Niger'), ('OG', 'Ogun'), ('ON', 'Ondo'), ('OS', 'Osun'),
  ('OY', 'Oyo'), ('PL', 'Plateau'), ('RI', 'Rivers'), ('SO', 'Sokoto'), ('TA', 'Taraba'),
  ('YO', 'Yobe'), ('ZA', 'Zamfara');

-- TechShop itself, as the first-party seller.
insert into sellers (type, display_name, slug, status, state_code, city)
values ('first_party', 'TechShop', 'techshop', 'active', 'LA', 'Lagos');

-- Platform ledger accounts (seller payable accounts are created per seller).
insert into ledger_accounts (code, name, kind) values
  ('clearing:paystack', 'Paystack clearing', 'asset'),
  ('clearing:opay', 'OPay clearing', 'asset'),
  ('clearing:moniepoint', 'Moniepoint clearing', 'asset'),
  ('bank:operating', 'Operating bank account', 'asset'),
  ('receivable:b2b', 'B2B invoices receivable', 'asset'),
  ('revenue:first_party_sales', 'TechShop own-stock sales', 'revenue'),
  ('revenue:commission', 'Marketplace commission', 'revenue'),
  ('revenue:delivery', 'Delivery fees', 'revenue'),
  ('liability:vat', 'VAT payable', 'liability'),
  ('liability:customer_refunds', 'Refunds owed to customers', 'liability'),
  ('expense:payment_fees', 'Payment provider fees', 'expense');
