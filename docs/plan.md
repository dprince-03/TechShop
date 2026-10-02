# TechShop — Plans & Decisions

Append-only log of plans and decisions. **Never delete or rewrite past entries.** To change a decision, add a new dated entry that supersedes it and names the entry it replaces. Newest entries go at the bottom.

Entry format: `## YYYY-MM-DD — Title`, followed by **Decision**, **Why**, and (optionally) **Supersedes** / **Follow-ups**.

---

## 2026-10-01 — Initial stack and monorepo layout *(backfilled 2026-10-02)*

**Decision:** Monorepo with `frontend/` (Next.js App Router, TypeScript, custom CSS — no Tailwind) and `backend/` (Go + Gin), PostgreSQL for data, local Postgres via `docker-compose.yml`. Go layout follows the standard `cmd/` + `internal/` convention.
**Why:** User's requested stack; `cmd/internal` is the most widely adopted Go service layout in industry.

## 2026-10-01 — Drizzle placed in the frontend *(backfilled; superseded)*

**Decision:** Drizzle ORM installed inside `frontend/` with a server-side DB client.
**Why:** Drizzle is TypeScript-only and couldn't run in the Go backend.
**Superseded by:** "Frontend must never access the database" (below). This conflict should have been raised before building.

## 2026-10-01 — Frontend must never access the database *(backfilled)*

**Decision:** Removed all DB access from `frontend/` (driver, ORM, `DATABASE_URL`, DB health route). The frontend talks only to the Go API. Drizzle temporarily moved to a standalone `database/` package for schema/migrations only.
**Why:** User requirement — the API must be the single gateway to data.

## 2026-10-01 — Replace Drizzle with goose + sqlc in the backend *(backfilled)*

**Decision:** Removed `database/` and Drizzle. The Go backend owns the database: goose for SQL migrations (`backend/db/migrations`, embedded in the binary, optional `DB_MIGRATE_ON_STARTUP`), sqlc for type-safe query code (`backend/internal/store`). Both pinned as Go tools in `go.mod`.
**Why:** A TS schema tool next to a Go service meant two sources of truth, a Node dependency for migrations, and no compile-time link between schema and Go queries. goose + sqlc is the most common Go-native setup.
**Supersedes:** "Frontend must never access the database" (its standalone Drizzle package only).

## 2026-10-01 — Root env files removed; no `pipeline/` folder *(backfilled)*

**Decision:** Deleted root `.env`/`.env.example` (compose has built-in defaults); each app owns its env (`backend/.env.example`, `frontend/.env.example`). Deleted empty `pipeline/`. `infra/` deferred until a hosting target is chosen.
**Why:** CI/CD tools require fixed locations (`.github/workflows/`), so `pipeline/` would never be picked up; infra-as-code is deployment work, out of current scope.

## 2026-10-01 — CI only; CD in a separate workflow *(backfilled)*

**Decision:** `.github/workflows/ci.yml` runs lint/typecheck/test/build for both apps (plus migrations against a Postgres service). Deployment, when added, goes in its own workflow (e.g. `cd.yml`) gated on CI passing.
**Why:** User requirement; standard separation of verification from release.

## 2026-10-01 — Secrets management deferred *(backfilled)*

**Decision:** Keep `.env` files for local dev for now. Recommended Infisical (open source, cloud-agnostic) or Doppler once real secrets exist; cloud-native managers once a cloud is chosen. App code already reads env vars, so no code change will be needed.
**Why:** Current values are local dev defaults only.

---

## 2026-10-02 — Plan & decision logging policy

**Decision:** Every plan and decision is logged here with a date, append-only. Actions, sources, and verification results are logged chronologically in `docs/log.md`. Recorded in `CLAUDE.md` so every session follows it.
**Why:** User requirement — keep a permanent, auditable record.

## 2026-10-02 — Global design system

**Plan:** Study six reference stores (Apple Store, Samsung, MSI Store, Amazon, Jumia, Gucci) for principles only — no copying of text, assets, or branding — then implement a global design system in custom CSS.

**Decisions:**

