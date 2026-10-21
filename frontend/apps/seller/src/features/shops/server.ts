import "server-only";

import { bff } from "@/lib/bff";

/** Nigerian states for address pickers (public reference data). */
export async function stateOptions() {
  const states = await bff
    .publicApi()
    .get("/api/v1/states", { init: { next: { revalidate: 86400 } } });
  return (states ?? []).map((s) => ({ value: s.code, label: s.name }));
}

/** Plain-language status of a shop. */
export const shopStatus = {
  pending_kyc: { label: "Waiting for verification", tone: "warning" },
  active: { label: "Active", tone: "success" },
  suspended: { label: "Suspended", tone: "danger" },
  closed: { label: "Closed", tone: "neutral" },
} as const;
