import "server-only";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const TTL = { next: { revalidate: 300 } } as const;

/** Published market help articles (empty when none or the API is unavailable). */
export async function listArticles() {
  try {
    return (
      (await bff.publicApi().get("/api/v1/content/help/{site}", {
        path: { site: "market" },
        init: TTL,
      })) ?? []
    );
  } catch {
    return [];
  }
}

export async function getArticle(slug: string) {
  try {
    return await bff.publicApi().get("/api/v1/content/help/{site}/{slug}", {
      path: { site: "market", slug },
      init: TTL,
    });
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) return null;
    throw e;
  }
}
