-- Marketing: promotions, flash deals (concurrency-safe counters), coupons and unique codes,
-- audiences, campaigns (A/B, holdout, dual approval), journeys and tracked links.
-- Plan: docs/messaging-marketing.md §4, docs/backend.md §4.4.

-- +goose Up
create table marketing.promotions (
  id          uuid primary key default gen_random_uuid(),
  name        text not null,
  kind        text not null check (kind in ('percentage', 'fixed_amount', 'free_delivery', 'flash_price')),
  value       integer check (value > 0),
  funded_by   text not null default 'techshop' check (funded_by in ('techshop', 'seller')),
  starts_at   timestamptz not null,
  ends_at     timestamptz not null,
  status      text not null default 'draft' check (status in ('draft', 'scheduled', 'live', 'ended', 'cancelled')),
  created_by  uuid not null references identity.users (id),
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now(),
  check (ends_at > starts_at),
  check (kind in ('free_delivery', 'flash_price') or value is not null),
  check (kind <> 'percentage' or value <= 90)
);
comment on table marketing.promotions is '[marketing] Deals with real start and end times (countdowns read ends_at; no fake timers).';

create table marketing.promotion_targets (
  id            uuid primary key default gen_random_uuid(),
  promotion_id  uuid not null references marketing.promotions (id) on delete cascade,
  category_id   uuid references catalog.categories (id),
  listing_id    uuid references catalog.listings (id),
  check ((category_id is null) <> (listing_id is null))
);
comment on table marketing.promotion_targets is '[marketing] What a promotion applies to (a category or a listing); none = sitewide.';

create table marketing.promotion_listing_prices (
  promotion_id    uuid not null references marketing.promotions (id) on delete cascade,
  listing_id      uuid not null references catalog.listings (id),
  price_kobo      bigint not null check (price_kobo > 0),
  stock_limit     integer check (stock_limit > 0),
  claimed         integer not null default 0 check (claimed >= 0),
  per_user_limit  integer check (per_user_limit > 0),
  primary key (promotion_id, listing_id),
  check (stock_limit is null or claimed <= stock_limit)
);
comment on table marketing.promotion_listing_prices is '[marketing] Flash-deal prices per listing; claimed can never exceed stock_limit (conditional update).';

create table marketing.flash_claims (
  id            uuid primary key default gen_random_uuid(),
  promotion_id  uuid not null,
  listing_id    uuid not null,
  user_id       uuid not null references identity.users (id),
  order_id      uuid not null references sales.orders (id) on delete cascade,
  quantity      integer not null check (quantity > 0),
  created_at    timestamptz not null default now(),
  unique (promotion_id, listing_id, order_id),
  foreign key (promotion_id, listing_id) references marketing.promotion_listing_prices (promotion_id, listing_id)
);
comment on table marketing.flash_claims is '[marketing] Who claimed flash-deal units, for per-user limits and release on cancellation.';
create index flash_claims_user_idx on marketing.flash_claims (promotion_id, listing_id, user_id);

create table marketing.coupons (
  id               uuid primary key default gen_random_uuid(),
  code             citext unique,
  promotion_id     uuid not null references marketing.promotions (id) on delete cascade,
  is_unique_codes  boolean not null default false,
  max_redemptions  integer check (max_redemptions > 0),
  redeemed_count   integer not null default 0 check (redeemed_count >= 0),
  per_user_limit   integer not null default 1 check (per_user_limit > 0),
  min_order_kobo   bigint check (min_order_kobo >= 0),
  created_at       timestamptz not null default now(),
  check (max_redemptions is null or redeemed_count <= max_redemptions),
  check (is_unique_codes or code is not null)
);
comment on table marketing.coupons is '[marketing] Coupons: one shared code, or many single-use codes (coupon_codes). Counters are concurrency-safe.';

create table marketing.coupon_codes (
  id                 uuid primary key default gen_random_uuid(),
  coupon_id          uuid not null references marketing.coupons (id) on delete cascade,
  code               citext not null unique,
  issued_to_user_id  uuid references identity.users (id),
  notification_id    uuid references messaging.notifications (id),
  expires_at         timestamptz,
  redeemed_at        timestamptz,
  created_at         timestamptz not null default now()
);
comment on table marketing.coupon_codes is '[marketing] Unique single-use codes issued to one person (e.g. CART-7KQ2M9); a leaked code works once.';

