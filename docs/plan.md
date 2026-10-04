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
