-- Content: CMS pages, banners, help articles.

-- name: ContentGetPublishedPage :one
select * from content.cms_pages where site = $1 and slug = $2 and status = 'published';

-- name: ContentListPages :many
select * from content.cms_pages where (sqlc.narg('site')::text is null or site = sqlc.narg('site')) order by site, slug;

-- name: ContentUpsertPage :one
insert into content.cms_pages (site, slug, title, body, status, updated_by) values ($1, $2, $3, $4, 'draft', $5)
on conflict (site, slug) do update set title = excluded.title, body = excluded.body, status = 'draft', updated_by = excluded.updated_by, approved_by = null
returning *;

-- name: ContentGetPage :one
select * from content.cms_pages where id = $1;

-- name: ContentSetPageStatus :one
update content.cms_pages set status = $2, approved_by = sqlc.narg('approved_by'),
  published_at = case when $2 = 'published' then now() else published_at end
where id = $1 returning *;

-- name: ContentLiveBanners :many
select * from content.banners
where site = $1 and (sqlc.narg('placement')::text is null or placement = sqlc.narg('placement')) and status = 'live'
  and (starts_at is null or starts_at <= now()) and (ends_at is null or ends_at > now())
order by placement, position;

-- name: ContentUpsertBanner :one
insert into content.banners (site, placement, title, image_file_id, link_url, starts_at, ends_at, position, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *;

-- name: ContentSetBannerStatus :one
update content.banners set status = $2 where id = $1 returning *;

-- name: ContentListBanners :many
select * from content.banners order by site, placement, position;

-- name: ContentListHelp :many
select id, site, topic, slug, title, published_at from content.help_articles where site = $1 and status = 'published' order by topic, title;

-- name: ContentGetHelp :one
select * from content.help_articles where site = $1 and slug = $2 and status = 'published';

-- name: ContentUpsertHelp :one
insert into content.help_articles (site, topic, slug, title, body, status, published_at)
values ($1, $2, $3, $4, $5, $6, case when $6 = 'published' then now() end)
on conflict (site, slug) do update set topic = excluded.topic, title = excluded.title, body = excluded.body, status = excluded.status,
  published_at = case when excluded.status = 'published' then now() else content.help_articles.published_at end
returning *;
