-- Platform tables: files, audit log, transactional outbox, idempotency keys, feature flags,
-- public holidays, encryption keys. Plan: docs/backend.md §1.3, §2.3, §5.

-- +goose Up
create table platform.files (
  id             uuid primary key default gen_random_uuid(),
  storage_key    text not null unique,
  purpose        text not null check (purpose in ('product_image', 'car_image', 'kyc_document', 'delivery_proof', 'invoice_pdf',
                   'receipt_pdf', 'return_photo', 'ticket_attachment', 'privacy_export', 'car_document', 'inspection_report',
                   'brand_logo', 'banner', 'review_image', 'dispute_evidence', 'other')),
  status         text not null default 'pending' check (status in ('pending', 'ready', 'rejected')),
  original_name  text,
  content_type   text not null,
  size_bytes     bigint not null check (size_bytes > 0),
  checksum       text,
  width          integer check (width > 0),
  height         integer check (height > 0),
  blurhash       text,
  visibility     text not null default 'private' check (visibility in ('public', 'private')),
  uploaded_by    uuid references identity.users (id),
  deleted_at     timestamptz,
  created_at     timestamptz not null default now()
);
comment on table platform.files is '[platform] Uploaded files (photos, KYC documents, delivery proofs, PDFs). Bytes live in object storage via presigned uploads.';

alter table identity.privacy_requests
  add constraint privacy_requests_file_id_fkey foreign key (file_id) references platform.files (id);

create table platform.audit_log (
  id             bigint generated always as identity primary key,
  actor_user_id  uuid references identity.users (id),
  actor_kind     text not null default 'user' check (actor_kind in ('user', 'staff', 'seller', 'system', 'service')),
  action         text not null,
  entity_type    text not null,
  entity_id      text not null,
  changes        jsonb,
  request_id     text,
  ip             inet,
  user_agent     text,
  created_at     timestamptz not null default now()
);
comment on table platform.audit_log is '[platform] Append-only record of who changed what (staff and seller actions, sensitive reads such as ID reveals).';
create index audit_log_entity_idx on platform.audit_log (entity_type, entity_id, created_at desc);

create table platform.outbox_events (
  id              bigint generated always as identity primary key,
  aggregate_type  text not null,
  aggregate_id    uuid not null,
  event_type      text not null,
  payload         jsonb not null,
  trace_context   jsonb,
  created_at      timestamptz not null default now(),
  published_at    timestamptz
);
comment on table platform.outbox_events is '[platform] Transactional outbox: events written in the same transaction as the change, then relayed to jobs.';
create index outbox_events_unpublished_idx on platform.outbox_events (id) where published_at is null;

-- Wakes the outbox relay as soon as a transaction with new events commits.
-- +goose StatementBegin
create function platform.notify_outbox() returns trigger
language plpgsql as $$
begin
  perform pg_notify('outbox', '');
  return null;
end $$;
-- +goose StatementEnd
create trigger outbox_events_notify after insert on platform.outbox_events
  for each statement execute function platform.notify_outbox();

create table platform.idempotency_keys (
  scope          text not null,
  key            text not null,
  user_id        uuid references identity.users (id),
  method         text not null,
  path           text not null,
  request_hash   bytea not null,
  response_code  smallint,
  response_body  jsonb,
  locked_at      timestamptz,
  completed_at   timestamptz,
  expires_at     timestamptz not null,
  created_at     timestamptz not null default now(),
  primary key (scope, key)
);
comment on table platform.idempotency_keys is '[platform] Remembers responses to retried requests (checkout, payments, refunds) so they run once; locked_at detects in-flight duplicates.';
create index idempotency_keys_expires_idx on platform.idempotency_keys (expires_at);

create table platform.feature_flags (
  key          text primary key,
  description  text not null,
  enabled      boolean not null default false,
  rules        jsonb,
  updated_by   uuid references identity.users (id),
  updated_at   timestamptz not null default now()
);
comment on table platform.feature_flags is '[platform] Runtime feature switches (IT tools module), e.g. the live recommender service.';

create table platform.public_holidays (
  date  date primary key,
  name  text not null
);
comment on table platform.public_holidays is '[platform] Nigerian public holidays, so payout runs skip non-business days.';

create table platform.encryption_keys (
  id           uuid primary key default gen_random_uuid(),
  purpose      text not null check (purpose in ('kyc_id', 'bank_account', 'mfa_secret', 'trade_in_bank')),
  wrapped_dek  bytea not null,
  kms_key_id   text not null,
  created_at   timestamptz not null default now(),
  retired_at   timestamptz
);
comment on table platform.encryption_keys is '[platform] Envelope encryption: data keys wrapped by the KMS key. Only one active key per purpose.';
create unique index encryption_keys_one_active on platform.encryption_keys (purpose) where retired_at is null;

alter table identity.user_mfa_factors
  add constraint user_mfa_factors_encryption_key_id_fkey foreign key (encryption_key_id) references platform.encryption_keys (id);

-- Indexes on foreign-key columns (every FK is indexed).
create index audit_log_actor_user_id_idx on platform.audit_log (actor_user_id);
create index feature_flags_updated_by_idx on platform.feature_flags (updated_by);
create index files_uploaded_by_idx on platform.files (uploaded_by);
create index idempotency_keys_user_id_idx on platform.idempotency_keys (user_id);

-- Keep updated_at current.
create trigger feature_flags_set_updated_at before update on platform.feature_flags
  for each row execute function platform.set_updated_at();

-- +goose Down
alter table identity.user_mfa_factors drop constraint if exists user_mfa_factors_encryption_key_id_fkey;
alter table identity.privacy_requests drop constraint if exists privacy_requests_file_id_fkey;
drop table if exists platform.encryption_keys, platform.public_holidays, platform.feature_flags,
  platform.idempotency_keys, platform.outbox_events, platform.audit_log, platform.files cascade;
drop function if exists platform.notify_outbox();
