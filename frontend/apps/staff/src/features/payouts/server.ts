import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function listBatches(cursor?: string) {
  return (await bff.api()).get("/api/v1/staff/finance/payout-batches", {
    query: { cursor },
    init: NO_STORE,
  });
}

export async function getBatch(id: string) {
  return (await bff.api()).get("/api/v1/staff/finance/payout-batches/{id}", {
    path: { id },
    init: NO_STORE,
  });
}
