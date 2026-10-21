import { ApiError } from "@techshop/api-client";

/** What a Server Action returns to its form or button. */
export type ActionState = {
  ok?: boolean;
  message?: string;
  /** Field name → error, shown next to the field. */
  errors?: Record<string, string>;
  /** The API wants a fresh authenticator code first (step-up). */
  stepUp?: boolean;
};

/**
 * Turns an API failure into something a form can show: field errors next to the fields,
 * otherwise the API's own explanation for 4xx, otherwise `fallback`. Never leaks 5xx detail.
 */
export function failure(
  e: unknown,
  fallback = "Something went wrong. Please try again.",
): ActionState {
  if (e instanceof ApiError) {
    if (e.isUnavailable)
      return {
        ok: false,
        message:
          "We couldn’t reach TechShop. Check your connection and try again.",
      };
    if (e.code === "step_up_required")
      return {
        ok: false,
        stepUp: true,
        message: "Confirm with your authenticator code to continue.",
      };
    if (e.status === 401)
      return { ok: false, message: "Your session ended. Sign in again." };
    const errors: Record<string, string> = {};
    for (const f of e.fieldErrors) {
      const name = (f.location ?? "").replace(/^body\./, "");
      if (name) errors[name] = f.message ?? "Check this field.";
    }
    if (Object.keys(errors).length)
      return { ok: false, message: "Check the highlighted fields.", errors };
    if (e.status < 500 && e.detail) return { ok: false, message: e.detail };
  }
  return { ok: false, message: fallback };
}
