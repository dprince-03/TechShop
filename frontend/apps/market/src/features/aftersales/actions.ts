"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import { failure, submitContact, type ActionState } from "@techshop/bff";

import { bff } from "@/lib/bff";

const text = (form: FormData, name: string, max = 200) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);

const invalid = (errors: Record<string, string>): ActionState => ({
  ok: false,
  message: "Check the highlighted fields.",
  errors,
});

async function signedIn() {
  return (await bff.getSession({ refresh: true })) ? bff.api() : null;
}

const SESSION_ENDED: ActionState = {
  ok: false,
  message: "Your session ended. Sign in again.",
};

const REASONS = [
  "faulty",
  "damaged",
  "not_as_described",
  "wrong_item",
  "change_of_mind",
] as const;
type Reason = (typeof REASONS)[number];

/** Returns items from a delivered order. Each line's quantity comes from `qty:<lineId>`. */
export async function requestReturn(
  number: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const items = [...form.entries()]
    .filter(([k]) => k.startsWith("qty:"))
    .map(([k, v]) => ({
      orderLineId: k.slice(4),
      quantity: Number.parseInt(String(v), 10) || 0,
    }))
    .filter((i) => i.quantity > 0);
  const reason = text(form, "reason") as Reason;
  const method = text(form, "method") === "pickup" ? "pickup" : "store_dropoff";
  const errors: Record<string, string> = {};
  if (!items.length)
    return {
      ok: false,
      message: "Choose at least one item to return (set how many).",
    };
  if (!REASONS.includes(reason)) errors.reason = "Choose a reason.";
  if (reason === "change_of_mind" && form.get("unopened") !== "on")
    errors.unopened =
      "Change-of-mind returns must be unopened with the seals intact.";
  if (Object.keys(errors).length) return invalid(errors);
  const api = await signedIn();
  if (!api) return SESSION_ENDED;
  let rma: string;
  try {
    const r = await api.post("/api/v1/me/orders/{number}/returns", {
      path: { number },
      body: {
        items,
        reason,
        method,
        details: text(form, "details", 1000) || undefined,
        unopened: form.get("unopened") === "on" || undefined,
      },
    });
    rma = r.rma;
  } catch (e) {
    return failure(e);
  }
  redirect(`/account/returns/${rma}?new=1`);
}

/** A warranty claim for a covered item. */
export async function makeClaim(
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const orderLineId = text(form, "orderLineId", 40);
  const imeiOrSerial = text(form, "imeiOrSerial", 40).replace(/\s+/g, "");
  const fault = text(form, "fault", 1000);
  const errors: Record<string, string> = {};
  if (!orderLineId) errors.orderLineId = "Choose the item.";
  if (imeiOrSerial.length < 3)
    errors.imeiOrSerial =
      "Enter the IMEI (dial *#06#) or the serial number on the box.";
  if (fault.length < 5)
    errors.fault = "Tell us what's wrong and when it started.";
  if (Object.keys(errors).length) return invalid(errors);
  const api = await signedIn();
  if (!api) return SESSION_ENDED;
  let number: string;
  try {
    number = (
      await api.post("/api/v1/me/warranty-claims", {
        body: { orderLineId, imeiOrSerial, fault },
      })
    ).number;
  } catch (e) {
    return failure(e);
  }
  redirect(`/account/warranty/${number}?new=1`);
}

const DEVICE_TYPES = [
  "phone",
  "laptop",
  "tablet",
  "console",
  "smartwatch",
] as const;
const CONDITIONS = ["like_new", "good", "fair", "faulty"] as const;

/** Starts a trade-in: a catalogue model gets an instant estimate; anything else is priced by staff. */
export async function startTradeIn(
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const deviceType = text(form, "deviceType") as (typeof DEVICE_TYPES)[number];
  const condition = text(form, "condition") as (typeof CONDITIONS)[number];
  const variantId = text(form, "variantId", 40);
  const brand = text(form, "brand", 60);
  const model = text(form, "model", 80);
  const imei = text(form, "imei", 20).replace(/\s+/g, "");
  const errors: Record<string, string> = {};
  if (!DEVICE_TYPES.includes(deviceType))
    errors.deviceType = "Choose the type of device.";
  if (!CONDITIONS.includes(condition))
    errors.condition = "Choose its condition.";
  if (!variantId && (!brand || !model))
    errors.model =
      "Choose your model from the list, or type the brand and model.";
  if (form.get("owner") !== "on")
    errors.owner = "Only the owner can trade a device in.";
  if (imei && !/^\d{15}$/.test(imei))
    errors.imei = "An IMEI is 15 digits. Dial *#06# to see it.";
  if (Object.keys(errors).length) return invalid(errors);
  const api = await signedIn();
  if (!api)
    return { ok: false, message: "Sign in to get your estimate saved." };
  let ref: string;
  try {
    ref = (
      await api.post("/api/v1/me/trade-ins", {
        body: {
          deviceType,
          condition,
          variantId: variantId || undefined,
          brand: brand || undefined,
          model: model || undefined,
          storage: text(form, "storage", 20) || undefined,
          imei: imei || undefined,
        },
      })
    ).reference;
  } catch (e) {
    return failure(e);
  }
  redirect(`/account/trade-ins/${ref}?new=1`);
}

