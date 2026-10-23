// Server-only entry. `@techshop/bff/proxy` is the separate entry for each app's proxy.ts.
export { loadConfig } from "./config";
export type { Audience, BffConfig } from "./config";
export { safeReturnPath, signInPath } from "./paths";
export { createAuthFlow, maskPhone } from "./auth";
export type { AuthStep } from "./auth";
export { assertSameOrigin, createBff } from "./server";
export type { Bff } from "./server";
export type { Session } from "./session";
export { failure } from "./forms";
export { submitContact, type ContactSite } from "./contact";
export type { ActionState } from "./forms";
