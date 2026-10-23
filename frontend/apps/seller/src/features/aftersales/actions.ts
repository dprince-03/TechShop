"use server";

import { redirect } from "next/navigation";

import { failure, type ActionState } from "@techshop/bff";

import { act, text } from "@/lib/actions";
import { bff } from "@/lib/bff";

const invalid = (errors: Record<string, string>): ActionState => ({
  ok: false,
  message: "Check the highlighted fields.",
  errors,
});

/** Disputes TechShop's inspection of the shop's returned items (within 3 days). */
export async function disputeReturn(
  shopId: string,
  rma: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const reason = text(form, "reason", 1000);
  if (reason.length < 10)
    return invalid({
      reason: "Explain what's wrong with the inspection (at least a sentence).",
    });
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/returns/{rma}/dispute", {
        path: { shopId, rma },
        body: { reason },
      }),
    "Dispute sent. A support lead will decide and you'll see the outcome here.",
  );
}

const TOPICS = [
  "order",
  "delivery",
  "payment",
  "refund",
  "return",
  "warranty",
  "account",
  "selling",
  "other",
] as const;

export async function openShopTicket(
  shopId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const topic = TOPICS.find((t) => t === text(form, "topic"));
  const subject = text(form, "subject", 140);
  const message = text(form, "message", 5000);
  const errors: Record<string, string> = {};
  if (!topic) errors.topic = "Choose what it's about.";
  if (subject.length < 3) errors.subject = "Give it a short subject.";
  if (!message) errors.message = "Tell us what's happening.";
  if (Object.keys(errors).length || !topic) return invalid(errors);
  if (!(await bff.getSession({ refresh: true })))
    return { ok: false, message: "Your session ended. Sign in again." };
  let number: string;
  try {
    number = (
      await (
        await bff.api()
      ).post("/api/v1/seller/shops/{shopId}/tickets", {
        path: { shopId },
        body: {
          topic,
          subject,
          message,
          orderNumber: text(form, "orderNumber", 20) || undefined,
        },
      })
    ).number;
  } catch (e) {
    return failure(e);
  }
  redirect(`/centre/${shopId}/support/${number}`);
}

export async function replyShopTicket(
  shopId: string,
  number: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const message = text(form, "message", 5000);
  if (!message) return invalid({ message: "Write a message." });
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/tickets/{number}/messages", {
        path: { shopId, number },
        body: { message },
      }),
    "Sent. We'll reply here.",
  );
}
