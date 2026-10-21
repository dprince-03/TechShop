-- Row-level security on seller-scoped tables (docs/database/RLS-POLICIES.md). Seller Centre
-- requests set app.seller_id for their transaction (set_config(..., true)); rows of any other
-- shop are then invisible and can't be written, even if a query forgets its seller filter.
-- Requests without that setting (staff, customers, jobs) are unaffected. The table owner bypasses
-- RLS, which is why the API logs in as a techshop_rw member (00024).

-- +goose Up
alter table sellers.sellers enable row level security;
create policy seller_scope on sellers.sellers
  using (coalesce(current_setting('app.seller_id', true), '') = '' or id = nullif(current_setting('app.seller_id', true), '')::uuid);

alter table sellers.seller_members enable row level security;
create policy seller_scope on sellers.seller_members
  using (coalesce(current_setting('app.seller_id', true), '') = '' or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);

alter table sellers.seller_kyc_submissions enable row level security;
create policy seller_scope on sellers.seller_kyc_submissions
  using (coalesce(current_setting('app.seller_id', true), '') = '' or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);

alter table sellers.kyc_documents enable row level security;
create policy seller_scope on sellers.kyc_documents
  using (coalesce(current_setting('app.seller_id', true), '') = '' or exists (
    select 1 from sellers.seller_kyc_submissions k
    where k.id = submission_id and k.seller_id = nullif(current_setting('app.seller_id', true), '')::uuid));

alter table sellers.seller_bank_accounts enable row level security;
create policy seller_scope on sellers.seller_bank_accounts
  using (coalesce(current_setting('app.seller_id', true), '') = '' or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid);

-- +goose Down
drop policy seller_scope on sellers.seller_bank_accounts;
alter table sellers.seller_bank_accounts disable row level security;
drop policy seller_scope on sellers.kyc_documents;
alter table sellers.kyc_documents disable row level security;
drop policy seller_scope on sellers.seller_kyc_submissions;
alter table sellers.seller_kyc_submissions disable row level security;
drop policy seller_scope on sellers.seller_members;
alter table sellers.seller_members disable row level security;
drop policy seller_scope on sellers.sellers;
alter table sellers.sellers disable row level security;