create table marketing.coupon_redemptions (
  id              uuid primary key default gen_random_uuid(),
  coupon_id       uuid not null references marketing.coupons (id),
  coupon_code_id  uuid unique references marketing.coupon_codes (id),
  order_id        uuid not null unique references sales.orders (id),
  user_id         uuid references identity.users (id),
  discount_kobo   bigint not null check (discount_kobo > 0),
  status          text not null default 'applied' check (status in ('applied', 'released')),
  redeemed_at     timestamptz not null default now()
);
comment on table marketing.coupon_redemptions is '[marketing] Each coupon use (one per order); released when the order is cancelled.';

create table marketing.segments (
  id               uuid primary key default gen_random_uuid(),
  name             text not null,
  rules            jsonb not null,
  created_by       uuid not null references identity.users (id),
  last_count       integer check (last_count >= 0),
  last_counted_at  timestamptz,
  created_at       timestamptz not null default now(),
  updated_at       timestamptz not null default now()
);
comment on table marketing.segments is '[marketing] Saved audiences: JSON rules compiled to SQL over an allowlist of fields.';

create table marketing.campaigns (
  id                  uuid primary key default gen_random_uuid(),
  name                text not null,
  status              text not null default 'draft' check (status in ('draft', 'in_review', 'scheduled', 'sending', 'paused',
                        'sent', 'cancelled')),
  segment_id          uuid not null references marketing.segments (id),
  exclusions          jsonb not null default '{}'::jsonb,
  channels            text[] not null check (cardinality(channels) >= 1 and channels <@ array['sms', 'email', 'push', 'whatsapp']::text[]),
  scheduled_at        timestamptz,
  holdout_pct         numeric(4, 2) not null default 5 check (holdout_pct between 0 and 50),
  ab_test             jsonb,
  cost_estimate_kobo  bigint check (cost_estimate_kobo >= 0),
  created_by          uuid not null references identity.users (id),
  approved_by         uuid references identity.users (id),
  approved_at         timestamptz,
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now(),
  -- The creator can't approve their own campaign.
  check (approved_by is distinct from created_by or approved_by is null),
  check (status <> 'scheduled' or scheduled_at is not null)
);
comment on table marketing.campaigns is '[marketing] One-off sends. Large or costly campaigns need approval by someone other than the creator.';

create table marketing.campaign_variants (
  id                   uuid primary key default gen_random_uuid(),
  campaign_id          uuid not null references marketing.campaigns (id) on delete cascade,
  label                text not null,
  channel              text not null check (channel in ('sms', 'email', 'push', 'whatsapp')),
  template_version_id  uuid references messaging.template_versions (id),
  share_pct            numeric(5, 2) not null check (share_pct > 0 and share_pct <= 100),
  is_winner            boolean not null default false,
  unique (campaign_id, label, channel)
);
comment on table marketing.campaign_variants is '[marketing] Content variants for A/B tests (winner by clicks).';

create table marketing.campaign_recipients (
  campaign_id  uuid not null references marketing.campaigns (id) on delete cascade,
  user_id      uuid not null references identity.users (id) on delete cascade,
  variant_id   uuid references marketing.campaign_variants (id),
  holdout      boolean not null default false,
  status       text not null default 'pending' check (status in ('pending', 'sent', 'skipped', 'failed')),
  primary key (campaign_id, user_id),
  check (not holdout or variant_id is null)
);
comment on table marketing.campaign_recipients is '[marketing] Recipient snapshot taken when sending starts (stable and auditable); holdout rows get nothing.';

