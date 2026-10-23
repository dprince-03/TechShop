import "server-only";

import type { Query } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function supportSummary() {
  return (await bff.api()).get("/api/v1/staff/support/summary", {
    init: NO_STORE,
  });
}

export async function tickets(f: {
  status?: string;
  priority?: string;
  mine?: boolean;
  q?: string;
}) {
  return (await bff.api()).get("/api/v1/staff/support/tickets", {
    query: {
      status: (f.status || undefined) as Query<"staffSupportTickets">["status"],
      priority: (f.priority ||
        undefined) as Query<"staffSupportTickets">["priority"],
      mine: f.mine || undefined,
      q: f.q || undefined,
    },
    init: NO_STORE,
  });
}

export async function ticket(number: string) {
  return (await bff.api()).get("/api/v1/staff/support/tickets/{number}", {
    path: { number },
    init: NO_STORE,
  });
}

export async function agents() {
  return (await bff.api()).get("/api/v1/staff/support/agents", {
    init: { next: { revalidate: 120 } },
  });
}
