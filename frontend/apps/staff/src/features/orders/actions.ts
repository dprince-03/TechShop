"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const ORDER = /^TS-\d{1,12}$/;

export async function cancelOrder(
  number: string,
  reason: string,
): Promise<StaffResult> {
  if (!ORDER.test(number)) return { ok: false, message: "Unknown order." };
  const why = reason.trim().slice(0, 200);
  if (why.length < 3)
    return {
      ok: false,
      message: "Give a reason (it goes in the order history).",
    };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/orders/{number}/cancel",
        { path: { number }, body: { reason: why } },
      ),
    "Order cancelled. The customer is told and any payment goes to refunds.",
  );
}

export async function cancelLine(
  number: string,
  lineId: string,
  reason: string,
): Promise<StaffResult> {
  if (!ORDER.test(number) || !UUID.test(lineId))
    return { ok: false, message: "Unknown order line." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/orders/{number}/lines/{lineId}/cancel",
        {
          path: { number, lineId },
          body: { reason: reason.trim().slice(0, 200) || "Cancelled by staff" },
        },
      ),
    "Line cancelled.",
  );
}

export async function verifyPayment(paymentId: string): Promise<StaffResult> {
  if (!UUID.test(paymentId)) return { ok: false, message: "Unknown payment." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/finance/payments/{id}/verify",
        { path: { id: paymentId } },
      ),
    "Checked with the provider.",
  );
}
