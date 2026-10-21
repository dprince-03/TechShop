-- Sellers (marketplace shops, members, KYC, bank accounts). Plan: trust-safety.md §2.

-- name: SellersGetFirstParty :one
select * from sellers.sellers where type = 'first_party';

-- name: SellersGet :one
select * from sellers.sellers where id = $1;

-- name: SellersGetBySlug :one
select * from sellers.sellers where slug = $1;

-- name: SellersCreate :one
insert into sellers.sellers (type, display_name, slug, owner_user_id, status, state_code, city, rating_avg, rating_count)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
on conflict (slug) do update set display_name = excluded.display_name, status = excluded.status
returning *;

-- name: SellersAddMember :exec
insert into sellers.seller_members (seller_id, user_id, role) values ($1, $2, $3)
on conflict (seller_id, user_id) do update set role = excluded.role;

-- ---- Seller Centre (M4): shops, KYC, bank accounts. Seller requests run with app.seller_id set,
-- so row-level security (00025) also scopes these tables to the caller's shop.

-- name: SellersRegisterShop :one
insert into sellers.sellers (type, display_name, slug, owner_user_id, status, state_code, city)
values ($1, $2, $3, $4, 'pending_kyc', $5, $6)
returning *;

-- name: SellersMemberRole :one
select role from sellers.seller_members where seller_id = $1 and user_id = $2;

-- name: SellersShopsForUser :many
select s.*, m.role as member_role from sellers.sellers s
join sellers.seller_members m on m.seller_id = s.id
where m.user_id = $1 and s.status <> 'closed'
order by s.created_at;

-- name: SellersCountOwnedByUser :one
select count(*) from sellers.sellers where owner_user_id = $1 and status <> 'closed';

-- name: SellersSlugTaken :one
select exists (select 1 from sellers.sellers where slug = $1);

-- name: SellersLatestKyc :one
select * from sellers.seller_kyc_submissions where seller_id = $1 order by created_at desc limit 1;

-- name: SellersGetKyc :one
select * from sellers.seller_kyc_submissions where id = $1;

-- name: SellersGetKycForUpdate :one
select * from sellers.seller_kyc_submissions where id = $1 for update;

-- name: SellersInsertKyc :one
insert into sellers.seller_kyc_submissions (seller_id, id_type, id_number_encrypted, id_number_hmac, id_number_last4, cac_rc_number, tin, verification_provider)
values ($1, $2, $3, $4, $5, $6, $7, 'manual')
returning *;

-- name: SellersInsertKycDocument :exec
insert into sellers.kyc_documents (submission_id, kind, file_id) values ($1, $2, $3);

-- name: SellersKycDocuments :many
select d.id, d.kind, d.file_id, f.content_type, f.original_name, d.created_at
from sellers.kyc_documents d join platform.files f on f.id = d.file_id
where d.submission_id = $1 order by d.created_at;

-- name: SellersReviewKyc :one
update sellers.seller_kyc_submissions
set status = $2, reviewed_by = $3, reviewed_at = now(), rejection_reason = $4
where id = $1
returning *;

-- name: SellersOtherSellersWithIDNumber :one
select count(distinct seller_id) from sellers.seller_kyc_submissions where id_number_hmac = $1 and seller_id <> $2;

-- name: SellersSetStatus :one
update sellers.sellers set status = $2 where id = $1 returning *;

-- name: SellersInsertBankAccount :one
insert into sellers.seller_bank_accounts (seller_id, bank_code, account_name, resolved_account_name, account_number_encrypted,
  account_number_hmac, account_number_last4, provider, verified_at, is_default)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
returning *;

-- name: SellersBankAccounts :many
select * from sellers.seller_bank_accounts where seller_id = $1 order by is_default desc, created_at;

-- name: SellersBankAccountExists :one
select exists (select 1 from sellers.seller_bank_accounts where seller_id = $1 and account_number_hmac = $2);

-- name: SellersOtherSellersWithAccount :one
select count(distinct seller_id) from sellers.seller_bank_accounts where account_number_hmac = $1 and seller_id <> $2;

