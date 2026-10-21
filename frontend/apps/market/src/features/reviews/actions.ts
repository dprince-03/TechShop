"use server";

import { refresh } from "next/cache";

import { failure, type ActionState } from "@techshop/bff";

import { bff } from "@/lib/bff";

const text = (form: FormData, name: string, max: number) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);

const stars = (form: FormData) => {
  const n = Number.parseInt(text(form, "rating", 1), 10);
  return n >= 1 && n <= 5 ? n : null;
};

async function signedIn() {
  return (await bff.getSession({ refresh: true })) ? bff.api() : null;
}

/** Reviews an item the customer received (verified purchase). */
export async function reviewItem(
  lineId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const rating = stars(form);
  if (!rating)
    return {
      ok: false,
      message: "Check the highlighted fields.",
      errors: { rating: "Choose 1 to 5 stars." },
    };
  const api = await signedIn();
  if (!api) return { ok: false, message: "Your session ended. Sign in again." };
  try {
    const r = await api.post("/api/v1/me/order-lines/{lineId}/review", {
      path: { lineId },
      body: {
        rating,
        title: text(form, "title", 120) || undefined,
        body: text(form, "body", 3000) || undefined,
      },
    });
    return {
      ok: true,
      message:
        r.status === "published"
          ? "Thanks! Your review is live."
          : "Thanks! Your review will appear after a quick check.",
    };
  } catch (e) {
    return failure(e);
  }
}

/** Rates the seller of a delivered parcel. */
export async function rateSeller(
  fulfilmentId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const rating = stars(form);
  if (!rating)
    return {
      ok: false,
      message: "Check the highlighted fields.",
      errors: { rating: "Choose 1 to 5 stars." },
    };
  const api = await signedIn();
  if (!api) return { ok: false, message: "Your session ended. Sign in again." };
  try {
    await api.post("/api/v1/me/deliveries/{fulfilmentId}/rating", {
      path: { fulfilmentId },
      body: { rating, comment: text(form, "comment", 500) || undefined },
    });
  } catch (e) {
    return failure(e);
  }
  refresh();
  return { ok: true, message: "Thanks for rating the seller." };
}

const REASONS = [
  "abuse",
  "personal_data",
  "off_topic",
  "fake",
  "other",
] as const;

/** Reports a review that breaks the rules; a moderator decides. */
export async function reportReview(
  reviewId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const reason = REASONS.find((r) => r === text(form, "reason", 20));
  if (!reason) return { ok: false, message: "Choose a reason." };
  const api = await signedIn();
  if (!api) return { ok: false, message: "Sign in to report a review." };
  try {
    await api.post("/api/v1/reviews/{id}/report", {
      path: { id: reviewId },
      body: { reason },
    });
  } catch (e) {
    return failure(e);
  }
  return { ok: true, message: "Thanks. A moderator will look at it." };
}
