-- Identity & access: users, sessions, one-time codes, MFA, RBAC, staff, consents.
-- Plan: docs/identity-access.md. Addresses are created in 00005 (they can belong to a business).

-- +goose Up
create table identity.nigerian_states (
  code  text primary key,
  name  text not null unique
);
comment on table identity.nigerian_states is '[identity] Lookup: the 36 states and the FCT, used by addresses, zones and listings.';

create table identity.nigerian_lgas (
  id          uuid primary key default gen_random_uuid(),
  state_code  text not null references identity.nigerian_states (code),
  name        text not null,
  unique (state_code, name)
);
comment on table identity.nigerian_lgas is '[identity] Lookup: local government areas per state, so delivery fees match consistent spellings.';

create table identity.users (
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
  check (email is not null or phone is not null or status = 'deleted')
);
comment on table identity.users is '[identity] Every person who signs in: customers, seller staff, business buyers, riders and TechShop staff. Phone in E.164 (+234…).';

-- One row per refresh token. Rotation links tokens in the same family; reuse of a replaced
-- token revokes the whole family (docs/identity-access.md §5).
create table identity.user_sessions (
  id                  uuid primary key default gen_random_uuid(),
  user_id             uuid not null references identity.users (id) on delete cascade,
  family_id           uuid not null,
  refresh_token_hash  bytea not null unique,
  replaced_by         uuid references identity.user_sessions (id),
  aud                 text not null check (aud in ('market', 'wholesale', 'seller', 'staff', 'customer_app', 'logistics')),
  platform            text not null default 'web' check (platform in ('web', 'ios', 'android')),
  amr                 text[] not null default '{}',
  device_id           text,
  device_name         text,
  app_version         text,
  user_agent          text,
  ip                  inet,
  mfa_at              timestamptz,
  last_used_at        timestamptz not null default now(),
  idle_expires_at     timestamptz not null,
  expires_at          timestamptz not null,
  revoked_at          timestamptz,
  revoked_reason      text check (revoked_reason in ('logout', 'logout_all', 'reuse_detected', 'password_changed', 'admin', 'staff_exited', 'expired')),
  created_at          timestamptz not null default now(),
  check (idle_expires_at <= expires_at),
  check (revoked_at is null or revoked_reason is not null)
);
comment on table identity.user_sessions is '[identity] Refresh-token sessions per device and app. Tokens stored only as SHA-256 hashes; rotation tracked by family.';
create index user_sessions_family_idx on identity.user_sessions (family_id);
create index user_sessions_active_idx on identity.user_sessions (user_id, last_used_at desc) where revoked_at is null;

create table identity.verification_codes (
  id           uuid primary key default gen_random_uuid(),
  user_id      uuid references identity.users (id) on delete cascade,
  channel      text not null check (channel in ('sms', 'email')),
  destination  text not null,
  purpose      text not null check (purpose in ('signup', 'login', 'password_reset', 'phone_change', 'email_change', 'new_device', 'staff_mfa')),
  code_hash    bytea not null,
  attempts     smallint not null default 0 check (attempts between 0 and 5),
  ip           inet,
  user_agent   text,
  expires_at   timestamptz not null,
  consumed_at  timestamptz,
  created_at   timestamptz not null default now(),
  check (expires_at > created_at)
);
comment on table identity.verification_codes is '[identity] One-time codes sent by SMS or email. Stored as HMAC; at most 5 attempts. Delivery codes live on logistics.delivery_jobs.';
create index verification_codes_destination_idx on identity.verification_codes (destination, created_at desc);

create table identity.roles (
  id           uuid primary key default gen_random_uuid(),
  key          text not null unique check (key ~ '^[a-z][a-z0-9_]*$'),
  name         text not null,
  description  text,
  is_system    boolean not null default false,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now()
);
comment on table identity.roles is '[identity] Staff roles. Proposed templates in docs/identity-access.md §7.2; the owner approves the final list.';

create table identity.permissions (
  id           uuid primary key default gen_random_uuid(),
  key          text not null unique check (key ~ '^[a-z_-]+\.[a-z_]+$'),
  module       text not null,
  description  text,
  is_sensitive boolean not null default false
);
comment on table identity.permissions is '[identity] Actions per staff module (module.action). Sensitive ones need a recent MFA step-up.';

create table identity.role_permissions (
  role_id        uuid not null references identity.roles (id) on delete cascade,
  permission_id  uuid not null references identity.permissions (id) on delete cascade,
  primary key (role_id, permission_id)
);
comment on table identity.role_permissions is '[identity] Which permissions each role grants.';

