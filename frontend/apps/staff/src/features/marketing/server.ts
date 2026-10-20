import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function marketingData() {
  const api = await bff.api();
  const [promotions, coupons, segments, campaigns, journeys] =
    await Promise.all([
      api.get("/api/v1/staff/marketing/promotions", { init: NO_STORE }),
      api.get("/api/v1/staff/marketing/coupons", { init: NO_STORE }),
      api.get("/api/v1/staff/marketing/segments", { init: NO_STORE }),
      api.get("/api/v1/staff/marketing/campaigns", { init: NO_STORE }),
      api.get("/api/v1/staff/marketing/journeys", { init: NO_STORE }),
    ]);
  return {
    promotions: promotions ?? [],
    coupons: coupons ?? [],
    segments: segments ?? [],
    campaigns: campaigns ?? [],
    journeys: journeys ?? [],
  };
}
