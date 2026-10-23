import "server-only";

import type { Query } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function riskSummary() {
  return (await bff.api()).get("/api/v1/staff/risk/summary", {
    init: NO_STORE,
  });
}

export async function cases(status?: string, kind?: string) {
  return (await bff.api()).get("/api/v1/staff/risk/cases", {
    query: {
      status: (status || undefined) as Query<"staffRiskCases">["status"],
      kind: (kind || undefined) as Query<"staffRiskCases">["kind"],
    },
    init: NO_STORE,
  });
}

export async function riskCase(ref: string) {
  return (await bff.api()).get("/api/v1/staff/risk/cases/{ref}", {
    path: { ref },
    init: NO_STORE,
  });
}

export async function blockedDevices(imei?: string) {
  return (await bff.api()).get("/api/v1/staff/risk/devices", {
    query: { imei: imei || undefined },
    init: NO_STORE,
  });
}

export async function checkImei(imei: string) {
  return (await bff.api()).get("/api/v1/staff/risk/devices/check", {
    query: { imei },
    init: NO_STORE,
  });
}

export async function ruleSets() {
  return (await bff.api()).get("/api/v1/staff/risk/rule-sets", {
    init: NO_STORE,
  });
}
