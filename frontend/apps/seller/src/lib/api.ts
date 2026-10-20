import "server-only";

import { createApiClient } from "@techshop/api-client";

/**
 * Server-side client for the Go API's public endpoints. Created per call so `next build` never
 * needs API_URL (pages that read content render per request).
 */
export function api() {
  const baseUrl = process.env.API_URL;
  if (!baseUrl) throw new Error("API_URL is required");
  return createApiClient({ baseUrl });
}
