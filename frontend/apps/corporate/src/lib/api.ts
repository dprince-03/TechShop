import "server-only";

import { createApiClient } from "@techshop/api-client";

import { env } from "./env";

// Server-side client for the Go API. Use in Server Components, Route Handlers, and Server Actions.
export const api = createApiClient({ baseUrl: env.API_URL });
