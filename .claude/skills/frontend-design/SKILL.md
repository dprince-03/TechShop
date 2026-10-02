---
name: frontend-design
description: TechShop visual design system and the reference-store research behind it (Apple Store, Samsung, MSI Store, Amazon, Jumia, Gucci). Use when building or styling any frontend UI — pages, layouts, product cards, buttons, navigation, forms — or when choosing colors, type, spacing, radius, or breakpoints in frontend/.
---

# TechShop frontend design system

The design system has two parts:
- **Tokens:** `shared/design-tokens/src/index.ts` is the single source of truth for web **and** mobile.
- **Web CSS:** `frontend/packages/ui` holds the global CSS system. Every web app loads it once in its root layout with `import "@techshop/ui/styles.css"`.

Build every UI on it. No Tailwind or CSS frameworks on the web.

## Hard rules

- **Use tokens, never raw values.** Colors, sizes, spacing, radius, shadows and durations come from the tokens: CSS variables on the web, `tokens` from `@/theme` on mobile. If a value is missing, add it to `shared/design-tokens/src/index.ts` and run `make tokens`. Never edit `shared/design-tokens/css/tokens.css` by hand; it's generated, and CI fails if it's stale.
- **Global primitives first, CSS Modules second.** Compose `.container`, `.section`, `.stack`, `.cluster`, `.grid`, `.btn`, `.card`, and the rest. Feature-specific styles go in `Component.module.css` next to the component, and they must also use tokens.
- **Don't add global classes casually.** Add one to `layout.css`, `components.css` or `utilities.css` only when it's a generic primitive reused across features. Product cards, headers and similar pieces are feature components and use CSS Modules.
- **Cascade layers:** `reset < tokens < base < layout < components < utilities`. CSS Modules are unlayered, so they win automatically. Never use `!important` to fight the cascade (the only exceptions are the visibility utilities and the reduced-motion reset).
- **Mobile-first.** Use `min-width` media queries at the fixed breakpoints below. Don't invent new ones.
- **Accessibility floor:**
  - Text contrast ≥ 4.5:1, and UI boundaries such as form controls ≥ 3:1. The tokens were chosen to pass; re-check any new color.
  - Touch targets ≥ 44px (`--control-md`).
  - Keep the `:focus-visible` ring.
  - Icon-only buttons need an `aria-label`.
  - Respect `prefers-reduced-motion`; the reset handles this globally.
- **One primary action per view.** `.btn--primary` is reserved for the main conversion action (add to cart, checkout). Everything else is secondary, inverse or ghost.
- **Brand independence.** The references below inform principles only. Never copy their text, images, icons, logos, fonts or exact brand colors.

## File map

| File | Contains |
|---|---|
| `shared/design-tokens/src/index.ts` | All design tokens (px numbers, light/dark colors). The source of truth. |
| `shared/design-tokens/css/tokens.css` | **Generated** CSS variables and light/dark themes; `[data-theme]` overrides the OS preference |
| `frontend/packages/ui/src/index.css` | Entry point: declares the layer order, imports the tokens and the files below. Add nothing else here. |
| `frontend/packages/ui/src/reset.css` | Modern reset and the reduced-motion guard |
| `frontend/packages/ui/src/base.css` | Element defaults: body type, h1–h6 scale, links, focus, selection |
| `frontend/packages/ui/src/layout.css` | `.container`, `.full-bleed`, `.section`, `.stack`, `.cluster`, `.grid`, `.grid--catalog`, `.with-sidebar`, `.scroller`, `.frame`, `.center` |
| `frontend/packages/ui/src/components.css` | `.btn` (+ variants and sizes), `.link`, `.card`, `.badge`, `.divider` |
| `frontend/packages/ui/src/utilities.css` | Text helpers, `.eyebrow`, `.line-clamp`, `.tabular-nums`, `.strike`, `.visually-hidden`, `.skip-link`, `.hide-mobile`/`.hide-desktop` |
| `mobile/apps/*/src/theme/index.ts` | Mobile access to the tokens: `useColors()` and a navigation theme built from the tokens |

## Mobile (React Native)

