# Row-level security (RLS)

## Current state: no RLS policies

TechShop has **no Postgres row-level security policies today** (no `enable row level security` or `create policy` in any migration). This is deliberate, not an oversight: access control is enforced in the Go API, the only thing that connects to the database.

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

## Known gap

Locally and in the current deploy plan, the API and the migrations use **one database user** that owns every table. That breaks least privilege (`.claude/rules/01` §6: separate users for the app, migrations and read-only analytics). It should be fixed before production, and it is a prerequisite for RLS anyway: table owners bypass RLS.

## Recommendation

Use RLS as **defence in depth** for the few places where a missed `where` clause would leak another party's data. Don't put it everywhere: it adds a per-request setup step and makes queries harder to reason about.

1. **First, split database roles** (one migration):
   - `techshop_owner`: owns tables, used only by `cmd/migrate`;
   - `techshop_app`: `select/insert/update/delete` on application tables, no DDL, used by the API and worker;
   - `techshop_readonly`: `select` on non-personal tables, for analytics.
2. **Then RLS on seller-scoped tables** (marketplace, M4 onwards), because sellers are separate businesses:
   - `sellers.*`, `catalog.listings`, `sales.fulfilments`, `finance.payouts` / `payout_items`;
   - the API sets `set local app.seller_id = '<uuid>'` inside each seller request's transaction;
   - policy: `using (seller_id = current_setting('app.seller_id', true)::uuid or current_setting('app.role', true) = 'staff')`.
3. **Optionally on customer personal data** (`identity.addresses`, `sales.orders`) the same way with `app.user_id`.

Example policy (not applied):

```sql
alter table catalog.listings enable row level security;
create policy listings_seller_scope on catalog.listings
  using (
    current_setting('app.role', true) in ('staff', 'system', 'public_read')
    or seller_id = nullif(current_setting('app.seller_id', true), '')::uuid
  );
```

**Decision needed from the owner** before any of this is built: approve step 1 (role split) now, and step 2 as part of M4 or later.
