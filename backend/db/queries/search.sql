-- Search merchandising and analytics (docs/search-catalogue.md §4–§5).

-- name: SearchListSynonyms :many
select * from search.search_synonyms where active order by created_at desc;

-- name: SearchCreateSynonym :one
insert into search.search_synonyms (terms, target, kind, created_by) values ($1, $2, $3, $4) returning *;

-- name: SearchDeactivateSynonym :execrows
update search.search_synonyms set active = false where id = $1;

-- name: SearchGetRedirect :one
select * from search.search_redirects where query_norm = $1 and (expires_at is null or expires_at > now());

-- name: SearchUpsertRedirect :one
insert into search.search_redirects (query_norm, url, created_by, expires_at) values ($1, $2, $3, $4)
on conflict (query_norm) do update set url = excluded.url, expires_at = excluded.expires_at
returning *;

-- name: SearchListRedirects :many
select * from search.search_redirects order by created_at desc;

-- name: SearchActivePins :many
select * from search.search_pins where query_norm = $1 and now() between starts_at and ends_at;

-- name: SearchCreatePin :one
insert into search.search_pins (query_norm, product_id, action, position, starts_at, ends_at, created_by)
values ($1, $2, $3, $4, $5, $6, $7) returning *;

-- name: SearchLogQuery :exec
insert into search.search_queries (session_id, user_id, query_norm, filters, results_count, corrected_to, latency_ms, channel)
values ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: SearchZeroResults :many
select query_norm, count(*)::int as searches, max(created_at) as last_seen
from search.search_queries
where results_count = 0 and created_at >= now() - interval '7 days'
group by query_norm order by searches desc limit $1;

-- name: SearchTopQueries :many
select query_norm, count(*)::int as searches, avg(results_count)::float8 as avg_results, avg(latency_ms)::float8 as avg_latency_ms
from search.search_queries where created_at >= now() - interval '7 days'
group by query_norm order by searches desc limit $1;

-- name: SearchStats :one
select count(*)::int as searches,
       coalesce(avg((results_count = 0)::int), 0)::float8 as zero_rate,
       coalesce(percentile_cont(0.95) within group (order by latency_ms), 0)::float8 as p95_latency_ms
from search.search_queries where created_at >= now() - interval '7 days';

-- name: SearchUpsertGap :one
insert into search.catalogue_gaps (query_norm, searches_30d, status, owner_id, note) values ($1, $2, $3, $4, $5)
on conflict (query_norm) do update set status = excluded.status, note = excluded.note, owner_id = excluded.owner_id
returning *;

-- name: SearchListGaps :many
select * from search.catalogue_gaps order by searches_30d desc;

-- name: SearchRebuildSuggestions :exec
-- Nightly: popular queries with results become autocomplete entries keyed by 1–4 letter prefixes.
with popular as (
  select query_norm, count(*) as n from search.search_queries
  where results_count > 0 and created_at >= now() - interval '30 days'
  group by query_norm having count(*) >= 2
), prefixes as (
  select left(query_norm, l) as prefix, query_norm as suggestion, n from popular, generate_series(1, 4) l
  where length(query_norm) >= l
)
insert into search.search_suggestions (prefix, suggestion, weight)
select prefix, suggestion, n from prefixes
on conflict (prefix, suggestion) do update set weight = excluded.weight;
