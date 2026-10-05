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
