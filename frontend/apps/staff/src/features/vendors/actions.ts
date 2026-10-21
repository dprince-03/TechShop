"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

type Decision = "approved" | "rejected" | "needs_more_info";
const DECISIONS: Decision[] = ["approved", "rejected", "needs_more_info"];

/** A KYC decision; anything but approval needs a note the seller will read. */
export async function reviewKyc(
  submissionId: string,
  decision: string,
  note: string,
): Promise<StaffResult> {
  const d = DECISIONS.find((x) => x === decision);
  if (!UUID.test(submissionId) || !d)
    return { ok: false, message: "Unknown submission or decision." };
  if (d !== "approved" && note.trim().length < 3)
    return { ok: false, message: "Tell the seller why, so they can fix it." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/vendors/kyc/{id}/review",
        {
          path: { id: submissionId },
          body: { decision: d, note: note.trim().slice(0, 500) || undefined },
        },
      ),
    d === "approved"
      ? "Approved. The shop is active and the seller has been emailed."
      : "Decision sent to the seller.",
  );
}

/**
 * Shows the full ID number to a reviewer who gives a reason. The API demands a recent
 * authenticator check (the button asks for it) and audits every reveal.
 */
export async function revealId(
  submissionId: string,
  reason: string,
): Promise<StaffResult> {
  if (!UUID.test(submissionId))
    return { ok: false, message: "Unknown submission." };
  if (reason.trim().length < 10)
    return {
      ok: false,
      message: "Give a reason of at least 10 characters (it’s audited).",
    };
  let number = "";
  const r = await runStaffAction(async () => {
    const out = await (
      await bff.api({ refresh: true })
    ).post("/api/v1/staff/vendors/kyc/{id}/reveal", {
      path: { id: submissionId },
      body: { reason: reason.trim().slice(0, 300) },
    });
    number = out.idNumber;
  }, "");
  return r.ok ? { ok: true, message: `Full ID number: ${number}` } : r;
}
