-- Search merchandising and analytics. Plan: docs/search-catalogue.md §4–§5.

-- +goose Up
create table search.search_synonyms (
  id          uuid primary key default gen_random_uuid(),
  terms       text[] not null check (cardinality(terms) >= 1),
  target      text not null,
  kind        text not null check (kind in ('one_way', 'two_way')),
  active      boolean not null default true,
  created_by  uuid references identity.users (id),
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now()
);
comment on table search.search_synonyms is '[search] Query rewrites, e.g. tokunbo → UK-used, ps5 → Play 5. Reloaded instantly via NOTIFY.';

create table search.search_redirects (
  query_norm  text primary key,
  url         text not null check (url ~ '^/'),
  created_by  uuid references identity.users (id),
  expires_at  timestamptz,
  created_at  timestamptz not null default now()
);
comment on table search.search_redirects is '[search] Queries that go straight to a page (e.g. "cars" → /c/cars). Internal paths only.';

create table search.search_pins (
  id          uuid primary key default gen_random_uuid(),
  query_norm  text not null,
  product_id  uuid not null references catalog.products (id) on delete cascade,
  action      text not null check (action in ('pin', 'bury')),
  position    smallint check (position between 1 and 50),
  starts_at   timestamptz not null default now(),
  ends_at     timestamptz not null,
  created_by  uuid not null references identity.users (id),
  created_at  timestamptz not null default now(),
  check (ends_at > starts_at),
  check (action = 'bury' or position is not null)
);
comment on table search.search_pins is '[search] Time-boxed, audited pins or burials of a product for a query.';
create index search_pins_query_idx on search.search_pins (query_norm, ends_at);

create table search.search_queries (
  id             bigint generated always as identity,
  session_id     text,
  user_id        uuid references identity.users (id),
  query_norm     text not null,
  filters        jsonb not null default '{}'::jsonb,
  results_count  integer not null check (results_count >= 0),
  corrected_to   text,
  latency_ms     integer check (latency_ms >= 0),
  channel        text not null default 'retail' check (channel in ('retail', 'wholesale')),
  created_at     timestamptz not null default now(),
  primary key (id, created_at)
) partition by range (created_at);
comment on table search.search_queries is '[search] Every search with result count and latency (zero-result queue, insights). Partitioned monthly; 13-month retention.';
create table search.search_queries_default partition of search.search_queries default;
create index search_queries_zero_idx on search.search_queries (query_norm, created_at) where results_count = 0;

create table search.search_suggestions (
  prefix      text not null,
  suggestion  text not null,
  weight      real not null default 0,
  primary key (prefix, suggestion)
);
comment on table search.search_suggestions is '[search] Autocomplete entries built nightly from popular queries.';

create table search.catalogue_gaps (
  query_norm    text primary key,
  searches_30d  integer not null default 0 check (searches_30d >= 0),
  status        text not null default 'new' check (status in ('new', 'sourcing', 'wont_stock', 'resolved')),
  owner_id      uuid references identity.users (id),
  note          text,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table search.catalogue_gaps is '[search] Things customers search for that TechShop doesn''t sell yet; demand data for purchasing.';

-- Indexes on foreign-key columns (every FK is indexed).
create index catalogue_gaps_owner_id_idx on search.catalogue_gaps (owner_id);
create index search_pins_created_by_idx on search.search_pins (created_by);
create index search_pins_product_id_idx on search.search_pins (product_id);
create index search_queries_user_id_idx on search.search_queries (user_id);
create index search_redirects_created_by_idx on search.search_redirects (created_by);
create index search_synonyms_created_by_idx on search.search_synonyms (created_by);

-- Keep updated_at current.
create trigger catalogue_gaps_set_updated_at before update on search.catalogue_gaps
  for each row execute function platform.set_updated_at();
create trigger search_synonyms_set_updated_at before update on search.search_synonyms
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists search.catalogue_gaps, search.search_suggestions, search.search_queries,
  search.search_pins, search.search_redirects, search.search_synonyms cascade;
