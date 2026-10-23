"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });
const text = (form: FormData, name: string, max = 200) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);

const STATUSES = ["open", "pending_customer", "resolved", "closed"] as const;
const PRIORITIES = ["low", "normal", "high", "urgent"] as const;

/** A public reply (the requester is notified) or an internal note (never shown to them). */
export async function replyToTicket(
  number: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const body = text(form, "body", 5000);
  if (!body)
    return {
      ok: false,
      message: "Write a message.",
      errors: { body: "Write a message." },
    };
  const internal = form.get("internal") === "on";
  const status = STATUSES.find((s) => s === text(form, "status"));
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/support/tickets/{number}/messages", {
        path: { number },
        body: {
          body,
          internal: internal || undefined,
          status: internal
            ? undefined
            : (status as "open" | "pending_customer" | "resolved" | undefined),
        },
      }),
    internal
      ? "Note added (only staff see it)."
      : "Reply sent. The customer has been notified.",
  );
}

export async function claimTicket(number: string): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/support/tickets/{number}/claim", {
        path: { number },
      }),
    "It's yours.",
  );
}

export async function assignTicket(
  number: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const agent = text(form, "agentId", 40);
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/support/tickets/{number}/assign", {
        path: { number },
        body: { agentId: UUID.test(agent) ? agent : undefined },
      }),
    UUID.test(agent) ? "Assigned." : "Unassigned.",
  );
}

export async function setTicketStatus(
  number: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const status = STATUSES.find((s) => s === text(form, "status"));
  const priority = PRIORITIES.find((p) => p === text(form, "priority"));
  if (!status) return { ok: false, message: "Choose a status." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/support/tickets/{number}/status", {
        path: { number },
        body: { status, priority },
      }),
    "Ticket updated.",
  );
}
