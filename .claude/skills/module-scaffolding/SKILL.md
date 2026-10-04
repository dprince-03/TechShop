---
name: module-scaffolding
description: Generate a complete new feature module, endpoint, resource, or service end-to-end — routes/handlers, service, repository, validation schemas, DTOs, migration, tests, and API docs — following the project's layered architecture and data-scoping rules (tenant or ownership). Use this skill whenever the user asks to "add a new module/feature/resource/endpoint/CRUD", "create an API for X", "scaffold", or "build the backend for X" in Go/Gin, Node/Express, Next.js, or Java projects, even if they only mention one layer.
---

# Module Scaffolding

Build a new feature as a complete vertical slice so nothing is forgotten (validation, authZ, data scoping, tests, docs).

## Step 1 — Gather the spec (ask only for what's missing)

- Resource name (singular + plural) and fields with types, required/optional, constraints
- Operations needed (list, get, create, update, delete, custom actions)
- Who can do what (roles/permissions per operation)
- How is data scoped? Multi-tenant apps: by `tenant_id` (default **yes**). Single-company apps: by ownership.
  TechShop is **not** multi-tenant (one company). Scope data by **ownership** instead: `customer_user_id` for customers, seller membership for Seller Centre, business membership for B2B, staff permissions plus data scopes (store, warehouse, zone) for staff. Put the scope in the SQL (`WHERE … AND customer_user_id = $2`) — see `docs/identity-access.md` §7 and `project-domain-knowledge/references/techshop.md`.
- Relationships to existing entities
- Soft delete or hard delete? Match the project (TechShop mostly uses status columns such as `cancelled` or `archived` plus append-only history, not `deleted_at`).

If the user gave enough to proceed, state assumptions inline instead of asking.

## Step 2 — Study the codebase

Find an existing, well-built module and mirror it exactly: folder layout, file names, error helpers, response envelope, test style. Read `references/<stack>.md` for the default templates when no module exists yet:
- Go/Gin → `references/go-gin-module.md`
- Node/Express → `references/express-module.md`
- Next.js (full-stack) → `references/nextjs-feature.md`

## Step 3 — Generate, in this order

1. **Migration** (table/collection, indexes, constraints, timestamps; `tenant_id` and `deleted_at` only if the project uses them). TechShop: a goose migration in `backend/db/migrations/` (`make migrate-create`), queries in `backend/db/queries/`, then `make sqlc` — never hand-edit `internal/store/`.
2. **Model / entity**
3. **Validation schemas / DTOs** (create, update, query/filter) — allow-list fields only
4. **Repository / queries** — every query applies the scope (`tenant_id`, or the ownership condition); parameterized; pagination with a max page size
5. **Service** — business rules, authorization checks, transactions where multiple writes happen
6. **Handler / controller** — bind → validate → call service → map to response DTO
7. **Routes** — register with auth + permission middleware
8. **Tests** — service unit tests + API integration tests, including a cross-tenant or cross-owner access test and a forbidden-role test
9. **API docs** — OpenAPI path entries (or update existing spec)
10. **Wiring** — register module in router/DI container

## Non-negotiables for every module

- [ ] Authentication on all routes unless explicitly public
- [ ] Authorization per operation (role/permission or ownership)
- [ ] Data scoping on every read and write (tenant or ownership)
- [ ] Input validated at the boundary; unknown fields stripped
- [ ] Response DTOs — never return raw DB rows/documents
- [ ] List endpoints paginated with a hard maximum
- [ ] Unique constraints at DB level for business-unique fields (include `tenant_id` only in multi-tenant apps)
- [ ] Audit log entry for create/update/delete of sensitive resources
- [ ] Tests include: happy path, validation failure, not found, forbidden, cross-tenant/cross-owner

## Standard REST shape (unless project differs)

**TechShop differs:** follow `docs/backend.md` §2.3 — Huma/OpenAPI, cursor pagination `{ data, page: { nextCursor, hasMore } }`, `application/problem+json` errors with `code` and field `errors`, idempotency keys on money-moving endpoints. Use the generic shape below only for other projects.

| Operation | Method & path | Success |
|---|---|---|
| List | `GET /api/v1/<plural>?page=&limit=&sort=&q=` | 200 `{ data: [], meta: { page, limit, total } }` |
| Get | `GET /api/v1/<plural>/:id` | 200 `{ data }` |
| Create | `POST /api/v1/<plural>` | 201 `{ data }` |
| Update | `PATCH /api/v1/<plural>/:id` | 200 `{ data }` |
| Delete | `DELETE /api/v1/<plural>/:id` | 204 |

Error envelope: `{ "error": { "code": "not_found", "message": "...", "details": [...] , "request_id": "..." } }`

## Finish with a summary

List files created/modified, assumptions made, and any follow-ups (e.g., "add permission `invoices:write` to the admin role seed").
