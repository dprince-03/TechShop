-- Platform queries: outbox, audit log, idempotency keys, rate limits, feature flags, files.

-- name: PlatformPing :one
select 1::int as ok;

-- name: PlatformInsertOutboxEvent :exec
insert into platform.outbox_events (aggregate_type, aggregate_id, event_type, payload, trace_context)
values ($1, $2, $3, $4, $5);

-- name: PlatformLockUnpublishedEvents :many
select id, aggregate_type, aggregate_id, event_type, payload, created_at
from platform.outbox_events
where published_at is null
order by id
limit $1
for update skip locked;

-- name: PlatformMarkEventsPublished :exec
update platform.outbox_events set published_at = now() where id = any(@ids::bigint[]);

-- name: PlatformOutboxLag :one
select count(*)::int as pending, coalesce(extract(epoch from now() - min(created_at)), 0)::float8 as oldest_seconds
from platform.outbox_events where published_at is null;

-- name: PlatformInsertAudit :exec
insert into platform.audit_log (actor_user_id, actor_kind, action, entity_type, entity_id, changes, request_id, ip, user_agent)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: PlatformListAudit :many
select * from platform.audit_log
where (sqlc.narg('entity_type')::text is null or entity_type = sqlc.narg('entity_type'))
  and (sqlc.narg('actor_user_id')::uuid is null or actor_user_id = sqlc.narg('actor_user_id'))
  and (sqlc.narg('before_id')::bigint is null or id < sqlc.narg('before_id'))
order by id desc
limit $1;

-- name: PlatformGetIdempotencyKey :one
select * from platform.idempotency_keys where scope = $1 and key = $2;

-- name: PlatformInsertIdempotencyKey :one
insert into platform.idempotency_keys (scope, key, user_id, method, path, request_hash, locked_at, expires_at)
values ($1, $2, $3, $4, $5, $6, now(), $7)
on conflict (scope, key) do nothing
returning *;

-- name: PlatformCompleteIdempotencyKey :exec
update platform.idempotency_keys
set response_code = $3, response_body = $4, completed_at = now(), locked_at = null
where scope = $1 and key = $2;

-- name: PlatformReleaseIdempotencyKey :exec
delete from platform.idempotency_keys where scope = $1 and key = $2 and completed_at is null;

-- name: PlatformDeleteExpiredIdempotencyKeys :execrows
delete from platform.idempotency_keys where expires_at < now();

-- name: PlatformHitRateLimit :one
-- Atomically counts one hit in the window and returns the new count.
insert into identity.auth_rate_limits (key, window_start, count)
values ($1, $2, 1)
on conflict (key, window_start) do update set count = identity.auth_rate_limits.count + 1
returning count;

-- name: PlatformCountRateLimit :one
select coalesce(sum(count), 0)::int from identity.auth_rate_limits where key = $1 and window_start >= $2;

-- name: PlatformDeleteOldRateLimits :execrows
delete from identity.auth_rate_limits where window_start < now() - interval '2 days';

-- name: PlatformListFeatureFlags :many
select * from platform.feature_flags order by key;

-- name: PlatformGetFeatureFlag :one
select * from platform.feature_flags where key = $1;

-- name: PlatformSetFeatureFlag :one
update platform.feature_flags set enabled = $2, rules = $3, updated_by = $4 where key = $1 returning *;

-- name: PlatformCreateFile :one
insert into platform.files (storage_key, purpose, status, original_name, content_type, size_bytes, visibility, uploaded_by)
values ($1, $2, 'pending', $3, $4, $5, $6, $7)
returning *;

-- name: PlatformGetFile :one
select * from platform.files where id = $1 and deleted_at is null;

-- name: PlatformMarkFileReady :one
update platform.files set status = 'ready', checksum = $2, width = $3, height = $4
where id = $1 and status = 'pending' returning *;

-- name: PlatformListPublicHolidays :many
select * from platform.public_holidays where date between $1 and $2 order by date;

-- name: PlatformListJobsSummary :many
select state::text as state, count(*)::int as count from river.river_job group by state order by state;
