// HTTP client
export { buildPath, buildQuery, createApiClient } from "./http/client";
export type { ApiClient, ApiClientOptions } from "./http/client";
export { ApiError, isApiError } from "./http/errors";
export type { FieldError } from "./http/errors";
export type { ApiResponse, CallOptions } from "./http/types";

// Contract
export type * from "./contract";

// Content
export { publishedPage } from "./content";
export type { Site } from "./content";

// Formatting
export {
  discountPercent,
  formatKobo,
  formatMoney,
  naira,
} from "./format/money";
export { normaliseNigerianPhone } from "./format/phone";

// UI view models
export type * from "./domain/types";
