-- Marketing: promotions, flash deals, coupons, segments, campaigns, journeys, tracked links
-- (docs/messaging-marketing.md §4, docs/backend.md §4.4).

-- name: MarketingCreatePromotion :one
insert into marketing.promotions (name, kind, value, funded_by, starts_at, ends_at, status, created_by)
values ($1, $2, $3, $4, $5, $6, $7, $8) returning *;

-- name: MarketingGetPromotion :one
select * from marketing.promotions where id = $1;

-- name: MarketingUpdatePromotion :one
update marketing.promotions set name = $2, value = $3, starts_at = $4, ends_at = $5, updated_at = now()
where id = $1 and status in ('draft', 'scheduled') returning *;

-- name: MarketingSetPromotionStatus :one
update marketing.promotions set status = $2, updated_at = now() where id = $1 returning *;

-- name: MarketingListPromotions :many
select * from marketing.promotions where (sqlc.narg('status')::text is null or status = sqlc.narg('status')) order by starts_at desc limit $1;

-- name: MarketingAdvancePromotions :many
-- Scheduled promotions go live at starts_at; live ones end at ends_at (runs every minute).
update marketing.promotions set status = case when now() >= ends_at then 'ended' else 'live' end, updated_at = now()
where (status = 'scheduled' and now() >= starts_at) or (status = 'live' and now() >= ends_at)
returning id, status;

-- name: MarketingAddTarget :one
insert into marketing.promotion_targets (promotion_id, category_id, listing_id) values ($1, $2, $3) returning *;

-- name: MarketingPromotionTargets :many
select * from marketing.promotion_targets where promotion_id = $1;

-- name: MarketingSetListingPrice :one
insert into marketing.promotion_listing_prices (promotion_id, listing_id, price_kobo, stock_limit, per_user_limit) values ($1, $2, $3, $4, $5)
on conflict (promotion_id, listing_id) do update set price_kobo = excluded.price_kobo, stock_limit = excluded.stock_limit, per_user_limit = excluded.per_user_limit
returning *;

-- name: MarketingPromotionListingPrices :many
select * from marketing.promotion_listing_prices where promotion_id = $1;

-- name: MarketingLiveAutoPromotions :many
-- Live automatic promotions (not coupon-only, not flash) with their targets; empty targets = sitewide.
select pr.id, pr.name, pr.kind, pr.value, pr.funded_by,
       coalesce(array_agg(t.category_id) filter (where t.category_id is not null), '{}')::uuid[] as category_ids,
       coalesce(array_agg(t.listing_id) filter (where t.listing_id is not null), '{}')::uuid[] as listing_ids
from marketing.promotions pr
left join marketing.promotion_targets t on t.promotion_id = pr.id
where pr.status = 'live' and now() between pr.starts_at and pr.ends_at and pr.kind in ('percentage', 'fixed_amount', 'free_delivery')
  and not exists (select 1 from marketing.coupons c where c.promotion_id = pr.id)
group by pr.id;

-- name: MarketingPromotionTargetsFor :one
select coalesce(array_agg(category_id) filter (where category_id is not null), '{}')::uuid[] as category_ids,
       coalesce(array_agg(listing_id) filter (where listing_id is not null), '{}')::uuid[] as listing_ids
from marketing.promotion_targets where promotion_id = $1;

-- name: MarketingCategoryAncestors :many
-- A category and its parents (a promotion on "Phones" covers "Phones › Android").
with recursive up as (
  select c.id, c.parent_id from catalog.categories c where c.id = $1
  union all
  select p.id, p.parent_id from catalog.categories p join up on up.parent_id = p.id
)
select id from up;

-- name: MarketingClaimFlash :one
-- Concurrency-safe: claimed never exceeds stock_limit (a 0-row result means sold out).
update marketing.promotion_listing_prices set claimed = claimed + @qty::int
where promotion_id = @promotion_id and listing_id = @listing_id and (stock_limit is null or claimed + @qty::int <= stock_limit)
returning claimed, stock_limit;

-- name: MarketingUserFlashClaimed :one
select coalesce(sum(quantity), 0)::int from marketing.flash_claims where promotion_id = $1 and listing_id = $2 and user_id = $3;

-- name: MarketingInsertFlashClaim :exec
insert into marketing.flash_claims (promotion_id, listing_id, user_id, order_id, quantity) values ($1, $2, $3, $4, $5);

-- name: MarketingReleaseFlashClaims :many
-- Cancellation gives flash units back to the pool.
with gone as (delete from marketing.flash_claims where order_id = $1 returning promotion_id, listing_id, quantity)
update marketing.promotion_listing_prices p set claimed = greatest(p.claimed - g.quantity, 0)
from gone g where p.promotion_id = g.promotion_id and p.listing_id = g.listing_id
returning p.promotion_id, p.listing_id, p.claimed;

-- name: MarketingCreateCoupon :one
insert into marketing.coupons (code, promotion_id, is_unique_codes, max_redemptions, per_user_limit, min_order_kobo)
values ($1, $2, $3, $4, $5, $6) returning *;

