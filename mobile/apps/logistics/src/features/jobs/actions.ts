import { ApiError, type Schema } from "@techshop/api-client";

import type { MobileSession } from "@techshop/mobile-auth";
import { enqueue } from "@/features/sync/queue";
import { newEventId, queueStore } from "@/lib/storage";

type Action = Schema<"Action">;
export type ActionInput = Omit<Action, "clientEventId" | "occurredAt" | "jobId">;

export type ActionOutcome = { kind: "done" } | { kind: "queued" } | { kind: "refused"; message: string };

/** Actions that can wait for signal. A delivery code can't be checked offline, so it never queues. */
const QUEUEABLE = new Set<Action["kind"]>(["picked_up", "en_route", "attempt_failed", "note"]);

const refusal: Record<string, string> = {
  label_mismatch: "That label isn’t this parcel. Check you have the right one.",
  otp_wrong: "That code isn’t right. Ask the customer to read it again.",
  otp_locked: "Too many wrong codes. Call dispatch.",
  proof_required: "Enter the customer’s delivery code.",
  job_reassigned: "This job was given to another rider. Hand the parcel back at the hub.",
  bad_transition: "That step isn’t possible now. Refresh your jobs.",
  not_on_shift: "Start your shift first.",
  off_shift: "Start your shift first.",
};

/**
 * Sends one action. Online: straight to the API, so the rider sees a wrong label or code at once.
 * No signal: queued with a client event ID (applied exactly once when it syncs).
 */
export async function performAction(session: MobileSession, jobId: string, input: ActionInput, position?: { latitude: number; longitude: number; accuracyM?: number }): Promise<ActionOutcome> {
  const event: Action = { ...input, ...position, jobId, clientEventId: newEventId(), occurredAt: new Date().toISOString() };
  try {
    await session.api.post("/api/v1/rider/jobs/{id}/actions", { path: { id: jobId }, body: event });
    return { kind: "done" };
  } catch (e) {
    if (e instanceof ApiError && e.isUnavailable) {
      if (!QUEUEABLE.has(input.kind)) return { kind: "refused", message: "Delivery needs signal to check the customer’s code. Move to where you have signal and try again." };
      await enqueue(queueStore, { ...event, offline: true });
      return { kind: "queued" };
    }
    if (e instanceof ApiError) return { kind: "refused", message: refusal[e.code] ?? e.detail ?? "The action was refused." };
    return { kind: "refused", message: "Something went wrong. Try again." };
  }
}
