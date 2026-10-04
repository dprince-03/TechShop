---
name: database-migrations
description: Design database schemas and write safe migrations for PostgreSQL, MySQL, MongoDB, and Redis data models — naming, keys, indexes, constraints, multi-tenant columns, soft deletes, zero-downtime changes, backfills, and rollback plans. Use this skill whenever the user asks to create or change tables/collections/columns/indexes, write a migration, design a schema or data model, optimize queries with indexes, or plan a data backfill — and always before running anything destructive like DROP, RENAME, or type changes.
---

# Database Schema & Migrations

## Before writing a migration

1. Find the migration tool in use (golang-migrate, goose, Atlas, Prisma Migrate, Drizzle Kit, Knex, TypeORM, Flyway, Liquibase, migrate-mongo) and follow its file naming and format.
2. Read the latest few migrations to match style.
3. Run `scripts/check_migration.py <file.sql>` on SQL migrations to flag risky statements.

## TechShop specifics

- Tool: **goose**, run as `go tool goose` or the Make targets (`make migrate-create name=…`, `migrate-up`, `migrate-status`). Never run `down`/`reset` against a shared or production database.
- The reviewed design is `docs/database/schema.sql`; proposed changes wait in `docs/schema-changes.md` for owner approval before any migration is written.
- After changing `db/migrations` or `db/queries`, run `make sqlc`; never hand-edit `internal/store/`.
- **Not multi-tenant:** no `tenant_id`. Rows are scoped by ownership columns (`customer_user_id`, `seller_id`, `business_id`). Money is `bigint` kobo; statuses are `text` + `CHECK`; ledger and audit tables are append-only.

## Schema conventions (defaults)

- Tables: plural `snake_case` (`invoice_items`); columns `snake_case`.
- Primary keys: `id UUID` (v4 or v7) or `BIGINT GENERATED ALWAYS AS IDENTITY` for internal-only tables — match project.
- Every business table: `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`; `deleted_at TIMESTAMPTZ NULL` for soft delete.
- Multi-tenant tables only: `tenant_id UUID NOT NULL REFERENCES tenants(id)`, indexed, and **included in unique constraints** (`UNIQUE (tenant_id, email)`).
- Foreign keys with explicit `ON DELETE` behaviour (prefer `RESTRICT`; `CASCADE` only for true child rows).
- Money: `BIGINT` minor units + `CHAR(3)` currency, or `NUMERIC(19,4)`. Never `FLOAT`.
- Enums: `TEXT` + `CHECK` constraint (easier to evolve than native enums) unless project uses enums.
- Use `TIMESTAMPTZ`, not `TIMESTAMP`.
- Name constraints and indexes explicitly: `idx_invoices_tenant_id_created_at`, `uq_users_tenant_id_email`, `fk_invoice_items_invoice_id`.

## Indexing

- Index foreign keys and columns used in `WHERE`, `JOIN`, `ORDER BY`.
- Composite index order: equality columns first (usually `tenant_id`), then range/sort columns.
- Partial indexes for soft delete: `WHERE deleted_at IS NULL`.
- Verify with `EXPLAIN (ANALYZE, BUFFERS)` on realistic data; don't add indexes speculatively on write-heavy tables.

## Safe migration rules (zero-downtime)

| Change | Safe approach |
|---|---|
| Add column | Add as nullable or with a constant default (fast in PG 11+). |
| Add NOT NULL column | Add nullable → backfill in batches → add `CHECK ... NOT VALID` → `VALIDATE` → set NOT NULL. |
| Add index (PG) | `CREATE INDEX CONCURRENTLY` (outside a transaction). |
| Add foreign key | Add `NOT VALID`, then `VALIDATE CONSTRAINT` separately. |
| Rename column/table | Expand–contract: add new, dual-write, backfill, switch reads, drop old in a later release. |
| Change column type | New column + backfill + switch; avoid in-place rewrite on large tables. |
| Drop column/table | Stop using it in code first (deploy), then drop in a later migration; back up first. |
| Large backfill | Batches (e.g., 1–10k rows) with pauses; run as a job, not inside the schema migration. |

- Set `lock_timeout` (e.g., `SET lock_timeout = '5s'`) so migrations fail fast instead of blocking traffic.
- Every migration has a down/rollback, or a documented reason it can't (destructive).
- Take/confirm a backup before destructive migrations in production.

## Output format

When producing a migration, include:
1. The up migration
2. The down migration
3. Risk notes (locks, duration on large tables, deploy ordering with code)
4. Any app-code changes required and in which release

## References

- PostgreSQL specifics, RLS, PITR → `references/postgres.md`
- MongoDB schema design & validation → `references/mongodb.md`
- Redis key design → `references/redis.md`
