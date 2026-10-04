---
name: frontend-design-system
description: Build consistent, accessible, responsive UI — design tokens (colors, spacing, typography, radii, shadows), component library structure (custom CSS on design tokens — no Tailwind or CSS frameworks), forms, tables, modals, loading/empty/error states, dark mode, responsive layouts, and WCAG 2.2 AA accessibility. Use this skill whenever the user asks to build or restyle pages, components, dashboards, admin panels, landing pages, or forms in React/Next.js (or similar), or asks to make UI consistent, accessible, responsive, or "look better".
---

# Frontend Design System

## First, find the system

**TechShop:** read the `frontend-design` skill and `.claude/rules/frontend-design.md` first.
- **Tokens** come from `shared/design-tokens` (generated `css/tokens.css`; run `make tokens`, never hand-edit).
- **Components** live in `frontend/packages/ui`. Global CSS uses cascade layers: reset < tokens < base < layout < components < utilities. Feature components use **CSS Modules + tokens**.
- **Mobile** uses the same token values through `shared/design-tokens`.
- **No Tailwind, shadcn/ui or other CSS frameworks**, and no new UI libraries without asking.

Any project: look for existing tokens (CSS variables, theme files), a components package and Storybook before writing UI. **Reuse existing components and tokens**; never hardcode one-off colours or spacing when a token exists.

## Tokens (only when a project has none)

Define them as CSS custom properties on `:root`, with dark values under `prefers-color-scheme` and a `[data-theme]` override:
- **Colour roles**, not raw colours: background, surface, text, muted, border, accent, on-accent, link, success, warning, danger, focus. Light and dark values for each.
- **Spacing**: 4px base scale (4, 8, 12, 16, 24, 32, 48, 64).
- **Typography**: 1–2 font families; a scale such as 12/14/16/18/20/24/30/36; body 16px; line-height 1.5 for body, 1.2 for headings.
- **Radius**: sm, md, lg, pill.
- **Shadows**: 2–3 elevation levels at most.
- **Breakpoints**: content-driven, with container queries where they fit; test at 360, 768 and 1280px.

## Component structure

```
packages/ui/src/       # shared primitives and patterns: Button, Field, Dialog, Badge, Card, Table, Skeleton, Tabs…
  styles/              # layered global CSS (reset, base, layout, components, utilities)
apps/<app>/app/…       # routes; feature components beside them with *.module.css
```
- Primitives take a `className` and a small set of variant props, mapped to CSS classes or data attributes (`data-variant="primary"`), not utility-class strings.
- Accessible behaviour (dialogs, menus, tabs) uses native elements first (`<dialog>`, `<details>`, `<button>`). If a headless primitive library is really needed, **ask before adding it**.

## Every data view needs 4 states

1. **Loading** — skeletons matching final layout (not just a spinner)
2. **Empty** — explanation + primary action ("No invoices yet — Create invoice")
3. **Error** — human message + retry; no raw error dumps
4. **Success** — the data

## Forms

- Label every field (visible label, not just placeholder); mark required fields.
- Inline validation on blur/submit; error text below field linked with `aria-describedby`.
- Disable submit + show progress while pending; prevent double submission.
- Preserve user input on error.
- Destructive actions need confirmation dialogs naming the object ("Delete invoice INV-0042?").

## Tables & lists

- Pagination or infinite scroll; sortable headers with `aria-sort`; sticky header for long tables.
- Horizontal scroll container on mobile, or switch to card layout under `md`.
- Row actions in a menu; bulk actions when selection exists.

## Responsive

- Mobile-first CSS; test at 360px, 768px, 1280px; no horizontal page scroll.
- Touch targets ≥ 44×44px; no hover-only interactions.
- Sidebars collapse to drawer on mobile.

## Accessibility (WCAG 2.2 AA)

- Semantic HTML (`button` for actions, `a` for navigation, headings in order, landmarks).
- Contrast ≥ 4.5:1 text, 3:1 large text and UI components.
- Full keyboard support; visible focus ring (`focus-visible`); focus trapped in modals and returned on close.
- `alt` text for meaningful images, `alt=""` for decorative.
- Don't convey meaning by color alone (add icon/text).
- Respect `prefers-reduced-motion`.
- Announce async results (toasts with `role="status"`/`aria-live`).

## Visual quality checklist

- [ ] Consistent spacing rhythm and alignment
- [ ] Clear visual hierarchy (one primary action per view)
- [ ] Tokens only, no stray hex values
- [ ] Dark mode works if supported
- [ ] Loading/empty/error states present
- [ ] Keyboard + screen reader friendly
- [ ] Looks right at mobile, tablet, desktop
