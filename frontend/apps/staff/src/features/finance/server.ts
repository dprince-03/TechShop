import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function financeOverview() {
  const api = await bff.api();
  const [dashboard, providers] = await Promise.all([
    api.get("/api/v1/staff/finance/dashboard", { init: NO_STORE }),
    api.get("/api/v1/staff/finance/providers", { init: NO_STORE }),
  ]);
  return { dashboard, providers };
}

type RefundStatus =
  "requested" | "approved" | "processing" | "succeeded" | "failed" | "rejected";
const REFUND_STATUSES: RefundStatus[] = [
  "requested",
  "approved",
  "processing",
  "succeeded",
  "failed",
  "rejected",
];

export async function listRefunds(raw?: string) {
  const status = REFUND_STATUSES.find((s) => s === raw);
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/finance/refunds", {
      query: { status },
      init: NO_STORE,
    })) ?? []
  );
}

export async function trialBalance() {
  return (await bff.api()).get("/api/v1/staff/finance/trial-balance", {
    init: NO_STORE,
  });
}

export async function listJournals(cursor?: string) {
  return (await bff.api()).get("/api/v1/staff/finance/journals", {
    query: { cursor, limit: 50 },
    init: NO_STORE,
  });
}

export async function listExceptions(
  status?: "open" | "resolved" | "escalated",
) {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/finance/reconciliation/exceptions", {
      query: { status },
      init: NO_STORE,
    })) ?? []
  );
}

export async function listPeriods() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/finance/periods", { init: NO_STORE })) ?? []
  );
}
