import "server-only";

import type { Query } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;
const api = () => bff.api();

type ReturnStatus = Query<"staffListReturns">["status"];

export async function returnsList(
  status?: string,
  disputed = false,
  cursor?: string,
) {
  return (await api()).get("/api/v1/staff/returns", {
    query: {
      status: (status || undefined) as ReturnStatus,
      disputed: disputed || undefined,
      cursor,
    },
    init: NO_STORE,
  });
}

export async function returnDetail(rma: string) {
  return (await api()).get("/api/v1/staff/returns/{rma}", {
    path: { rma },
    init: NO_STORE,
  });
}

/** Warehouses and stores (receiving and restocking destinations). */
export async function locations() {
  return (await api()).get("/api/v1/staff/warranty/stores", {
    init: { next: { revalidate: 300 } },
  });
}

export async function warrantySummary() {
  return (await api()).get("/api/v1/staff/warranty/summary", {
    init: NO_STORE,
  });
}

export async function claims(status?: string, q?: string) {
  return (await api()).get("/api/v1/staff/warranty/claims", {
    query: {
      status: (status || undefined) as Query<"staffWarrantyClaims">["status"],
      q: q || undefined,
    },
    init: NO_STORE,
  });
}

export async function claim(number: string) {
  return (await api()).get("/api/v1/staff/warranty/claims/{number}", {
    path: { number },
    init: NO_STORE,
  });
}

export async function repairs(status?: string, purpose?: string) {
  return (await api()).get("/api/v1/staff/warranty/repairs", {
    query: {
      status: (status || undefined) as Query<"staffRepairQueue">["status"],
      purpose: (purpose || undefined) as Query<"staffRepairQueue">["purpose"],
    },
    init: NO_STORE,
  });
}

export async function tradeInSummary() {
  return (await api()).get("/api/v1/staff/trade-ins/summary", {
    init: NO_STORE,
  });
}

export async function tradeIns(status?: string, q?: string) {
  return (await api()).get("/api/v1/staff/trade-ins", {
    query: {
      status: (status || undefined) as Query<"staffTradeIns">["status"],
      q: q || undefined,
    },
    init: NO_STORE,
  });
}

export async function tradeIn(ref: string) {
  return (await api()).get("/api/v1/staff/trade-ins/{ref}", {
    path: { ref },
    init: NO_STORE,
  });
}

/** Catalogue models a trade-in can be matched to. */
export async function tradeInModels(type: string) {
  return (await bff.publicApi()).get("/api/v1/trade-ins/devices", {
    query: { type: type as Query<"tradeInDevices">["type"] },
    init: { next: { revalidate: 300 } },
  });
}
