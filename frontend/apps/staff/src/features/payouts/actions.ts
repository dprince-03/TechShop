"use server";

import { formatKobo } from "@techshop/api-client";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });

const holdText: Record<string, string> = {
  kyc_incomplete: "not verified",
  enforcement: "suspended",
  no_bank_account: "no verified payout account",
  bank_account_changed: "payout account under 24 hours old",
  credentials_changed: "sign-in details changed in the last 24 hours",
  risk_case: "open risk case",
  negative_balance: "negative balance",
  manual: "held by finance",
  below_minimum: "below the minimum",
};

/** Builds a draft batch from every seller who can be paid now, and says who was left out. */
export async function prepareBatch(): Promise<StaffResult> {
  let message = "";
  const r = await runStaffAction(async () => {
    const d = await (await api()).post("/api/v1/staff/finance/payout-batches");
    const left = (d.skipped ?? []).map(
      (s) =>
        `${s.name} (${(s.reasons ?? []).map((x) => holdText[x] ?? x).join(", ")})`,
    );
    message = `Draft batch: ${d.batch.count} payout${d.batch.count === 1 ? "" : "s"}, ${formatKobo(d.batch.total)}.`;
    if (left.length) message += ` Left out: ${left.join("; ")}.`;
  }, "");
  return r.ok ? { ok: true, message } : r;
}

export async function submitBatch(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown batch." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/finance/payout-batches/{id}/submit", {
        path: { id },
      }),
    "Submitted. A finance manager must approve it.",
  );
}

export async function approveBatch(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown batch." };
  let released = false;
  const r = await runStaffAction(async () => {
    released = (
      await (
        await api()
      ).post("/api/v1/staff/finance/payout-batches/{id}/approve", {
        path: { id },
      })
    ).released;
  }, "");
  if (!r.ok) return r;
  return {
    ok: true,
    message: released
      ? "Approved. The transfers are on their way."
      : "Approved. This batch is above the threshold, so a second manager must approve it too.",
  };
}

export async function cancelBatch(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown batch." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/finance/payout-batches/{id}/cancel", {
        path: { id },
      }),
    "Batch cancelled. Its earnings go into the next run.",
  );
}

export async function excludePayout(
  batchId: string,
  payoutId: string,
  note: string,
  hold: boolean,
): Promise<StaffResult> {
  if (!UUID.test(batchId) || !UUID.test(payoutId))
    return { ok: false, message: "Unknown payout." };
  const why = note.trim().slice(0, 500);
  if (why.length < 3)
    return { ok: false, message: "Say why this seller is left out." };
  return runStaffAction(
    async () =>
      (await api()).post(
        "/api/v1/staff/finance/payout-batches/{id}/payouts/{payoutId}/exclude",
        { path: { id: batchId, payoutId }, body: { note: why, hold } },
      ),
    hold ? "Left out, and later payouts are held." : "Left out of this batch.",
  );
}
