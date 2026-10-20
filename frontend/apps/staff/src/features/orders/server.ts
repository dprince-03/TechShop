import "server-only";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function listOrders(q: {
  status?: string;
  q?: string;
  cursor?: string;
}) {
  return (await bff.api()).get("/api/v1/staff/orders", {
    query: { ...q, limit: 50 },
    init: NO_STORE,
  });
}

export async function getOrder(number: string) {
  try {
    return await (
      await bff.api()
    ).get("/api/v1/staff/orders/{number}", {
      path: { number },
      init: NO_STORE,
    });
  } catch (e) {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      return null;
    throw e;
  }
}
