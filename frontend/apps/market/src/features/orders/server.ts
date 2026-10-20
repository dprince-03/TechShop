import "server-only";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function listMyOrders(cursor?: string) {
  const api = await bff.api();
  return api.get("/api/v1/me/orders", {
    query: { cursor, limit: 20 },
    init: NO_STORE,
  });
}

/** The customer's order, or null when it isn't theirs or doesn't exist (same answer: no leaks). */
export async function getMyOrder(number: string) {
  const api = await bff.api();
  try {
    return await api.get("/api/v1/me/orders/{number}", {
      path: { number },
      init: NO_STORE,
    });
  } catch (e) {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      return null;
    throw e;
  }
}
