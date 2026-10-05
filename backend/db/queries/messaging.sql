-- Messaging queries (docs/messaging-marketing.md §3, §5–§6).

-- name: MessagingInsertNotification :one
insert into messaging.notifications (user_id, channel, category, priority, template, template_version_id, locale, payload,
  idempotency_key, campaign_id, journey_enrollment_id, to_hash, scheduled_for)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
on conflict (idempotency_key) do nothing
returning *;

-- name: MessagingGetNotification :one
select * from messaging.notifications where id = $1;

-- name: MessagingMarkSent :exec
update messaging.notifications
set status = 'sent', provider = $2, provider_message_id = $3, cost_kobo = $4, sent_at = now(), payload = payload || @extra::jsonb
where id = $1;

-- name: MessagingMarkSkipped :exec
update messaging.notifications set status = 'skipped', skip_reason = $2 where id = $1;

-- name: MessagingMarkFailed :exec
update messaging.notifications set status = 'failed', error = $2 where id = $1;

-- name: MessagingReschedule :exec
update messaging.notifications set scheduled_for = $2 where id = $1;

-- name: MessagingMarkDeliveredByProvider :one
update messaging.notifications set status = 'delivered', delivered_at = now()
where provider = $1 and provider_message_id = $2 and status in ('sent', 'queued')
returning id;

-- name: MessagingMarkClicked :exec
update messaging.notifications set clicked_at = coalesce(clicked_at, now()) where id = $1;

-- name: MessagingInsertEvent :exec
insert into messaging.message_events (notification_id, kind, meta) values ($1, $2, $3);

-- name: MessagingListInbox :many
select * from messaging.notifications
where user_id = $1 and channel = 'in_app'
  and (sqlc.narg('before')::timestamptz is null or created_at < sqlc.narg('before'))
order by created_at desc
limit $2;

-- name: MessagingUnreadCount :one
select count(*)::int from messaging.notifications where user_id = $1 and channel = 'in_app' and read_at is null;

-- name: MessagingMarkRead :execrows
update messaging.notifications set read_at = now(), status = 'read' where id = $1 and user_id = $2 and channel = 'in_app';

-- name: MessagingMarkAllRead :execrows
update messaging.notifications set read_at = now(), status = 'read' where user_id = $1 and channel = 'in_app' and read_at is null;

-- name: MessagingIsSuppressed :one
select exists (select 1 from messaging.message_suppressions where channel = $1 and address_hash = $2) as suppressed;

-- name: MessagingSuppress :exec
insert into messaging.message_suppressions (channel, address_hash, reason) values ($1, $2, $3)
on conflict (channel, address_hash) do nothing;

-- name: MessagingListSuppressions :many
select * from messaging.message_suppressions order by created_at desc limit $1;

-- name: MessagingGetPreferences :many
select * from messaging.notification_preferences where user_id = $1;

-- name: MessagingSetPreference :exec
insert into messaging.notification_preferences (user_id, category, channel, enabled) values ($1, $2, $3, $4)
on conflict (user_id, category, channel) do update set enabled = excluded.enabled, updated_at = now();

-- name: MessagingGetSettings :one
select * from messaging.notification_settings where user_id = $1;

-- name: MessagingUpsertSettings :one
insert into messaging.notification_settings (user_id, quiet_start, quiet_end, quiet_hours) values ($1, $2, $3, $4)
on conflict (user_id) do update set quiet_start = excluded.quiet_start, quiet_end = excluded.quiet_end, quiet_hours = excluded.quiet_hours, updated_at = now()
returning *;

-- name: MessagingCountMarketingToday :one
select count(*)::int from messaging.notifications
where user_id = $1 and channel = $2 and priority = 'marketing' and status in ('sent', 'delivered', 'read') and sent_at >= $3;

-- name: MessagingUpsertPushToken :one
insert into messaging.push_tokens (user_id, app, platform, token, app_version, locale)
values ($1, $2, $3, $4, $5, $6)
on conflict (token) do update set user_id = excluded.user_id, app = excluded.app, platform = excluded.platform,
  app_version = excluded.app_version, locale = excluded.locale, last_seen_at = now(), revoked_at = null
returning *;

-- name: MessagingActivePushTokens :many
select * from messaging.push_tokens where user_id = $1 and revoked_at is null and (sqlc.narg('app')::text is null or app = sqlc.narg('app'));

-- name: MessagingRevokePushToken :exec
update messaging.push_tokens set revoked_at = now() where token = $1;

-- name: MessagingRevokeUserPushTokens :exec
update messaging.push_tokens set revoked_at = now() where user_id = $1 and revoked_at is null;

-- name: MessagingDevRecent :many
-- Development only: recent messages to an address, so tests and local dev can read codes.
select id, channel, template, payload, status, created_at from messaging.notifications
where payload->>'to' = $1 order by created_at desc limit 10;

-- name: MessagingTrimBodies :execrows
update messaging.notifications set payload = payload - 'body' - 'data'
where created_at < now() - interval '90 days' and payload ? 'body';

-- name: MessagingListTemplates :many
select t.*, (select max(version) from messaging.template_versions v where v.template_id = t.id)::int as latest_version
from messaging.message_templates t order by key, channel;

-- name: MessagingCreateTemplate :one
insert into messaging.message_templates (key, channel, category, locale, status) values ($1, $2, $3, $4, 'draft') returning *;

-- name: MessagingAddTemplateVersion :one
insert into messaging.template_versions (template_id, version, subject, blocks, created_by, published_at)
values ($1, coalesce((select max(version) from messaging.template_versions where template_id = $1), 0) + 1, $2, $3, $4, now())
returning *;

-- name: MessagingGetTemplateVersion :one
select * from messaging.template_versions where id = $1;
