import { ApiError } from "@techshop/api-client";

/** A short message for a failed review action (the API's own words for 4xx, never 5xx detail). */
export function problem(e: unknown, fallback = "Something went wrong. Try again."): string {
  if (e instanceof ApiError) {
    if (e.isUnavailable) return "No connection. Try again when you’re online.";
    if (e.status === 401) return "Sign in again to continue.";
    if (e.status < 500 && e.detail) return e.detail;
  }
  return fallback;
}
