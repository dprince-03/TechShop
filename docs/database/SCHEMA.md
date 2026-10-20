# Schema reference

The live schema is defined by the migrations in [`backend/db/migrations/`](../../backend/db/migrations/); they are the source of truth. This page is the map. For the plain-language model see [`DATA-MODEL.md`](DATA-MODEL.md).

## Where to find what

| You want | Look at |
|---|---|
| Every table with columns and relationships, as diagrams | [`erd.md`](erd.md) (generated, one diagram per domain) |
| The exact DDL | the migration that creates the table (table below) |
| What a table is for | `\d+ schema.table` in psql, or the `comment on table` line in its migration (every table has a `[domain] purpose` comment) |
| Status values and allowed changes | [`states.md`](states.md) |
| Who may read and write what | [`access.md`](access.md), [`RLS-POLICIES.md`](RLS-POLICIES.md) |
| The queries the API runs | [`backend/db/queries/`](../../backend/db/queries/) (sqlc generates Go from these into `backend/internal/store/`) |

## Size

- **205 tables** in **22 domain schemas** plus `river` (job queue).
- Largest schemas: `identity` 22, `catalog` 20, `risk` 15, `marketing` 15, `personalisation` 13.

## Tables by migration

| Migration | Schema(s) | Tables | Contents |
|---|---|---|---|
| `00003_identity` | `identity` | 21 | Users, sessions, OTP codes, MFA, roles and permissions, staff, consents, privacy requests, auth events |
| `00004_platform` | `platform` | 7 | Files, audit log, outbox, idempotency keys, feature flags, holidays, encryption keys |
| `00005_partners` | `sellers`, `b2b`, `identity` | 9 | Sellers, members, KYC, bank accounts; businesses, members, credit; addresses |
| `00006_catalog` | `catalog` | 17 | Categories, brands, attributes, products, variants, images, listings, prices, moderation, search read model |
| `00007_search` | `search` | 7 | Synonyms, redirects, pins, query log, suggestions, catalogue gaps |
| `00008_inventory` | `inventory`, `purchasing` | 14 | Warehouses, bins, levels, device units, movements, transfers, counts, costs; suppliers, purchase orders |
| `00009_pos` | `pos` | 2 | Tills, shifts |
| `00010_sales` | `sales`, `catalog`, `inventory` | 12 | Carts, quotes, orders, fulfilments, lines, history, reservations, reviews, ratings |
| `00011_logistics` | `logistics` | 11 | Zones, rates, riders, devices, shifts, locations, delivery jobs, events |
| `00012_fulfilment` | `fulfilment` | 10 | Parcels, pick lists, packing, manifests, carriers, shipments, seller SLA events |
| `00013_payments` | `payments` | 11 | Payments, webhooks, refunds, transfer accounts, disputes, provider health, reconciliation |
| `00014_finance` | `finance` | 10 | Invoices, ledger, periods, commission, payout batches, payouts, holds |
| `00015_aftersales` | `aftersales` | 6 | Returns, inspections, warranty, repairs, trade-ins |
| `00016_messaging` | `messaging` | 9 | Message log and inbox, provider events, suppressions, preferences, templates, push tokens |
| `00017_marketing` | `marketing` | 15 | Promotions, flash claims, coupons, segments, campaigns, journeys, tracked links |
| `00018_content_support` | `content`, `support` | 6 | CMS pages, banners, help articles; tickets |
| `00019_risk` | `risk` | 15 | Rules, decisions, cases, KYC checks, link graph, lists, blocklist, seller metrics, enforcement, reports |
| `00020_cars` | `cars` | 6 | Listings, photos, inspections, documents, viewings, financing |
| `00021_personalisation` | `personalisation` | 13 | Behaviour events, identity links, recommendation outputs and staging |
| `00023_river_jobs` | `river` | 4 | River job queue (vendored River migrations) |

`00001` (baseline), `00002` (extensions, schemas, helper functions) and `00022` (reference data) create no tables.

## Conventions

| Thing | Convention |
|---|---|
| Names | `snake_case`, plural tables, always schema-qualified (`sales.orders`) |
| Primary keys | `id uuid default gen_random_uuid()` |
| Human references | Sequence-based codes via `platform.next_ref`: `TS-10482` (orders), `QT-`, `INV-`, `PO-`, `RMA-`, `TKT-`, `WR-`, `TI-`, `CAR-` |
| Money | `bigint` kobo, column suffix `_kobo`, plus `currency char(3) default 'NGN'` where needed |
| Statuses | `text` + `CHECK (status IN (…))` |
| Time | `timestamptz`; `created_at default now()`, `updated_at` by trigger |
| Foreign keys | Every FK column has its own index |
| Encrypted fields | `bytea` (`*_enc`), `*_hmac` blind index for lookups, `*_last4` for display |
| Search | Generated `tsvector` + GIN; `pg_trgm` on names |
| High-volume logs | Partitioned by month with a default partition (`user_events`, `auth_events`, `message_events`, `risk_decisions`, `search_queries`, `rider_location_pings`) |

## Database functions

| Function | Does |
|---|---|
| `platform.set_updated_at()` | Trigger: sets `updated_at` on update |
| `platform.forbid_change()` | Trigger: makes a table append-only |
| `platform.notify_outbox()` | Trigger: `NOTIFY` the worker when an outbox event is written |
| `platform.next_ref(prefix)` | Next human reference number |
| `platform.ensure_monthly_partitions()` | Creates upcoming monthly partitions (run by the worker) |
| `finance.assert_journal_balanced()` | Deferred trigger: each journal sums to zero |
| `finance.assert_period_open()` | Trigger: no postings into a closed period |
