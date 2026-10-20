import "server-only";

import { refresh } from "next/cache";

import { ApiError } from "@techshop/api-client";

/** What every staff mutation returns to its button or form. */
export type StaffResult = {
  ok: boolean;
  message: string;
  /** The API wants a fresh authenticator code before this action (step-up). */
  stepUp?: boolean;
  errors?: Record<string, string>;
};

/** Plain-language text for the API's common refusal codes. */
const CODE_MESSAGES: Record<string, string> = {
  forbidden: "You don’t have permission to do that.",
  separation_of_duties:
    "Someone else must approve this (you created or requested it).",
  bad_transition:
    "That isn’t possible in its current state. Refresh and try again.",
  retry: "Someone changed this at the same time. Try again.",
  rate_limited: "Too many attempts. Wait a moment and try again.",
  device_not_approved: "That phone hasn’t been approved yet.",
};

/** Runs an API call for a Server Action and turns failures into a message the UI can show. */
export async function runStaffAction(
  run: () => Promise<unknown>,
  success: string,
): Promise<StaffResult> {
  try {
    await run();
    refresh();
    return { ok: true, message: success };
  } catch (e) {
    if (e instanceof ApiError) {
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
      const message =
        CODE_MESSAGES[e.code] ??
        (e.status === 403 ? CODE_MESSAGES.forbidden : undefined) ??
        (e.status < 500 && e.detail ? e.detail : undefined) ??
        "Something went wrong. Please try again.";
      return {
        ok: false,
        message,
        errors: Object.keys(errors).length ? errors : undefined,
      };
    }
    return { ok: false, message: "Something went wrong. Please try again." };
  }
}

export const UUID =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
