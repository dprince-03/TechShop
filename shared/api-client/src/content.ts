import type { ApiClient } from "./http/client";

export type Site = "corporate" | "market" | "wholesale" | "seller";

/**
 * A published CMS page, or null when it isn't published yet (404) or the API can't be reached.
 * Callers fall back to their written outline, so a page never breaks because content is missing.
 */
export async function publishedPage(
  api: ApiClient,
  site: Site,
  slug: string,
  init?: RequestInit,
) {
  try {
    return await api.get("/api/v1/content/pages/{site}/{slug}", {
      path: { site, slug },
      init: init ?? ({ next: { revalidate: 60 } } as RequestInit),
    });
  } catch {
    return null;
  }
}
