-- Sellers (marketplace shops, KYC, payout bank accounts), business buyers (B2B) and addresses.
-- Plans: docs/trust-safety.md §2 (KYC), docs/identity-access.md §7.3 (member roles).

-- +goose Up
create table sellers.sellers (
  id                       uuid primary key default gen_random_uuid(),
  type                     text not null check (type in ('first_party', 'business', 'individual')),
  display_name             text not null,
  slug                     text not null unique check (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
  owner_user_id            uuid references identity.users (id),
  status                   text not null default 'pending_kyc' check (status in ('pending_kyc', 'active', 'suspended', 'closed')),
  state_code               text references identity.nigerian_states (code),
  city                     text,
  rating_avg               numeric(3, 2) check (rating_avg between 1 and 5),
  rating_count             integer not null default 0 check (rating_count >= 0),
  commission_override_bps  integer check (commission_override_bps between 0 and 10000),
  created_at               timestamptz not null default now(),
  updated_at               timestamptz not null default now(),
  check (type = 'first_party' or owner_user_id is not null)
);
comment on table sellers.sellers is '[sellers] Everyone who sells: TechShop itself (first_party), marketplace businesses and individuals.';
create unique index sellers_one_first_party on sellers.sellers ((true)) where type = 'first_party';

create table sellers.seller_members (
  seller_id   uuid not null references sellers.sellers (id) on delete cascade,
  user_id     uuid not null references identity.users (id) on delete cascade,
  role        text not null check (role in ('owner', 'manager', 'staff', 'finance')),
  created_at  timestamptz not null default now(),
  primary key (seller_id, user_id)
);
comment on table sellers.seller_members is '[sellers] Users who act for a seller in Seller Centre. The finance role needs TOTP.';
create index seller_members_user_idx on sellers.seller_members (user_id);

create table sellers.seller_kyc_submissions (
  id                     uuid primary key default gen_random_uuid(),
  seller_id              uuid not null references sellers.sellers (id) on delete cascade,
  status                 text not null default 'submitted' check (status in ('submitted', 'in_review', 'approved', 'rejected', 'needs_more_info')),
  id_type                text not null check (id_type in ('nin', 'nin_slip', 'passport', 'drivers_licence', 'voters_card')),
  id_number_encrypted    bytea not null,
  id_number_hmac         bytea not null,
  id_number_last4        text not null check (length(id_number_last4) = 4),
  encryption_key_id      uuid references platform.encryption_keys (id),
  cac_rc_number          text,
  tin                    text,
  verification_provider  text,
  verification_ref       text,
  verified_at            timestamptz,
  reviewed_by            uuid references identity.users (id),
  reviewed_at            timestamptz,
  rejection_reason       text,
  created_at             timestamptz not null default now(),
  updated_at             timestamptz not null default now(),
  check (status not in ('approved', 'rejected') or (reviewed_by is not null and reviewed_at is not null)),
  check (status <> 'rejected' or rejection_reason is not null)
);
comment on table sellers.seller_kyc_submissions is '[sellers] Identity and business verification. ID numbers encrypted (NDPA); the HMAC finds the same ID across sellers.';
create index seller_kyc_submissions_hmac_idx on sellers.seller_kyc_submissions (id_number_hmac);

create table sellers.kyc_documents (
  id             uuid primary key default gen_random_uuid(),
  submission_id  uuid not null references sellers.seller_kyc_submissions (id) on delete cascade,
  kind           text not null check (kind in ('id_front', 'id_back', 'cac_certificate', 'utility_bill', 'selfie')),
  file_id        uuid not null references platform.files (id),
  created_at     timestamptz not null default now()
);
comment on table sellers.kyc_documents is '[sellers] Documents uploaded with a KYC submission (private files; every view audited).';

create table sellers.seller_bank_accounts (
  id                        uuid primary key default gen_random_uuid(),
  seller_id                 uuid not null references sellers.sellers (id) on delete cascade,
  bank_code                 text not null,
  account_name              text not null,
  resolved_account_name     text,
  account_number_encrypted  bytea not null,
  account_number_hmac       bytea not null,
  account_number_last4      text not null check (length(account_number_last4) = 4),
  encryption_key_id         uuid references platform.encryption_keys (id),
  provider                  text check (provider in ('paystack', 'moniepoint', 'opay')),
  provider_recipient_code   text,
  verified_at               timestamptz,
  is_default                boolean not null default false,
  created_at                timestamptz not null default now(),
  updated_at                timestamptz not null default now()
);
comment on table sellers.seller_bank_accounts is '[sellers] Payout bank accounts (NUBAN encrypted, last 4 shown). A new account triggers a 24-hour payout hold.';
create unique index seller_bank_accounts_one_default on sellers.seller_bank_accounts (seller_id) where is_default;
create index seller_bank_accounts_hmac_idx on sellers.seller_bank_accounts (account_number_hmac);

create table b2b.businesses (
  id                  uuid primary key default gen_random_uuid(),
  legal_name          text not null,
  rc_number           text unique,
  tin                 text,
  type                text not null check (type in ('office_sme', 'school', 'reseller', 'government', 'ngo', 'healthcare', 'other')),
  size_band           text check (size_band in ('1-10', '11-50', '51-200', '201-1000', '1000+')),
  status              text not null default 'pending_verification' check (status in ('pending_verification', 'active', 'suspended')),
  account_manager_id  uuid references identity.staff_members (user_id),
  created_at          timestamptz not null default now(),
  updated_at          timestamptz not null default now()
);
comment on table b2b.businesses is '[b2b] Business buyers (wholesale site): trade pricing, VAT invoices, credit.';

create table b2b.business_members (
  business_id          uuid not null references b2b.businesses (id) on delete cascade,
  user_id              uuid not null references identity.users (id) on delete cascade,
  role                 text not null check (role in ('admin', 'buyer', 'approver')),
  approval_limit_kobo  bigint check (approval_limit_kobo >= 0),
  created_at           timestamptz not null default now(),
  primary key (business_id, user_id)
);
comment on table b2b.business_members is '[b2b] Users who buy or approve for a business. Buyers above approval_limit_kobo need an approver.';
create index business_members_user_idx on b2b.business_members (user_id);

create table b2b.credit_accounts (
  id           uuid primary key default gen_random_uuid(),
  business_id  uuid not null unique references b2b.businesses (id) on delete cascade,
  limit_kobo   bigint not null default 0 check (limit_kobo >= 0),
  currency     char(3) not null default 'NGN',
  term_days    integer not null default 30 check (term_days between 1 and 180),
  status       text not null default 'applied' check (status in ('applied', 'approved', 'on_hold', 'suspended', 'closed')),
  approved_by  uuid references identity.users (id),
  approved_at  timestamptz,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now(),
  check (status <> 'approved' or (approved_by is not null and approved_at is not null))
);
comment on table b2b.credit_accounts is '[b2b] Pay-on-invoice credit for approved businesses (limit and payment term).';

create table identity.addresses (
  id              uuid primary key default gen_random_uuid(),
  user_id         uuid references identity.users (id) on delete cascade,
  business_id     uuid references b2b.businesses (id) on delete cascade,
  label           text,
  recipient_name  text not null,
  phone           text not null check (phone ~ '^\+234[0-9]{10}$'),
  line1           text not null,
  line2           text,
  landmark        text,
  city            text not null,
  lga             text,
  state_code      text not null references identity.nigerian_states (code),
  latitude        numeric(9, 6) check (latitude between -90 and 90),
  longitude       numeric(9, 6) check (longitude between -180 and 180),
  is_default      boolean not null default false,
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  check ((user_id is null) <> (business_id is null))
);
comment on table identity.addresses is '[identity] Saved delivery and business addresses, owned by exactly one user or one business.';
create unique index addresses_one_default_per_user on identity.addresses (user_id) where is_default and user_id is not null;
create unique index addresses_one_default_per_business on identity.addresses (business_id) where is_default and business_id is not null;

-- Indexes on foreign-key columns (every FK is indexed).
create index businesses_account_manager_id_idx on b2b.businesses (account_manager_id);
create index credit_accounts_approved_by_idx on b2b.credit_accounts (approved_by);
create index addresses_state_code_idx on identity.addresses (state_code);
create index kyc_documents_file_id_idx on sellers.kyc_documents (file_id);
create index kyc_documents_submission_id_idx on sellers.kyc_documents (submission_id);
create index seller_bank_accounts_encryption_key_id_idx on sellers.seller_bank_accounts (encryption_key_id);
create index seller_kyc_submissions_encryption_key_id_idx on sellers.seller_kyc_submissions (encryption_key_id);
create index seller_kyc_submissions_reviewed_by_idx on sellers.seller_kyc_submissions (reviewed_by);
create index seller_kyc_submissions_seller_id_idx on sellers.seller_kyc_submissions (seller_id);
create index sellers_owner_user_id_idx on sellers.sellers (owner_user_id);
create index sellers_state_code_idx on sellers.sellers (state_code);

-- Keep updated_at current.
create trigger businesses_set_updated_at before update on b2b.businesses
  for each row execute function platform.set_updated_at();
create trigger credit_accounts_set_updated_at before update on b2b.credit_accounts
  for each row execute function platform.set_updated_at();
create trigger addresses_set_updated_at before update on identity.addresses
  for each row execute function platform.set_updated_at();
create trigger seller_bank_accounts_set_updated_at before update on sellers.seller_bank_accounts
  for each row execute function platform.set_updated_at();
create trigger seller_kyc_submissions_set_updated_at before update on sellers.seller_kyc_submissions
  for each row execute function platform.set_updated_at();
create trigger sellers_set_updated_at before update on sellers.sellers
  for each row execute function platform.set_updated_at();

-- +goose Down
drop table if exists identity.addresses, b2b.credit_accounts, b2b.business_members, b2b.businesses,
  sellers.seller_bank_accounts, sellers.kyc_documents, sellers.seller_kyc_submissions,
  sellers.seller_members, sellers.sellers cascade;
