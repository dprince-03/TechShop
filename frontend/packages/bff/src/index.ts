// Server-only entry. `@techshop/bff/proxy` is the separate entry for each app's proxy.ts.
export { loadConfig } from "./config";
export type { Audience, BffConfig } from "./config";
export { safeReturnPath, signInPath } from "./paths";
export { assertSameOrigin, createBff } from "./server";
export type { Bff } from "./server";
export type { Session } from "./session";
