import type { CategorySlug, Product, Seller } from "@techshop/api-client";
import {
  categoryBySlug,
  type ProductQuery,
  type SortKey,
  sortOptions,
} from "@techshop/fixtures";

type SearchParams = Record<string, string | string[] | undefined>;

const list = (v: string | string[] | undefined) =>
  v === undefined ? [] : Array.isArray(v) ? v : [v];
const first = (v: string | string[] | undefined) => list(v)[0];

const CONDITIONS: Product["condition"][] = ["new", "uk-used", "refurbished"];
const SELLERS: Seller["type"][] = ["techshop", "vendor"];

const num = (v: string | undefined) => {
  if (!v) return undefined;
  const n = Number(v.replace(/[,\s₦]/g, ""));
  return Number.isFinite(n) && n >= 0 ? n : undefined;
};

/**
 * Parses listing URL params into a product query, ignoring anything invalid
 * (Postel's Law: accept "₦250,000" or "250000" for prices).
 */
export function parseQuery(sp: SearchParams): ProductQuery {
  const sort = first(sp.sort) as SortKey | undefined;
  const category = first(sp.category) as CategorySlug | undefined;
  return {
    q: first(sp.q)?.slice(0, 100),
    category: category && category in categoryBySlug ? category : undefined,
    condition: list(sp.condition).filter((c): c is Product["condition"] =>
      CONDITIONS.includes(c as never),
    ),
    seller: list(sp.seller).filter((s): s is Seller["type"] =>
      SELLERS.includes(s as never),
    ),
    min: num(first(sp.min)),
    max: num(first(sp.max)),
    sort: sortOptions.some((o) => o.value === sort) ? sort : "relevance",
  };
}

/** Number of filters the shopper has applied (sort excluded). */
export const activeFilterCount = (q: ProductQuery) =>
  (q.condition?.length ?? 0) +
  (q.seller?.length ?? 0) +
  (q.min !== undefined ? 1 : 0) +
  (q.max !== undefined ? 1 : 0);
