# TechShop

Commerce platform for phones, laptops, accessories, gaming, smart home, office, workstations, cars, and more. TechShop sells its own stock, runs a marketplace for local vendors and individuals, and supplies businesses (B2C and B2B).

## Apps

| App | Path | Stack | Local URL |
| --- | --- | --- | --- |
| Company site | `frontend/apps/corporate` | Next.js | http://techshop.localhost · :3005 |
| Customer marketplace | `frontend/apps/market` | Next.js | http://market.techshop.localhost · :3001 |
| Wholesale & retail | `frontend/apps/wholesale` | Next.js | http://wholesale.techshop.localhost · :3002 |
| Seller centre | `frontend/apps/seller` | Next.js | http://seller.techshop.localhost · :3003 |
| Staff portal (role-based) | `frontend/apps/staff` | Next.js | http://staff.techshop.localhost · :3004 |
| Customer app | `mobile/apps/customer` | Expo (React Native) | — |
| Delivery & dispatch app | `mobile/apps/logistics` | Expo (React Native) | — |
| API | `backend/` | Go, Gin, pgx, goose, sqlc | http://api.techshop.localhost · :8080 |

## Architecture

```
Web apps ─┐
          ├──HTTP──► Go API ──pgx──► PostgreSQL
Mobile ───┘
```

- Only the Go API touches the database. Web and mobile apps call the API through `shared/api-client`.
- One brand definition: `shared/design-tokens` generates the web CSS variables and is imported directly by the mobile apps.

## Repository layout

```
backend/    Go API (migrations in db/migrations, queries in db/queries)
frontend/   Web apps — npm workspaces + Turborepo
  apps/       corporate, market, wholesale, seller, staff
  packages/   ui (global CSS design system)
mobile/     Mobile apps — npm workspaces, Expo
  apps/       customer, logistics
shared/     Platform-neutral code used by web and mobile
  design-tokens/, api-client/
infra/      Local infrastructure
  docker/     compose.yml (Postgres, nginx)
  nginx/      Local hostname routing
docs/       plan.md (decisions), log.md (work log)
```

Web and mobile have separate `node_modules`: Expo pins its own React version, which differs from Next.js.

## Getting started

Prerequisites: Node.js 24+, Go 1.26+, Docker.

```bash
cp backend/.env.example backend/.env
for app in frontend/apps/*; do cp "$app/.env.example" "$app/.env.local"; done
for app in mobile/apps/*; do cp "$app/.env.example" "$app/.env.local"; done

make install      # web, mobile, and Go dependencies
make infra-up     # Postgres + nginx
make migrate-up   # apply database migrations

make dev-api      # Go API
make dev-web      # all five web apps (or: make dev-market, make dev-staff, …)
make dev-customer # customer mobile app (Expo)
```

`make check` runs everything CI runs.

## Database workflow

1. `make migrate-create name=create_users` → edit the new file in `backend/db/migrations/`
2. `make migrate-up`
3. Write queries in `backend/db/queries/*.sql` (annotated, e.g. `-- name: GetUser :one`)
4. `make sqlc` → typed Go functions appear in `backend/internal/store/`

## Design tokens

Edit `shared/design-tokens/src/index.ts`, then run `make tokens` to regenerate the web CSS. CI fails if the generated file is out of date.

## Planning docs

| Doc | What it covers |
| --- | --- |
| [`docs/plan.md`](docs/plan.md) / [`docs/log.md`](docs/log.md) | Every decision (dated, append-only) and every action with verification |
| [`docs/database.md`](docs/database.md) | Data model overview; schema, ER diagrams, state machines, flows and access in [`docs/database/`](docs/database/) |
| [`docs/schema-changes.md`](docs/schema-changes.md) | Proposed schema changes awaiting approval |
| [`docs/backend.md`](docs/backend.md) | Go API plan: architecture, endpoints, data layer, security, milestones |
| [`docs/mobile.md`](docs/mobile.md) | Customer and logistics app plans |
| [`docs/recommendations.md`](docs/recommendations.md) | Hybrid recommendation system (Python + Go) |
| [`docs/architecture-decisions.md`](docs/architecture-decisions.md) | Decisions reconciling the plans |
| [`docs/messaging-marketing.md`](docs/messaging-marketing.md) | SMS, email, push and in-app messaging; campaigns, audiences, journeys, consent |
| [`docs/identity-access.md`](docs/identity-access.md) | Sign-in, tokens and sessions, MFA, staff roles and permissions, abuse protection |
| [`docs/payments-finance.md`](docs/payments-finance.md) | Payment methods, ledger, reconciliation, refunds, chargebacks, seller payouts |
| [`docs/search-catalogue.md`](docs/search-catalogue.md) | Product model, listing moderation, search, ranking, buy box |
| [`docs/orders-fulfilment.md`](docs/orders-fulfilment.md) | Orders, stock reservation, warehouse operations, fulfilment models, returns |
| [`docs/trust-safety.md`](docs/trust-safety.md) | KYC, risk engine, cases, stolen-device checks, seller enforcement |
| [`docs/mockups/`](docs/mockups/) | Interactive mockups and system simulations (open any `.html` in a browser) |
