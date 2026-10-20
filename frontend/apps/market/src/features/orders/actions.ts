"use server";

import { refresh } from "next/cache";
import { redirect } from "next/navigation";

import {
  ApiError,
  normaliseNigerianPhone,
  type Schema,
} from "@techshop/api-client";

import { bff } from "@/lib/bff";

import { safePaymentUrl } from "./payment";

const ORDER_NUMBER = /^TS-\d{1,12}$/;
const PROVIDERS = ["paystack", "opay", "moniepoint"] as const;
type Provider = (typeof PROVIDERS)[number];

export type ActionResult = { ok: boolean; message?: string };

export async function cancelOrder(
  number: string,
  reason: string,
): Promise<ActionResult> {
  if (!ORDER_NUMBER.test(number))
    return { ok: false, message: "Unknown order." };
  const api = await bff.api({ refresh: true });
  try {
    await api.post("/api/v1/me/orders/{number}/cancel", {
      path: { number },
      body: { reason: reason.trim().slice(0, 200) || "Cancelled by customer" },
    });
    refresh();
    return { ok: true, message: "Your order was cancelled." };
  } catch (e) {
    if (e instanceof ApiError && e.code === "too_late_to_cancel")
      return {
        ok: false,
        message:
          "This order is already on its way, so it can’t be cancelled here. Contact support.",
      };
    return {
      ok: false,
      message: "We couldn’t cancel the order. Please try again.",
    };
  }
}

/** Starts a new payment for an unpaid order and sends the customer to the provider. */
export async function retryPayment(
  number: string,
  provider: string,
): Promise<ActionResult> {
  if (!ORDER_NUMBER.test(number))
    return { ok: false, message: "Unknown order." };
  if (!PROVIDERS.includes(provider as Provider))
    return { ok: false, message: "Choose a payment method." };
  const api = await bff.api({ refresh: true });
  let action: Schema<"PaymentAction">;
  try {
    action = await api.post("/api/v1/me/orders/{number}/payments", {
      path: { number },
      body: { provider: provider as Provider },
    });
  } catch (e) {
    if (e instanceof ApiError && e.code === "provider_unavailable")
      return {
        ok: false,
        message:
          "That payment method is unavailable right now. Try another one.",
      };
    return {
      ok: false,
      message: "We couldn’t start the payment. Please try again.",
    };
  }
  const url = safePaymentUrl(action.checkoutUrl, bff.config().apiUrl);
  if (url) redirect(url);
  refresh();
  return {
    ok: true,
    message: action.message ?? "Follow the payment instructions below.",
  };
}

export type TrackResult =
  | { state: "idle" }
  | { state: "error"; message: string }
  | { state: "found"; tracking: Schema<"Tracking"> };

/** Guest tracking: order number + the phone used at checkout (the API rate-limits this). */
export async function trackOrder(
  _prev: TrackResult,
  form: FormData,
): Promise<TrackResult> {
  const number = String(form.get("number") ?? "")
    .trim()
    .toUpperCase();
  const phone = normaliseNigerianPhone(String(form.get("phone") ?? ""));
  if (!ORDER_NUMBER.test(number))
    return {
      state: "error",
      message: "Enter your order number, e.g. TS-10482.",
    };
  if (!phone)
    return {
      state: "error",
      message: "Enter the mobile number you used at checkout.",
    };
  try {
    const tracking = await (
      await bff.clientApi()
    ).get("/api/v1/track/{number}", {
      path: { number },
      query: { phone },
      init: { cache: "no-store" },
    });
    return { state: "found", tracking };
  } catch (e) {
    if (e instanceof ApiError && e.code === "rate_limited")
      return {
        state: "error",
        message: "Too many lookups. Please wait a few minutes.",
      };
    return {
      state: "error",
      message: "We couldn’t find an order with that number and phone.",
    };
  }
}