create table marketing.journeys (
  id            uuid primary key default gen_random_uuid(),
  name          text not null,
  trigger       text not null,
  entry_rules   jsonb not null default '{}'::jsonb,
  reentry_days  integer check (reentry_days >= 0),
  goal_event    text,
  status        text not null default 'draft' check (status in ('draft', 'active', 'paused', 'archived')),
  created_by    uuid not null references identity.users (id),
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
comment on table marketing.journeys is '[marketing] Automations (abandoned cart, back in stock, welcome…): a trigger plus steps.';

create table marketing.journey_steps (
  id          uuid primary key default gen_random_uuid(),
  journey_id  uuid not null references marketing.journeys (id) on delete cascade,
  position    integer not null check (position >= 0),
  kind        text not null check (kind in ('wait', 'condition', 'send', 'issue_coupon', 'branch', 'exit')),
  config      jsonb not null default '{}'::jsonb,
  unique (journey_id, position)
);
comment on table marketing.journey_steps is '[marketing] Ordered steps of a journey.';

create table marketing.journey_enrollments (
  id            uuid primary key default gen_random_uuid(),
  journey_id    uuid not null references marketing.journeys (id) on delete cascade,
  user_id       uuid not null references identity.users (id) on delete cascade,
  current_step  integer not null default 0,
  next_run_at   timestamptz,
  status        text not null default 'active' check (status in ('active', 'exited', 'goal_reached', 'completed', 'failed')),
  exit_reason   text,
  context       jsonb not null default '{}'::jsonb,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  check (status <> 'active' or next_run_at is not null)
);
comment on table marketing.journey_enrollments is '[marketing] People inside a journey; a worker runs due steps every minute; goal events exit immediately.';
create index journey_enrollments_due_idx on marketing.journey_enrollments (next_run_at) where status = 'active';
create unique index journey_enrollments_one_active on marketing.journey_enrollments (journey_id, user_id) where status = 'active';

create table marketing.tracked_links (
  id               uuid primary key default gen_random_uuid(),
  url              text not null check (url ~ '^https://([a-z0-9-]+\.)*techshop\.ng(/|$)'),
  notification_id  uuid references messaging.notifications (id) on delete cascade,
  campaign_id      uuid references marketing.campaigns (id) on delete cascade,
  created_at       timestamptz not null default now()
);
comment on table marketing.tracked_links is '[marketing] Signed click-redirect targets (only techshop.ng URLs, so it can never be an open redirect).';

alter table messaging.notifications
  add constraint notifications_campaign_id_fkey foreign key (campaign_id) references marketing.campaigns (id),
  add constraint notifications_journey_enrollment_id_fkey foreign key (journey_enrollment_id) references marketing.journey_enrollments (id);

-- Indexes on foreign-key columns (every FK is indexed).
create index campaign_recipients_user_id_idx on marketing.campaign_recipients (user_id);
create index campaign_recipients_variant_id_idx on marketing.campaign_recipients (variant_id);
create index campaign_variants_template_version_id_idx on marketing.campaign_variants (template_version_id);
create index campaigns_approved_by_idx on marketing.campaigns (approved_by);
create index campaigns_created_by_idx on marketing.campaigns (created_by);
create index campaigns_segment_id_idx on marketing.campaigns (segment_id);
create index coupon_codes_coupon_id_idx on marketing.coupon_codes (coupon_id);
create index coupon_codes_issued_to_user_id_idx on marketing.coupon_codes (issued_to_user_id);
create index coupon_codes_notification_id_idx on marketing.coupon_codes (notification_id);
create index coupon_redemptions_coupon_id_idx on marketing.coupon_redemptions (coupon_id);
create index coupon_redemptions_user_id_idx on marketing.coupon_redemptions (user_id);
create index coupons_promotion_id_idx on marketing.coupons (promotion_id);
create index flash_claims_order_id_idx on marketing.flash_claims (order_id);
create index flash_claims_user_id_idx on marketing.flash_claims (user_id);
create index journey_enrollments_user_id_idx on marketing.journey_enrollments (user_id);
create index journeys_created_by_idx on marketing.journeys (created_by);
create index promotion_listing_prices_listing_id_idx on marketing.promotion_listing_prices (listing_id);
create index promotion_targets_category_id_idx on marketing.promotion_targets (category_id);
create index promotion_targets_listing_id_idx on marketing.promotion_targets (listing_id);
create index promotion_targets_promotion_id_idx on marketing.promotion_targets (promotion_id);
create index promotions_created_by_idx on marketing.promotions (created_by);
create index segments_created_by_idx on marketing.segments (created_by);
create index tracked_links_campaign_id_idx on marketing.tracked_links (campaign_id);
create index tracked_links_notification_id_idx on marketing.tracked_links (notification_id);

-- Keep updated_at current.
create trigger campaigns_set_updated_at before update on marketing.campaigns
  for each row execute function platform.set_updated_at();
create trigger journey_enrollments_set_updated_at before update on marketing.journey_enrollments
  for each row execute function platform.set_updated_at();
create trigger journeys_set_updated_at before update on marketing.journeys
  for each row execute function platform.set_updated_at();
create trigger promotions_set_updated_at before update on marketing.promotions
  for each row execute function platform.set_updated_at();
create trigger segments_set_updated_at before update on marketing.segments
  for each row execute function platform.set_updated_at();

-- +goose Down
alter table messaging.notifications
  drop constraint if exists notifications_journey_enrollment_id_fkey,
  drop constraint if exists notifications_campaign_id_fkey;
drop table if exists marketing.tracked_links, marketing.journey_enrollments, marketing.journey_steps,
  marketing.journeys, marketing.campaign_recipients, marketing.campaign_variants, marketing.campaigns,
  marketing.segments, marketing.coupon_redemptions, marketing.coupon_codes, marketing.coupons,
  marketing.flash_claims, marketing.promotion_listing_prices, marketing.promotion_targets,
  marketing.promotions cascade;
