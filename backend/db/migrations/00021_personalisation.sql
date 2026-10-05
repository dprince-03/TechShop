-- Personalisation: behaviour events (with consent), identity links, and the recommendation model
-- outputs imported by Go from the Python batch job (Python never connects to the database).
-- Plan: docs/recommendations.md §4–§6.

-- +goose Up
create table personalisation.user_events (
  id              bigint generated always as identity,
  occurred_at     timestamptz not null,
  received_at     timestamptz not null default now(),
  anonymous_id    uuid not null,
  user_id         uuid references identity.users (id),
  session_id      text,
  event_type      text not null check (event_type in ('product_view', 'list_impression', 'rec_impression', 'rec_click',
                    'search', 'search_click', 'add_to_cart', 'remove_from_cart', 'save', 'unsave', 'begin_checkout',
                    'purchase', 'car_view', 'car_enquiry', 'share')),
  app             text not null check (app in ('market', 'wholesale', 'customer_app', 'server')),
  surface         text,
  product_id      uuid references catalog.products (id) on delete set null,
  listing_id      uuid references catalog.listings (id) on delete set null,
  car_listing_id  uuid references cars.car_listings (id) on delete set null,
  category_id     uuid references catalog.categories (id) on delete set null,
  query_norm      text,
  results_count   integer check (results_count >= 0),
  position        smallint check (position >= 0),
  rec_request_id  text,
  strategy        text,
  state_code      text,
  device_class    text check (device_class in ('mobile', 'tablet', 'desktop')),
  properties      jsonb not null default '{}'::jsonb,
  primary key (id, received_at)
) partition by range (received_at);
comment on table personalisation.user_events is '[personalisation] Behaviour events from consented visitors (views, clicks, carts, searches). Partitioned monthly.';
create table personalisation.user_events_default partition of personalisation.user_events default;
create index user_events_received_brin on personalisation.user_events using brin (received_at);
create index user_events_subject_idx on personalisation.user_events (anonymous_id, occurred_at);

create table personalisation.identity_links (
  anonymous_id  uuid primary key,
  user_id       uuid not null references identity.users (id) on delete cascade,
  linked_at     timestamptz not null default now()
);
comment on table personalisation.identity_links is '[personalisation] Joins a guest''s anonymous history to their account at sign-in.';

create table personalisation.rec_model_versions (
  id            uuid primary key default gen_random_uuid(),
  model_kind    text not null default 'hybrid' check (model_kind in ('hybrid', 'popularity', 'content', 'cooccurrence', 'als', 'bpr',
                  'item2vec', 'two_tower', 'sasrec', 'ranker')),
  status        text not null default 'staged' check (status in ('staged', 'active', 'retired', 'failed')),
  trained_at    timestamptz not null,
  data_cutoff   timestamptz not null,
  code_version  text not null,
  metrics       jsonb not null default '{}'::jsonb,
  row_counts    jsonb not null default '{}'::jsonb,
  activated_at  timestamptz,
  created_at    timestamptz not null default now(),
  check (status <> 'active' or activated_at is not null)
);
comment on table personalisation.rec_model_versions is '[personalisation] Imported model versions; one active at a time, the previous kept for one-click rollback.';
create unique index rec_model_versions_one_active on personalisation.rec_model_versions (model_kind) where status = 'active';

create table personalisation.rec_item_neighbours (
  model_version_id      uuid not null references personalisation.rec_model_versions (id) on delete cascade,
  kind                  text not null check (kind in ('similar', 'bought_together', 'also_viewed')),
  product_id            uuid not null references catalog.products (id) on delete cascade,
  neighbour_product_id  uuid not null references catalog.products (id) on delete cascade,
  score                 real not null,
  rank                  smallint not null check (rank > 0),
  primary key (model_version_id, kind, product_id, rank),
  check (neighbour_product_id <> product_id)
);
comment on table personalisation.rec_item_neighbours is '[personalisation] Similar, bought-together and also-viewed products per product.';

create table personalisation.rec_subject_candidates (
  model_version_id  uuid not null references personalisation.rec_model_versions (id) on delete cascade,
  user_id           uuid not null references identity.users (id) on delete cascade,
  product_id        uuid not null references catalog.products (id) on delete cascade,
  score             real not null,
  rank              smallint not null check (rank > 0),
  primary key (model_version_id, user_id, rank)
);
comment on table personalisation.rec_subject_candidates is '[personalisation] "For you" candidates per consented user.';

