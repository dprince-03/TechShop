"use server";

import { type ActionState } from "@techshop/bff";

import { act, text } from "@/lib/actions";

/** The shop's one public answer to a review. */
export async function respond(
  shopId: string,
  reviewId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const reply = text(form, "text", 1000);
  if (reply.length < 2)
    return {
      ok: false,
      message: "Check the highlighted fields.",
      errors: { text: "Write your answer." },
    };
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/reviews/{reviewId}/response", {
        path: { shopId, reviewId },
        body: { text: reply },
      }),
    "Answer posted.",
  );
}