create table identity.staff_members (
  user_id      uuid primary key references identity.users (id),
  employee_no  text not null unique,
  department   text not null,
  job_title    text not null,
  status       text not null default 'active' check (status in ('active', 'on_leave', 'suspended', 'exited')),
  hired_on     date,
  exited_on    date,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now(),
  check (exited_on is null or hired_on is null or exited_on >= hired_on),
  check (status <> 'exited' or exited_on is not null)
);
comment on table identity.staff_members is '[identity] TechShop employees. A staff member is a user with an employee record.';

create table identity.staff_roles (
  user_id     uuid not null references identity.staff_members (user_id) on delete cascade,
  role_id     uuid not null references identity.roles (id) on delete cascade,
  granted_by  uuid references identity.users (id),
  granted_at  timestamptz not null default now(),
  primary key (user_id, role_id),
  check (granted_by is distinct from user_id)
);
comment on table identity.staff_roles is '[identity] Roles held by each staff member. Nobody grants a role to themselves.';

create table identity.staff_role_scopes (
  user_id     uuid not null,
  role_id     uuid not null,
  scope_type  text not null check (scope_type in ('store', 'warehouse', 'zone')),
  scope_id    uuid not null,
  primary key (user_id, role_id, scope_type, scope_id),
  foreign key (user_id, role_id) references identity.staff_roles (user_id, role_id) on delete cascade
);
comment on table identity.staff_role_scopes is '[identity] Limits a staff role to specific stores, warehouses or delivery zones. No rows = unrestricted.';

create table identity.staff_invites (
  id             uuid primary key default gen_random_uuid(),
  staff_user_id  uuid not null references identity.staff_members (user_id) on delete cascade,
  token_hash     bytea not null unique,
  created_by     uuid not null references identity.users (id),
  expires_at     timestamptz not null,
  used_at        timestamptz,
  created_at     timestamptz not null default now()
);
comment on table identity.staff_invites is '[identity] Single-use invite links for new staff (set password, enrol TOTP).';

create table identity.user_mfa_factors (
  id                uuid primary key default gen_random_uuid(),
  user_id           uuid not null references identity.users (id) on delete cascade,
  kind              text not null default 'totp' check (kind in ('totp')),
  secret_encrypted  bytea not null,
  encryption_key_id uuid,
  last_used_step    bigint,
  confirmed_at      timestamptz,
  created_at        timestamptz not null default now()
);
comment on table identity.user_mfa_factors is '[identity] Authenticator-app (TOTP) secrets, envelope-encrypted. last_used_step stops a code being used twice.';
create unique index user_mfa_factors_one_totp on identity.user_mfa_factors (user_id) where kind = 'totp';

create table identity.user_recovery_codes (
  id          uuid primary key default gen_random_uuid(),
  user_id     uuid not null references identity.users (id) on delete cascade,
  code_hash   bytea not null,
  used_at     timestamptz,
  created_at  timestamptz not null default now(),
  unique (user_id, code_hash)
);
comment on table identity.user_recovery_codes is '[identity] Single-use MFA recovery codes, stored hashed.';

create table identity.webauthn_credentials (
  id             uuid primary key default gen_random_uuid(),
  user_id        uuid not null references identity.users (id) on delete cascade,
  credential_id  bytea not null unique,
  public_key     bytea not null,
  sign_count     bigint not null default 0 check (sign_count >= 0),
  transports     text[] not null default '{}',
  name           text,
  last_used_at   timestamptz,
  created_at     timestamptz not null default now()
);
comment on table identity.webauthn_credentials is '[identity] Passkeys (planned for v2; created now so there are no migration surprises later).';

create table identity.auth_events (
  id          bigint generated always as identity,
  user_id     uuid references identity.users (id),
  kind        text not null check (kind in (
                'otp_requested', 'otp_failed', 'login_succeeded', 'login_failed', 'mfa_enrolled', 'mfa_removed',
                'mfa_failed', 'step_up_required', 'step_up_succeeded', 'refresh_reuse_detected', 'session_revoked',
                'password_changed', 'phone_changed', 'email_changed', 'role_granted', 'role_revoked',
                'permission_changed', 'reveal_id', 'rate_limited', 'access_denied', 'staff_exited')),
  ip          inet,
  user_agent  text,
  device_id   text,
  state_code  text,
  meta        jsonb not null default '{}'::jsonb,
  created_at  timestamptz not null default now(),
  primary key (id, created_at)
) partition by range (created_at);
comment on table identity.auth_events is '[identity] Append-only security events (sign-ins, failures, MFA, token reuse, role changes). Partitioned monthly.';
create table identity.auth_events_default partition of identity.auth_events default;
create index auth_events_user_idx on identity.auth_events (user_id, created_at desc);

