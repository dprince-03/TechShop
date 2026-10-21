import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

/** Product search for staff (the catalogue's public index; no staff list endpoint exists). */
export async function searchProducts(q?: string, cursor?: string) {
  return bff
    .publicApi()
    .get("/api/v1/search", { query: { q, cursor, limit: 50 }, init: NO_STORE });
}

export async function moderationQueue() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/catalog/moderation", { init: NO_STORE })) ?? []
  );
}

export async function suggestions() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/catalog/suggestions", { init: NO_STORE })) ?? []
  );
}

export async function categories() {
  return (
    (await bff
      .publicApi()
      .get("/api/v1/categories", { init: { next: { revalidate: 300 } } })) ?? []
  );
}

/** Held reviews and reviews with open reports, oldest first. */
export async function reviewQueue() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/reviews/queue", { init: NO_STORE })) ?? []
  );
}
