import { createApiClient } from "@techshop/api-client";

// EXPO_PUBLIC_* values are inlined at build time. Never put secrets in them.
const baseUrl = process.env.EXPO_PUBLIC_API_URL;
if (!baseUrl) throw new Error("EXPO_PUBLIC_API_URL is not set (see .env.example)");

/** Public catalogue client (sign-in and cart arrive with the customer session, next). */
export const api = createApiClient({ baseUrl: baseUrl.replace(/\/$/, "") });
