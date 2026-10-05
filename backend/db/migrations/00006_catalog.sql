-- Catalogue: categories, brands, typed attributes, products, variants, images, listings (seller
-- offers), price tiers and history, saved items, compatibility, moderation and the search read model.
-- Plan: docs/search-catalogue.md. Conditions everywhere: new, uk_used, refurbished, open_box.

-- +goose Up
create table catalog.categories (
  id               uuid primary key default gen_random_uuid(),
  parent_id        uuid references catalog.categories (id),
  slug             text not null unique,
  name             text not null,
  description      text,
  kind             text not null default 'product' check (kind in ('product', 'vehicle')),
  repurchase_days  integer check (repurchase_days > 0),
  position         integer not null default 0,
  is_active        boolean not null default true,
  created_at       timestamptz not null default now(),
  updated_at       timestamptz not null default now(),
  check (parent_id is distinct from id)
);
comment on table catalog.categories is '[catalog] Category tree (Phones → Android…). kind = vehicle routes to cars; repurchase_days stops re-recommending phones.';

create table catalog.brands (
  id            uuid primary key default gen_random_uuid(),
  slug          text not null unique,
  name          text not null,
  logo_file_id  uuid references platform.files (id),
  created_at    timestamptz not null default now()
);
comment on table catalog.brands is '[catalog] Manufacturers and brands.';

create table catalog.attribute_definitions (
  id               uuid primary key default gen_random_uuid(),
  category_id      uuid not null references catalog.categories (id) on delete cascade,
  key              text not null check (key ~ '^[a-z][a-z0-9_]*$'),
  label            text not null,
  data_type        text not null check (data_type in ('text', 'number', 'boolean', 'enum')),
  unit             text,
  options          jsonb,
  is_required      boolean not null default false,
  is_filterable    boolean not null default false,
  is_variant_axis  boolean not null default false,
  position         integer not null default 0,
  unique (category_id, key),
  check (data_type <> 'enum' or jsonb_typeof(options) = 'array')
);
comment on table catalog.attribute_definitions is '[catalog] Spec fields per category (RAM, storage, GPU…): drive spec tables, filters and variant axes.';

create table catalog.products (
  id             uuid primary key default gen_random_uuid(),
  category_id    uuid not null references catalog.categories (id),
  brand_id       uuid references catalog.brands (id),
  slug           text not null unique,
  name           text not null,
  model          text,
  description    text,
  status         text not null default 'draft' check (status in ('draft', 'in_review', 'active', 'archived')),
  search_vector  tsvector generated always as (
                   setweight(to_tsvector('simple', coalesce(name, '') || ' ' || coalesce(model, '')), 'A') ||
                   setweight(to_tsvector('simple', coalesce(description, '')), 'C')
                 ) stored,
  created_by     uuid references identity.users (id),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now()
);
comment on table catalog.products is '[catalog] Catalogue entries at model level. TechShop owns titles, images and specs; sellers sell through listings on variants.';
create index products_search_idx on catalog.products using gin (search_vector);
create index products_name_trgm_idx on catalog.products using gin (name gin_trgm_ops);

create table catalog.product_attribute_values (
  product_id    uuid not null references catalog.products (id) on delete cascade,
  attribute_id  uuid not null references catalog.attribute_definitions (id) on delete cascade,
  value_text    text,
  value_number  numeric,
  value_bool    boolean,
  primary key (product_id, attribute_id),
  check (num_nonnulls(value_text, value_number, value_bool) = 1)
);
comment on table catalog.product_attribute_values is '[catalog] Spec values for a product (exactly one typed value per attribute).';