/** Takes the offer, paid as store credit or to a bank account (checked with the bank). */
export async function acceptTradeIn(
  ref: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const payoutMethod =
    text(form, "payoutMethod") === "bank" ? "bank" : "store_credit";
  const bankCode = text(form, "bankCode", 10);
  const accountNumber = text(form, "accountNumber", 10).replace(/\D/g, "");
  if (payoutMethod === "bank") {
    const errors: Record<string, string> = {};
    if (!bankCode) errors.bankCode = "Choose your bank.";
    if (accountNumber.length !== 10)
      errors.accountNumber = "Enter your 10-digit account number.";
    if (Object.keys(errors).length) return invalid(errors);
  }
  const api = await signedIn();
  if (!api) return SESSION_ENDED;
  try {
    await api.post("/api/v1/me/trade-ins/{ref}/accept", {
      path: { ref },
      body:
        payoutMethod === "bank"
          ? {
              payoutMethod,
              bankCode,
              accountNumber,
              accountName: text(form, "accountName", 120) || undefined,
            }
          : { payoutMethod },
    });
  } catch (e) {
    return failure(e);
  }
  refresh();
  return {
    ok: true,
    message:
      payoutMethod === "bank"
        ? "Accepted. We'll send the money within 2 working days."
        : "Accepted. The credit is in your account now.",
  };
}

export async function declineTradeIn(ref: string): Promise<ActionState> {
  const api = await signedIn();
  if (!api) return SESSION_ENDED;
  try {
    await api.post("/api/v1/me/trade-ins/{ref}/decline", { path: { ref } });
  } catch (e) {
    return failure(e);
  }
  refresh();
  return { ok: true, message: "Declined. Collect your device from the store." };
}

export async function cancelTradeIn(ref: string): Promise<ActionState> {
  const api = await signedIn();
  if (!api) return SESSION_ENDED;
  try {
    await api.post("/api/v1/me/trade-ins/{ref}/cancel", { path: { ref } });
  } catch (e) {
    return failure(e);
  }
  refresh();
  return { ok: true, message: "Trade-in cancelled." };
}

const TOPICS = [
  "order",
  "delivery",
  "payment",
  "refund",
  "return",
  "warranty",
  "trade_in",
  "account",
  "selling",
  "business",
  "other",
] as const;
type Topic = (typeof TOPICS)[number];

/** Opens a support ticket for the signed-in customer. */
export async function openTicket(
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const topic = text(form, "topic") as Topic;
  const subject = text(form, "subject", 140);
  const message = text(form, "message", 5000);
  const errors: Record<string, string> = {};
  if (!TOPICS.includes(topic)) errors.topic = "Choose what it's about.";
  if (subject.length < 3) errors.subject = "Give it a short subject.";
  if (!message) errors.message = "Tell us what's happening.";
  if (Object.keys(errors).length) return invalid(errors);
  const api = await signedIn();
  if (!api) return SESSION_ENDED;
  let number: string;
  try {
    number = (
      await api.post("/api/v1/me/tickets", {
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
  redirect(`/account/support/${number}`);
}

export async function replyTicket(
  number: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const message = text(form, "message", 5000);
  if (!message) return invalid({ message: "Write a message." });
  const api = await signedIn();
  if (!api) return SESSION_ENDED;
  try {
    await api.post("/api/v1/me/tickets/{number}/messages", {
      path: { number },
      body: { message },
    });
  } catch (e) {
    return failure(e);
  }
  refresh();
  return { ok: true, message: "Sent. We'll reply here and let you know." };
}

/** The public contact form (no account needed). */
export async function contactUs(
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  return submitContact("market", form);
}
