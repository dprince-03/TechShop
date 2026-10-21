"use server";

import { type ActionState } from "@techshop/bff";

import { act, text } from "@/lib/actions";

const reasons = ["out_of_stock", "damaged", "price_error", "other"] as const;
type Reason = (typeof reasons)[number];

export async function acceptOrder(
  shopId: string,
  orderId: string,
): Promise<ActionState> {
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/orders/{orderId}/accept", {
        path: { shopId, orderId },
      }),
    "Accepted. Pack it before the deadline.",
  );
}

export async function packOrder(
  shopId: string,
  orderId: string,
): Promise<ActionState> {
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/orders/{orderId}/pack", {
        path: { shopId, orderId },
      }),
    "Packed. Write the code shown below on the parcel and drop it at a hub.",
  );
}

/** Cancels before packing; the customer is refunded and it counts against the shop. */
export async function cancelOrder(
  shopId: string,
  orderId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const reason = text(form, "reason") as Reason;
  if (!reasons.includes(reason))
    return {
      ok: false,
      message: "Check the highlighted fields.",
      errors: { reason: "Choose a reason." },
    };
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/orders/{orderId}/cancel", {
        path: { shopId, orderId },
        body: { reason, note: text(form, "note", 300) || undefined },
      }),
    "Order cancelled. The customer will be refunded.",
  );
}