-- name: MarketingListCoupons :many
select c.*, p.name as promotion_name, p.kind, p.value, p.status as promotion_status, p.ends_at
from marketing.coupons c join marketing.promotions p on p.id = c.promotion_id order by c.created_at desc limit $1;

-- name: MarketingCreateCouponCode :one
insert into marketing.coupon_codes (coupon_id, code, issued_to_user_id, notification_id, expires_at) values ($1, $2, $3, $4, $5) returning *;

-- name: MarketingFindCoupon :one
-- Resolve a typed code: a shared coupon code, or a unique single-use code.
select c.id as coupon_id, c.promotion_id, c.max_redemptions, c.redeemed_count, c.per_user_limit, c.min_order_kobo,
       cc.id as code_id, cc.issued_to_user_id, cc.expires_at as code_expires_at, cc.redeemed_at as code_redeemed_at,
       p.kind, p.value, p.funded_by, p.status as promotion_status, p.starts_at, p.ends_at
from marketing.coupons c
join marketing.promotions p on p.id = c.promotion_id
left join marketing.coupon_codes cc on cc.coupon_id = c.id and cc.code = @code::citext
where (c.code = @code::citext and not c.is_unique_codes) or cc.id is not null
limit 1;

-- name: MarketingUserRedemptions :one
select count(*)::int from marketing.coupon_redemptions where coupon_id = $1 and user_id = $2 and status = 'applied';

-- name: MarketingIncrementCoupon :one
-- Concurrency-safe counter: never passes max_redemptions.
update marketing.coupons set redeemed_count = redeemed_count + 1
where id = $1 and (max_redemptions is null or redeemed_count < max_redemptions) returning redeemed_count;

-- name: MarketingRedeemCode :execrows
update marketing.coupon_codes set redeemed_at = now() where id = $1 and redeemed_at is null;

-- name: MarketingInsertRedemption :exec
insert into marketing.coupon_redemptions (coupon_id, coupon_code_id, order_id, user_id, discount_kobo) values ($1, $2, $3, $4, $5);

-- name: MarketingReleaseRedemption :one
update marketing.coupon_redemptions set status = 'released' where order_id = $1 and status = 'applied' returning *;

-- name: MarketingDecrementCoupon :exec
update marketing.coupons set redeemed_count = greatest(redeemed_count - 1, 0) where id = $1;

-- name: MarketingUnredeemCode :exec
update marketing.coupon_codes set redeemed_at = null where id = $1;

-- name: MarketingCreateSegment :one
insert into marketing.segments (name, rules, created_by) values ($1, $2, $3) returning *;

-- name: MarketingGetSegment :one
select * from marketing.segments where id = $1;

-- name: MarketingListSegments :many
select * from marketing.segments order by created_at desc limit 200;

-- name: MarketingSetSegmentCount :exec
update marketing.segments set last_count = $2, last_counted_at = now(), updated_at = now() where id = $1;

-- name: MarketingCreateCampaign :one
insert into marketing.campaigns (name, segment_id, exclusions, channels, holdout_pct, ab_test, cost_estimate_kobo, created_by)
values ($1, $2, $3, $4, $5, $6, $7, $8) returning *;

-- name: MarketingGetCampaign :one
select * from marketing.campaigns where id = $1;

-- name: MarketingGetCampaignForUpdate :one
select * from marketing.campaigns where id = $1 for update;

-- name: MarketingListCampaigns :many
select * from marketing.campaigns order by created_at desc limit $1;

-- name: MarketingSetCampaignStatus :one
update marketing.campaigns set status = $2, scheduled_at = coalesce(sqlc.narg('scheduled_at'), scheduled_at),
  approved_by = coalesce(sqlc.narg('approved_by'), approved_by), approved_at = case when sqlc.narg('approved_by')::uuid is not null then now() else approved_at end,
  updated_at = now()
where id = $1 returning *;

-- name: MarketingDueCampaigns :many
select * from marketing.campaigns where status = 'scheduled' and scheduled_at <= now() limit 10;

-- name: MarketingAddVariant :one
insert into marketing.campaign_variants (campaign_id, label, channel, template_version_id, share_pct) values ($1, $2, $3, $4, $5) returning *;

-- name: MarketingCampaignVariants :many
select * from marketing.campaign_variants where campaign_id = $1 order by label, channel;

-- name: MarketingInsertRecipient :exec
insert into marketing.campaign_recipients (campaign_id, user_id, variant_id, holdout) values ($1, $2, $3, $4) on conflict do nothing;

-- name: MarketingPendingRecipients :many
select r.*, u.phone, u.email from marketing.campaign_recipients r join identity.users u on u.id = r.user_id
where r.campaign_id = $1 and r.status = 'pending' and not r.holdout limit $2;

-- name: MarketingSetRecipientStatus :exec
update marketing.campaign_recipients set status = $3 where campaign_id = $1 and user_id = $2;