Mobile can't use the CSS, but it follows the same rules through the tokens:
- **Colors:** `useColors()` for the current light/dark scheme.
- **Spacing and radius:** `tokens.space[n]` and `tokens.radius.*` (px numbers).
- **Font sizes:** fixed sizes are numbers. Fluid sizes (`3xl`–`6xl`) are `{ min, max }`; use `min` on phones.
- **Letter spacing:** `tokens.tracking.*` is in em, so multiply by the font size (e.g. `tokens.tracking.tight * size`).
- **Touch targets:** at least `tokens.control.md` (44).
- **Prices:** the same rules as the web, including tabular numbers (`fontVariant: ["tabular-nums"]`).

## Token reference

**Color.** The palette is neutral, with one blue accent and semantic colors for commerce.
- Surfaces and text:
  - `--color-bg` (white / near-black) and `--color-surface` (light gray bands and media backdrops).
  - `--color-surface-raised` (cards) and `--color-surface-inverse`.
  - `--color-text`, `--color-text-muted` (secondary), `--color-text-subtle` (captions).
- Borders: `--color-border` is a decorative hairline. `--color-border-strong` is the 3:1 border for form controls.
- Accent: `--color-accent` and `--color-accent-hover` are action fills with `--color-on-accent` text on top. `--color-accent-subtle` is a tint, and `--color-link` is for link text.
- Commerce:
  - `--color-sale` (discounts, price drops) and `--color-promo` (deals, countdowns) are text colors.
  - `--color-sale-fill` and `--color-promo-fill` are badge backgrounds that take white text in both themes.
  - `--color-rating` is for stars only, never text. Also `--color-success`, `--color-warning`, `--color-danger`.
- **Dark mode:** never use `--color-sale` or `--color-promo` as a background behind white text. In dark mode they are light tints for text; use the `-fill` tokens instead.

**Type.**
- Font is Geist (`--font-sans`) with a system fallback.
- Fixed UI sizes: `--text-xs` 12, `sm` 14, `base` 16, `lg` 18, `xl` 21, `2xl` 24.
- Fluid headline sizes: `--text-3xl` 28→32, `4xl` 32→40, `5xl` 40→56, `6xl` 48→72.
- Elements map to sizes: h1 = 5xl, h2 = 4xl, h3 = 3xl, h4 = 2xl. `.text-display` uses 6xl for hero headlines.
- Weights: 400 for body, 600 for headings and buttons, 700 for display only.
- Line height: tight 1.1 (display), snug 1.25 (headings), normal 1.5 (body).
- Tracking: tight −0.022em on large headlines, snug −0.011em on section headings, wide 0.08em for uppercase eyebrows only.
- `--measure: 65ch` is the maximum line length for paragraphs.

**Spacing.** A 4px base scale: `--space-1` to `--space-10` = 4, 8, 12, 16, 24, 32, 48, 64, 96, 128. `--gutter` is the fluid page side padding (16→40). `--section-space` is the fluid vertical rhythm between page bands (48→96).

**Containers.** `--container-narrow` 720 (forms, checkout, articles) · `--container-md` 1024 (product detail copy) · `--container` 1280 (default) · `--container-wide` 1440 (catalogue, hero bands). Use `.container`, `.container--narrow`, `.container--md` or `.container--wide`.

**Radius.** `xs` 4 (badges) · `sm` 8 (inputs) · `md` 12 (menus and popovers) · `lg` 18 (cards, media frames) · `xl` 28 (hero panels) · `pill` (all buttons and chips).

**Elevation.** Shadows are minimal: `--shadow-sm`, `md`, `lg`. Use them for hover lift, dropdowns and modals only. At rest, cards rely on a hairline border or a surface contrast, not a shadow.

**Motion.** `--duration-fast` 150ms (hover and color), `base` 250ms (lift, expand), `slow` 400ms (larger transitions). `--ease-standard` is the default easing and `--ease-emphasized` is for entering elements.

**Breakpoints.** These are literals, because CSS variables can't be used in media queries: `40rem` (640) sm · `48rem` (768) md · `64rem` (1024) lg · `80rem` (1280) xl.

**Layering.** `--z-dropdown` 100 · `--z-sticky` 200 · `--z-overlay` 300 · `--z-modal` 400 · `--z-toast` 500.

## Primitive usage

