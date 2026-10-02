# TechShop — Web apps

Five Next.js apps sharing one design system, managed with npm workspaces and Turborepo.

| App | Port | Purpose |
| --- | --- | --- |
| `apps/corporate` | 3000 | Company main site |
| `apps/market` | 3001 | Multi-vendor customer marketplace |
| `apps/wholesale` | 3002 | TechShop's own stock: retail and bulk/trade buying |
| `apps/seller` | 3003 | Seller centre for marketplace vendors (not indexed) |
| `apps/staff` | 3004 | Staff portal with role-based sections (not indexed) |

| Package | Purpose |
| --- | --- |
| `packages/ui` | Global CSS design system: `import "@techshop/ui/styles.css"` |
| `../shared/design-tokens` | Brand tokens; generates the CSS variables `ui` imports |
| `../shared/api-client` | Typed Go API client; each app wraps it in `src/lib/api.ts` |

```bash
npm install
npm run dev              # all apps
npm run dev:market       # one app
npm run lint | typecheck | build | format
```

To add an app, copy an existing one in `apps/`, then change its `name`, port, and metadata.

Apps never access the database. All data goes through the Go API.