-- name: MarketingCampaignStats :one
select count(*)::int as recipients,
       count(*) filter (where holdout)::int as holdout,
       count(*) filter (where status = 'sent')::int as sent,
       count(*) filter (where status = 'skipped')::int as skipped,
       (select count(*) from messaging.notifications n where n.campaign_id = $1 and n.clicked_at is not null)::int as clicked,
       (select count(*) from sales.orders o join messaging.notifications n on n.id = o.attributed_notification_id where n.campaign_id = $1)::int as orders
from marketing.campaign_recipients where campaign_id = $1;

-- name: MarketingCreateJourney :one
insert into marketing.journeys (name, trigger, entry_rules, reentry_days, goal_event, created_by) values ($1, $2, $3, $4, $5, $6) returning *;

-- name: MarketingGetJourney :one
select * from marketing.journeys where id = $1;

-- name: MarketingListJourneys :many
select * from marketing.journeys order by created_at desc limit 100;

-- name: MarketingSetJourneyStatus :one
update marketing.journeys set status = $2, updated_at = now() where id = $1 returning *;

-- name: MarketingAddJourneyStep :one
insert into marketing.journey_steps (journey_id, position, kind, config) values ($1, $2, $3, $4) returning *;

-- name: MarketingJourneySteps :many
select * from marketing.journey_steps where journey_id = $1 order by position;

-- name: MarketingActiveJourneysByTrigger :many
select * from marketing.journeys where trigger = $1 and status = 'active';

-- name: MarketingRecentlyEnrolled :one
-- Re-entry guard: was this person in the journey within reentry_days?
select exists (select 1 from marketing.journey_enrollments where journey_id = $1 and user_id = $2 and created_at > $3)::bool;

-- name: MarketingEnroll :one
insert into marketing.journey_enrollments (journey_id, user_id, next_run_at, context) values ($1, $2, now(), $3)
on conflict (journey_id, user_id) where status = 'active' do nothing returning *;

-- name: MarketingDueEnrollments :many
select e.*, j.goal_event from marketing.journey_enrollments e join marketing.journeys j on j.id = e.journey_id
where e.status = 'active' and e.next_run_at <= now() and j.status = 'active'
order by e.next_run_at limit 100
for update of e skip locked;

-- name: MarketingAdvanceEnrollment :exec
update marketing.journey_enrollments set current_step = $2, next_run_at = $3, updated_at = now() where id = $1;

-- name: MarketingFinishEnrollment :exec
update marketing.journey_enrollments set status = $2, exit_reason = $3, next_run_at = null, updated_at = now() where id = $1 and status = 'active';

-- name: MarketingGoalReached :exec
-- A goal event (e.g. order placed) exits that person from journeys waiting for it.
update marketing.journey_enrollments e set status = 'goal_reached', next_run_at = null, updated_at = now()
from marketing.journeys j
where j.id = e.journey_id and e.user_id = $1 and e.status = 'active' and j.goal_event = $2;

-- name: MarketingJourneyStats :many
select status, count(*)::int as n from marketing.journey_enrollments where journey_id = $1 group by status;

-- name: MarketingCreateLink :one
insert into marketing.tracked_links (url, notification_id, campaign_id) values ($1, $2, $3) returning *;

-- name: MarketingGetLink :one
select * from marketing.tracked_links where id = $1;

-- name: MarketingAbandonedCarts :many
-- Signed-in carts idle for an hour with items, not yet enrolled in the abandoned-cart journey.
select c.id as cart_id, c.user_id, count(ci.id)::int as items, sum(ci.quantity * l.price_kobo)::bigint as value_kobo
from sales.carts c
join sales.cart_items ci on ci.cart_id = c.id
join catalog.listings l on l.id = ci.listing_id
where c.status = 'active' and c.user_id is not null and c.updated_at < now() - interval '1 hour' and c.updated_at > now() - interval '3 days'
group by c.id limit 500;

-- name: MarketingCampaignNotification :one
select id from messaging.notifications where campaign_id = $1 and user_id = $2 order by created_at limit 1;

-- name: MarketingLastClick :one
-- Attribution: the latest marketing message this person clicked within the window.
select id from messaging.notifications
where user_id = $1 and clicked_at > $2 and (campaign_id is not null or journey_enrollment_id is not null)
order by clicked_at desc limit 1;

-- name: MarketingUserContact :one
select id, phone, email, first_name from identity.users where id = $1 and status = 'active';

-- name: MarketingCartStillActive :one
select exists (select 1 from sales.carts c join sales.cart_items ci on ci.cart_id = c.id where c.user_id = $1 and c.status = 'active')::bool;

-- name: MarketingOrderedSince :one
select exists (select 1 from sales.orders where customer_user_id = $1 and placed_at > $2 and status <> 'cancelled')::bool;

-- name: MarketingCampaignLink :one
select * from marketing.tracked_links where campaign_id = $1 order by created_at limit 1;

-- name: MarketingSetEnrollmentContext :exec
update marketing.journey_enrollments set context = $2, updated_at = now() where id = $1;