create table identity.auth_rate_limits (
  key           text not null,
  window_start  timestamptz not null,
  count         integer not null default 0 check (count >= 0),
  primary key (key, window_start)
);
comment on table identity.auth_rate_limits is '[identity] Window counters for OTP sends and sign-in attempts (per number, IP, device, global). No Redis needed.';

create table identity.service_clients (
  id           uuid primary key default gen_random_uuid(),
  name         text not null unique,
  secret_hash  bytea not null,
  scopes       text[] not null default '{}',
  status       text not null default 'active' check (status in ('active', 'disabled')),
  rotated_at   timestamptz not null default now(),
  created_at   timestamptz not null default now()
);
comment on table identity.service_clients is '[identity] Machine clients (the Python recommender) using client credentials on the internal listener.';

create table identity.consents (
  id              bigint generated always as identity primary key,
  user_id         uuid references identity.users (id),
  anonymous_id    uuid,
  purpose         text not null check (purpose in ('analytics', 'personalisation', 'marketing_sms', 'marketing_email',
                    'marketing_push', 'marketing_whatsapp', 'location_tracking', 'terms', 'privacy')),
  granted         boolean not null,
  policy_version  text not null,
  source          text not null check (source in ('web', 'app', 'unsubscribe_link', 'sms_stop', 'staff', 'import')),
  recorded_at     timestamptz not null default now(),
  check (user_id is not null or anonymous_id is not null)
);
comment on table identity.consents is '[identity] Append-only record of consent given or withdrawn (NDPA). The latest row per subject and purpose wins.';
create index consents_user_purpose_idx on identity.consents (user_id, purpose, recorded_at desc) where user_id is not null;
create index consents_anon_purpose_idx on identity.consents (anonymous_id, purpose, recorded_at desc) where anonymous_id is not null;

create table identity.privacy_requests (
  id            uuid primary key default gen_random_uuid(),
  user_id       uuid not null references identity.users (id),
  kind          text not null check (kind in ('export', 'delete')),
  status        text not null default 'requested' check (status in ('requested', 'cancelled', 'processing', 'completed', 'failed')),
  cancel_until  timestamptz,
  file_id       uuid,
  completed_at  timestamptz,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  check (status <> 'completed' or completed_at is not null)
);
comment on table identity.privacy_requests is '[identity] Data export and account deletion requests (deletion has a 7-day cancel window, then anonymisation).';

-- Indexes on foreign-key columns (every FK is indexed).
create index privacy_requests_file_id_idx on identity.privacy_requests (file_id);
create index privacy_requests_user_id_idx on identity.privacy_requests (user_id);
create index role_permissions_permission_id_idx on identity.role_permissions (permission_id);
create index staff_invites_created_by_idx on identity.staff_invites (created_by);
create index staff_invites_staff_user_id_idx on identity.staff_invites (staff_user_id);
create index staff_roles_granted_by_idx on identity.staff_roles (granted_by);
create index staff_roles_role_id_idx on identity.staff_roles (role_id);
create index user_mfa_factors_encryption_key_id_idx on identity.user_mfa_factors (encryption_key_id);
create index user_sessions_replaced_by_idx on identity.user_sessions (replaced_by);
create index verification_codes_user_id_idx on identity.verification_codes (user_id);
create index webauthn_credentials_user_id_idx on identity.webauthn_credentials (user_id);

-- Keep updated_at current.
create trigger privacy_requests_set_updated_at before update on identity.privacy_requests
  for each row execute function platform.set_updated_at();
create trigger roles_set_updated_at before update on identity.roles
  for each row execute function platform.set_updated_at();
create trigger staff_members_set_updated_at before update on identity.staff_members
  for each row execute function platform.set_updated_at();
create trigger users_set_updated_at before update on identity.users
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists identity.privacy_requests, identity.consents, identity.service_clients,
  identity.auth_rate_limits, identity.auth_events, identity.webauthn_credentials,
  identity.user_recovery_codes, identity.user_mfa_factors, identity.staff_invites,
  identity.staff_role_scopes, identity.staff_roles, identity.staff_members, identity.role_permissions,
  identity.permissions, identity.roles, identity.verification_codes, identity.user_sessions,
  identity.users, identity.nigerian_lgas, identity.nigerian_states cascade;
