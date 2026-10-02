# TechShop — Work Log

Append-only, chronological record of work done, sources consulted, and verification results. Decisions and their rationale live in [`plan.md`](plan.md). **Never delete or rewrite past entries.**

Format: `### YYYY-MM-DD` heading, then one bullet per action.

---

### 2026-10-01 *(backfilled 2026-10-02)*

- Scaffolded `frontend/` with `create-next-app` (TS, App Router, `src/`, ESLint, no Tailwind).
- Scaffolded `backend/` (Gin, pgx, `caarlos0/env`, godotenv, gin-contrib/cors); `/health` endpoint; slog request logging; graceful shutdown.
- Added `docker-compose.yml` (Postgres 17), root `Makefile`, `.editorconfig`, READMEs.
- Added Drizzle to frontend → removed it from frontend → moved to `database/` → removed it entirely in favor of goose + sqlc in `backend/`. Verified migrate-up/status/create, auto-migrate on startup, and sqlc codegen with a temporary table/query.
- Removed root `.env`/`.env.example` and empty `pipeline/`.
- Added `.github/workflows/ci.yml`; ran every CI step locally — all pass. golangci-lint caught an unchecked `Close()` error in `internal/database/migrate.go`; fixed.
- Created `CLAUDE.md`; later trimmed to non-derivable rules.

### 2026-10-02

- Created `docs/plan.md` (decisions) and `docs/log.md` (this file); backfilled 2026-10-01 entries.
- **Design research — sources:**
  - https://www.apple.com/store — HTML + linked/inline CSS downloaded and measured (fonts, sizes, weights, tracking, colors, radius, breakpoints, shadows, transitions); page structure via fetch.
  - https://www.samsung.com/us/mobile/ — HTML + CSS measured; page structure via fetch.
  - https://us-store.msi.com/ — HTML + CSS measured (Bootstrap-based).
  - https://www.amazon.com/ — direct download blocked (HTTP 202, empty); markup and CSS variables partially analysed via fetch.
  - https://www.jumia.com.ng/ — direct download blocked (HTTP 403); page structure analysed via fetch.
  - https://www.gucci.com/us/en/ — blocked for all automated access (HTTP 403); notes based on known public design conventions, not measured.
- Computed WCAG contrast ratios for the candidate palette; all text pairings ≥ 4.5:1, control borders ≥ 3:1.
- Added `frontend/src/styles/{tokens,reset,base,layout,components,utilities}.css`; rewrote `frontend/src/app/globals.css` as the layered entry point.
- `npm run build` succeeded; confirmed compiled CSS keeps `@layer` order (~17 KB).
- Rendered a scratch preview (outside the repo) with headless Chrome at 390px/1440px, light/dark. Fixed: badge stretching in flex-column cards; white-on-light-red badge contrast in dark mode (new `-fill` tokens). Re-rendered — fixed.
- Created skill `.claude/skills/frontend-design/SKILL.md` with all research findings, synthesis table, token reference, and usage rules.
- Added the logging rule to `CLAUDE.md`.

- Recorded business context from the owner (model, payments, catalogue, planned sites and apps) in `plan.md`; listed architecture implications as pending proposals.
- Logged the multi-app monorepo restructure decisions in `plan.md`. Checked versions: Expo 57 pins its own React (differs from Next's 19.2), which is why web and mobile use separate install roots.
- Restructure **not started**: the first move command was denied by the permission prompt and the usage limit was reached. No files were moved; the repo is unchanged since the design-system entry.
- Restructure implemented:
  - Created `shared/design-tokens` (TS tokens + CSS generator) and `shared/api-client`. Generated `tokens.css` has exactly the same variable set as the previous hand-written file (diffed).
  - Created `frontend/` workspace (npm workspaces + Turborepo): `packages/ui` (moved CSS), apps `corporate` (3000), `market` (3001), `wholesale` (3002), `seller` (3003), `staff` (3004).
  - Created `mobile/` workspace: Expo 57 apps `customer` and `logistics` from the default template, trimmed to an empty shell; theme hooks built from shared tokens; ESLint configured (`eslint-config-expo`).
  - Moved `docker-compose.yml` → `infra/docker/compose.yml`; added nginx service and `infra/nginx/` routing config for `*.techshop.localhost`.
  - Updated backend CORS origins (`internal/config`, `.env.example`), root `Makefile`, CI workflow, READMEs, `CLAUDE.md`, and the `frontend-design` skill.
- Verification:
  - Web: typecheck, lint, Prettier, and production build pass for all 5 apps; each compiled CSS contains the generated tokens with correct layer order.
  - Mobile: typecheck and lint pass; `expo-doctor` 21/21 for both apps (after fixing duplicate React via `overrides` + clean reinstall); Android JS bundles exported for both apps, confirming Metro resolves `shared/`.
  - Infra: `docker compose config` valid; `nginx -t` passes.
  - End to end: started Postgres + nginx, API, market and staff dev servers; via nginx `api.techshop.localhost/health` → `{"status":"ok","database":"up"}`, market/staff served their pages (staff with `noindex`), unknown host → 404. Services stopped afterwards.
  - `make check` (everything CI runs) passes.
- Not done (blocked by permission rules): deleting the temporary `_web_old/` folder; updating the developer's local `backend/.env` CORS line.