-- name: SellersClearDefaultBank :exec
update sellers.seller_bank_accounts set is_default = false where seller_id = $1 and is_default;

-- name: SellersSetDefaultBank :execrows
update sellers.seller_bank_accounts set is_default = true where id = $1 and seller_id = $2;

-- name: SellersDeleteBankAccount :execrows
delete from sellers.seller_bank_accounts where id = $1 and seller_id = $2 and not is_default;

-- name: SellersStaffList :many
select s.*, k.status as kyc_status, k.created_at as kyc_submitted_at
from sellers.sellers s
left join sellers.seller_kyc_submissions k on k.seller_id = s.id
  and k.created_at = (select max(created_at) from sellers.seller_kyc_submissions x where x.seller_id = s.id)
where s.type <> 'first_party'
  and (sqlc.narg('status')::text is null or s.status = sqlc.narg('status'))
  and (sqlc.narg('kyc_status')::text is null or k.status = sqlc.narg('kyc_status'))
  and (sqlc.narg('q')::text is null or s.display_name ilike '%' || sqlc.narg('q') || '%' or s.slug ilike '%' || sqlc.narg('q') || '%')
  and (sqlc.narg('before')::timestamptz is null or s.created_at < sqlc.narg('before'))
order by s.created_at desc limit $1;

-- name: SellersMembers :many
select m.user_id, m.role, m.created_at, u.first_name, u.last_name
from sellers.seller_members m join identity.users u on u.id = m.user_id
where m.seller_id = $1 order by m.created_at;

-- ---- Team: members and invites (identity-access.md §7.3)

-- name: SellersTeam :many
select m.user_id, m.role, m.created_at, u.first_name, u.last_name, u.phone, u.email,
       exists (select 1 from identity.user_mfa_factors f where f.user_id = m.user_id and f.kind = 'totp' and f.confirmed_at is not null)::boolean as has_authenticator
from sellers.seller_members m join identity.users u on u.id = m.user_id
where m.seller_id = $1 order by m.created_at;

-- name: SellersHasAuthenticator :one
select exists (select 1 from identity.user_mfa_factors where user_id = $1 and kind = 'totp' and confirmed_at is not null)::boolean;

-- name: SellersOwnerCount :one
select count(*)::int from sellers.seller_members where seller_id = $1 and role = 'owner';

-- name: SellersSetMemberRole :one
update sellers.seller_members set role = $3 where seller_id = $1 and user_id = $2 returning *;

-- name: SellersRemoveMember :execrows
delete from sellers.seller_members where seller_id = $1 and user_id = $2;

-- name: SellersCreateInvite :one
insert into sellers.seller_invites (seller_id, phone, email, role, token_hash, invited_by, expires_at)
values ($1, $2, $3, $4, $5, $6, $7) returning *;

-- name: SellersOpenInvites :many
select i.*, u.first_name as invited_by_first_name from sellers.seller_invites i join identity.users u on u.id = i.invited_by
where i.seller_id = $1 and i.accepted_at is null and i.revoked_at is null order by i.created_at desc;

-- name: SellersGetInvite :one
select * from sellers.seller_invites where id = $1 and seller_id = $2;

-- name: SellersInviteByToken :one
select i.*, s.display_name, s.status as shop_status from sellers.seller_invites i join sellers.sellers s on s.id = i.seller_id
where i.token_hash = $1;

-- name: SellersRenewInvite :one
update sellers.seller_invites set token_hash = $2, expires_at = $3 where id = $1 and accepted_at is null and revoked_at is null returning *;

-- name: SellersRevokeInvite :execrows
update sellers.seller_invites set revoked_at = now() where id = $1 and seller_id = $2 and accepted_at is null and revoked_at is null;

-- name: SellersAcceptInvite :execrows
update sellers.seller_invites set accepted_by = $2, accepted_at = now()
where id = $1 and accepted_at is null and revoked_at is null and expires_at > now();
