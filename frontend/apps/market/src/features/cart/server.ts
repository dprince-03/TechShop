import "server-only";

import { cookies } from "next/headers";

import type { Schema } from "@techshop/api-client";

import { bff } from "@/lib/bff";

export type Cart = Schema<"CartView">;

/** Guest carts are keyed by an API-issued token, kept in an httpOnly cookie (never in JS). */
const CART_COOKIE = "ts_cart";
const CART_MAX_AGE = 30 * 86_400;

async function guestToken(): Promise<string | undefined> {
  return (await cookies()).get(CART_COOKIE)?.value || undefined;
}

/** Saves a new guest cart token (Server Functions only). */
export async function rememberCartToken(token: string | null | undefined) {
  if (!token) return;
  const store = await cookies();
  if (store.get(CART_COOKIE)?.value === token) return;
  store.set(CART_COOKIE, token, {
    httpOnly: true,
    secure: bff.config().secure,
    sameSite: "lax",
    path: "/",
    maxAge: CART_MAX_AGE,
  });
}

export async function forgetCartToken() {
  (await cookies()).delete(CART_COOKIE);
}

/**
 * Who the cart belongs to: the signed-in customer (bearer token), else the guest cart token.
 * Server Functions pass `refresh: true` so an expiring session is renewed first.
 */
export async function cartCaller(opts: { refresh?: boolean } = {}) {
  const session = await bff.getSession(opts);
  const token = session ? undefined : await guestToken();
  return {
    api: await bff.api(opts),
    headers: token ? { "X-Cart-Token": token } : undefined,
    signedIn: Boolean(session),
  };
}

/** The current cart, or null when there is none yet (or the API is unavailable). */
export async function getCart(): Promise<Cart | null> {
  const { api, headers, signedIn } = await cartCaller();
  if (!signedIn && !headers) return null;
  try {
    return await api.get("/api/v1/cart", {
      headers,
      init: { cache: "no-store" },
    });
  } catch {
    // An expired guest token or an API outage both read as "no cart": the page still renders.
    return null;
  }
}

export const itemCount = (cart: Cart | null) =>
  (cart?.lines ?? []).reduce((n, l) => n + l.quantity, 0);

/** Moves a guest cart into the customer's cart after sign-in, then drops the guest token. */
export async function mergeGuestCart() {
  const token = await guestToken();
  if (!token) return;
  try {
    const api = await bff.api();
    await api.post("/api/v1/cart/merge", { body: { cartToken: token } });
  } catch {
    // The guest cart stays reachable by its token if merging fails; nothing is lost.
    return;
  }
  await forgetCartToken();
}
