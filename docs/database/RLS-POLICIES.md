# Row-level security (RLS)

## Current state (2026-10-21)

Row-level security is **on for seller-scoped tables** (M4), as defence in depth behind the API's own checks. Seller Centre requests run in a transaction that sets `app.seller_id` (`set_config(..., true)`); requests without it (staff, customers, jobs) are unaffected.

| Table | Migration | Seller requests can read | Seller requests can write |
|---|---|---|---|
| `sellers.sellers`, `seller_members`, `seller_kyc_submissions`, `kyc_documents`, `seller_bank_accounts` | 00025 | own shop only | own shop only |
| `catalog.listings` | 00026 | own + every **active** listing (public on the market; needed by the price-outlier check) | own only |
| `sales.fulfilments` | 00027 | all (the order status is derived from every fulfilment of an order; seller queries filter by shop) | own only |

Policy shape: `coalesce(current_setting('app.seller_id', true), '') = '' or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid`.

Verified in Postgres as `techshop_app` with another shop's ID set: other shops' private rows invisible, `UPDATE`/`DELETE` affect 0 rows, `INSERT` refused by policy (`docs/log.md`, 2026-10-21).

The rest of this page describes how access is enforced everywhere else, and what's left.

## How access is enforced instead

| Layer | Enforces |
|---|---|
| **Network** | Postgres is only reachable on the private Docker network; no client app has a database driver or credentials |
| **API authentication** | Every request carries a verified JWT; the audience (`market`, `wholesale`, `customer_app`, `seller`, `staff`, `logistics`, `internal`) limits which routes it can call |
| **Permissions (RBAC)** | Each staff route declares a permission (e.g. `dispatch.view`); `httpx.Register` checks it before the handler runs |
| **Ownership scoping** | Customer, seller and business queries always filter by the caller's ID from the token (e.g. `where customer_user_id = $1`, `where seller_id = $1`), never by an ID the client sends |
| **Response types** | Handlers return purpose-built DTOs, so internal columns (hashes, encrypted fields) never serialise |
| **Database constraints** | Append-only tables, balanced ledger, closed periods, separation-of-duties checks (see [`DATA-MODEL.md`](DATA-MODEL.md)) |
| **Audit** | Sensitive reads and every admin action write `platform.audit_log` |

Per-table data classification and who may see what: [`access.md`](access.md).

## Database users (done, 2026-10-21, migration 00024)

| Login / role | Used by | Can do |
|---|---|---|
| table owner (e.g. `techshop`) | `cmd/migrate` only, via `DATABASE_MIGRATE_URL` | Everything, including DDL |
| `techshop_app` (member of `techshop_rw`) | API and worker, via `DATABASE_URL` | Read and write rows; **no DDL**; **cannot update or delete** `platform.audit_log`, ledger journals and entries, stock movements, delivery events |
| `techshop_ro` (group, no login yet) | Future reporting | `select` on `catalog`, `search`, `inventory`, `purchasing`, `content` only (no personal data) |

- Locally: `make db-users APP_DB_PASSWORD=…` (repo root) after migrating creates `techshop_app`.
- Outside development the API and worker **refuse to start** if `DATABASE_URL` is a superuser or owns the tables (`internal/platform/database/roles.go`).
- `platform.ensure_monthly_partitions` runs as its owner (`security definer`, pinned `search_path`) so the worker can add partitions without DDL rights.

## Next candidates

- Customer personal data (`identity.addresses`, `sales.orders` for customers) with an `app.user_id` setting, the same way.
- Business buyers (`b2b.*`) when M6 starts.

