# React / Next.js Conventions

## Next.js App Router layout
```
app/
  (marketing)/            # route groups
  (dashboard)/
    layout.tsx
    page.tsx
  api/                    # route handlers
components/
  ui/                     # primitives (button, input, dialog)
  <feature>/              # feature components
lib/                      # api client, auth, utils
hooks/
server/                   # server-only code (import 'server-only')
types/
```

## Components
- Function components + hooks only; PascalCase filenames for components.
- Server Components by default; add `'use client'` only when you need state, effects, or browser APIs. Push `'use client'` as far down the tree as possible.
- Props typed with interfaces; avoid `any`.
- Keep components small; extract hooks for logic (`useInvoices`).

## Data fetching
- Server Components fetch on the server; use React Query/SWR for client-side fetching and caching.
- Server Actions: validate input with zod and check auth + authorization **inside every action** — they are public endpoints.
- Never pass secrets or full DB objects to Client Components.

## State
- Local state first; lift only when needed. URL state (search params) for filters/pagination. Context for low-frequency global values (theme, session). Zustand/Redux only for genuinely complex client state.

## Forms
- Use the form approach the project already has. TechShop: server actions and native form validation plus shared validators (e.g. `normaliseNigerianPhone` in `@techshop/api-client`). Ask before adding a form or schema library.
- Show field-level errors; disable submit while pending.

## Styling
- TechShop: no Tailwind or CSS frameworks. Global styles live in the cascade-layered CSS of `frontend/packages/ui`; feature components use CSS Modules that consume design tokens (`var(--…)`), with no hard-coded colours or spacing.

## Env vars
- Only `NEXT_PUBLIC_*` reaches the browser — never put secrets there.

## Accessibility
- Semantic HTML first; label every input; keyboard-operable interactive elements; visible focus states.

## Performance
- `next/image` for images, `next/font` for fonts, dynamic import heavy client components, avoid unnecessary `'use client'`.