create table catalog.product_variants (
  id            uuid primary key default gen_random_uuid(),
  product_id    uuid not null references catalog.products (id) on delete cascade,
  sku           text not null unique,
  name          text not null,
  axis_values   jsonb not null default '{}'::jsonb,
  gtin          text,
  mpn           text,
  weight_grams  integer check (weight_grams > 0),
  length_mm     integer check (length_mm > 0),
  width_mm      integer check (width_mm > 0),
  height_mm     integer check (height_mm > 0),
  is_serialised boolean not null default false,
  is_active     boolean not null default true,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table catalog.product_variants is '[catalog] Purchasable configurations (256GB · Titanium). Serialised variants are tracked per unit (IMEI/serial).';
create index product_variants_gtin_idx on catalog.product_variants (gtin) where gtin is not null;

create table catalog.product_images (
  id          uuid primary key default gen_random_uuid(),
  product_id  uuid not null references catalog.products (id) on delete cascade,
  variant_id  uuid references catalog.product_variants (id) on delete cascade,
  file_id     uuid not null references platform.files (id),
  alt         text not null,
  position    integer not null default 0
);
comment on table catalog.product_images is '[catalog] Product photos (optionally per variant); alt text required.';

create table catalog.listings (
  id                  uuid primary key default gen_random_uuid(),
  seller_id           uuid not null references sellers.sellers (id),
  variant_id          uuid not null references catalog.product_variants (id),
  condition           text not null check (condition in ('new', 'uk_used', 'refurbished', 'open_box')),
  grade               text check (grade in ('A', 'B', 'C')),
  price_kobo          bigint not null check (price_kobo > 0),
  compare_at_kobo     bigint,
  currency            char(3) not null default 'NGN',
  fulfilled_by        text not null default 'techshop' check (fulfilled_by in ('techshop', 'seller')),
  seller_stock        integer check (seller_stock >= 0),
  min_order_quantity  integer check (min_order_quantity > 1),
  handling_days       smallint not null default 1 check (handling_days between 0 and 10),
  warranty_provider   text not null default 'manufacturer' check (warranty_provider in ('manufacturer', 'techshop', 'seller', 'none')),
  warranty_months     smallint not null default 0 check (warranty_months between 0 and 60),
  condition_notes     text,
  battery_health_pct  smallint check (battery_health_pct between 1 and 100),
  status              text not null default 'draft' check (status in ('draft', 'auto_check', 'in_review', 'changes_requested',
                        'active', 'paused', 'suspended', 'rejected')),
  risk_score          smallint check (risk_score between 0 and 100),
  published_at        timestamptz,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  unique (seller_id, variant_id, condition),
  check (compare_at_kobo is null or compare_at_kobo > price_kobo),
  check (condition = 'new' or condition_notes is not null),
  check (fulfilled_by = 'techshop' or seller_stock is not null),
  check (status <> 'active' or published_at is not null)
);
comment on table catalog.listings is '[catalog] A seller''s offer on a variant: condition, price, stock, warranty. TechShop stock is a first_party listing.';
create index listings_variant_status_idx on catalog.listings (variant_id, status);
create index listings_moderation_idx on catalog.listings (risk_score desc, created_at) where status = 'in_review';

create table catalog.price_tiers (
  id               uuid primary key default gen_random_uuid(),
  listing_id       uuid not null references catalog.listings (id) on delete cascade,
  min_quantity     integer not null check (min_quantity > 1),
  unit_price_kobo  bigint not null check (unit_price_kobo > 0),
  unique (listing_id, min_quantity)
);
comment on table catalog.price_tiers is '[catalog] Wholesale quantity breaks: unit price from min_quantity upward (business accounts only).';

create table catalog.listing_price_history (
  id               uuid primary key default gen_random_uuid(),
  listing_id       uuid not null references catalog.listings (id) on delete cascade,
  price_kobo       bigint not null check (price_kobo > 0),
  compare_at_kobo  bigint,
  changed_by       uuid references identity.users (id),
  changed_at       timestamptz not null default now()
);
comment on table catalog.listing_price_history is '[catalog] Every price change. Proves discounts are genuine (FCCPA 30-day maximum) and feeds price-drop alerts.';
create index listing_price_history_listing_idx on catalog.listing_price_history (listing_id, changed_at desc);

create table catalog.listing_reviews (
  id           uuid primary key default gen_random_uuid(),
  listing_id   uuid not null references catalog.listings (id) on delete cascade,
  reviewer_id  uuid references identity.users (id),
  decision     text not null check (decision in ('auto_passed', 'approved', 'changes_requested', 'rejected', 'suspended')),
  reasons      text[] not null default '{}',
  note         text,
  created_at   timestamptz not null default now(),
  check (decision = 'auto_passed' or reviewer_id is not null)
);
comment on table catalog.listing_reviews is '[catalog] Moderation decisions on listings (automatic checks and human reviewers), visible to the seller.';

create table catalog.product_suggestions (
  id            uuid primary key default gen_random_uuid(),
  seller_id     uuid not null references sellers.sellers (id),
  product_id    uuid references catalog.products (id),
  kind          text not null check (kind in ('new_product', 'edit')),
  payload       jsonb not null,
  status        text not null default 'pending' check (status in ('pending', 'accepted', 'rejected')),
  reviewed_by   uuid references identity.users (id),
  reviewed_at   timestamptz,
  created_at    timestamptz not null default now(),
  check (kind = 'new_product' or product_id is not null),
  check (status = 'pending' or reviewed_by is not null)
);
comment on table catalog.product_suggestions is '[catalog] Seller-proposed new products or edits, waiting for the catalogue team.';

create table catalog.image_hashes (
  product_image_id  uuid primary key references catalog.product_images (id) on delete cascade,
  phash             bigint not null
);
comment on table catalog.image_hashes is '[catalog] Perceptual hashes of product photos, to catch photos stolen from other sellers.';
create index image_hashes_phash_idx on catalog.image_hashes (phash);

create table catalog.saved_items (
  user_id     uuid not null references identity.users (id) on delete cascade,
  product_id  uuid not null references catalog.products (id) on delete cascade,
  created_at  timestamptz not null default now(),
  primary key (user_id, product_id)
);
comment on table catalog.saved_items is '[catalog] Customers'' saved products (the heart button).';

create table catalog.stock_alerts (
  user_id      uuid not null references identity.users (id) on delete cascade,
  product_id   uuid not null references catalog.products (id) on delete cascade,
  created_at   timestamptz not null default now(),
  notified_at  timestamptz,
  primary key (user_id, product_id)
);
comment on table catalog.stock_alerts is '[catalog] "Notify me when back in stock" requests.';

create table catalog.product_compatibility (
  id                    uuid primary key default gen_random_uuid(),
  accessory_product_id  uuid not null references catalog.products (id) on delete cascade,
  device_product_id     uuid references catalog.products (id) on delete cascade,
  rule                  jsonb,
  source                text not null check (source in ('manual', 'attribute_rule', 'inferred')),
  verified              boolean not null default false,
  created_at            timestamptz not null default now(),
  check (num_nonnulls(device_product_id, rule) = 1),
  check (device_product_id is distinct from accessory_product_id)
);
comment on table catalog.product_compatibility is '[catalog] Which accessories fit which devices (explicit pair or attribute rule), for "fits your phone" shelves.';

-- Search read model: one row per product × channel, rebuilt from the outbox within seconds.
create table catalog.product_offer_summary (
  product_id        uuid not null references catalog.products (id) on delete cascade,
  channel           text not null check (channel in ('retail', 'wholesale')),
  slug              text not null,
  title             text not null,
  category_ids      uuid[] not null,
  brand_id          uuid references catalog.brands (id),
  best_listing_id   uuid references catalog.listings (id) on delete set null,
  min_price_kobo    bigint check (min_price_kobo > 0),
  max_price_kobo    bigint check (max_price_kobo > 0),
  compare_at_kobo   bigint,
  promo_price_kobo  bigint,
  discount_pct      smallint check (discount_pct between 0 and 100),
  offer_count       integer not null default 0 check (offer_count >= 0),
  conditions        text[] not null default '{}',
  seller_types      text[] not null default '{}',
  in_stock          boolean not null default false,
  stock_band        text check (stock_band in ('out', 'low', 'in')),
  rating_avg        numeric(3, 2),
  rating_count      integer not null default 0,
  sales_30d         integer not null default 0,
  return_rate_90d   numeric(5, 4),
  attrs             jsonb not null default '{}'::jsonb,
  facet_keys        text[] not null default '{}',
  search_vector     tsvector not null,
  refreshed_at      timestamptz not null default now(),
  primary key (product_id, channel),
  check (max_price_kobo is null or min_price_kobo is null or max_price_kobo >= min_price_kobo)
);
comment on table catalog.product_offer_summary is '[catalog] Search read model: offers, buy box winner, facets and search vector per product and channel. Never used for pricing at checkout.';
create index product_offer_summary_search_idx on catalog.product_offer_summary using gin (search_vector);
create index product_offer_summary_title_trgm_idx on catalog.product_offer_summary using gin (title gin_trgm_ops);
create index product_offer_summary_attrs_idx on catalog.product_offer_summary using gin (attrs jsonb_path_ops);
create index product_offer_summary_facets_idx on catalog.product_offer_summary using gin (facet_keys);
create index product_offer_summary_categories_idx on catalog.product_offer_summary using gin (category_ids);
create index product_offer_summary_browse_idx on catalog.product_offer_summary (channel, in_stock, sales_30d desc);

-- Indexes on foreign-key columns (every FK is indexed).
create index brands_logo_file_id_idx on catalog.brands (logo_file_id);
create index categories_parent_id_idx on catalog.categories (parent_id);
create index listing_price_history_changed_by_idx on catalog.listing_price_history (changed_by);
create index listing_reviews_listing_id_idx on catalog.listing_reviews (listing_id);
create index listing_reviews_reviewer_id_idx on catalog.listing_reviews (reviewer_id);
create index product_attribute_values_attribute_id_idx on catalog.product_attribute_values (attribute_id);
create index product_compatibility_accessory_product_id_idx on catalog.product_compatibility (accessory_product_id);
create index product_compatibility_device_product_id_idx on catalog.product_compatibility (device_product_id);
create index product_images_file_id_idx on catalog.product_images (file_id);
create index product_images_product_id_idx on catalog.product_images (product_id);
create index product_images_variant_id_idx on catalog.product_images (variant_id);
create index product_offer_summary_best_listing_id_idx on catalog.product_offer_summary (best_listing_id);
create index product_offer_summary_brand_id_idx on catalog.product_offer_summary (brand_id);
create index product_suggestions_product_id_idx on catalog.product_suggestions (product_id);
create index product_suggestions_reviewed_by_idx on catalog.product_suggestions (reviewed_by);
create index product_suggestions_seller_id_idx on catalog.product_suggestions (seller_id);
create index product_variants_product_id_idx on catalog.product_variants (product_id);
create index products_brand_id_idx on catalog.products (brand_id);
create index products_category_id_idx on catalog.products (category_id);
create index products_created_by_idx on catalog.products (created_by);
create index saved_items_product_id_idx on catalog.saved_items (product_id);
create index stock_alerts_product_id_idx on catalog.stock_alerts (product_id);

-- Keep updated_at current.
create trigger categories_set_updated_at before update on catalog.categories
  for each row execute function platform.set_updated_at();
create trigger listings_set_updated_at before update on catalog.listings
  for each row execute function platform.set_updated_at();
create trigger product_variants_set_updated_at before update on catalog.product_variants
  for each row execute function platform.set_updated_at();
create trigger products_set_updated_at before update on catalog.products
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists catalog.product_offer_summary, catalog.product_compatibility, catalog.stock_alerts,
  catalog.saved_items, catalog.image_hashes, catalog.product_suggestions, catalog.listing_reviews,
  catalog.listing_price_history, catalog.price_tiers, catalog.listings, catalog.product_images,
  catalog.product_variants, catalog.product_attribute_values, catalog.products,
  catalog.attribute_definitions, catalog.brands, catalog.categories cascade;
