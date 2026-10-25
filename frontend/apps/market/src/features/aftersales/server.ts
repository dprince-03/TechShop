import "server-only";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

/** A missing or someone else's record reads as "not found" (no leaks). */
async function orNull<T>(p: Promise<T>): Promise<T | null> {
  try {
    return await p;
  } catch (e) {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      return null;
    throw e;
  }
}

export async function myReturns() {
  return (await bff.api()).get("/api/v1/me/returns", { init: NO_STORE });
}

export async function myReturn(rma: string) {
  const api = await bff.api();
  return orNull(
    api.get("/api/v1/me/returns/{rma}", { path: { rma }, init: NO_STORE }),
  );
}

export async function coveredItems() {
  return (await bff.api()).get("/api/v1/me/warranty/items", { init: NO_STORE });
}

export async function myClaims() {
  return (await bff.api()).get("/api/v1/me/warranty-claims", {
    init: NO_STORE,
  });
}

export async function myClaim(number: string) {
  const api = await bff.api();
  return orNull(
    api.get("/api/v1/me/warranty-claims/{number}", {
      path: { number },
      init: NO_STORE,
    }),
  );
}

export async function tradeInDevices(
  type: "phone" | "laptop" | "tablet" | "console" | "smartwatch",
) {
  return (await bff.api()).get("/api/v1/trade-ins/devices", {
    query: { type },
    init: { next: { revalidate: 300 } },
  });
}

export async function myTradeIns() {
  return (await bff.api()).get("/api/v1/me/trade-ins", { init: NO_STORE });
}

export async function myTradeIn(ref: string) {
  const api = await bff.api();
  return orNull(
    api.get("/api/v1/me/trade-ins/{ref}", { path: { ref }, init: NO_STORE }),
  );
}

export async function myTickets() {
  return (await bff.api()).get("/api/v1/me/tickets", { init: NO_STORE });
}

export async function myTicket(number: string) {
  const api = await bff.api();
  return orNull(
    api.get("/api/v1/me/tickets/{number}", {
      path: { number },
      init: NO_STORE,
    }),
  );
}

export async function storeCredit() {
  return (await bff.api()).get("/api/v1/me/store-credit", { init: NO_STORE });
}

/** Banks for trade-in payouts (same list sellers use). */
export async function banks() {
  try {
    return await (
      await bff.api()
    ).get("/api/v1/me/banks", { init: { next: { revalidate: 3600 } } });
  } catch {
    return [];
  }
}
