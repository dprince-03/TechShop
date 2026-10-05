-- Behaviour events and identity links (docs/recommendations.md; consent-gated).

-- name: PersonalisationInsertEvent :copyfrom
insert into personalisation.user_events (occurred_at, anonymous_id, user_id, session_id, event_type, app, surface, product_id, listing_id,
  category_id, query_norm, results_count, position, rec_request_id, strategy, state_code, device_class, properties)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);

-- name: PersonalisationLinkIdentity :exec
insert into personalisation.identity_links (anonymous_id, user_id) values ($1, $2) on conflict (anonymous_id) do nothing;
