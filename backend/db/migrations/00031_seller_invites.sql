-- Seller team invites (identity-access.md §7.3; docs/plan.md 2026-10-21 decision 4): an owner
-- or manager invites someone by phone or email to a role. The link is single use and expires
-- after 7 days (re-sending issues a new link); only its hash is stored. Finance members must
-- have an authenticator before they can join or use money screens.

-- +goose Up
create table sellers.seller_invites (
  id           uuid primary key default gen_random_uuid(),
  seller_id    uuid not null references sellers.sellers (id) on delete cascade,
  phone        text,
  email        text,
  role         text not null check (role in ('owner', 'manager', 'staff', 'finance')),
  token_hash   bytea not null unique,
  invited_by   uuid not null references identity.users (id),
  expires_at   timestamptz not null,
  accepted_by  uuid references identity.users (id),
  accepted_at  timestamptz,
  revoked_at   timestamptz,
  created_at   timestamptz not null default now(),
  check ((phone is null) <> (email is null)),
  check ((accepted_by is null) = (accepted_at is null))
);
comment on table sellers.seller_invites is '[sellers] Invitations to join a shop''s team (single-use link, 7 days, token hashed).';
create index seller_invites_seller_idx on sellers.seller_invites (seller_id, created_at desc);
create index seller_invites_invited_by_idx on sellers.seller_invites (invited_by);
create index seller_invites_accepted_by_idx on sellers.seller_invites (accepted_by);
-- One open invite per person per shop: re-sending replaces the link instead of stacking invites.
create unique index seller_invites_one_open on sellers.seller_invites (seller_id, coalesce(phone, email))
  where accepted_at is null and revoked_at is null;

alter table sellers.seller_invites enable row level security;
create policy seller_scope on sellers.seller_invites
  using (coalesce(current_setting('app.seller_id', true), '') = '' or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);

-- +goose Down
drop table sellers.seller_invites;
