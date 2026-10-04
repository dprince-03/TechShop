# Proposed schema changes

All database changes proposed by the backend, mobile and recommendation plans (§1–§3) and the system plans (§4–§9: messaging, identity, payments, search, orders, trust and safety), in one list, with duplicates merged. **Nothing here is applied yet.** [`database/schema.sql`](database/schema.sql) stays as reviewed until the owner approves these. Each approved item then goes into the goose migration of the domain that owns it ([`backend.md`](backend.md) §4.1), and `docs/database/tools/generate.sh` regenerates the diagrams.

**Source key:** **B** = [`backend.md`](backend.md), **M** = [`mobile.md`](mobile.md), **R** = [`recommendations.md`](recommendations.md). Naming conflicts are resolved in [`architecture-decisions.md`](architecture-decisions.md).

## 1. New tables

| # | Table | Domain | Columns (summary) | Source | Why |
|---|---|---|---|---|---|
| 1 | `user_events` | personalisation | `id bigint identity, occurred_at, received_at, anonymous_id, user_id null, session_id, event_type (check list), app, surface, product_id, listing_id, car_listing_id, category_id, query_norm, results_count, position, rec_request_id, strategy, state_code, device_class, properties jsonb`. Partitioned monthly; BRIN on `received_at` | B M R | Behaviour tracking for analytics and recommendations, from launch |
| 2 | `identity_links` | personalisation | `anonymous_id pk, user_id, linked_at` | R | Join guest history to the account at sign-in |
| 3 | `consents` | identity | `subject (user_id or anonymous_id), purpose, granted, policy_version, source, recorded_at`, append-only | B M R | NDPA proof of consent for analytics, personalisation, marketing and location tracking |
| 4 | `privacy_requests` | identity | `user_id, kind (export, delete), status, completed_at, file_id` | B | Track data export and deletion requests |
| 5 | `push_tokens` | identity | `id, user_id, app (customer, logistics), platform, token unique, provider, app_version, locale, last_seen_at, revoked_at` | B M | Push notifications |
| 6 | `user_mfa_factors` | identity | `user_id, kind 'totp', secret_encrypted, confirmed_at` | B | Staff TOTP |
| 7 | `user_recovery_codes` | identity | `user_id, code_hash, used_at` | B | MFA recovery |
| 8 | `encryption_keys` | platform | `id, purpose, wrapped_dek, kms_key_id, created_at, retired_at` | B | Envelope encryption for NIN and bank numbers |
| 9 | `public_holidays` | platform | `date pk, name` | B | Payout business days (movable holidays) |
| 10 | `nigerian_lgas` | identity | `state_code, name` (unique together) | M | Consistent LGA spelling, so delivery fee lookups work |
| 11 | `product_offer_summary` | catalogue | One row per product × channel: min price, compare-at, promo price, conditions, seller types, in stock, stock band, rating, discount, category path, brand, filterable attributes, search vector, 30-day sales | B | Fast search, filters and facets (read model) |
| 12 | `stock_reservations` | inventory | `order_id, order_line_id, warehouse_id, listing_id, variant_id, condition, owner_seller_id, quantity, status (reserved, committed, released), expires_at` | B | Release exactly what was reserved |
| 13 | `flash_claims` | marketing | `promotion_id, listing_id, user_id, order_id, quantity` (unique per user) | B | Per-user limits on flash deals |
| 14 | `rider_locations` | logistics | `rider_id pk, lat, lng, accuracy_m, recorded_at, job_id` | B M | Latest rider position for tracking and the dispatch map |
| 15 | `rider_location_pings` (optional) | logistics | `rider_id, lat, lng, accuracy_m, speed, heading, recorded_at, received_at`. Partitioned; ~90-day retention | M | Full trail if needed; otherwise a sampled trail lives in `delivery_events` |
| 16 | `rider_shifts` | logistics | `rider_id, started_at, ended_at, start_lat/lng, device_session_id, location_consent_at` | M | Shift history and hours |
| 17 | `rider_zones` | logistics | `rider_id, zone_id` (+ dispatcher zone scopes) | M | Assign by zone |
| 18 | `parcels` | logistics | `id, fulfilment_id, label_code unique, weight_grams, scanned_at` | M | Pickup scanning needs a printed label code |
| 19 | `rec_model_versions` | personalisation | `id, model_kind, status (staged, active, retired, failed), trained_at, data_cutoff, code_version, metrics jsonb, row_counts jsonb` | R | Versioned, rollback-able recommendation imports |
| 20 | `rec_item_neighbours` | personalisation | `model_version_id, kind (similar, bought_together, also_viewed), product_id, neighbour_product_id, score, rank` | R | Similar and bought-together shelves |
| 21 | `rec_subject_candidates` | personalisation | `model_version_id, user_id, product_id, score, rank` (consented users only) | R | "For you" |
| 22 | `rec_item_vectors` | personalisation | `model_version_id, product_id, vector real[]` | R | Live session re-ranking in Go |
| 23 | `rec_popularity` | personalisation | `computed_at, window (7d, 30d), category_id null, state_code null, product_id, score` | R | Trending by category and state |
| 24 | `rec_overrides` | personalisation | `surface, anchor_product_id null, product_id, action (pin, block), starts_at, ends_at, created_by` | R | Staff merchandising control (audited) |
| 25 | `rec_staging_*` | personalisation | Mirrors of tables 20–23 | R | Validate imports before activation |
| 26 | `product_compatibility` | catalogue | `accessory_product_id, device_product_id null, rule jsonb null, source (manual, attribute_rule, inferred), verified` | R | "Accessories that fit your phone" |
| 27 | `stock_alerts` (nice to have) | catalogue | `user_id, product_id, created_at, notified_at` | M | "Notify me when back in stock" |
| 28 | `review_images` (nice to have) | sales | `review_id, file_id` | M | Photo reviews |

