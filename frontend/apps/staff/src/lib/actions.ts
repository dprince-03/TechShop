import "server-only";

import { refresh } from "next/cache";

import { ApiError } from "@techshop/api-client";
import { failure } from "@techshop/bff";

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
    // Staff-specific wording for the API's refusal codes; everything else as for any form.
    if (
      e instanceof ApiError &&
      e.code !== "forbidden" &&
      CODE_MESSAGES[e.code]
    )
      return { ok: false, message: CODE_MESSAGES[e.code] };
    const f = failure(e);
    if (e instanceof ApiError && e.status === 403 && !e.detail)
      return { ok: false, message: CODE_MESSAGES.forbidden };
    return {
      ok: false,
      message: f.message ?? "Something went wrong. Please try again.",
      stepUp: f.stepUp,
      errors: f.errors,
    };
  }
}

export const UUID =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
