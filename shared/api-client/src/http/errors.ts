import type { components } from "../generated/schema";

type Problem = components["schemas"]["Problem"];
export type FieldError = components["schemas"]["ErrorDetail"];

/**
 * A failed API call. The Go API answers errors as RFC 9457 problem+json with a stable `code`
 * (for example `out_of_stock`, `price_changed`); screens branch on `code`, never on `detail`.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly detail?: string;
  readonly fieldErrors: FieldError[];
  /** Correlation ID from `X-Request-Id`, worth showing on 5xx screens. */
  readonly requestId?: string;

  constructor(status: number, problem: Partial<Problem>, requestId?: string) {
    super(problem.detail || problem.title || `Request failed with status ${status}`);
    this.name = "ApiError";
    this.status = status;
    this.code = problem.code || (status >= 500 ? "server_error" : "request_failed");
    this.detail = problem.detail;
    this.fieldErrors = problem.errors ?? [];
    this.requestId = requestId;
  }

  /** True when the request never reached the API or the API is down. */
  get isUnavailable(): boolean {
    return this.status === 0 || this.status === 502 || this.status === 503 || this.status === 504;
  }
}

export const isApiError = (e: unknown): e is ApiError => e instanceof ApiError;

/** Network failures (DNS, refused connection, abort) are reported as status 0. */
export function networkError(cause: unknown): ApiError {
  const err = new ApiError(0, { code: "network_error", detail: "Could not reach the TechShop API." });
  (err as { cause?: unknown }).cause = cause;
  return err;
}
