// Named access to the generated OpenAPI contract (regenerate with `npm run api:types` in frontend/).
import type { components, operations, paths } from "./generated/schema";

export type { components, operations, paths };

/** A schema from the contract, e.g. `Schema<"ProductDetail">`. */
export type Schema<K extends keyof components["schemas"]> = components["schemas"][K];
