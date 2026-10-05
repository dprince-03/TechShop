-- Identity & access queries (docs/identity-access.md).

-- name: IdentityGetUser :one
select * from identity.users where id = $1;

-- name: IdentityGetUserByPhone :one
select * from identity.users where phone = $1;

-- name: IdentityGetUserByEmail :one
select * from identity.users where email = $1;

-- name: IdentityCreateUser :one
insert into identity.users (email, phone, password_hash, first_name, last_name, phone_verified_at, email_verified_at)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: IdentityUpdateProfile :one
update identity.users
set first_name = coalesce(sqlc.narg('first_name'), first_name),
    last_name = coalesce(sqlc.narg('last_name'), last_name),
    marketing_opt_in = coalesce(sqlc.narg('marketing_opt_in'), marketing_opt_in)
where id = $1
returning *;

-- name: IdentitySetPassword :exec
update identity.users set password_hash = $2 where id = $1;

-- name: IdentitySetPhone :exec
update identity.users set phone = $2, phone_verified_at = now() where id = $1;

-- name: IdentityMarkPhoneVerified :exec
update identity.users set phone_verified_at = coalesce(phone_verified_at, now()) where id = $1;

-- name: IdentityTouchLogin :exec
update identity.users set last_login_at = now() where id = $1;

-- name: IdentityAnonymiseUser :exec
update identity.users
set email = null, phone = null, password_hash = null, first_name = 'Deleted', last_name = 'user',
    status = 'deleted', marketing_opt_in = false
where id = $1;

-- name: IdentityListUsers :many
select * from identity.users
where (sqlc.narg('q')::text is null or email ilike '%' || sqlc.narg('q') || '%' or phone like '%' || sqlc.narg('q') || '%'
       or (first_name || ' ' || last_name) ilike '%' || sqlc.narg('q') || '%')
  and (sqlc.narg('after_created')::timestamptz is null or (created_at, id) < (sqlc.narg('after_created'), sqlc.narg('after_id')::uuid))
order by created_at desc, id desc
limit $1;

