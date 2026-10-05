-- Messaging: the per-recipient message log (and in-app inbox), provider events, suppressions,
-- preferences, quiet hours, templates and push tokens. Plan: docs/messaging-marketing.md §3, §5–§6.

-- +goose Up
create table messaging.push_tokens (
  id            uuid primary key default gen_random_uuid(),
  user_id       uuid not null references identity.users (id) on delete cascade,
  app           text not null check (app in ('customer', 'logistics')),
  platform      text not null check (platform in ('ios', 'android')),
  token         text not null unique,
  provider      text not null default 'expo' check (provider in ('expo', 'fcm', 'apns')),
  app_version   text,
  locale        text,
  last_seen_at  timestamptz not null default now(),
  revoked_at    timestamptz,
  created_at    timestamptz not null default now()
);
comment on table messaging.push_tokens is '[messaging] Device push tokens per app; dead tokens are revoked by the receipts job.';
create index push_tokens_user_idx on messaging.push_tokens (user_id) where revoked_at is null;

create table messaging.message_templates (
  id          uuid primary key default gen_random_uuid(),
  key         text not null,
  channel     text not null check (channel in ('sms', 'email', 'push', 'in_app', 'whatsapp')),
  category    text not null check (category in ('security', 'delivery', 'orders', 'account', 'seller', 'b2b', 'deals',
                'recommendations', 'reminders')),
  locale      text not null default 'en',
  status      text not null default 'draft' check (status in ('draft', 'active', 'archived')),
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now(),
  unique (key, channel, locale)
);
comment on table messaging.message_templates is '[messaging] Marketing-authored templates (transactional ones live in code).';

create table messaging.template_versions (
  id            uuid primary key default gen_random_uuid(),
  template_id   uuid not null references messaging.message_templates (id) on delete cascade,
  version       integer not null check (version > 0),
  subject       text,
  blocks        jsonb not null,
  created_by    uuid not null references identity.users (id),
  published_at  timestamptz,
  created_at    timestamptz not null default now(),
  unique (template_id, version)
);
comment on table messaging.template_versions is '[messaging] Immutable versions of a template; each message records the version it was rendered from.';

create table messaging.notifications (
  id                     uuid primary key default gen_random_uuid(),
  user_id                uuid references identity.users (id) on delete cascade,
  channel                text not null check (channel in ('sms', 'email', 'push', 'in_app', 'whatsapp')),
  category               text not null check (category in ('security', 'delivery', 'orders', 'account', 'seller', 'b2b', 'deals',
                           'recommendations', 'reminders')),
  priority               text not null check (priority in ('critical', 'transactional', 'marketing')),
  template               text not null,
  template_version_id    uuid references messaging.template_versions (id),
  locale                 text not null default 'en',
  payload                jsonb not null default '{}'::jsonb,
  idempotency_key        text not null unique,
  campaign_id            uuid,
  journey_enrollment_id  uuid,
  to_hash                bytea,
  provider               text,
  provider_message_id    text,
  cost_kobo              bigint check (cost_kobo >= 0),
  status                 text not null default 'queued' check (status in ('queued', 'skipped', 'sent', 'delivered', 'bounced', 'failed', 'read')),
  skip_reason            text check (skip_reason in ('no_consent', 'suppressed', 'capped', 'quiet_hours', 'preference', 'invalid_address', 'duplicate')),
  error                  text,
  scheduled_for          timestamptz,
  sent_at                timestamptz,
  delivered_at           timestamptz,
  clicked_at             timestamptz,
  read_at                timestamptz,
  created_at             timestamptz not null default now(),
  check (status <> 'skipped' or skip_reason is not null),
  check (user_id is not null or to_hash is not null)
);
comment on table messaging.notifications is '[messaging] Every message to a recipient, any channel. Doubles as the in-app inbox. Bodies trimmed after 90 days.';
create index notifications_inbox_idx on messaging.notifications (user_id, created_at desc) where channel = 'in_app';
create index notifications_unread_idx on messaging.notifications (user_id) where channel = 'in_app' and read_at is null;
create index notifications_provider_idx on messaging.notifications (provider, provider_message_id) where provider_message_id is not null;

alter table sales.orders
  add constraint orders_attributed_notification_id_fkey foreign key (attributed_notification_id) references messaging.notifications (id);

create table messaging.message_events (
  id               bigint generated always as identity,
  notification_id  uuid not null references messaging.notifications (id) on delete cascade,
  kind             text not null check (kind in ('delivered', 'bounced', 'complained', 'opened', 'clicked', 'unsubscribed', 'failed')),
  meta             jsonb not null default '{}'::jsonb,
  occurred_at      timestamptz not null default now(),
  primary key (id, occurred_at)
) partition by range (occurred_at);
comment on table messaging.message_events is '[messaging] Append-only provider callbacks and clicks per message. Partitioned monthly; 13-month retention.';
create table messaging.message_events_default partition of messaging.message_events default;
create index message_events_notification_idx on messaging.message_events (notification_id);

create table messaging.message_suppressions (
  id            uuid primary key default gen_random_uuid(),
  channel       text not null check (channel in ('sms', 'email', 'push', 'whatsapp')),
  address_hash  bytea not null,
  reason        text not null check (reason in ('hard_bounce', 'complaint', 'unsubscribed', 'invalid', 'manual')),
  created_at    timestamptz not null default now(),
  unique (channel, address_hash)
);
comment on table messaging.message_suppressions is '[messaging] Addresses never to send to on a channel (bounces, complaints, unsubscribes).';

create table messaging.notification_preferences (
  user_id   uuid not null references identity.users (id) on delete cascade,
  category  text not null check (category in ('orders', 'seller', 'deals', 'recommendations', 'reminders')),
  channel   text not null check (channel in ('sms', 'email', 'push', 'whatsapp')),
  enabled   boolean not null,
  updated_at timestamptz not null default now(),
  primary key (user_id, category, channel)
);
comment on table messaging.notification_preferences is '[messaging] Per category and channel choices. Security and delivery messages cannot be turned off, so they are not listed.';

create table messaging.notification_settings (
  user_id      uuid primary key references identity.users (id) on delete cascade,
  quiet_start  time not null default '21:00',
  quiet_end    time not null default '08:00',
  quiet_hours  boolean not null default true,
  timezone     text not null default 'Africa/Lagos',
  updated_at   timestamptz not null default now()
);
comment on table messaging.notification_settings is '[messaging] Quiet hours per user (marketing only).';

-- Indexes on foreign-key columns (every FK is indexed).
create index notifications_campaign_id_idx on messaging.notifications (campaign_id);
create index notifications_journey_enrollment_id_idx on messaging.notifications (journey_enrollment_id);
create index notifications_template_version_id_idx on messaging.notifications (template_version_id);
create index template_versions_created_by_idx on messaging.template_versions (created_by);

-- Keep updated_at current.
create trigger message_templates_set_updated_at before update on messaging.message_templates
  for each row execute function platform.set_updated_at();
create trigger notification_preferences_set_updated_at before update on messaging.notification_preferences
  for each row execute function platform.set_updated_at();
create trigger notification_settings_set_updated_at before update on messaging.notification_settings
  for each row execute function platform.set_updated_at();

-- +goose Down
alter table sales.orders drop constraint if exists orders_attributed_notification_id_fkey;
drop table if exists messaging.notification_settings, messaging.notification_preferences,
  messaging.message_suppressions, messaging.message_events, messaging.notifications,
  messaging.template_versions, messaging.message_templates, messaging.push_tokens cascade;
