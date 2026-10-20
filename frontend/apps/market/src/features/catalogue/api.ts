import "server-only";

import { cache } from "react";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";

import { categoryTree, type MarketCategory } from "./mappers";
import type { ListingQuery } from "./query";

/** Catalogue structure changes rarely: cache it for five minutes across requests. */
const CATEGORY_TTL = { next: { revalidate: 300 } } as const;

/** The category tree, or an empty list when the API can't be reached (the shell still renders). */
export const getCategories = cache(async (): Promise<MarketCategory[]> => {
  try {
    const rows = await bff
      .publicApi()
      .get("/api/v1/categories", { init: CATEGORY_TTL });
    return categoryTree(rows ?? []);
  } catch {
    return [];
  }
});

export async function getCategory(slug: string) {
  return (await getCategories()).find((c) => c.slug === slug);
}

export const LISTING_PAGE_SIZE = 24;

type ListingKind = "search" | "deals" | "best-sellers";

const PATHS = {
  search: "/api/v1/search",
  deals: "/api/v1/deals",
  "best-sellers": "/api/v1/best-sellers",
} as const;

/** One page of products for a listing (search, category, deals, best sellers). */
export async function listProducts(
  q: ListingQuery,
  kind: ListingKind = "search",
) {
  return bff.publicApi().get(PATHS[kind], {
    query: {
      q: q.q,
      category: q.category,
      brand: q.brands.join(",") || undefined,
      condition: q.conditions.join(",") || undefined,
      sellerType: q.sellerType,
      minPrice: q.minKobo,
      maxPrice: q.maxKobo,
      inStock: q.inStock || undefined,
      sort: q.sort,
      cursor: q.cursor,
      limit: q.limit ?? LISTING_PAGE_SIZE,
    },
    init: { cache: "no-store" },
  });
}

/** Product detail, or null when it doesn't exist. Other failures propagate to the error page. */
export async function getProduct(slug: string, shipTo?: string) {
  try {
    return await bff.publicApi().get("/api/v1/products/{slug}", {
      path: { slug },
      query: { shipTo },
      init: { cache: "no-store" },
    });
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) return null;
    throw e;
  }
}

export async function getDeliveryEstimate(
  state: string,
  weightGrams?: number,
  fulfilledBy?: "techshop" | "seller",
) {
  try {
    return await bff.publicApi().get("/api/v1/delivery/estimate", {
      query: { state, weightGrams, fulfilledBy },
      init: { next: { revalidate: 300 } },
    });
  } catch {
    return null;
  }
}

export const getStates = cache(async () => {
  try {
    return (
      (await bff
        .publicApi()
        .get("/api/v1/states", { init: { next: { revalidate: 86_400 } } })) ??
      []
    );
  } catch {
    return [];
  }
});
