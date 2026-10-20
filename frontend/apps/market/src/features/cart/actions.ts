"use server";

import { refresh } from "next/cache";

import { ApiError } from "@techshop/api-client";

import { cartCaller, itemCount, rememberCartToken } from "./server";

export type CartActionResult =
  { ok: true; count: number; message: string } | { ok: false; message: string };

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function failure(e: unknown, fallback: string): CartActionResult {
  if (e instanceof ApiError) {
    if (e.code === "out_of_stock" || e.code === "quantity_too_large")
      return {
        ok: false,
        message: "Sorry, there isn’t enough stock for that quantity.",
      };
    if (e.code === "unavailable")
      return { ok: false, message: "This item is no longer available." };
    if (e.status === 422) return { ok: false, message: e.detail ?? fallback };
    if (e.isUnavailable)
      return {
        ok: false,
        message: "We couldn’t reach the shop. Please try again.",
      };
  }
  return { ok: false, message: fallback };
}

const validQuantity = (q: number) => Number.isInteger(q) && q >= 1 && q <= 99;

export async function addToCart(
  listingId: string,
  quantity: number,
): Promise<CartActionResult> {
  if (!UUID.test(listingId) || !validQuantity(quantity))
    return { ok: false, message: "Choose a quantity between 1 and 99." };
  try {
    const { api, headers } = await cartCaller({ refresh: true });
    const r = await api.request.post("/api/v1/cart/items", {
      headers,
      body: { listingId, quantity },
    });
    await rememberCartToken(r.headers.get("x-cart-token") ?? r.data.cartToken);
    refresh();
    return {
      ok: true,
      count: itemCount(r.data),
      message: "Added to your cart.",
    };
  } catch (e) {
    return failure(e, "We couldn’t add that to your cart.");
  }
}

export async function setQuantity(
  itemId: string,
  quantity: number,
): Promise<CartActionResult> {
  if (!UUID.test(itemId) || !validQuantity(quantity))
    return { ok: false, message: "Choose a quantity between 1 and 99." };
  try {
    const { api, headers } = await cartCaller({ refresh: true });
    const cart = await api.patch("/api/v1/cart/items/{id}", {
      headers,
      path: { id: itemId },
      body: { quantity },
    });
    refresh();
    return { ok: true, count: itemCount(cart), message: "Quantity updated." };
  } catch (e) {
    return failure(e, "We couldn’t update the quantity.");
  }
}

export async function removeItem(itemId: string): Promise<CartActionResult> {
  if (!UUID.test(itemId))
    return { ok: false, message: "That item isn’t in your cart." };
  try {
    const { api, headers } = await cartCaller({ refresh: true });
    const cart = await api.delete("/api/v1/cart/items/{id}", {
      headers,
      path: { id: itemId },
    });
    refresh();
    return {
      ok: true,
      count: itemCount(cart),
      message: "Removed from your cart.",
    };
  } catch (e) {
    return failure(e, "We couldn’t remove that item.");
  }
}