-- name: IdentityCreateCode :one
insert into identity.verification_codes (id, user_id, channel, destination, purpose, code_hash, ip, user_agent, expires_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *;

-- name: IdentityGetCode :one
select * from identity.verification_codes where id = $1 for update;

-- name: IdentityBumpCodeAttempts :one
update identity.verification_codes set attempts = attempts + 1 where id = $1 returning attempts;

-- name: IdentityConsumeCode :exec
update identity.verification_codes set consumed_at = now() where id = $1 and consumed_at is null;

-- name: IdentityCreateSession :one
insert into identity.user_sessions (user_id, family_id, refresh_token_hash, aud, platform, amr, device_id, device_name,
  app_version, user_agent, ip, mfa_at, idle_expires_at, expires_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
returning *;

-- name: IdentityGetSessionByToken :one
select * from identity.user_sessions where refresh_token_hash = $1 for update;

-- name: IdentityGetSession :one
select * from identity.user_sessions where id = $1;

-- name: IdentitySessionActive :one
select exists (
  select 1 from identity.user_sessions s
  where s.id = $1 and s.revoked_at is null and s.expires_at > now()
) as active;

-- name: IdentityMarkSessionReplaced :exec
update identity.user_sessions set replaced_by = $2, last_used_at = now() where id = $1;

-- name: IdentityRevokeFamily :execrows
update identity.user_sessions set revoked_at = now(), revoked_reason = $2
where family_id = $1 and revoked_at is null;

-- name: IdentityRevokeSession :execrows
update identity.user_sessions s set revoked_at = now(), revoked_reason = $3
where s.family_id = (select x.family_id from identity.user_sessions x where x.id = $1 and x.user_id = $2) and s.revoked_at is null;

-- name: IdentityRevokeOtherSessions :execrows
update identity.user_sessions set revoked_at = now(), revoked_reason = 'logout_all'
where user_id = $1 and family_id <> $2 and revoked_at is null;

-- name: IdentityRevokeAllSessions :execrows
update identity.user_sessions set revoked_at = now(), revoked_reason = $2
where user_id = $1 and revoked_at is null;

-- name: IdentityListActiveSessions :many
-- One row per device (the newest token of each family).
select distinct on (family_id) id, family_id, aud, platform, device_name, user_agent, ip, last_used_at, created_at
from identity.user_sessions
where user_id = $1 and revoked_at is null and replaced_by is null and expires_at > now()
order by family_id, created_at desc;

-- name: IdentitySetSessionMFA :exec
update identity.user_sessions set mfa_at = now() where family_id = $1 and revoked_at is null;

-- name: IdentityGetTOTPFactor :one
select * from identity.user_mfa_factors where user_id = $1 and kind = 'totp';

-- name: IdentityUpsertPendingTOTP :one
insert into identity.user_mfa_factors (user_id, kind, secret_encrypted)
values ($1, 'totp', $2)
on conflict (user_id) where kind = 'totp' do update set secret_encrypted = excluded.secret_encrypted, confirmed_at = null, last_used_step = null
returning *;

-- name: IdentityConfirmTOTP :exec
update identity.user_mfa_factors set confirmed_at = now(), last_used_step = $2 where id = $1;

-- name: IdentityUseTOTPStep :execrows
-- Records the step; fails (0 rows) if this step or a later one was already used (no code reuse).
update identity.user_mfa_factors set last_used_step = $2
where id = $1 and (last_used_step is null or last_used_step < $2);

-- name: IdentityDeleteTOTP :exec
delete from identity.user_mfa_factors where user_id = $1;

-- name: IdentityDeleteRecoveryCodes :exec
delete from identity.user_recovery_codes where user_id = $1;

-- name: IdentityInsertRecoveryCode :exec
insert into identity.user_recovery_codes (user_id, code_hash) values ($1, $2);

-- name: IdentityUseRecoveryCode :execrows
update identity.user_recovery_codes set used_at = now()
where user_id = $1 and code_hash = $2 and used_at is null;

-- name: IdentityInsertAuthEvent :exec
insert into identity.auth_events (user_id, kind, ip, user_agent, device_id, meta)
values ($1, $2, $3, $4, $5, $6);

-- name: IdentityListAuthEvents :many
select * from identity.auth_events
where (sqlc.narg('user_id')::uuid is null or user_id = sqlc.narg('user_id'))
  and (sqlc.narg('kind')::text is null or kind = sqlc.narg('kind'))
  and (sqlc.narg('before_id')::bigint is null or id < sqlc.narg('before_id'))
order by id desc
limit $1;

-- name: IdentityListStates :many
select * from identity.nigerian_states order by name;

-- name: IdentityListAddresses :many
select * from identity.addresses where user_id = $1 order by is_default desc, created_at desc;

-- name: IdentityGetAddress :one
select * from identity.addresses where id = $1 and user_id = $2;

-- name: IdentityCreateAddress :one
insert into identity.addresses (user_id, business_id, label, recipient_name, phone, line1, line2, landmark, city, lga, state_code, latitude, longitude, is_default)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
returning *;

-- name: IdentityUpdateAddress :one
update identity.addresses
set label = $3, recipient_name = $4, phone = $5, line1 = $6, line2 = $7, landmark = $8, city = $9, lga = $10, state_code = $11
where id = $1 and user_id = $2
returning *;

-- name: IdentityClearDefaultAddress :exec
update identity.addresses set is_default = false where user_id = $1 and is_default;

-- name: IdentitySetDefaultAddress :execrows
update identity.addresses set is_default = true where id = $1 and user_id = $2;

-- name: IdentityDeleteAddress :execrows
delete from identity.addresses where id = $1 and user_id = $2;

-- name: IdentityInsertConsent :exec
insert into identity.consents (user_id, anonymous_id, purpose, granted, policy_version, source)
values ($1, $2, $3, $4, $5, $6);

-- name: IdentityLatestConsents :many
select distinct on (purpose) purpose, granted, policy_version, source, recorded_at
from identity.consents
where (user_id = sqlc.narg('user_id') or anonymous_id = sqlc.narg('anonymous_id'))
order by purpose, recorded_at desc;

-- name: IdentityHasConsent :one
select coalesce((
  select granted from identity.consents
  where (user_id = sqlc.narg('user_id') or anonymous_id = sqlc.narg('anonymous_id')) and purpose = @purpose
  order by recorded_at desc limit 1), false)::bool as granted;

-- name: IdentityCreatePrivacyRequest :one
insert into identity.privacy_requests (user_id, kind, cancel_until) values ($1, $2, $3) returning *;

-- name: IdentityListPrivacyRequests :many
select * from identity.privacy_requests where user_id = $1 order by created_at desc;

-- name: IdentityDuePrivacyDeletions :many
select * from identity.privacy_requests
where kind = 'delete' and status = 'requested' and cancel_until < now()
limit 50;

-- name: IdentitySetPrivacyRequestStatus :exec
update identity.privacy_requests set status = $2, completed_at = case when $2 = 'completed' then now() else completed_at end where id = $1;

-- name: IdentityCancelPrivacyRequest :execrows
update identity.privacy_requests set status = 'cancelled' where id = $1 and user_id = $2 and status = 'requested';

-- name: IdentityGetServiceClient :one
select * from identity.service_clients where name = $1 and status = 'active';

-- name: IdentityUpsertServiceClient :exec
insert into identity.service_clients (name, secret_hash, scopes) values ($1, $2, $3)
on conflict (name) do update set secret_hash = excluded.secret_hash, scopes = excluded.scopes, rotated_at = now();

-- RBAC

-- name: IdentityUserPermissionKeys :many
select distinct p.key
from identity.staff_members sm
join identity.staff_roles sr on sr.user_id = sm.user_id
join identity.role_permissions rp on rp.role_id = sr.role_id
join identity.permissions p on p.id = rp.permission_id
where sm.user_id = $1 and sm.status = 'active';

-- name: IdentityUpsertPermission :exec
insert into identity.permissions (key, module, description, is_sensitive) values ($1, $2, $3, $4)
on conflict (key) do update set module = excluded.module, description = excluded.description, is_sensitive = excluded.is_sensitive;

-- name: IdentityListPermissions :many
select * from identity.permissions order by module, key;

-- name: IdentityListRoles :many
select r.*, coalesce(array_agg(p.key order by p.key) filter (where p.key is not null), '{}')::text[] as permission_keys
from identity.roles r
left join identity.role_permissions rp on rp.role_id = r.id
left join identity.permissions p on p.id = rp.permission_id
group by r.id
order by r.name;

-- name: IdentityGetRoleByKey :one
select * from identity.roles where key = $1;

-- name: IdentityUpsertRole :one
insert into identity.roles (key, name, description, is_system) values ($1, $2, $3, $4)
on conflict (key) do update set name = excluded.name, description = excluded.description
returning *;

-- name: IdentityClearRolePermissions :exec
delete from identity.role_permissions where role_id = $1;

-- name: IdentityAddRolePermissions :exec
insert into identity.role_permissions (role_id, permission_id)
select $1, id from identity.permissions where key = any(@keys::text[])
on conflict do nothing;

-- name: IdentityUsersWithRole :many
select user_id from identity.staff_roles where role_id = $1;

-- name: IdentityCreateStaffMember :one
insert into identity.staff_members (user_id, employee_no, department, job_title, hired_on)
values ($1, $2, $3, $4, $5)
returning *;

-- name: IdentityGetStaffMember :one
select * from identity.staff_members where user_id = $1;

-- name: IdentityListStaffMembers :many
select sm.*, u.first_name, u.last_name, u.email,
       coalesce(array_agg(r.key order by r.key) filter (where r.key is not null), '{}')::text[] as role_keys,
       exists (select 1 from identity.user_mfa_factors f where f.user_id = sm.user_id and f.confirmed_at is not null) as mfa_enabled
from identity.staff_members sm
join identity.users u on u.id = sm.user_id
left join identity.staff_roles sr on sr.user_id = sm.user_id
left join identity.roles r on r.id = sr.role_id
group by sm.user_id, u.id
order by u.first_name, u.last_name;

-- name: IdentityGrantRole :exec
insert into identity.staff_roles (user_id, role_id, granted_by) values ($1, $2, $3) on conflict do nothing;

-- name: IdentityRevokeRole :execrows
delete from identity.staff_roles where user_id = $1 and role_id = $2;

-- name: IdentityRevokeAllRoles :execrows
delete from identity.staff_roles where user_id = $1;

-- name: IdentitySetStaffStatus :exec
update identity.staff_members set status = $2, exited_on = case when $2 = 'exited' then current_date else exited_on end where user_id = $1;

-- name: IdentityCreateInvite :one
insert into identity.staff_invites (staff_user_id, token_hash, created_by, expires_at) values ($1, $2, $3, $4) returning *;

-- name: IdentityGetInvite :one
select * from identity.staff_invites where token_hash = $1 and used_at is null and expires_at > now() for update;

-- name: IdentityUseInvite :exec
update identity.staff_invites set used_at = now() where id = $1;

-- name: IdentityHasDeviceSession :one
select exists (select 1 from identity.user_sessions where user_id = $1 and device_id = $2) as known;

-- name: IdentityMemberships :one
-- Read-only summary of what an account may act as (documented cross-module read for /me).
select
  coalesce((select json_agg(json_build_object('sellerId', s.id, 'name', s.display_name, 'role', sm.role, 'status', s.status))
            from sellers.seller_members sm join sellers.sellers s on s.id = sm.seller_id where sm.user_id = $1), '[]')::jsonb as sellers,
  coalesce((select json_agg(json_build_object('businessId', b.id, 'name', b.legal_name, 'role', bm.role, 'status', b.status))
            from b2b.business_members bm join b2b.businesses b on b.id = bm.business_id where bm.user_id = $1), '[]')::jsonb as businesses,
  exists (select 1 from logistics.riders r where r.user_id = $1) as is_rider,
  exists (select 1 from identity.staff_members st where st.user_id = $1 and st.status = 'active') as is_staff,
  exists (select 1 from identity.user_mfa_factors f where f.user_id = $1 and f.confirmed_at is not null) as mfa_enabled;

-- name: IdentitySeedTOTP :exec
-- Dev seed only: a confirmed authenticator for a seeded staff account.
insert into identity.user_mfa_factors (user_id, kind, secret_encrypted, confirmed_at)
values ($1, 'totp', $2, now())
on conflict (user_id) where kind = 'totp' do update set secret_encrypted = excluded.secret_encrypted, confirmed_at = now(), last_used_step = null;
