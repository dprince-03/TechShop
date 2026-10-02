# @techshop/design-tokens

The single source of truth for TechShop's brand values: colours, type, spacing, radius, motion.

- **Mobile** imports the values directly: `import { tokens } from "@techshop/design-tokens"`.
- **Web** uses the generated `css/tokens.css` (CSS custom properties), imported by `frontend/packages/ui`.

Edit `src/index.ts`, then regenerate the CSS:

```bash
npm run build    # writes css/tokens.css
npm run check    # fails if css/tokens.css is out of date (used in CI)
```

Never edit `css/tokens.css` by hand.

Units: sizes are in px (numbers). The CSS generator converts them to rem. Fluid sizes are `{ min, max }` pairs that become `clamp()` on the web (360px → 1280px viewport); mobile should use `min`.