- **Page band:** `<section class="section [section--surface|section--inverse]"><div class="container">…`. Alternate white and surface bands to separate sections instead of using borders or rules.
- **Product listing:** `.grid--catalog` gives 2 → 3 → 4 → 5 columns at md, lg and xl, so rows line up across sections. Use `.grid` with `--grid-min` for intrinsic grids such as category tiles.
- **Filters + results:** `.with-sidebar`. The sidebar is set with `--sidebar-width` (default 16rem), and the layout stacks automatically when space runs out.
- **Carousel/shelf:** `.scroller` is a horizontal scroll-snap row with no JS. Items are 70% wide on mobile, 40% at md and 23% at lg, so the next item always peeks in to show the row scrolls. Override the item width with `--scroller-item`.
- **Product imagery:** wrap images in `.frame`. The default is a 1:1 ratio on `--color-surface` with `object-fit: contain`, so the whole product shows and nothing is cropped. Use `.frame--cover` for lifestyle or hero photos, `.frame--wide` for 16:9 and `.frame--portrait` for 4:5.
- **Card:** `.card` is a raised surface with a hairline border and an 18px radius. Use `.card--flat` on surface bands. `.card--interactive` lifts on hover, on hover-capable devices only.
- **Buttons:** `.btn` plus a variant: `--primary` (accent fill), `--secondary` (outline), `--inverse` (dark fill, the premium or neutral CTA), or `--ghost`.
  - Sizes: `--sm` 36px, the 44px default, and `--lg` 52px.
  - Width and shape: `--block` makes it full width, `--block-mobile` makes it full width below md only, and `--icon` makes it square (needs an `aria-label`).
  - Works on both `<button>` and `<a>`.
- **Links:** a plain `<a>` inherits its color (navigation and cards). Use `.link` for visible inline links, `.link--subtle` for footer and secondary nav, and `.link--arrow` for "see all ›" calls to action.
- **Badges:** `.badge` plus `--sale` (discount percentage), `--promo` (deal or limited), `--accent` (new) or `--outline` (spec tag).
- **Prices:** use `.tabular-nums`, with the current price in semibold `--color-text`. Show the old price with `.strike` next to a `.badge--sale` percentage. Never color the main price red; red is only for the discount.
- **Product titles:** clamp them with `.line-clamp` (`--lines`, default 2) so card heights stay consistent.
- **Section header:** an optional `.eyebrow`, then the heading, then a `.link--arrow` aligned right in a `.cluster--between`.

---

## Reference research (2026-10-02)

Sources studied: apple.com/store, samsung.com/us/mobile, us-store.msi.com, amazon.com, jumia.com.ng and gucci.com. Apple, Samsung and MSI were measured from their live CSS. Amazon was partly analysed from its markup and CSS variables. Jumia's page structure was analysed. Gucci blocks automated access (HTTP 403), so its notes come from its well-known public design conventions and were **not measured live**.

### Apple Store
- **Type:** SF Pro Text/Display with a Helvetica fallback. The most frequent sizes are 12, 14, 17, 21, 24, 28, 32, 40, 48 and 64px. Weights are almost all 600 (headings) and 400 (body). Letter spacing is consistently negative on large text (−0.022em, −0.016em, −0.01em) and slightly positive on small text (+0.007 to +0.011em). Line heights run 1.2–1.38.
- **Color:** near-black text `#1d1d1f`, gray surface `#f5f5f7`, muted `#6e6e73`/`#86868b`, hairlines `#d2d2d7`, and a single blue for actions and links (`#0071e3`/`#06c`). Very little else.
- **Shape:** 18–20px card radius, 999px pill buttons. Shadows are soft and rare (`2px 4px 16px` at low alpha).
- **Layout:** content width around 980–1024px inside bands up to 1440px. Breakpoints at 734/735, 833/834, 1023/1024 and 1440/1441. Horizontally scrolling "shelves" of cards. A row of category icons with labels. Large gaps between sections.
- **Navigation:** a slim global bar with a flat category list, search and bag icons, and quick task links (order status, help). The footer has multi-column link groups that collapse into accordions on mobile.
- **Links/buttons:** mostly text links with a "›" chevron, and few filled buttons. Calm and confident.
- **Transitions:** about 0.3s with `cubic-bezier(.25,.1,.3,1)`.

### Samsung (mobile store)
- **Type:** two families, a geometric display face for headings and a neutral sans for UI and body. Sizes are mostly 12, 14, 16, 18, 24 and 28, with 38/48/60 for display. Weight is almost always 700 or 400. Line height is about 1.33.
- **Color:** black/white base, `#f7f7f7` surface, `#ddd` borders, `#555`/`#757575` muted, a blue accent (`#2189ff`/`#006bea`) and red (`#d62e2e`) for sale.
- **Shape:** 20px pill buttons, 8px small radius, circular swatches. The focus ring is a white inner ring plus an outline (`0 0 0 2px #fff`).
- **Layout:** max width 1440 (up to 1920 for heroes). Breakpoints at 767/768, 1279/1280 and 1440. Full-width hero carousels with dual CTAs (filled "buy" plus outline "learn more").
- **Product cards:** image, color swatches, name, rating, price and savings, then CTAs.
- **Navigation:** a mega-menu per category and a tabbed sub-navigation.
- **Transitions:** 0.3s with `cubic-bezier(0.4,0,0.2,1)`.

