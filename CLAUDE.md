# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Rules

- **Only the Go API (`backend/`) touches PostgreSQL.** Web and mobile apps must never get a DB driver, ORM, or `DATABASE_URL`; they call the API through `@techshop/api-client` (`shared/api-client`). Schema, migrations, and queries all live in `backend/`.
- **Layout:** web apps in `frontend/apps/`, mobile apps in `mobile/apps/`, platform-neutral code used by both in `shared/` (no Next.js, React Native, or DOM-only APIs there), local infrastructure in `infra/`.
- **Separate install roots:** `frontend/` and `mobile/` each have their own `node_modules` and lockfile — Expo pins a different React version than Next.js. Never add a repo-root `package.json`/workspace. Mobile pins React via `overrides` in `mobile/package.json`; keep it matching the Expo SDK and verify with `npm run doctor`.
- **No Tailwind or CSS frameworks.** Web UI is custom CSS on the design system in `frontend/packages/ui`; tokens come from `shared/design-tokens` — load the `frontend-design` skill before any UI work (web or mobile). Feature components use CSS Modules + tokens.
- **Log every plan and decision** in `docs/plan.md` (dated, append-only — never delete or rewrite entries; supersede with a new entry). Log actions, sources, and verification in `docs/log.md` (same rules).
- **CI and CD are separate.** `.github/workflows/ci.yml` only checks code; deployment goes in its own workflow file. `make check` runs what CI runs.

## Shared packages

- `shared/design-tokens/css/tokens.css` is generated from `src/index.ts` — never hand-edit it; run `make tokens`. CI fails if it is stale.
- The token generator runs TypeScript directly with Node (type stripping) — requires Node 22.18+; CI uses Node 24.
- Shared packages are consumed via `file:` dependencies. Next apps need them in `transpilePackages` and `turbopack.root` set to the repo root (see any app's `next.config.ts`); Expo apps need `shared/` in Metro `watchFolders` (`metro.config.js`).

## Backend

- goose and sqlc are pinned as Go tools — run them via `go tool goose` / `go tool sqlc` (or the Make targets), not global binaries.
- New settings: add a tagged field to `internal/config.Config` and document it in `backend/.env.example`.
- Log with `log/slog` only; no `log` or `fmt.Println`.
- `internal/store/` is sqlc-generated — never hand-edit it (only `doc.go` is hand-written); change `db/queries/` or `db/migrations/` and run `make sqlc`.
- `make sqlc` intentionally skips while `db/queries/` has no `.sql` files — sqlc errors on an empty query set.
- New web origins must be added to `CORS_ALLOWED_ORIGINS` (default in `internal/config`, plus `.env.example`).

## Web

- This is Next.js 16 — APIs differ from older versions; check current docs rather than assuming Next 13/14 patterns.
- `npm run typecheck` runs `next typegen` first; plain `tsc` fails on a fresh clone because global types like `LayoutProps` are generated.
- Seller and staff apps must stay non-indexed (`robots: { index: false }` in their root layout metadata).
- New app: copy an existing one in `frontend/apps/`, then change its package `name`, port, metadata, and add its hostname to `infra/nginx/conf.d/techshop.conf` and the backend CORS origins.

## Mobile

- Expo SDK 57 with expo-router. Run `npx expo install <pkg>` (not plain `npm install`) so native module versions match the SDK.
- `EXPO_PUBLIC_*` env values are bundled into the app — never put secrets in them.
