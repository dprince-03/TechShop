import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export type VendorFilter = "to_review" | "all" | "active" | "suspended";

/** Marketplace sellers for the vendor list (newest first). */
export async function vendors(
  filter: VendorFilter,
  q?: string,
  cursor?: string,
) {
  const query = {
    q: q || undefined,
    cursor,
    kycStatus: filter === "to_review" ? ("submitted" as const) : undefined,
    status:
      filter === "active"
        ? ("active" as const)
        : filter === "suspended"
          ? ("suspended" as const)
          : undefined,
  };
  return (await bff.api()).get("/api/v1/staff/vendors", {
    query,
    init: NO_STORE,
  });
}

/** One seller with KYC, bank accounts and duplicate signals. */
export async function vendor(id: string) {
  return (await bff.api()).get("/api/v1/staff/vendors/{id}", {
    path: { id },
    init: NO_STORE,
  });
}
