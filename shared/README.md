# shared/

Code used by **both** the web apps (`frontend/`) and the mobile apps (`mobile/`). Nothing here may depend on a platform (no Next.js, no React Native, no DOM-only APIs).

| Package | Purpose |
| --- | --- |
| `design-tokens/` | Brand values (colour, type, spacing, radius, motion). Generates the web `tokens.css`. |
| `api-client/` | Typed client for the Go API. |

Apps consume these through `file:` dependencies, e.g. `"@techshop/api-client": "file:../../../shared/api-client"`.
