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

## 2026-10-02 — First homepages for all five web apps

**Owner's answers:** retail-overlap decision (market vs wholesale) deferred until the UI is visible; storefronts cover **all categories**; build homepages for **all 5 web apps**.

**Plan & decisions:**
- **UI/UX rulebook:** `.claude/rules/frontend-design.md` — laws of UX, Nielsen heuristics, WCAG 2.2 AA, visual/e-commerce/forms/motion/content/performance rules, and a definition of done. Scoped with `paths` to `frontend/**`, `mobile/**`, `shared/design-tokens/**` so it loads only for UI work. The `frontend-design` skill keeps the design-system *how*; the rule holds the *why*.
- **Shared React components** go in `@techshop/ui` (alongside the CSS), each with a CSS Module: logo, icons, category illustrations, price, rating, product card, section header, site header, site footer, countdown. *Why:* five apps must look and behave consistently (Jakob's Law, heuristic #4).
- **Domain types** (category, product, money, seller, car listing) go in `shared/api-client` — they are the future API contract, shared with mobile.
- **Sample data** lives in a separate `@techshop/fixtures` web package, clearly labelled as sample data, to be replaced by API calls. Fictional brands and products only — no real brand names, logos, or photos.
- **Money in kobo** (integer minor units), formatted as `₦1,250,000` via `Intl.NumberFormat("en-NG")`. *Why:* avoids floating-point errors; standard for NGN payment gateways (Paystack also uses kobo).
- **Product imagery:** original simple SVG category illustrations as placeholders until real photography exists.
- **Homepage concepts:**
  - *Market* — search-first header with category scope, category nav, hero + side promos, trust strip, shop by category, flash deals with a real countdown, per-category shelves (phones, laptops by type, gaming, accessories, smart home, office & workstations), cars as listing cards (no add-to-cart), marketplace "Sold by" on every card, sell-on-TechShop band, footer with payment methods.
  - *Wholesale & retail* — retail/wholesale mode, tiered price tables, MOQ, request-a-quote, business account CTA, how it works, who we serve.
  - *Corporate* — restrained company site: mission hero, our businesses, what we sell, why TechShop, careers, newsroom placeholders, contact.
  - *Seller centre* — vendor landing: benefits, how it works, requirements, FAQ (native `<details>`), start-selling CTA.
  - *Staff portal* — app shell: grouped sidebar of all 20 modules, top bar, dashboard with sample KPIs, recent orders table, attention queue, module shortcuts. No auth or RBAC yet.
- **No invented business claims presented as real:** policies, stats, and fees are placeholders marked as sample content.

## 2026-10-02 — Homepage implementation decisions

- **Sample-data honesty:** a `SampleDataNotice` banner appears on every page that shows sample products, prices or figures (market, wholesale, staff). Corporate and seller show no sample figures, so they don't have one. Their copy (payout timing, verification steps, policies) is **draft copy that the business must confirm** before launch.
- **New token** `--color-on-fill` (text on sale/promo fills), replacing raw `#ffffff` in badges and the countdown.
- **Staff dashboard renders per request** (`await connection()`) so dates and figures are never frozen at build time.
- **Staff orders table:** the order time sits under the order ID instead of in its own column, so the table fits beside the attention panel at 1440px without clipping.
- **`data-scroll-behavior="smooth"`** on every `<html>`: Next 16 no longer resets smooth scrolling on navigation by itself.
- **Navigation without JavaScript:** mobile menus and the FAQ use native `<details>`; search is a plain GET form.
- **Links point to routes that don't exist yet** (category, product, cart, account, help, …). They are the planned URL structure: `/c/[category]`, `/c/[category]/[sub]`, `/p/[slug]`, `/cars/[slug]`, `/deals`, `/search`.

**Follow-ups:** real product photography (replace `CategoryArt`); confirm all policy and marketing copy; decide the market/wholesale retail overlap after reviewing the UI; catalogue API endpoints so fixtures can be removed; role model for the staff portal.

## 2026-10-02 — Corporate dev port moved to 3005

**Decision:** The corporate app runs on **3005** instead of 3000. Updated its `package.json` scripts, nginx routing, backend CORS defaults and `.env.example`, and the READMEs. Added "Changing an app's port" instructions to `infra/README.md`.
**Why:** Port 3000 is taken by AdGuard on the owner's machine.
**Supersedes:** "Dev ports: corporate 3000" in "Multi-app monorepo restructure".

## 2026-10-02 — Market shopping journey (phase 1 of remaining pages)

**Owner's answers:** build the market shopping journey first (about 70 linked pages exist in total; the others follow in later phases). Sign-in, registration, cart and checkout are **interface only**: real forms and states, no submission, no auth, no payments.

**Pages:** `/c/[category]`, `/c/[category]/[sub]`, `/p/[slug]`, `/search`, `/deals`, `/best-sellers`, `/cars/[slug]` (cars listing at `/c/cars`), `/cart`, `/saved`, plus a market 404 page.

**Decisions:**
- **Filtering and sorting are server-side via URL query params** (`?condition=new&seller=techshop&min=…&max=…&sort=price-asc`), driven by a plain GET form. *Why:* works without JavaScript, is shareable and bookmarkable, back/forward behave correctly, and the same params map onto a future API query.
- **Mobile filters** collapse into a `<details>` disclosure above the results; desktop shows a sidebar (`.with-sidebar`).
- **Product and car pages are statically generated** from the catalogue (`generateStaticParams`, `dynamicParams = false`), so unknown slugs return 404. This will switch to on-demand rendering when the API exists.
- **Cars use a viewing/enquiry flow, never add-to-cart** (rulebook §6). The viewing form is interface only.
- **Cart is interface only:** it shows sample lines grouped by seller (marketplace rule), with client-side quantity and remove, and a summary marked with an estimated delivery fee. Nothing persists.
- **Product page:** breadcrumbs, gallery placeholder, price/condition/seller/warranty/delivery near the buy button, a specs table, related products, and on mobile a sticky buy bar in the thumb zone.
- **New shared components in `@techshop/ui`:** Breadcrumbs, EmptyState, QuantityStepper (client).

**Implementation notes (shopping journey):**
- New layout modifier `.section--page` (tight top padding for inner pages) added to the design system.
- Subcategory links show as a scrollable chip row on mobile and a list in the sidebar from 1024px.
- Search and the cart, saved, search and 404 pages are `noindex`. Product and car pages have per-item titles and descriptions.
- Sample-only details are labelled on the page: delivery times and fees, warranty wording, the inspection checklist, and the cart contents.
- Known gap: submitting filters can leave empty params in the URL (`?min=&max=`). Harmless (they're ignored), but it can be tidied later.

## 2026-10-02 — Remaining pages (content, wholesale, seller, staff)

**Owner's request:** "build the rest", meaning every page the five apps link to that doesn't exist yet. The interface-only rule for auth, forms and payments still applies. The staff portal has no role-based access yet (the role model is undecided).

**Approach: templates, not one-off pages.**
- **Shared templates in `@techshop/ui`:**
  - `PageHeader`: breadcrumbs, eyebrow, h1, lead, actions.
  - `Prose`: long-form text styling.
  - `FaqList`: native `<details>`.
  - `LegalDocument`: structured outline.
  - `PreviewForm`: a client form driven by a field config, with validation on blur and submit, focus on the first error, `aria-invalid`/`aria-describedby`, Nigerian phone normalisation, and an honest "nothing was sent" confirmation.
- **Phone normalisation moves to `shared/api-client`** (platform-neutral, so mobile reuses it).
- **Staff modules** use one `ModulePage` template (header, figures, tabs, sample table) configured per module. Orders also gets a detail page.

**Content honesty rules for these pages:**
- **Legal pages** (terms, privacy, cookies, seller agreement) are **structured outlines marked "draft — to be written by legal counsel"**, never invented legal terms. Nigeria's NDPA 2023 governs privacy.
- **Careers and newsroom:** no fake job posts or press releases. They show empty states plus a way to get in touch.
- **No invented company facts** (founding year, headcount, office addresses, phone numbers). Contact pages use forms and email-style placeholders marked as sample.
- **Fees and commissions** are not stated as numbers until the business sets them.

**Routes:**
- **Market:** `/account`, `/orders`, `/delivery`, `/help`, `/help/[topic]` (delivery, returns, warranty, contact), `/trade-in`, `/legal/[doc]`.
- **Wholesale:** `/retail`, `/wholesale`, `/categories`, `/c/[category]`, `/p/[slug]`, `/quote`, `/business/register`, `/business/sign-in`, `/business/credit`, `/business/invoices`, `/contact`, `/help`, `/legal/[doc]`, 404.
- **Corporate:** `/about`, `/partners`, `/careers`, `/news`, `/contact`, `/legal/[doc]`, 404.
- **Seller:** `/register`, `/sign-in`, `/fees`, `/policies`, `/policies/prohibited`, `/help`, `/help/contact`, `/legal/[doc]` (incl. seller agreement), 404.
- **Staff:** 19 module routes, `/orders/[id]`, `/notifications`, `/account`, `/search`, 404.

**Verification plan:** build all apps, then a crawler that follows every internal link in all five apps and must find zero broken links, an overflow check at 390px on every page, screenshots of each template, and interaction tests for `PreviewForm`.

**Implementation notes (remaining pages):**
- Added the `.card--roomy` modifier and the `NavDisclosure` component. Every mobile menu now closes after navigation and on Escape (previously a `<details>` menu stayed open over the new page).
- The staff sidebar is now path-aware (`usePathname`), so the current module gets `aria-current`.
- Sample orders now carry real line items whose quantities and prices sum exactly to the order totals (the order detail page previously showed unrelated items). Wholesale sample stock is attributed to TechShop.
- Corporate inner pages use bordered cards (borderless cards were invisible on white).
- **Still interface only:** every form, sign-in, quote, cart and checkout. **Still draft:** legal outlines, help/policy copy, credit and fees wording. **Still undecided:** the staff role model (all modules visible to everyone).

## 2026-10-02 — Database design plan (documentation only)

**Owner's request:** a full database plan with relationships, schemas and every diagram.

**Scope:** documentation in `docs/database/` only. No migrations and no changes to `backend/` (the owner's standing instruction). The SQL is a *proposal* to be turned into goose migrations later.

**Decisions (proposed, for owner review):**
- **PostgreSQL 17**, owned exclusively by the Go API (unchanged rule).
- **Primary keys:** `uuid` with `gen_random_uuid()`. Human-facing references (order `TS-10482`, invoices, RMAs, tickets…) come from sequences. Revisit UUIDv7 when moving to Postgres 18.
- **Money:** `bigint` kobo plus `char(3)` currency (default `NGN`), never floats. Matches the frontend `Money` type and Paystack's units.
- **Status fields:** `text` with `CHECK` constraints instead of Postgres enum types. *Why:* adding a value is a one-line migration, and sqlc handles text cleanly.
- **One marketplace model:** TechShop itself is a row in `sellers` (`type = first_party`). Products are catalogue entries; **listings** are seller offers (price, condition, stock) on a product variant. *Why:* one code path for own stock and vendors, Amazon-style.
- **Order split:** one `orders` row per checkout, one `fulfilments` row per seller, `order_lines` snapshot names and prices.
- **Money movements:** a **double-entry ledger** (`ledger_journals` and `ledger_entries`, balanced per journal by a deferred constraint trigger) for customer payments, commission, seller payables, payouts and refunds. *Why:* marketplace money must always reconcile.
- **Serialised devices:** `device_units` (IMEI/serial) for phones and laptops, plus an IMEI blocklist.
- **Cars:** their own tables (listings, inspections, documents, viewings, financing), separate from the product catalogue.
- **Staff access:** RBAC (`roles`, `permissions`, `role_permissions`, `staff_roles`). A default role set is **proposed only**, because the role model is still the owner's decision.
- **PII:** NIN, bank account numbers and similar are encrypted in the application (`bytea`), with only the last 4 digits stored in plain text (NDPA 2023).
- **Platform:** `audit_log` (append-only), `outbox_events` (reliable notifications/integrations), `idempotency_keys`, `feature_flags`.
- **Table comments carry the domain and purpose**, so the ER diagrams and the table index are generated from the live schema and can't drift.

**Deliverables:**
- `README.md`: principles, domain map, table index, migration phases.
- `schema.sql`.
- `erd.md`: generated ER diagrams.
- `states.md`: state machines.
- `flows.md`: sequence and money-flow diagrams.
- `access.md`: RBAC proposal and data ownership by app.
- `tools/` to regenerate, plus rendered SVGs.

**Verification:** load `schema.sql` into a throwaway Postgres 17 container, run constraint tests (ledger balance, money checks, uniqueness), and render every Mermaid diagram with mermaid-cli.

## 2026-10-03 — Database plan completed and verified

**Outcome:** the full documentation set from the 2026-10-02 "Database design plan" entry exists and is verified. The proposals in that entry stand unchanged. Still open for the owner: the staff role model (`access.md` §3 is a proposal), commission rates, VAT treatment (journal templates are placeholders) and the retail overlap. Schema size: 96 tables in 15 domains, 185 foreign keys, 229 check constraints.

## 2026-10-03 — Mobile, backend and recommendation-system plans (documentation only)

**Owner's request:** plan the complete mobile apps, the complete backend, and a Python recommendation system (three plans run in parallel). For recommendations: explain the integration and project impact first, answer every clarification, **use all industry approaches**, **combine all models (best of both worlds)**, and document everything.

**Decisions:**
1. **Recommendations follow the industry multi-stage architecture, with every pattern adopted:**
   - event logging from day one;
   - popularity and heuristics;
   - Amazon item-to-item collaborative filtering;
   - Netflix-style one strategy per shelf;
   - two-stage retrieval → ranking;
   - learning-to-rank;
   - business-rule re-ranking.
2. **Models: a hybrid ensemble.**
   - Generators: popularity, content kNN, co-occurrence, ALS + BPR, item2vec, two-tower and SASRec, blended by a LightGBM LambdaMART ranker (feature-weighted blending until click logs exist).
   - Every generator is built but **gated**: it is weighted only after meeting its data threshold and beating the baseline offline.
   - LightFM is excluded (unmaintained).
3. **Integration: combined, and "only Go touches the DB" is preserved.**
   - Python nightly batch via Go internal export/import endpoints (pseudonymised, no DB access).
   - Go live session re-ranking.
   - A feature-flagged Python live service (FastAPI) for real-time models, behind an 80 ms timeout, circuit breaker and Go fallback.
   - Read replica rejected.
   - A `CLAUDE.md` amendment is **proposed only**.
4. **Behaviour tracking at launch, with consent** (NDPA). Personal shelves only with consent.
5. **Fairness, combined policy:** neutral ranking, a configurable vendor exposure floor, an optional capped first-party boost (default 0; legal review), and an exposure dashboard.
6. **Plan reconciliations** are recorded in `docs/architecture-decisions.md`.
7. **Schema deltas** are consolidated in `docs/schema-changes.md`. `schema.sql` is unchanged until the owner approves.
8. **Not touched:** `backend/`, `frontend/`, `mobile/`, `shared/`, `CLAUDE.md`, `schema.sql`.

## 2026-10-03 — Interactive mockups (prototype + system simulation)

The owner asked for (1) a fully interactive mockup of every web app and both mobile apps, and (2) an interactive simulation of how the backend and recommendation system communicate.

**Decisions:**
1. Both are standalone private claude.ai artifact pages, **not** part of the repo or any app. They are throwaway visual aids; the real implementation still follows `frontend/`, `mobile/`, `backend.md`, `mobile.md` and `recommendations.md`.
2. **Prototype** (https://claude.ai/artifact/R9tuhiEChYWm1vvVk9NSDK):
   - All seven apps (Market, Wholesale, Corporate, Seller Centre, Staff portal, Customer app, Logistics app) share one in-memory state, so an order placed on one app shows up in the others.
   - Uses the real design tokens and the fictional fixture data. Payments, SMS and approvals are simulated; nothing is sent anywhere.
3. **Simulator** (https://claude.ai/artifact/LDLEc8UZvPVinrgcYpDv8N):
   - Six scenarios: checkout and payment, delivery and OTP, nightly recs training, live recommendations, seller payout, and return and refund.
   - Each has failure toggles and a recommender data-maturity selector. The flows mirror `backend.md` and `recommendations.md`, including the outbox, webhook verification, posting keys, gated generators, the 80 ms live-service fallback and consent.
4. **Not touched:** `backend/`, `frontend/`, `mobile/`, `shared/`, `CLAUDE.md`, `schema.sql`.

## 2026-10-03 — System plans, mockups and simulations (messaging, identity, payments, search, orders, trust)

The owner asked for complete plans, an interactive UI mockup and an interactive system simulation for the mailing and marketing service, the auth system and "all other big and important systems", with files created and documented.

**Decisions:**
1. **Scope: six systems.** Messaging & marketing, identity & access, payments & finance, search & catalogue, orders/inventory/fulfilment, trust & safety. Logistics, recommendations and the general backend already have plans (`mobile.md`, `recommendations.md`, `backend.md`).
2. **One plan doc per system** in `docs/` (status, summary, architecture, data model, API, screens, milestones, risks, open owner decisions), with Mermaid diagrams rendered to `docs/diagrams/`.
3. **One interactive page per system** with two tabs: a clickable UI mockup and an animated system simulation with failure toggles. All pages are built from one shared kit (tokens, click framework, simulation engine), saved in `docs/mockups/` (with the earlier prototype and system simulator), and published privately as claude.ai artifacts. They are visual aids, not app code.
4. **Key design choices:**
   - **Messaging:** one send pipeline (`notify.Send`) with critical, transactional and marketing queues; marketing only with per-channel consent; separate transactional and marketing sending domains; approval with separation of duties for large campaigns.
   - **Identity:** phone OTP for customers; password + TOTP for staff; 10/5-minute JWTs with no permissions inside; rotating refresh tokens with reuse detection; BFF cookies; step-up for sensitive permissions; Postgres-backed OTP limits against SMS pumping.
   - **Payments:** webhook + verify + sweep + daily three-way reconciliation; posting keys; reversals only; period close; dual-control refunds and payout batches; pay-by-transfer with under- and overpayment handling.
   - **Search:** Postgres read model with synonyms, intents, trigram fallback and an explainable ranking; buy box on one formula for TechShop and vendors; moderation with FCCPA was-price guard.
   - **Orders:** fulfilments per seller with derived order status; race-free conditional reservations; single stock code path; scan-driven warehouse; seller SLAs; returns with serial match and grading.
   - **Trust:** KYC behind a provider port; a Go rules engine with scores, holds and a test bench; case queue with a link graph; IMEI Luhn + blocklist at every intake; seller enforcement ladder with appeals.
5. **Cross-system reconciliations:** recorded as #16–#22 in `docs/architecture-decisions.md`.
6. **Schema:** 42 new proposals (#59–#100) added to `docs/schema-changes.md` §4–§9. **Not applied**; `schema.sql` unchanged until the owner approves.
7. **Open owner decisions:** listed at the end of each system doc (providers, thresholds, role list, return policy, carriers, KYC provider, …).
8. **Not touched:** `backend/`, `frontend/`, `mobile/`, `shared/`, `CLAUDE.md`, `schema.sql`.

## 2026-10-05 — Project `.claude/` setup reconciled

The owner added project-only Claude settings: `.claude/CLAUDE.md` working rules, security and compliance rules in `.claude/rules/` (00–10), 16 skills and `.claude/settings.json`.

**Decisions (owner's answers):**
1. **"The backend should never hit the DB directly"** means both of these:
   - only the Go backend's data layer (sqlc queries used from repository and service code) touches PostgreSQL, with no SQL in HTTP handlers and no ad-hoc database access;
   - connections go over the Docker network only.

   This refines, and doesn't replace, "only the Go API touches PostgreSQL".
2. **`.claude/CLAUDE.md` is rewritten** so it doesn't duplicate the root `CLAUDE.md`:
   - the root file stays the codebase guide;
   - `.claude/CLAUDE.md` holds the owner's working rules: ask before assuming or adding dependencies, simple explanations, no pushes or SSH, test every API, security and compliance via `.claude/rules/`, no weak secrets, Docker files in `infra/docker/`, dev-only test accounts.
3. **Skills adjusted to this project (only `.claude/` edited):**
   - **UI guidance:** Tailwind and shadcn advice replaced with TechShop's token + custom-CSS system.
   - **Tenancy:** made optional in the scaffolding, conventions and migration skills, with TechShop's ownership scoping, goose + sqlc workflow and API conventions added.
   - **DevOps:** project notes added.
   - **Domain knowledge:** `project-domain-knowledge/references/techshop.md` created.
4. **`.claude/settings.json` rewritten** to valid allow, ask and deny patterns so the owner's rules are actually enforced:
   - **Ask:** new dependencies, deletes and moves, pushes, migrations up, docker down.
   - **Deny:** git add/commit, force-push, hard reset, `rm -rf`, migrations down or reset, volume deletion, SSH/SCP/SFTP, reading or editing real `.env` files.

   The free-text entries (not supported by Claude Code) moved into `.claude/CLAUDE.md`.
5. **Backend items wait** until it's time for backend work: dev test accounts (seed, dev-only) and refusing weak or default secrets at startup.

## 2026-10-05 — Database built (migrations are now the source of truth)

The owner asked for the complete database, with "all schemas, tables and all".

**Owner decisions:**
1. **Scope:** build from the reviewed schema plus **all 100 proposals** in `docs/schema-changes.md`.
2. **Layout:** **one Postgres schema per domain**, 22 in total: `platform`, `identity`, `sellers`, `b2b`, `catalog`, `search`, `inventory`, `purchasing`, `pos`, `sales`, `fulfilment`, `logistics`, `payments`, `finance`, `aftersales`, `messaging`, `marketing`, `content`, `support`, `risk`, `cars`, `personalisation`.
3. **Backend work allowed:** write goose migrations and run them locally. Nothing else in `backend/`.

**Decisions taken while building:**
- **Migrations are the source of truth:** `backend/db/migrations/00002`–`00022`. `docs/database/schema.sql` is now a generated dump. This supersedes the phasing in `database.md` §9 and the order in `backend.md` §4.1.
- **Indexes and triggers are explicit:** every FK index and `updated_at` trigger is written out, not created by loops, as `backend.md` §4.1 asked.
- **High-volume logs are partitioned** monthly with a default partition: `user_events`, `auth_events`, `message_events`, `risk_decisions`, `search_queries`, `rider_location_pings`.
- **Deviations from the proposals:** listed in `schema-changes.md` §10. In short: deleted users may lack email and phone; reviews may be unverified; `open_box` condition added; some tables changed domain.
- **Deferred:**
  - River job tables (the River dependency needs owner approval);
  - LGAs and holidays data;
  - dev test accounts;
  - staff roles other than `admin`.

## 2026-10-05 — Backend build started (M0 → M7)

The owner approved building the complete backend in milestone order, with these packages: Huma v2, River, golang-jwt v5, x/crypto (argon2id), pquerna/otp, aws-sdk-go-v2 (S3) and maroto (PDFs).

**Decisions:**
1. **Structure:** Huma v2 on the existing Gin router.
   - **Modules:** `internal/<domain>` modules behind a small `kit.Module` interface, wired in `internal/app`.
   - **Transactions and jobs:** a unit of work (`internal/uow`) writes outbox events, audit rows and River jobs in the same commit.
   - **Background work:** River runs jobs in its own `river` schema (migration `00023_river_jobs.sql`, generated from River's own migration files). The outbox relay wakes on `pg_notify('outbox')`.
   - **Realtime:** SSE plus LISTEN/NOTIFY, which also handles cache invalidation for permissions and sessions.
2. **Permission catalogue:**
   - The catalogue (`internal/rbac/permissions.go`) is synced to the database at API start-up, together with the system `admin` role.
   - The other staff roles are **dev seed data only** until the owner approves the role list.
   - The B2B permission keys are named `business.*`, because the schema's key check doesn't allow digits.
3. **Secrets:** config refuses missing, short, default or placeholder secrets outside development. `LOCAL_KEK`, `PAYMENTS_FAKE` and the log providers are development-only.
4. **5xx errors** are logged with the request ID and never expose internals.
5. **Search** is the one sanctioned dynamic query: `internal/catalog/search_query.go`, fully parameterised. Every other query is sqlc.
6. **Testing:** everything runs in throwaway containers on a private Docker network. Postgres, the API and the end-to-end test binary are all containers, so nothing connects to a database from the host. Binaries are copied in with `docker cp`, because Docker Desktop can't mount the scratchpad.

## 2026-10-05 — Backend M2 (commerce) decisions

1. **Module shape:** `ledger` → `marketing` → `sales` (cart, pricing, checkout, orders) → `payments`. Sales reaches payments only through a `PaymentPort` interface, so imports point one way.
2. **Pricing order:** list price (tier price on the wholesale channel only) → flash price (signed-in only, within stock and per-person limits) → the **single best automatic promotion** (no stacking) → **at most one coupon** on what remains → delivery per seller → VAT → `priceHash`. Checkout reprices inside the order transaction; a different hash is `409 price_changed`.
3. **VAT and order totals (PLACEHOLDER until finance confirms):** consumer prices are VAT-inclusive; VAT is computed only on TechShop's own (first-party) lines at `VAT_BPS`. Because the schema requires `total = subtotal + delivery − discount + vat`, `orders.subtotal_kobo` stores the items total **net of VAT**; screens show `itemsKobo` (gross) from the API.
4. **Who pays for a discount:** without a schema change, each line's discount has exactly one funder.
   - Seller-funded promotions must target only that seller's listings and can't carry coupons; a TechShop coupon never stacks on a line with a seller-funded discount.
   - At payment the ledger asks "did a seller-funded promotion target this listing when the order was placed?".
   - TechShop-funded discounts on marketplace items post to a new account **`expense:promotions`** (sellers are still paid their full price). This adds one account to the chart in `payments-finance.md` §4.1.
5. **Delivery fees** are flat config values by who ships and whether the parcel crosses a state line, until logistics zones and rates arrive (M3).
6. **Payments:**
   - every provider is a port; **fakes are the default** and include a development checkout page (`/dev/pay/...`) that sends signed webhooks;
   - only Paystack has a live adapter; config refuses Monnify and OPay outside development until theirs are built and confirmed against the sandboxes;
   - webhooks are checked on the raw body, stored once per provider event and processed by a job that **re-verifies** with the provider;
   - card amount mismatches go to `pending_review` with a reconciliation exception; transfer under- and overpayments follow §4.3;
   - the expiry job asks the provider before cancelling; a 5-minute sweep re-verifies stale payments;
   - circuit breaker per provider (5 failures open it for a minute).
7. **Late payment (PLACEHOLDER, owner decision #57):** `LATE_PAYMENT_POLICY=reinstate_or_refund` — reinstate if all stock can be reserved again (savepoint), otherwise book the money to `liability:customer_refunds` and raise a refund request. Audited.
8. **Refunds:** every refund, including automatic ones (cancellations, overpayments), waits for a human approver. Approver ≠ requester; above `REFUND_MANAGER_THRESHOLD_KOBO` needs `refunds.approve`; step-up required. Refunds of excess money are marked with the reason prefix `Excess payment:` and post no approve journal.
9. **Ledger:** typed templates only (`Sale`, `Commission`, `RefundApprove`, `RefundPaid`, `Reverse`, plus `Adjust` behind `ledger.adjust` and step-up); `posting_key` makes every posting idempotent; a refund reverses the sale's credits pro rata; nightly checks raise reconciliation exceptions.
10. **Marketing:**
    - promotion values: `percentage` = whole percent (≤ 90), `fixed_amount` = kobo;
    - audiences are a typed allowlist compiled to parameterised SQL (the second sanctioned dynamic query, after search);
    - every campaign needs approval by someone other than its creator, with `marketing.approve` and step-up;
    - recipients are snapshotted at send time with a stable hash-based holdout and A/B split;
    - journeys run every minute (wait, condition, branch, send, issue coupon, exit); triggers are sign-up, abandoned cart (hourly scan), order paid and order delivered;
    - orders are attributed to the last marketing click within 7 days.
11. **Guest checkout:** carts work for guests (`X-Cart-Token`, merged on sign-in), but placing an order needs a signed-in customer. This follows the recommended "phone OTP at checkout" (architecture decision #15 is still for the owner to confirm).

## 2026-10-06 — Correction queued: backend layout (do after M3)

The owner flagged that the backend ignored their folders and lumped each feature into large `http.go` + `service.go` files.

**Cause:**
- The owner's `backend/pkg/` (empty) and an intended `internal/modules/` were not checked before work started.
- The `coding-conventions` skill note said `docs/backend.md` "takes precedence", so its layered per-feature layout was not followed.

**Correction, to do after M3 finishes (owner: not now, to save tokens):**
1. Restructure every feature into `internal/modules/<feature>/`, with handler, routes, dto, service, repository (wraps sqlc), model, errors and tests per feature. Shared app plumbing goes in `internal/platform/`. Exact layout to confirm with the owner before starting.
2. `pkg/` holds reusable, non-TechShop code: money (kobo), validation, crypto helpers, HTTP error helpers, pagination cursors, and **the Postgres connection client**. Design the DB clients so **Redis and MongoDB** can be added later. That reverses the earlier "no Redis" decision (architecture-decisions #12), so it needs an owner decision when it happens.
3. Remove the "takes precedence" note from the coding-conventions skill so the layered layout is the default.
4. Re-run all unit and end-to-end tests after the move (same behaviour).

## 2026-10-05 — Backend M3 (logistics and fulfilment) decisions; build paused after M3

**Scope change from the owner:** stop after M3. M4–M7 are not started.

1. **Modules:** `logistics` (zones, rates, riders, devices, shifts, delivery jobs, rider app, dispatch) and `wms` (waves, picking, packing, manifests, carriers). Both move fulfilments only through `sales.Move`, the single fulfilment state machine (`docs/orders-fulfilment.md` §2.1). Delivery posts the marketplace commission journal and tells the customer.
2. **Zone pricing:** sales asks logistics through a `DeliveryQuoter` port (imports stay one-way). The zone for the address (LGA-specific first, then state-wide) and the smallest weight band that fits set the fee. With no zone, the M2 flat fees still apply.
3. **Rider or carrier is decided at packing:** an address inside a delivery zone gets an own-rider delivery job; anywhere else a carrier shipment is booked. Carriers sit behind a port with **only a fake carrier** until the owner chooses couriers (open decision #9).
4. **Riders:**
   - riders are staff members who sign in with phone OTP on the logistics app;
   - a shift needs location consent and a phone approved by someone else (database check). The device is read from the signed-in session, never from a client header;
   - shifts close automatically after 12 hours; raw location points are deleted after 90 days.
5. **Delivery proof:**
   - picking up needs the scanned parcel label;
   - the customer gets a 6-digit delivery code (HMAC-hashed, 24 hours, 5 tries; wrong tries are counted even though the action is rejected);
   - photo proof is the fallback, marked `photo_offline` when recorded without signal.
6. **Offline sync:** one batch endpoint applies each action once per `clientEventId`. Each item runs in its own savepoint, so one rejected item (for example `job_reassigned`) never blocks the rest.
7. **Privacy:**
   - riders see the recipient's first name, address and landmark — never the phone number;
   - customers see the rider's position only while their own parcel is en route; dispatch sees riders on shift.
8. **Warehouse:**
   - a wave takes paid TechShop lines committed in that warehouse;
   - serialised devices must be picked by IMEI or serial: the unit must be in stock in that warehouse for the right model and condition, and it is bound to the order line as sold;
   - packing checks the weight against the catalogue (flagged when more than 10% off);
   - manifests list the parcels leaving with one rider or one carrier; the signature hands them over;
   - carrier tracking webhooks are signed and stored once per event.

## 2026-10-06: Recommendation project scaffolded

- **Location:** uses the owner's root folder `recommendation/` (the plan said `recommender/`; the owner's folder wins).
- **Structure:** follows `docs/recommendations.md` §9.
- **Dockerfile:** placed at `infra/docker/recommendation.Dockerfile`, per the rule that all Docker files go in `infra/docker/`.
- **Dependencies:** declared in `pyproject.toml` from the approved plan; nothing installed yet (`uv sync` when work starts).

## 2026-10-06: `pkg/` done (step 2 of the queued layout correction)

- **Done:** reusable code now lives in `backend/pkg/` (money, validate, crypto, pagination, logging, `database/postgres`), with `redis/` and `mongo/` reserved. Adding Redis still needs the owner's decision, because it reverses architecture decision #12.
- **Still queued:** the per-feature `internal/modules/<feature>/` restructure.

## 2026-10-06 — Recommender gets its own database; "be logical and critical" rule

1. **Owner decision:** the Python recommender owns a private Postgres (`recs-postgres` in `infra/docker/compose.recs.yml`, no published port, required password).
   - A Go job pulls results through the recommender's authenticated API and stores them in `personalisation.rec_*`; the site serves from Go with fallbacks.
   - No service connects to another's database. This is now a rule in `CLAUDE.md` and `.claude/CLAUDE.md`, and `architecture-decisions.md` #29.
2. **Postgres chosen over MongoDB** (model outputs are tabular). Redis remains undecided (it would reverse decision #12).
3. **New working rule** in `.claude/CLAUDE.md` §1: be logical and critical; challenge ideas (including the owner's), give trade-offs, then recommend.

## 2026-10-06 — No-Redis rule dropped

- **Owner decision:** Redis is allowed. Recorded as `architecture-decisions.md` #30, which supersedes #12.
- **Limits:** Redis is for caching and counters only; Postgres remains the source of truth for money, orders, stock, sessions, audit logs and jobs.
- **Docs and skills** updated (backend.md §5.4 note, techshop.md, the devops and coding-conventions skills, `pkg/database/doc.go`).
- **Nothing built yet:** the Redis client, compose service and the move of counters/caches happen when that work is scheduled.

## 2026-10-06 — Recommender database confirmed: Postgres

The owner asked whether MongoDB suits ML better. Answer: training data lives in files (Parquet), the database only holds tabular results and model bookkeeping, and Postgres covers flexibility (`jsonb`) and vectors (`pgvector`). The owner kept Postgres (decision #29 stands).

## 2026-10-20 — Complete frontend build, then mobile (overnight session)

The owner asked for the complete frontend, then mobile, built feature by feature, each endpoint tested before moving on, pushed to GitHub one piece at a time, with no co-author lines. Work starts 01:30 and stops at 07:30 even if unfinished. The owner said to skip any rule that requires asking for permission and to research instead of guessing.

**Plan (in order):**
1. **Typed API client** in `shared/api-client`, generated from `backend/api/openapi.yaml` (the pipeline in `backend.md` §2.1).
2. **Web foundation:** a shared BFF session kit (encrypted `httpOnly` cookie, refresh, sign-out) as `identity-access.md` §5.1 describes.
3. **Market** wired to the API feature by feature: catalogue → sign-in → cart → checkout and payment → orders and tracking → account.
4. **Staff portal:** sign-in (password + TOTP), then orders, catalogue, finance, marketing, dispatch, warehouse, inventory, content, admin.
5. **Wholesale, corporate, seller:** wire what the backend already serves. B2B (M6) and seller (M4) endpoints don't exist yet, so those screens keep clearly labelled sample data.
6. **Mobile:** the logistics app first (its backend, M3, is complete), then the customer app.

**Decisions:**
- **Branch, not `main`:** work goes to `feat/frontend-integration` and is pushed commit by commit. `main` stays reviewable and a bad commit never lands there directly.
- **Generated types, hand-written client:** `openapi-typescript@7.13.0` (dev dependency in `frontend/`) generates `shared/api-client/src/generated/schema.ts`. The runtime client is ~150 lines written here instead of adding `openapi-fetch`. *Why:* `shared/` sits outside both install roots, so a runtime dependency there would not resolve from `frontend/node_modules` or `mobile/node_modules` without a third install root. Types-only generation avoids that and adds no shipped code.
- **Client shape:** `api.get(path, { path, query, body, token, idempotencyKey, init })` returns the success body; `api.request.*` also returns status and headers. Errors are `ApiError` carrying the problem+json `code`, field errors and `X-Request-Id`. Required path, query and body parameters are compile-time errors when missing.
- **Tests without new tools:** Node's built-in test runner (`node --test`, native TypeScript stripping). Unit tests run in CI; contract tests call a real seeded local API (`test/contract/`), because CI has no running API.
- **CI:** the frontend job fails if the generated types are stale.
- **Known audit findings (not fixed):** 5 "high" `npm audit` findings, all in ESLint's `braces`/`micromatch` chain under `eslint-config-next`. They existed before this work, are dev-only and are never shipped. They clear when `eslint-config-next` updates.

## 2026-10-20 — Web session kit (`@techshop/bff`)

Implements `identity-access.md` §5.1 as a shared package in `frontend/packages/bff`, used by all five web apps.

- **Cookie:** one per app, `__Host-ts_session` over HTTPS (`ts_session` on plain-HTTP localhost, where the `__Host-` prefix can't be used). `HttpOnly`, `SameSite=Lax`, `Path=/`, no `Domain`, lifetime = refresh-token expiry. The value is AES-256-GCM sealed (Web Crypto, no dependency) and holds the session and user IDs plus both tokens.
- **Refresh lives in `proxy.ts`:** Next 16 Server Components can read cookies but not set them (checked in the bundled Next docs), and the API rotates the refresh token on every use. A page that refreshed without saving the new token would make the next request reuse the old one, and the API would revoke the session family. So each app's `proxy.ts` refreshes when the access token is within 60 s of expiry, sets the new cookie, and forwards it to the same render. Server Functions can also refresh (`getSession({ refresh: true })`), because they can write cookies.
- **Failure handling:** a 4xx from `/auth/refresh` signs the user out (cookie cleared). A 5xx or network error keeps the cookie, so an API blip doesn't sign everyone out.
- **The browser never calls the API.** `NEXT_PUBLIC_API_URL` was removed from every `.env.example` (no code used it). Server code calls the API with `Authorization: Bearer`.
- **CSRF:** Server Actions check Origin themselves. Custom Route Handlers that change state call `assertSameOrigin()` (`Sec-Fetch-Site`, then `Origin` vs `Host`).
- **Open redirects:** `?next=` accepts only plain relative paths (`//x`, `/\x`, `https:x` and control characters are rejected).
- **Secrets:** `SESSION_SECRET` (≥ 32 chars, no placeholder words) is required when `NODE_ENV=production`, which includes local `next start`. `next dev` uses a fixed, clearly named development-only key, so proxy and render agree without setup.

## 2026-10-20 — Market wired to the API

**Layout (owner rule: group by feature, reusables in one place):**
- `apps/market/src/app/` holds routes only. Each page is a thin server component.
- `apps/market/src/features/<feature>/` holds `api.ts`/`server.ts` (server-only loaders), `actions.ts` (Server Actions), `components/` and the mappers. Features: `catalogue`, `product`, `cart`, `auth`, `checkout`, `orders`, `account`, `saved`, `help`, `cars`, `shell`.
- Reusables live in `frontend/packages/`: `ui` (components and CSS primitives), `bff` (session, sign-in flow, client IP), `testing` (e2e kit), `fixtures` (sample data, still used for cars).

**Decisions:**
- **Every data page renders per request** (`await connection()`). The layout already reads the session cookie, and builds must not need a running API (CI has none). The category tree is cached for 5 minutes; listings and cart are `no-store`.
- **Graceful degradation:** if the API is down, the header still renders with no category row, listings show a "we can't load products" state with a retry link, and the home page drops its empty shelves.
- **Cards from search results:** the API's `ProductSummary` has no category or image, so cards take their placeholder art from the listing's category (a neutral box elsewhere). They show "Sold by TechShop" only when TechShop is among the sellers and "From N sellers" when there are several. Nothing is guessed.
- **Guest cart:** the API's `X-Cart-Token` is kept in an `httpOnly` cookie (`ts_cart`, 30 days) and merged into the customer's cart on sign-in.
- **Sign-in is phone-first** (OTP; the same code signs up a new number), with email and password as the alternative. The account page has profile, addresses (state from the API list) and signed-in devices with remote sign-out.
- **Checkout:** the address, coupon and payment method are re-priced by `/checkout/preview`. The order is placed with the quote's `priceHash` and one idempotency key per attempt (a new key after a price change). On `409 price_changed` the new total is shown and nothing is charged. Card and wallet payments redirect to the provider's page (only HTTPS, or the local API's `/dev/pay/` page in development); bank transfers show the one-time account in place. `/checkout/return?order=` sends the customer to the order page.
- **Orders:** signed-in customers see their orders and order pages (pay again, cancel while unpaid, parcels, history). Guests track with order number and checkout phone, posted by a Server Action so the phone never appears in a URL.
- **Honesty fixes:** removed the "deals end at midnight" countdown (no real deadline backs it) and the hard-coded "Deliver to Lagos". The delivery page is now a live fee checker. Cars are the only sample data left on the market, and every car view says so.
- **Client IP:** found while testing. The API rate-limits OTP and tracking per IP, but every BFF call came from the Next server's IP, so in production all shoppers would share one limit. The BFF now forwards the proxy-set `X-Real-IP` as `X-Forwarded-For`, only when `CLIENT_IP_HEADER` is configured (secure default).
  - **Needed at deployment (not a code change):** set `CLIENT_IP_HEADER=x-real-ip` on each web app, and list the web servers' addresses in the API's `TRUSTED_PROXIES`.

## 2026-10-20 — Staff portal on the API (part 1: foundation, orders, finance, catalogue)

- **Route group `(portal)`:** every page except `/sign-in` sits in it. The group layout and each page call `requireStaff(permission)` themselves. Next renders layouts and pages in parallel (seen in testing: a page ran before the layout's redirect), so a layout check alone is not a guard. The API enforces every permission regardless.
- **Sign-in:** work email + password, then the authenticator code or a recovery code (shared `createAuthFlow`). The MFA token stays in the sealed flow cookie.
- **Sidebar shows only modules the person may open** (permissions come from `/me`). Modules whose backend isn't built (M4–M7) are marked "Preview" and say so in a sample-data notice naming the milestone.
- **Step-up:** any action the API answers with `step_up_required` shows an authenticator-code prompt in place, calls `/auth/mfa/step-up` (the BFF swaps in the fresh tokens) and retries the action (`components/ActionButton`).
- **Staff building blocks** (`apps/staff/src/components`): `Table` (typed rows), `Tabs` (link tabs), `Figures`, `Status` (dot + words), `PageHead`, `ActionButton`; `lib/actions.ts` turns API refusals into plain messages.
- **Catalogue:** no staff product-list endpoint exists, so the product tab uses the catalogue's search index. The create form makes the product, its first variant and (optionally) TechShop's own listing. If the listing call fails, the product still exists as a draft and the message says so.
- **Backend issues found (not changed):**
  - `GET /dispatch/board` returns each job's `otpHash` (the hashed delivery code) to dispatch staff. The UI never shows it, but it should be removed from the response.
  - Staff catalogue has create/update endpoints but no list.

## 2026-10-20 — Staff portal part 2 (operations, growth, organisation)

- **Live modules added:**
  - dispatch (board with live counts, assign/unassign/return, riders suspend/reinstate, phone approvals, zones and rate bands);
  - warehouse (release a wave, pick lists with claim and scan, IMEI/serial for serialised devices, pack by weight, manifests and signed hand-over);
  - inventory (stock by warehouse, adjustments with reason, receiving serialised units, movements);
  - marketing (promotions end/cancel, coupons, campaigns submit → approve → pause/resume/cancel, audiences, journeys);
  - content (pages, help articles, banners);
  - search tuning (insights, zero-result searches, catalogue gaps, synonyms);
  - admin (staff roles grant/revoke, roles, people search, audit log, sign-in events, authenticator reset);
  - HR (staff list, invite, exit).
- **Content is edited as plain text** and stored as JSON blocks: blank lines separate paragraphs, `## ` starts a heading, `- ` starts a list item. No HTML can be stored or injected.
- **Separation of duties shows in the UI as well as the API:**
  - a page's last editor sees "Someone else must publish your edit" instead of a Publish button;
  - nobody edits their own roles or exits themselves;
  - campaign approval stays with `marketing.approve` holders, who need step-up.
- **Error messages prefer the API's own explanation** ("Someone other than the last editor must publish.") over a generic "no permission".
- **Receiving devices:** 15-digit codes are IMEIs (the API checks the Luhn digit); anything else is a serial number.

## 2026-10-20 — Mobile: shared UI package and the customer app's first slice

- **`mobile/packages/ui` (`@techshop/mobile-ui`):** the logistics building blocks (Screen, Title, Body, Button, Field, Notice, Card, `useColors`) moved here, so both apps share one set (`docs/mobile.md` §2). Controls stay 52 px in both apps: the customer minimum is 44 px, and larger targets are fine. The navigation theme stays per app, because it needs expo-router.
- **Root mobile scripts** run lint, typecheck and test with `--if-present`, so a package without a script doesn't fail the workspace run.
- **Customer app, first slice (browse only, no new packages):** home (search, categories, deals), search (did-you-mean), category and product (every offer, warranty, delivery fee, specs).
  - Buying says plainly "coming soon in the app; order on the website". There's no dead or fake button.
  - Lists use `FlatList` for now; FlashList (in the plan) can replace it when the dependency is added.
- **Next for the customer app:** phone sign-in (reuse the rider `session.ts` once it moves to a shared mobile package), cart, checkout with an `expo-web-browser` payment hand-off, orders and tracking.

## 2026-10-20 — Customer app checkout

- **Payment hand-off as planned (`docs/mobile.md` §1):** the API returns the provider's checkout URL; the app opens it with `expo-web-browser` (added). Card details never touch the app. On return, the app checks `/me/payments/{id}` for about 15 s, then opens the order page. The webhook is what marks the order paid. Bank transfers show the one-time account in place.
- **Same safety as the web:** the order is placed with the quote's price hash and one idempotency key per attempt (a new key after a price change). On `409 price_changed`, the new total is shown and nothing is charged.
- **Addresses:** checkout picks a saved address or adds one (the state is typed and matched against the API's list, with suggestions). Full address management on the account screen comes later.
- **Device testing note:** in development, the fake provider's page is served by the API at `APP_BASE_URL`. For an Android emulator, run the API with `APP_BASE_URL=http://10.0.2.2:8080` so the in-app browser can reach it.

## 2026-10-20 — Web dev ports moved to 4101–4105

**Owner decision:** corporate **4101**, market **4102**, wholesale **4103**, seller **4104**, staff **4105**. API stays on 8080.

**Why:** other projects on the owner's machine use the 3000s (AdGuard on 3000, another app on 3003, which clashed with the seller app). A dedicated block avoids clashes. 6000 was considered and rejected: browsers block it (and 6665–6669) as unsafe ports.

**Changed:**
- each app's `dev`/`start` scripts;
- the e2e test defaults;
- the nginx routes;
- the READMEs;
- the backend defaults for `CORS_ALLOWED_ORIGINS` and `PAYMENT_CALLBACK_URL` (now `http://localhost:4102/checkout/return`) in `internal/config` and `.env.example`. These backend config changes were approved by the owner.

**Supersedes** the ports in "Multi-app monorepo restructure" and "Corporate dev port moved to 3005".

**Owner action:** a local `backend/.env` that sets `CORS_ALLOWED_ORIGINS` or `PAYMENT_CALLBACK_URL` with the old ports needs the new ones. Claude doesn't read or edit `.env` files.

## 2026-10-20 — Claude Code project settings made valid again

The owner had commented out the `ask` and `deny` lists in `.claude/settings.json` with `//`. JSON doesn't allow comments, so Claude Code ignored the whole file, including `allow`. Owner decision:
- keep `allow`;
- drop `ask`;
- keep a short `deny` for irreversible actions and the `.claude/CLAUDE.md` "never" rules: `.env` read/edit/write, force-push, `reset --hard`, `clean -f`, `branch -D`, `rm -rf`, migration rollbacks, Docker volume deletion, `psql … DROP`, `ssh`/`scp`/`sftp`/`rsync` to remotes.

`git add` and `git commit` are no longer denied, so commits can be made when asked. Deny rules are enforced by Claude Code itself; Claude can't override them.

## 2026-10-21 — Backend restructure (supersedes the layout part of the 2026-10-06 correction)

Owner approved ("go") the proposed layout and these choices:
- **Move, don't rewrite.** Packages move with `git mv` (history kept); import paths are fixed by script; `go build`, unit and e2e tests prove behaviour is unchanged.
- `internal/modules/<feature>/`: identity, catalog, content, files, inventory, wms, sales, payments, ledger, logistics, marketing, notify, tracking.
- `internal/platform/`: app, config, auth, rbac, httpx, middleware, kit, uow, outbox, jobs, idempotency, ratelimit, realtime, storage, database.
- `internal/store/` (sqlc output) and `internal/seed/` stay where they are. `pkg/` is unchanged.
- **Large files split by cutting** into `routes.go`, `handler.go`, `dto.go`, `service.go` per module, no logic changes.
- **Repository layer only where it earns its place**: new M4 code and complex modules; no pass-through wrappers over sqlc everywhere.
- **Redis / MongoDB:** `pkg/database/` keeps room for `redis/` and `mongo/` beside `postgres/`; nothing added until a feature needs it.
- Then M4 (marketplace) starts on top of this layout.

## 2026-10-21 — M4 Marketplace: slices and slice 1 decisions

Owner: "add RLS in M4"; work until 07:00, slice by slice (code, test, fix).

**Slices** (each tested and committed on `feat/m4-marketplace`):
0. Least-privilege DB users (done).
1. Seller onboarding: open a shop, KYC submission, payout bank accounts, staff KYC review + RLS on `sellers.*`.
2. Listings: sellers list offers on catalogue variants, price and stock; moderation; RLS on `catalog.listings`.
3. Seller fulfilment: accept/ship orders, SLA events; RLS on `sales.fulfilments`.
4. Reviews and ratings.
5. Payouts (batches, dual approval, holds, ledger).
6. Monnify and OPay adapters.
7. Seller Centre and staff screens wired.

**Slice 1 decisions** (from trust-safety.md §2, payments-finance.md §9, identity-access.md §7.3; nothing new invented):
- One open shop per owner. Shops start `pending_kyc`; approval activates them.
- Seller routes are scoped by path (`/seller/shops/{shopId}/…`). A non-member gets **404**, so other shops' existence isn't revealed.
- KYC launch adapter is **manual review** (provider `manual`), as the plan says; a verification provider plugs in later.
- Required documents: ID + selfie; businesses also CAC certificate and number. NIN format checked (11 digits); other IDs 6–20 letters/digits.
- ID and account numbers: encrypted (sealer, AAD bound to the shop), blind index (`BLIND_INDEX_KEY`) for duplicate detection, last 4 shown. Full ID only via `vendors.reveal_id` (sensitive → MFA) with a reason; audited.
- Bank accounts: name enquiry via Paystack `/bank/resolve` before saving (no network call inside a DB transaction). The name match (`match/partial/mismatch`) is **shown to reviewers, not auto-blocking**. New accounts carry a 24-hour payout hold (`payoutsFrom`).
- Approval needs a verified bank account; reviewer may not be a member of the shop. The owner gets the existing `seller_kyc_result` email.
- **RLS policy semantics:** rows are restricted only when `app.seller_id` is set (seller requests). Staff, customer and job requests are unaffected. This is defence in depth for seller endpoints, not a replacement for the API's checks.
- Dev fake name enquiry: echoes the typed name; numbers starting 9 return a different holder; `0000000000` is not found.

## 2026-10-21 — M4 slices 2–3 decisions (listings, seller orders)

**Slice 2: seller listings**
- The catalogue module keeps owning listing rules; Seller Centre (`sellers/listings.go`) scopes them to the shop.
- Marketplace sellers ship their own stock for now (`fulfilled_by = seller`, stock required). TechShop-fulfilled vendor stock needs inbound receiving, which isn't built.
- A live listing whose price is cut more than 40 % below the other sellers' median goes **back to review** (same scam signal as at creation; stops "approve then bait price").
- RLS on `catalog.listings`: sellers write only their own; they can read their own plus all **active** listings (public anyway, and the price-outlier check needs them).
- Bug fixed in existing code: automatic flags recorded a review with no reviewer, breaking a DB check. New decision value `auto_flagged` (migration 00026).

**Slice 3: seller-shipped orders**
- Accept within 24 h (existing `accept_by`); on accept, `pack_by` = now + the longest `handling_days` of the order's listings (doc: "pack within handling_days"; before it was a flat 48 h).
- Seller cancel (pending or accepted) and SLA breaches cancel the whole fulfilment, release stock and raise **one refund request** (existing listener). Existing rule kept: system refund requests still need finance approval (dual control).
- 3 late breaches (`late_accept`, `late_pack`, `late_handover`) in 30 days send the seller's live listings back to review (trust-safety.md §7, orders-fulfilment.md §4). Seller cancellations are recorded but not counted as late.
- Sellers see the recipient's **first name and city only**.
- RLS on `sales.fulfilments` restricts **writes** only: the order status is derived from all of an order's fulfilments, so reads stay open; seller queries filter by shop.
- New-order message to the shop owner on payment (existing `seller_new_order` template, push + email).

**Open question for the owner (blocks hand-over, delivery and seller payouts for seller-shipped orders):**
1. How do packed seller orders reach the customer at launch: (a) a TechShop rider picks up from the seller, (b) the seller drops off at a TechShop hub, or (c) the seller's own courier with a tracking number? (a) needs a seller pickup address, which no table holds today.
2. For (c), who confirms delivery: carrier tracking, the customer's delivery code, or staff? Delivery is what makes the seller's money payable, so a seller must not be able to confirm it alone.
3. When should a seller see the full delivery address and phone? Today: never (first name and city only).

## 2026-10-21 — M4 slice 4 decisions (reviews and seller ratings)

From trust-safety.md §6 unless noted:
- New module `internal/modules/reviews`. One review per user per product (existing unique index). Verified = tied to the reviewer's own **delivered**, not-cancelled order line for that product.
- **Unverified reviews count half** in the rating (doc: "weighted lower"; 0.5 chosen) — the product page summary and the search read model use the same weighting.
- **Held for moderation** (status `pending`) when: 5+ reviews from the account in 24 h; the same text (normalised, 20+ characters) posted by another account; incentivised-review phrases; phone numbers, emails or links. Reasons are stored for moderators and never shown to the reviewer. Other reviews publish immediately.
- Moderation (`catalog.moderate`): publish / hide / reject; hiding or rejecting upholds open reports, publishing dismisses them. Nobody moderates their own review.
- Reports: any signed-in customer or seller, once per review while open, not your own.
- Sellers answer a review once, only for items **they shipped** (owner or manager); they can't remove reviews. New columns `seller_response`, `seller_response_at`, `moderation_reasons` (migration 00028).
- Seller ratings: once per **delivered** fulfilment of your own order; the shop's `rating_avg`/`rating_count` are recomputed on each rating.

## 2026-10-21 — M5 (after-sales) started: slice 1 returns — decisions

Owner asked for the next milestone after M4. M5 slices: 1 returns (RMA) → 2 warranty claims and repairs → 3 support tickets and contact → 4 trade-ins → 5 risk rules and cases; then web/app screens.

**Slice 1 decisions** (orders-fulfilment.md §7 unless noted):
- New module `internal/modules/aftersales`.
- Returns for **faulty, damaged, wrong item, not as described within 7 days of delivery**. **Change-of-mind returns are refused for now**: the plan says the owner decides that policy (eligible categories, restocking fee).
- **Auto-approval** for those reasons when the customer is in good standing, read as: no open risk case, and returns ≤ 30 % of delivered orders in 90 days counting this one (trust-safety.md §3's return rule). A first return is always in good standing. Otherwise support (`returns.manage`) approves or rejects with a reason the customer sees.
- Only items **TechShop shipped**; seller-shipped items wait for the hand-over decision (customers are told to contact support).
- Methods: `store_dropoff` or `pickup` (status `awaiting_pickup`). **Not built:** creating the rider return-pickup job; staff mark the item received when it arrives.
- Inspection per item: grade A/B/C/faulty and disposition. Serialised devices must be scanned: **a mismatched IMEI or serial rejects the whole return and opens a risk case (`kind = return`) — both committed**, the inspector gets 409 `serial_mismatch`. `restock_new` / `restock_open_box` put the unit back in stock (movement reason `return`; device unit back `in_stock` with the new condition).
- **Refund** = the price paid for the returned units (line total share, so discounts are respected); faulty returns have no restocking fee; the delivery fee is not refunded (not stated in the plan; flag for the owner). It's a refund request linked to the return, approved by finance as usual.

## 2026-10-21 — Open M4/M5 questions decided by research (supersedes the "open question" entries above)

Owner: "research and pick the best options" (standing instruction; don't wait on these). Decisions:

1. **Seller hand-over: seller drop-off at a TechShop hub** (Ikeja warehouse, Wuse hub to start). Hub staff scan the packed parcel in (`handed_over`), then it follows TechShop's own path: rider in delivery zones, carrier elsewhere. Delivery is confirmed only by TechShop (customer code or carrier tracking), never by the seller, because delivery releases the seller's money. Rider pickup from seller addresses is a later addition (needs a seller pickup address table). Source: Jumia's vendor drop-off model (Moveee 2025 guide; Jumia/Businesswire on Jumia Delivery).
2. **Change-of-mind returns: allowed within 7 days**, unopened with seals intact; excluded categories: hygiene items (underwear, cosmetics, in-ear audio); no restocking fee; customer pays return delivery. Source: Jumia vendor help (change of mind allowed with hygiene exclusions; sealed mobiles). FCCPA s.122 only mandates refunds for defective/wrong goods, so this is goodwill policy.
3. **Delivery fee refunded for our-fault returns** (faulty, damaged, wrong, not as described), not for change of mind. Source: FCCPA s.122 "full refund" (Templars 2025; Dubawa); no-refund terms held void (s.129). Delivery isn't addressed explicitly, so read in the customer's favour.
4. **Seller team invites expire after 7 days** (re-sendable); finance members must enrol an authenticator before seeing payouts or changing bank accounts. Source: Shopify staff invites (7 days); identity-access.md §7.3.
5. **Monnify and OPay:** adapters built from public docs and tested against local stand-ins; live keys are the owner's to provide.

Build order: 2–3 (returns) → 1 (hand-over, unblocks delivery, payouts, seller returns, seller review replies) → payouts → 4 → 5.

## 2026-10-21 — Seller hand-over deadline (decision 1, detail)

A packed seller parcel must be dropped at a TechShop warehouse or hub (not a retail store) within **24 hours of packing**. A deadline that falls on a Sunday (WAT) moves to Monday. If it's missed, the SLA job cancels the order as `late_handover`, refunds the customer, and the breach counts toward the three-strike rule. Source: Jumia Nigeria vendor FAQ, which gives 24 hours to drop off after ready-to-ship (vendor.jumia.com/black-friday-ops-faq; dated 2018). Senegal and Egypt allow 48 hours from the order (vendorhub.jumia.sn/?p=146, sell.jumia.com). 24 hours after packing sits inside both. Sunday is skipped because Jumia's own SLA excludes Sundays.

Later: rider pickup from the seller's address (needs seller pickup addresses), and a seller notification when the hub scans a parcel in.