## 2. Changes to existing tables

| # | Table | Change | Source | Why |
|---|---|---|---|---|
| 29 | `user_sessions` | Add `family_id, replaced_by, client, last_used_at, mfa_at, device_id, app, platform, app_version, device_name` | B M | Refresh-token rotation with reuse detection; signed-in devices list; remote logout of lost rider phones |
| 30 | `verification_codes` | Add `ip, user_agent`; index `(destination, created_at desc)`; purpose `staff_mfa`; **drop** purpose `delivery_otp` | B | OTP rate limits; the delivery OTP lives on the job (decision 8) |
| 31 | `idempotency_keys` | Primary key becomes `(scope, key)`; add `method, path, locked_at, completed_at, expires_at` | B | Safe replays and in-flight detection |
| 32 | `outbox_events` | Add `trace_context jsonb` + a `pg_notify` trigger | B | Low-latency relay; trace continuity |
| 33 | `audit_log` | Add `request_id, user_agent, actor_kind` | B | Linking audits to requests |
| 34 | `files` | Add `status (pending, ready, rejected), purpose, original_name, width, height, blurhash, deleted_at` | B M | Presigned upload lifecycle; image placeholders |
| 35 | `inventory_levels`, `stock_movements`, `device_units` | Add `owner_seller_id` (default TechShop); include it in the inventory key | B | Vendor stock held in TechShop warehouses must not mix with own stock |
| 36 | `promotion_listing_prices` | Add `claimed` (≤ `stock_limit`), `per_user_limit` | B | Concurrency-safe flash deals |
| 37 | `coupons` / `coupon_redemptions` | Add `redeemed_count` (≤ `max_redemptions`); redemption `status` | B | Concurrency-safe coupons; release on cancellation |
| 38 | `order_lines` | Add `vat_kobo` | B | Mixed own and vendor carts; per-line refunds and journals |
| 39 | `orders` | Add `payment_due_at` | B | Expiry of unpaid orders |
| 40 | `ledger_journals` | Add `posting_key text not null unique` | B | A money event can never be posted twice |
| 41 | `ledger_accounts` | Add `user_id unique` for `store_credit:<user>` accounts; seed `cash:store:<code>` per store | B | Trade-in store credit; POS cash |
| 42 | `refunds` | Add `requested_by, failure_reason`; check `approved_by <> requested_by` | B | Separation of duties |
| 43 | `payments` | Add `channel, fees_kobo, provider_status, checkout_url, expires_at` | B M | Fees in reconciliation; resume interrupted payments |
| 44 | `payment_events` | Add `payout_id, refund_id` | B | Transfer and refund webhooks |
| 45 | `seller_bank_accounts` | Add `account_number_hmac, resolved_account_name, provider, provider_recipient_code` | B | Duplicate detection; payouts |
| 46 | `payouts` | Add `approved_by, failure_reason` | B | Dual control |
| 47 | `seller_kyc_submissions` | Add `id_number_hmac, verification_provider, verification_ref, verified_at` | B | Duplicate NIN detection; KYC provider |
| 48 | `trade_ins` | Add encrypted bank details (`bank_code, account_number_encrypted, last4`), `paid_at` | B | Trade-in bank payouts |
| 49 | `support_tickets` | Add `contact_name, contact_email, contact_phone, site`; relax the "user or seller" check | B | Public contact forms (corporate, wholesale, partners) |
| 50 | `quotes` | `requested_by` nullable + contact fields (**only if** anonymous quotes are allowed) | B | Owner decision |
| 51 | `car_listings` | Add `slug unique` | B | URLs use the slug |
| 52 | `fulfilments` | Add `carrier, tracking_number, pickup_warehouse_id`, `label_code` (if no `parcels` table) | B M | Seller shipping; click and collect; scanning |
| 53 | `delivery_jobs` | Add `direction (delivery, pickup)` / `kind (delivery, return_pickup, seller_pickup)`, `return_request_id`; `fulfilment_id` nullable for pickups; add `otp_expires_at, otp_attempts, otp_verified_at, proof_method (otp, photo, photo_offline)`, `pickup_warehouse_id`/`pickup_address`, `route_sequence`, drop-off coordinates. **Change the delivered check** to require `otp_verified_at` or a proof photo | B M | Return pickups; fixes "delivered without checking the OTP" |
| 54 | `delivery_events` | Add `client_event_id uuid unique, occurred_at, reason, accuracy_m, file_id` | M | Idempotent offline sync; failure reasons |
| 55 | `categories` | Add `repurchase_days integer null` | R | Don't re-recommend phones; do re-recommend consumables |
| 56 | `pos_shifts` | Add `variance_kobo` | B | Cash reconciliation |

