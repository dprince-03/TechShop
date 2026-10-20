// Listing URL params ⇄ the API's search filters. Filters are a plain GET form, so every
// state is a shareable URL and works without JavaScript.
import type { ApiCondition } from "./mappers";

export type SearchParams = Record<string, string | string[] | undefined>;

export type Sort =
  | "relevance"
  | "price_asc"
  | "price_desc"
  | "rating"
  | "discount"
  | "best_selling";

export type SellerType = "first_party" | "business" | "individual";

export type ListingQuery = {
  q?: string;
  /** API category slug (a parent like "phones" or a child like "phones-android"). */
  category?: string;
  brands: string[];
  conditions: ApiCondition[];
  sellerType?: SellerType;
  /** Price bounds in naira as typed, and in kobo for the API. */
  min?: number;
  max?: number;
  minKobo?: number;
  maxKobo?: number;
  inStock: boolean;
  sort: Sort;
  cursor?: string;
  limit?: number;
};

export const sortOptions: { value: Sort; label: string }[] = [
  { value: "relevance", label: "Most relevant" },
  { value: "best_selling", label: "Best selling" },
  { value: "price_asc", label: "Price: low to high" },
  { value: "price_desc", label: "Price: high to low" },
  { value: "discount", label: "Biggest discount" },
  { value: "rating", label: "Top rated" },
];

export const conditionOptions: { value: ApiCondition; label: string }[] = [
  { value: "new", label: "New" },
  { value: "uk_used", label: "UK-used" },
  { value: "refurbished", label: "Refurbished" },
  { value: "open_box", label: "Open box" },
];

export const sellerOptions: { value: SellerType; label: string }[] = [
  { value: "first_party", label: "Sold by TechShop" },
  { value: "business", label: "Business sellers" },
  { value: "individual", label: "Individual sellers" },
];

const list = (v: string | string[] | undefined) =>
  v === undefined ? [] : Array.isArray(v) ? v : [v];
const first = (v: string | string[] | undefined) => list(v)[0];

/** Accepts "₦250,000", "250 000" or "250000" (Postel's Law). */
const naira = (v: string | undefined) => {
  if (!v) return undefined;
  const n = Number(v.replace(/[,\s₦]/g, ""));
  return Number.isFinite(n) && n >= 0 ? Math.floor(n) : undefined;
};

const slugLike = (v: string) => /^[a-z0-9-]{1,60}$/.test(v);

/** Parses listing URL params, ignoring anything invalid rather than failing the page. */
export function parseQuery(sp: SearchParams): ListingQuery {
  const sort = first(sp.sort);
  const seller = first(sp.seller);
  const category = first(sp.category);
  const min = naira(first(sp.min));
  const max = naira(first(sp.max));
  const cursor = first(sp.cursor);
  return {
    q: first(sp.q)?.trim().slice(0, 100) || undefined,
    category: category && slugLike(category) ? category : undefined,
    brands: list(sp.brand).filter(slugLike).slice(0, 20),
    conditions: list(sp.condition).filter((c): c is ApiCondition =>
      conditionOptions.some((o) => o.value === c),
    ),
    sellerType: sellerOptions.find((o) => o.value === seller)?.value,
    min,
    max,
    minKobo: min !== undefined ? min * 100 : undefined,
    maxKobo: max !== undefined ? max * 100 : undefined,
    inStock: first(sp.inStock) === "1",
    sort: sortOptions.find((o) => o.value === sort)?.value ?? "relevance",
    cursor: cursor && cursor.length < 500 ? cursor : undefined,
  };
}

/** Number of filters the shopper applied (sort, search text and paging excluded). */
export const activeFilterCount = (q: ListingQuery) =>
  q.brands.length +
  q.conditions.length +
  (q.sellerType ? 1 : 0) +
  (q.min !== undefined ? 1 : 0) +
  (q.max !== undefined ? 1 : 0) +
  (q.inStock ? 1 : 0);

/** The listing URL for a query, e.g. for "next page" links (cursor replaced). */
export function listingHref(
  path: string,
  q: ListingQuery,
  cursor?: string,
): string {
  const params = new URLSearchParams();
  if (q.q) params.set("q", q.q);
  q.brands.forEach((b) => params.append("brand", b));
  q.conditions.forEach((c) => params.append("condition", c));
  if (q.sellerType) params.set("seller", q.sellerType);
  if (q.min !== undefined) params.set("min", String(q.min));
  if (q.max !== undefined) params.set("max", String(q.max));
  if (q.inStock) params.set("inStock", "1");
  if (q.sort !== "relevance") params.set("sort", q.sort);
  if (cursor) params.set("cursor", cursor);
  const s = params.toString();
  return s ? `${path}?${s}` : path;
}
