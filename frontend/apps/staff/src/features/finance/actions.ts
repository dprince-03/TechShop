"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

export async function approveRefund(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown refund." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/finance/refunds/{id}/approve",
        { path: { id } },
      ),
    "Refund approved. The payout runs next.",
  );
}

export async function rejectRefund(
  id: string,
  note: string,
): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown refund." };
  const why = note.trim().slice(0, 300);
  if (why.length < 3) return { ok: false, message: "Say why it is rejected." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/finance/refunds/{id}/reject",
        { path: { id }, body: { note: why } },
      ),
    "Refund rejected.",
  );
}

export async function resolveException(
  id: string,
  status: "resolved" | "escalated",
  resolution: string,
): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown exception." };
  const text = resolution.trim().slice(0, 500);
  if (text.length < 3) return { ok: false, message: "Describe what was done." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/finance/reconciliation/exceptions/{id}/resolve",
        {
          path: { id },
          body: { status, resolution: text },
        },
      ),
    status === "resolved" ? "Marked resolved." : "Escalated.",
  );
}

export async function closePeriod(month: string): Promise<StaffResult> {
  if (!/^\d{4}-\d{2}$/.test(month))
    return { ok: false, message: "Unknown period." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/finance/periods/{month}/close",
        { path: { month } },
      ),
    `Period ${month} closed.`,
  );
}