## 3. States and rules

| # | Item | Source | Why |
|---|---|---|---|
| 57 | Document an audited `cancelled → paid` order transition (a late payment reinstates the order) **or** choose automatic refund | B | Payment succeeding after the order expired (owner decision) |
| 58 | Status labels and tones for every lifecycle in `shared/domain`, matching [`database/states.md`](database/states.md) | M | One source of status wording for web and mobile |

## 4. Messaging and marketing ([`messaging-marketing.md`](messaging-marketing.md) §6)

| # | Item | Why |
|---|---|---|
| 59 | `notifications` extended into the message log: `category, priority, locale, template_version_id, campaign_id, journey_enrollment_id, provider, provider_message_id, to_hash, cost_kobo, idempotency_key unique, skip_reason, delivered_at, clicked_at, error`; status adds `skipped, delivered, bounced` | One log for every message and the in-app inbox |
| 60 | `message_events` (append-only, partitioned) | Delivery, bounce, complaint, open, click, unsubscribe callbacks |
| 61 | `message_suppressions` (`channel, address_hash unique, reason`) | Never send to bounced, complaining or unsubscribed addresses |
| 62 | `notification_preferences`, `notification_settings` (quiet hours) | Preference centre |
| 63 | `message_templates`, `template_versions` | Marketing-authored content with versions |
| 64 | `segments`, `campaigns` (approved_by ≠ created_by), `campaign_variants`, `campaign_recipients` (partitioned) | Audiences, campaigns, A/B tests and holdouts |
| 65 | `journeys`, `journey_steps`, `journey_enrollments` | Automations |
| 66 | `tracked_links`, `coupon_codes`, `orders.attributed_notification_id` | Signed click redirects, unique coupons, attribution |

## 5. Identity and access ([`identity-access.md`](identity-access.md) §12)