### MSI Store
- **Base:** built on Bootstrap. The body font is Open Sans with a condensed display face for emphasis. Sizes are rem-based (0.875, 1, 1.25 and 1.5rem), with line height 1.5.
- **Color:** a red brand accent (`#e30613`), plus Bootstrap neutrals (`#212529`, `#6c757d`, `#dee2e6`, `#f8f9fa`) and Bootstrap semantic colors.
- **Shape:** a small 0.25rem radius with some 15–20px rounded elements. The focus ring is `0 0 0 .2rem` at 25% alpha.
- **Layout:** Bootstrap containers of 540, 720, 960 and 1140px, with breakpoints at 576, 768, 992 and 1200.
- **Product cards** list specs (CPU, GPU, display) under the name before the price, which suits spec-driven gadgets. The overall feel is denser and more "gamer".

### Amazon
- **Color:** a dark navy header (`#131921`/`#232f3e`) with a gold/orange accent (`#febd69`, `#ff9900`) used for search, focus and CTAs.
- **Header:** a full-width search bar in the center with a department dropdown on its left, and account and cart on the right. A second row holds department shortcuts.
- **Product cards** follow a strict order: image → title (clamped) → star rating with count → price (large, with superscript cents) → delivery or deal info → button.
- **Density:** high, built for scanning. Grids are 4–6 columns on desktop. Deal badges are red.

### Jumia
- **Header:** logo, a wide search bar, then account, help and cart. Below it, a rotating promo banner and a horizontally scrollable row of category shortcuts with icons.
- **Product cards:** image → clamped title → current price (bold) → struck-through old price → discount percentage badge. Some add a star rating and stock indicator.
- **Sections:** many titled carousels ("recently viewed", "top sellers", "limited stock"), each with a "see all" link. Flash-sale sections have a countdown timer in the header. Grids are 4–5 columns on desktop and scroll horizontally on mobile.
- **Color:** a strong orange accent for deals, on a white background with a light-gray page surface.

### Gucci (not measured live; blocked)
- **Luxury restraint:** a monochrome black/white palette, lots of whitespace and large edge-to-edge editorial imagery.
- **Type:** uppercase navigation and labels with wide letter spacing. A mix of serif and sans in a strong hierarchy, with little text on screen.
- **Product presentation:** large images on plain backgrounds, minimal card chrome (no borders or shadows), and just the name and price.
- **Buttons:** sharp or subtle radius in solid black or outline. The interface steps back so the product is the focus.

## How the references were synthesized

| Decision | Drawn from |
|---|---|
| Neutral palette with a single blue action color; color reserved for meaning (sale, deal, rating) | Apple, Samsung (neutrals and blue); Amazon, Jumia, MSI (semantic commerce colors) |
| Product shots on a light-gray surface, `contain`-fit, square frames | Apple, Samsung |
| Pill buttons, 18px cards | Apple, Samsung |
| Inverse black button variant, uppercase wide-tracked eyebrow, generous whitespace | Gucci |
| Negative tracking on headlines, 600 weight, fluid display sizes | Apple |
| Strict product-card order: image → badge → title (2-line clamp) → rating → price + strike + % | Amazon, Jumia, Samsung |
| Spec lines on cards for gadgets (via CSS Modules later) | MSI |
| 2→3→4→5 catalog columns; horizontal shelves with peeking items | Jumia, Amazon, Apple |
| Alternating white and surface section bands; "see all ›" section links | Apple, Jumia |
| 1280 default and 1440 wide containers; breakpoints at 768, 1024 and 1280 | Samsung, Apple, MSI (Bootstrap) |
| Visible double focus ring, 44px touch targets, reduced motion | Samsung, MSI, WCAG |
| Sticky search-first header with department shortcuts (to build later) | Amazon, Jumia |

The final palette is TechShop's own. Every text pairing was checked for WCAG AA contrast (see `docs/plan.md`, 2026-10-02).
