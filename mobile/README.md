# TechShop — Mobile apps

Expo (React Native) apps with expo-router, managed with npm workspaces.

| App | Purpose |
| --- | --- |
| `apps/customer` | Customer shopping app |
| `apps/logistics` | Delivery and dispatch teams (one app, role-based) |

Both import `@techshop/design-tokens` (via `src/theme`) and `@techshop/api-client` (via `src/lib/api.ts`) from `../shared`.

```bash
npm install
npm run customer        # start the customer app (Expo dev server)
npm run logistics
npm run lint | typecheck | doctor
```

Notes:
- React is pinned with `overrides` in `package.json` to the version Expo expects. Two React copies break native builds. Keep it in sync when upgrading the Expo SDK, then run `npm run doctor`.
- `EXPO_PUBLIC_API_URL` must be reachable from the device. See each app's `.env.example`.
- App icons and splash images are Expo placeholders. Replace them with TechShop artwork before release.
- Bundle IDs (`com.techshop.customer`, `com.techshop.logistics`) are placeholders. Confirm them before the first store submission; they can't be changed afterwards.