| # | Item | Why |
|---|---|---|
| 67 | `user_sessions` also adds `aud, idle_expires_at` (on top of #29) | Per-app audience and idle timeout |
| 68 | `auth_events` (append-only, partitioned) | Security event log for users and admins |
| 69 | `auth_rate_limits` (`key, window_start, count`) | OTP and login limits without Redis |
| 70 | `staff_role_scopes`, `staff_invites` | Data scopes for roles; invite-based onboarding |
| 71 | `rider_devices` | Logistics device binding |
| 72 | `service_clients` | Client credentials for the Python recommender |
| 73 | `webauthn_credentials` (v2) | Passkeys |
| 74 | `seller_members.role` values (`owner, manager, staff, finance`), `business_members.role` (`admin, buyer, approver`) + `approval_limit_kobo` | Seller and business roles |

## 6. Payments and finance ([`payments-finance.md`](payments-finance.md) §14)

| # | Item | Why |
|---|---|---|
| 75 | `virtual_accounts` | Dynamic (per order) and reserved (per business) transfer accounts |
| 76 | `settlement_reports`, `settlement_lines`, `bank_statement_lines`, `reconciliation_exceptions` | Daily three-way reconciliation |
| 77 | `disputes`, `dispute_evidence` | Chargebacks |
| 78 | `payout_batches` (approved_by ≠ prepared_by), `payouts.batch_id`, `payout_holds` | Dual-control payout runs and holds |
| 79 | `accounting_periods` + trigger rejecting postings into closed periods | Month-end close |
| 80 | `ledger_journals.reverses_journal_id, period` (with #40 `posting_key`) | Reversals instead of edits |
| 81 | `provider_health` | Shared circuit-breaker state for checkout |

## 7. Search and catalogue ([`search-catalogue.md`](search-catalogue.md) §7)

| # | Item | Why |
|---|---|---|
| 82 | `listings` adds `moderation_status, risk_score, handling_days, warranty_months, battery_health_pct, grade` | Moderation and condition details |
| 83 | `listing_reviews`, `product_suggestions` | Moderation decisions; seller-proposed products |
| 84 | `search_synonyms`, `search_redirects`, `search_pins` | Merchandising controls |
| 85 | `search_queries` (partitioned), `search_suggestions`, `catalogue_gaps` | Search analytics, autocomplete, demand gaps |
| 86 | `image_hashes` | Stolen-image detection |

## 8. Orders, inventory and fulfilment ([`orders-fulfilment.md`](orders-fulfilment.md) §10)

| # | Item | Why |
|---|---|---|
| 87 | `fulfilments.status` adds `accepted, picking`; adds `accept_by, pack_by, collection_store_id` | Seller SLAs, picking, click and collect |
| 88 | `listings.stock_quantity` (`>= 0`) | Seller-held stock reservations |
| 89 | `pick_lists`, `pick_list_items`, `pack_records`, `manifests`, `manifest_parcels` | Scan-driven warehouse work |
| 90 | `carriers`, `shipments`, `shipment_events` | Interstate carriers |
| 91 | `bin_locations` (optional), `inventory_counts`, `inventory_count_lines` | Bins and cycle counts |
| 92 | `inventory_costs`; `device_units` adds `cost_kobo, grade, owner_seller_id` | Costing and COGS |
| 93 | `return_inspections`, `seller_sla_events` | Return grading and dispositions; SLA tracking |

## 9. Trust and safety ([`trust-safety.md`](trust-safety.md) §9)

| # | Item | Why |
|---|---|---|
| 94 | `kyc_checks` | Results of each automated KYC check |
| 95 | `risk_rule_sets`, `risk_rules` (versioned) | Editable, auditable rules |
| 96 | `risk_decisions` (partitioned) | Every decision with score, reasons and version |
| 97 | `risk_cases` adds `money_at_risk_kobo, decision_id, sla_due_at, label`; status adds `in_review, escalated` | Case queue and labels |
| 98 | `risk_entities`, `risk_links`, `risk_lists` | Link graph and ban lists (hashed values) |
| 99 | `velocity_counters` (may share the `auth_rate_limits` design) | Velocity features |
| 100 | `seller_metrics_daily`, `enforcement_actions`, `listing_reports`, `review_reports` | Seller performance, enforcement and reports |

## Next steps after approval

1. Apply the approved items to `schema.sql` and re-run `docs/database/tools/generate.sh`: validation, 33+ constraint tests, regenerated diagrams.
2. Add constraint tests for the new rules: delivered requires a verified OTP or photo; flash claims within the limit; refund approver ≠ requester; unique `posting_key`.
3. Log the decision in `docs/plan.md`.