create table personalisation.rec_item_vectors (
  model_version_id  uuid not null references personalisation.rec_model_versions (id) on delete cascade,
  product_id        uuid not null references catalog.products (id) on delete cascade,
  vector            real[] not null check (cardinality(vector) between 8 and 512),
  primary key (model_version_id, product_id)
);
comment on table personalisation.rec_item_vectors is '[personalisation] Item embeddings used by Go for live session re-ranking.';

create table personalisation.rec_popularity (
  id           uuid primary key default gen_random_uuid(),
  computed_at  timestamptz not null,
  "window"     text not null check ("window" in ('7d', '30d')),
  category_id  uuid references catalog.categories (id) on delete cascade,
  state_code   text references identity.nigerian_states (code),
  product_id   uuid not null references catalog.products (id) on delete cascade,
  score        real not null,
  rank         smallint not null check (rank > 0)
);
comment on table personalisation.rec_popularity is '[personalisation] Trending products by window, category and state (the last-resort fallback for every shelf).';
create index rec_popularity_lookup_idx on personalisation.rec_popularity ("window", category_id, state_code, rank);

create table personalisation.rec_overrides (
  id                 uuid primary key default gen_random_uuid(),
  surface            text not null,
  anchor_product_id  uuid references catalog.products (id) on delete cascade,
  product_id         uuid not null references catalog.products (id) on delete cascade,
  action             text not null check (action in ('pin', 'block')),
  starts_at          timestamptz not null default now(),
  ends_at            timestamptz,
  created_by         uuid not null references identity.users (id),
  created_at         timestamptz not null default now(),
  check (ends_at is null or ends_at > starts_at)
);
comment on table personalisation.rec_overrides is '[personalisation] Staff merchandising pins and blocks per shelf (audited, time-boxed).';

-- Staging copies: imports land here, are validated, then swapped in by activating the version.
create table personalisation.rec_staging_item_neighbours (like personalisation.rec_item_neighbours including all);
comment on table personalisation.rec_staging_item_neighbours is '[personalisation] Staging copy of rec_item_neighbours for validating imports.';
create table personalisation.rec_staging_subject_candidates (like personalisation.rec_subject_candidates including all);
comment on table personalisation.rec_staging_subject_candidates is '[personalisation] Staging copy of rec_subject_candidates for validating imports.';
create table personalisation.rec_staging_item_vectors (like personalisation.rec_item_vectors including all);
comment on table personalisation.rec_staging_item_vectors is '[personalisation] Staging copy of rec_item_vectors for validating imports.';
create table personalisation.rec_staging_popularity (like personalisation.rec_popularity including all);
comment on table personalisation.rec_staging_popularity is '[personalisation] Staging copy of rec_popularity for validating imports.';

-- Indexes on foreign-key columns (every FK is indexed).
create index identity_links_user_id_idx on personalisation.identity_links (user_id);
create index rec_item_neighbours_neighbour_product_id_idx on personalisation.rec_item_neighbours (neighbour_product_id);
create index rec_item_neighbours_product_id_idx on personalisation.rec_item_neighbours (product_id);
create index rec_item_vectors_product_id_idx on personalisation.rec_item_vectors (product_id);
create index rec_overrides_anchor_product_id_idx on personalisation.rec_overrides (anchor_product_id);
create index rec_overrides_created_by_idx on personalisation.rec_overrides (created_by);
create index rec_overrides_product_id_idx on personalisation.rec_overrides (product_id);
create index rec_popularity_category_id_idx on personalisation.rec_popularity (category_id);
create index rec_popularity_product_id_idx on personalisation.rec_popularity (product_id);
create index rec_popularity_state_code_idx on personalisation.rec_popularity (state_code);
create index rec_subject_candidates_product_id_idx on personalisation.rec_subject_candidates (product_id);
create index rec_subject_candidates_user_id_idx on personalisation.rec_subject_candidates (user_id);
create index user_events_car_listing_id_idx on personalisation.user_events (car_listing_id);
create index user_events_category_id_idx on personalisation.user_events (category_id);
create index user_events_listing_id_idx on personalisation.user_events (listing_id);
create index user_events_product_id_idx on personalisation.user_events (product_id);
create index user_events_user_id_idx on personalisation.user_events (user_id);

-- +goose Down
drop table if exists personalisation.rec_staging_popularity, personalisation.rec_staging_item_vectors,
  personalisation.rec_staging_subject_candidates, personalisation.rec_staging_item_neighbours,
  personalisation.rec_overrides, personalisation.rec_popularity, personalisation.rec_item_vectors,
  personalisation.rec_subject_candidates, personalisation.rec_item_neighbours,
  personalisation.rec_model_versions, personalisation.identity_links, personalisation.user_events cascade;
