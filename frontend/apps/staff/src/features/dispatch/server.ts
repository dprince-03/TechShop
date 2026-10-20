import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export type JobStatus =
  | "unassigned"
  | "assigned"
  | "picked_up"
  | "en_route"
  | "delivered"
  | "failed"
  | "returned"
  | "cancelled";
export const JOB_STATUSES: JobStatus[] = [
  "unassigned",
  "assigned",
  "picked_up",
  "en_route",
  "delivered",
  "failed",
  "returned",
  "cancelled",
];

export async function board(status?: JobStatus) {
  return (await bff.api()).get("/api/v1/dispatch/board", {
    query: { status },
    init: NO_STORE,
  });
}

export async function riders() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/dispatch/riders", { init: NO_STORE })) ?? []
  );
}

export async function pendingDevices() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/dispatch/devices", {
      query: { pending: true },
      init: NO_STORE,
    })) ?? []
  );
}

export async function zones() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/dispatch/zones", { init: NO_STORE })) ?? []
  );
}

export async function zoneRates(id: string) {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/dispatch/zones/{id}/rates", {
      path: { id },
      init: NO_STORE,
    })) ?? []
  );
}
