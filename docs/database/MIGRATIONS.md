# Migrations

Schema changes are SQL files in [`backend/db/migrations/`](../../backend/db/migrations/), run by **goose** (pinned as a Go tool). Only the Go backend runs them.

## Commands (from `backend/`)

| Command | Does |
|---|---|
| `make migrate-create name=add_x` | New timestamped file `db/migrations/<ts>_add_x.sql` with Up/Down sections |
| `make migrate-up` | Applies pending migrations as the table owner (`DATABASE_MIGRATE_URL`, else `DATABASE_URL`) |
| `make db-users APP_DB_PASSWORD=…` (repo root) | Creates the API's least-privilege login `techshop_app` after migrating |
| `make migrate-status` | Lists applied and pending migrations |
| `make migrate-down` | Rolls back one migration. **Local only** |
| `make migrate-reset` | Rolls back everything. **Local only, destroys data** |
| `make sqlc` | Regenerates `internal/store/` after changing queries or tables |
| `go run ./cmd/migrate` | Same as `migrate-up`, as a binary (used in containers and deploys) |

`DB_MIGRATE_ON_STARTUP=false` by default: the API doesn't migrate itself; deploys run `cmd/migrate` as a separate step first.

## Current migrations

| # | File | What it adds |
|---|---|---|
| 1 | `00001_init` | Empty baseline |
| 2 | `00002_foundation` | Extensions (`citext`, `pg_trgm`), the domain schemas, helper functions |
| 3 | `00003_identity` | Users, sessions, codes, MFA, RBAC, staff, consents |
| 4 | `00004_platform` | Files, audit log, outbox, idempotency, flags, holidays, encryption keys |
| 5 | `00005_partners` | Sellers, KYC, bank accounts; B2B businesses and credit; addresses |
| 6 | `00006_catalog` | Catalogue, listings, prices, moderation, search read model |
| 7 | `00007_search` | Search merchandising and analytics |
| 8 | `00008_inventory` | Inventory and purchasing |
| 9 | `00009_pos` | Store tills and shifts |
| 10 | `00010_sales` | Carts, quotes, orders, fulfilments, reservations, reviews |
| 11 | `00011_logistics` | Zones, rates, riders, delivery jobs |
| 12 | `00012_fulfilment` | Warehouse work, carriers, shipments, seller SLAs |
| 13 | `00013_payments` | Payments, webhooks, refunds, disputes, reconciliation |
| 14 | `00014_finance` | Ledger, invoices, commission, payouts |
| 15 | `00015_aftersales` | Returns, warranty, repairs, trade-ins |
| 16 | `00016_messaging` | Message log, preferences, templates, push tokens |
| 17 | `00017_marketing` | Promotions, coupons, campaigns, journeys |
| 18 | `00018_content_support` | CMS and support tickets |
| 19 | `00019_risk` | Trust and safety |
| 20 | `00020_cars` | Car sales |
| 21 | `00021_personalisation` | Behaviour events and recommendation tables |
| 22 | `00022_reference_data` | Nigerian states, TechShop as first-party seller, chart of accounts, open periods, admin role, default flags |
| 23 | `00023_river_jobs` | River job-queue tables |
| 24 | `00024_database_roles` | `techshop_rw` / `techshop_ro` group roles and grants; append-only history protected from the app |
| 25 | `00025_seller_rls` | Row-level security on seller tables |
| 26 | `00026_listing_moderation` | `auto_flagged` review decision; listings RLS |
| 27 | `00027_fulfilment_rls` | Fulfilments RLS (writes) |
| 28 | `00028_review_responses` | Seller answers and moderation reasons on reviews |

The full schema was created up front (all milestones); later milestones add **code**, and new migrations only for changes.

## Rules for writing a migration

1. **One change per file**, with a full `-- +goose Down` that reverses it.
2. **Schema-qualified names** and a `comment on table … is '[domain] purpose'` for every new table.
3. **Index every foreign key**; add an explicit `platform.set_updated_at()` trigger for every `updated_at`.
4. **Money in kobo** (`bigint`), statuses as `text` + `CHECK`, times as `timestamptz`.
5. **Never edit a migration that has run anywhere shared** (staging or production). Write a new one.
6. **Zero-downtime changes** on big or live tables:
   - add columns nullable (or with a constant default), backfill in batches, then add `NOT NULL`;
   - `create index concurrently` (in its own migration, with `-- +goose NO TRANSACTION`);
   - rename or drop in two releases: stop using it first, remove it later.
7. **Destructive changes** (`DROP`, type changes, data deletes) need owner approval and a backup first.
8. After changing tables or queries, run `make sqlc` and commit the regenerated `internal/store/`.
9. CI runs every migration **up** against a fresh Postgres. Test each new `Down` locally (`make migrate-down` then `make migrate-up`) before committing.

## Production

- Back up (or confirm point-in-time recovery) before each release with migrations.
- Run `cmd/migrate` before starting the new API version; the old version must still work with the new schema (that's why rule 6 exists).
- **Never roll back a migration that has run in production.** Fix forward with a new migration.