1. **File structure** — `frontend/src/styles/{reset,tokens,base,layout,components,utilities}.css`, imported by `app/globals.css`. *Why:* one concern per file; `globals.css` stays a manifest.
2. **Cascade layers** — `@layer reset, tokens, base, layout, components, utilities;`. CSS Modules stay unlayered and therefore always override globals. *Why:* predictable precedence without specificity wars or `!important`.
3. **Color** — Neutral-dominant palette with a single blue action accent (`#1a56f0` light / `#2563eb` dark fill), plus semantic commerce colors: sale red, promo orange, rating amber, success, warning. Separate `-fill` tokens for badge backgrounds because the dark-mode text tints fail contrast under white text. Light/dark via `prefers-color-scheme`, overridable with `[data-theme]`. *Why:* Apple/Samsung restraint keeps product imagery dominant; Amazon/Jumia/MSI show commerce needs strong semantic signals. All text pairings verified ≥ 4.5:1 (WCAG AA); control borders ≥ 3:1.
4. **Typography** — Geist (already loaded) with system fallback; fixed UI sizes 12–24px, fluid `clamp()` headline sizes 28→72px; 400/600/700 weights; negative tracking on large headlines, wide tracking only for uppercase eyebrows; 65ch measure. *Why:* Apple's tracking/weight rhythm, Samsung's compact UI sizes, Gucci's uppercase labels.
5. **Spacing & containers** — 4px-base scale (4→128), fluid gutter (16→40px) and section rhythm (48→96px). Containers 720 / 1024 / 1280 (default) / 1440. *Why:* converges Apple (~980–1024 content), Samsung (1440 max), MSI (1140).
6. **Breakpoints** — mobile-first `min-width` at 640 / 768 / 1024 / 1280 (as rem literals). *Why:* matches the common steps across Samsung (768/1280), Apple (1024), MSI/Bootstrap (768/992/1200).
7. **Shape & elevation** — pill buttons; radius 4/8/12/18/28; hairline borders; shadows only for hover lift and overlays. *Why:* Apple/Samsung pills and 18–20px cards; minimal chrome (Gucci).
8. **Buttons** — `.btn` + primary/secondary/inverse/ghost, sizes 36/44/52px, block and icon variants, keyboard focus ring, disabled states. One primary per view. *Why:* Samsung filled-vs-outline CTA pairing, Gucci's black button, 44px touch targets.
9. **Links** — plain `<a>` inherits color (nav/cards); `.link`, `.link--subtle`, `.link--arrow` ("see all ›"). *Why:* most store links are navigation; Apple/Jumia "see all" chevron pattern.
10. **Layout primitives** — container, full-bleed, section bands, stack, cluster, intrinsic grid, catalog grid (2→3→4→5 cols), with-sidebar (filters), scroller (snap shelf, next item peeks), frame (aspect-ratio product imagery, `contain` on light surface), center. *Why:* catalog density from Amazon/Jumia; shelves from Apple/Jumia; product framing from Apple/Samsung.
11. **Scope kept out** — no header, product card, price, or rating components; those are feature components to be built later with CSS Modules on top of these primitives.
12. **Design knowledge as a skill** — research and rules captured in `.claude/skills/frontend-design/SKILL.md` (the user asked for `.claude/skills/frontend-design.md`; Claude Code only loads skills from `<name>/SKILL.md`).

**Verification:** Production build succeeds with layer order preserved (~17 KB CSS). Rendered a scratch preview with headless Chrome at 390px and 1440px, light and dark. Found and fixed two issues: badges stretching full-width inside flex-column cards (`width: fit-content`), and dark-mode badge contrast (introduced `--color-sale-fill` / `--color-promo-fill`).

**Limitations:** Gucci blocks automated access (HTTP 403); its notes are from known public conventions, not live measurement. Amazon was analysed partially (markup/CSS variables), Jumia structurally.

## 2026-10-02 — Business context captured

**Context (from the owner):**
- Real business, not a portfolio project.
- Payments: Paystack, OPay, Moniepoint.
- Catalogue: phones (all kinds); laptops (business, gaming, workstation); accessories; gaming; smart home; office; workstations; cars; more to come.
- Business model: sells its own stock **and** runs a marketplace for local vendors and individual sellers; also acts as a supplier. Both B2C and B2B; both single-seller store and multi-seller marketplace.
- Planned surfaces: customer marketplace site; wholesale & retail site; company main site; internal staff sites (admin, finance, etc. — modules to be chosen); customer mobile app; delivery & dispatch mobile app.

**Implications flagged (not yet decided):**
- The current single `frontend/` app can't host several sites + 2 mobile apps cleanly → proposed move to a workspace monorepo (`apps/*` + shared `packages/*`).
- A marketplace needs a **seller/vendor portal**, which wasn't on the list.
- The design system is CSS-only; mobile apps (React Native) can't use CSS → tokens should move to a shared source that generates both CSS variables and a mobile theme.
- Cars need a different flow from gadgets (inspection, documents, financing, viewings).
- Market is Nigeria-first (NGN pricing, local payment rails, mobile/data-light performance).

**Status:** Proposed. Awaiting the owner's choices on internal portal modules and the monorepo restructure.

