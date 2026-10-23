"use server";

import { redirect } from "next/navigation";

import { ApiError, type Schema } from "@techshop/api-client";

import { safePaymentUrl } from "@/features/orders/payment";
import { bff } from "@/lib/bff";

type Quote = Schema<"Quote">;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const PROVIDERS = ["paystack", "opay", "moniepoint", "store_credit"] as const;
type Provider = (typeof PROVIDERS)[number];

export type PreviewResult =
  | { ok: true; quote: Quote }
  | { ok: false; message: string; field?: "coupon" | "address" };

const couponMessage: Record<string, string> = {
  coupon_not_applicable: "That code doesn’t apply to the items in your cart.",
  coupon_min_order: "Your order is below the minimum for that code.",
};

function previewError(e: unknown): PreviewResult {
  if (e instanceof ApiError) {
    if (couponMessage[e.code])
      return { ok: false, message: couponMessage[e.code], field: "coupon" };
    if (e.code === "bad_address")
      return {
        ok: false,
        message: "Choose a delivery address.",
        field: "address",
      };
    if (e.code === "cart_empty")
      return { ok: false, message: "Your cart is empty." };
    if (e.status === 404 || e.status === 422)
      return {
        ok: false,
        message: e.detail ?? "Check your delivery address and code.",
        field: e.detail?.toLowerCase().includes("coupon")
          ? "coupon"
          : undefined,
      };
  }
  return {
    ok: false,
    message: "We couldn’t work out your total. Please try again.",
  };
}

/** Prices the cart for an address (and optional coupon): delivery, discounts, VAT, total. */
export async function previewCheckout(
  addressId: string,
  coupon: string,
): Promise<PreviewResult> {
  if (!UUID.test(addressId))
    return {
      ok: false,
      message: "Choose a delivery address.",
      field: "address",
    };
  const api = await bff.api({ refresh: true });
  try {
    const quote = await api.post("/api/v1/checkout/preview", {
      body: {
        addressId,
        couponCode: coupon.trim().toUpperCase().slice(0, 40) || undefined,
      },
    });
    return { ok: true, quote };
  } catch (e) {
    return previewError(e);
  }
}

export type PlaceResult =
  | { ok: false; message: string; quote?: Quote }
  | {
      ok: true;
      orderNumber: string;
      payment: Schema<"PaymentAction">;
      totalKobo: number;
    };

/**
 * Places the order with the client's idempotency key (a retry after a network error returns
 * the same order, never a second one), then sends the customer to the payment provider.
 */
export async function placeOrder(input: {
  addressId: string;
  coupon: string;
  priceHash: string;
  provider: string;
  idempotencyKey: string;
}): Promise<PlaceResult> {
  if (!UUID.test(input.addressId))
    return { ok: false, message: "Choose a delivery address." };
  if (!UUID.test(input.idempotencyKey))
    return { ok: false, message: "Please reload the page and try again." };
  if (!PROVIDERS.includes(input.provider as Provider))
    return { ok: false, message: "Choose how you’ll pay." };
  if (!input.priceHash || input.priceHash.length > 200)
    return { ok: false, message: "Please review your total again." };
  const couponCode =
    input.coupon.trim().toUpperCase().slice(0, 40) || undefined;

  const api = await bff.api({ refresh: true });
  let placed: Schema<"Placed">;
  try {
    placed = await api.post("/api/v1/orders", {
      idempotencyKey: input.idempotencyKey,
      body: {
        addressId: input.addressId,
        couponCode,
        priceHash: input.priceHash,
        provider: input.provider as Provider,
      },
    });
  } catch (e) {
    if (e instanceof ApiError && e.code === "price_changed") {
      const fresh = await previewCheckout(input.addressId, input.coupon);
      return {
        ok: false,
        message:
          "Prices or stock changed since you looked. Check the new total, then place your order.",
        quote: fresh.ok ? fresh.quote : undefined,
      };
    }
    if (
      e instanceof ApiError &&
      (e.code === "out_of_stock" || e.code === "cart_has_problems")
    )
      return {
        ok: false,
        message:
          "Some items are no longer available in that quantity. Review your cart.",
      };
    if (e instanceof ApiError && e.code === "provider_unavailable")
      return {
        ok: false,
        message:
          "That payment method is unavailable right now. Choose another one.",
      };
    if (e instanceof ApiError && e.code === "cart_empty")
      return { ok: false, message: "Your cart is empty." };
    if (e instanceof ApiError && e.code === "store_credit_insufficient")
      return {
        ok: false,
        message:
          e.detail ??
          "Your store credit doesn't cover this order. Choose another way to pay.",
      };
    return {
      ok: false,
      message:
        "We couldn’t place your order. You haven’t been charged. Please try again.",
    };
  }

  // Store credit settles at once: straight to the order.
  if (placed.payment.status === "succeeded")
    redirect(`/orders/${placed.orderNumber}?placed=1`);
  const url = safePaymentUrl(placed.payment.checkoutUrl, bff.config().apiUrl);
  if (url) redirect(url);
  // Bank transfers have no hosted page: the form shows the account to pay into.
  return {
    ok: true,
    orderNumber: placed.orderNumber,
    payment: placed.payment,
    totalKobo: placed.totalKobo,
  };
}