## 2026-10-02 — Multi-app monorepo restructure

**Decisions (owner's answers):**
1. **Staff portal scope:** all 20 modules — admin console, catalogue, orders, inventory/warehouse, finance, dispatch control, support, vendor management, B2B accounts, purchasing/suppliers, marketing, analytics, content, warranty/repairs, trade-ins/buybacks, car sales, POS, risk/fraud, HR, IT tools. Built incrementally; this entry only creates the app shell.
2. **One staff portal with role-based sections** (not one site per department). *Why:* one login, one deployment, shared code; modules can be split out later if one outgrows it (most likely POS, which may need offline/hardware support).
3. **Wholesale & retail site serves both:** the shopfront for TechShop's own stock (retail) and bulk/trade buying for businesses (wholesale). The market site is the multi-vendor marketplace.
4. **Top-level layout:** `frontend/` (all web apps), `mobile/` (all mobile apps), `backend/` (Go API), `shared/` (code used by both web and mobile), `infra/` with `docker/` and `nginx/`.
5. **Seller portal added** as its own web app (`frontend/apps/seller`) — required for the marketplace.

**Technical decisions:**
- **Web apps** (`frontend/apps/`): `corporate`, `market`, `wholesale`, `seller`, `staff` — all Next.js. npm workspaces + Turborepo inside `frontend/`. Shared web code in `frontend/packages/` (`ui` = CSS design system; base TS/ESLint config at the `frontend/` root).
- **Mobile apps** (`mobile/apps/`): `customer`, `logistics` (delivery + dispatch in one app, role-based) — Expo (React Native) with expo-router, npm workspaces inside `mobile/`.
- **Separate install roots for web and mobile.** *Why:* Expo pins its own React version (differs from Next's 19.2); one shared `node_modules` would cause duplicate-React conflicts.
- **`shared/`**: `design-tokens` (TypeScript source of truth → generates the web `tokens.css`; mobile imports the values directly) and `api-client` (typed client for the Go API). Consumed via `file:` dependencies. *Why:* one brand definition for web and mobile; one API contract.
- **Local hostnames via nginx:** `techshop.localhost` (corporate), `market.`, `wholesale.`, `seller.`, `staff.`, `api.techshop.localhost` → reverse proxy to the dev servers. `*.localhost` resolves to 127.0.0.1 without editing hosts files.
- **Dev ports:** corporate 3000, market 3001, wholesale 3002, seller 3003, staff 3004, API 8080.
- Docker scope for now: local Postgres + nginx only (moved from root `docker-compose.yml` to `infra/docker/`). Production Dockerfiles/deployment remain out of scope.

## 2026-10-02 — Restructure implementation details

**Decisions made while implementing "Multi-app monorepo restructure":**
- **Design tokens source of truth** moved from hand-written CSS to `shared/design-tokens/src/index.ts` (px numbers + light/dark colors). `css/tokens.css` is generated by `scripts/build-css.ts` and committed; CI fails if it is stale. *Supersedes* the token file location in "Global design system" (decision 1); token names and values are unchanged except that fluid sizes are now interpolated linearly between 360px and 1280px viewports (same min/max as before).
- **Web CSS** moved from `frontend/src/styles/` to `frontend/packages/ui/src/`; apps import `@techshop/ui/styles.css` in their root layout. *Supersedes* decision 1's file location.
- **Mobile React pin:** `mobile/package.json` `overrides` pins React to the Expo SDK's version (19.2.3). *Why:* Expo packages declare loose React peer deps; npm hoisted 19.3.0, giving two React copies, which breaks native builds (caught by `expo-doctor`).
- **Expo template trimmed** to an empty shell: demo screens, components, Expo logos, and the template's `.claude/settings.json` (which silently enabled an Expo Claude plugin) were removed. App icons/splash are still Expo placeholders.
- **Placeholder mobile identifiers:** `com.techshop.customer`, `com.techshop.logistics`. Must be confirmed before the first store submission (they can't change afterwards).
- **Seller and staff apps are non-indexed** (`robots: noindex`) — internal/partner tools must not appear in search results.
- **CI on Node 24** (token generator runs TypeScript natively, needs Node ≥ 22.18). New CI jobs: `design-tokens` (generated CSS up to date) and `mobile` (lint, typecheck, expo-doctor).
- **Backend CORS** default now allows all five web apps on their ports and `*.techshop.localhost` hostnames.

**Follow-ups:** replace app icons/splash with TechShop artwork; confirm bundle IDs; design the staff portal's role model (which roles see which of the 20 modules) before building the first module.
