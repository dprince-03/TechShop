// Turns catalogue API responses into the UI view models the shared components render.
import type { CategorySlug, Product, Schema } from "@techshop/api-client";

export type Subcategory = {
  /** URL segment under the parent, e.g. "android" for /c/phones/android. */
  slug: string;
  /** The API's slug, e.g. "phones-android". */
  apiSlug: string;
  name: string;
};

export type MarketCategory = {
  id: string;
  slug: string;
  name: string;
  description: string;
  /** "vehicle" categories (cars) are listings, not add-to-cart products. */
  kind: string;
  subcategories: Subcategory[];
};

const ART_SLUGS: readonly string[] = [
  "phones",
  "laptops",
  "accessories",
  "gaming",
  "smart-home",
  "office",
  "workstations",
  "cars",
];

/** The placeholder-art key for a category slug, when one exists. */
export const artFor = (slug?: string): CategorySlug | undefined =>
  slug && ART_SLUGS.includes(slug) ? (slug as CategorySlug) : undefined;

/** Builds the two-level tree from the flat category list (ordered by `position`). */
export function categoryTree(
  rows: Schema<"CatalogCategory">[],
): MarketCategory[] {
  const active = rows
    .filter((c) => c.isActive)
    .sort((a, b) => a.position - b.position);
  return active
    .filter((c) => !c.parentId)
    .map((parent) => ({
      id: parent.id,
      slug: parent.slug,
      name: parent.name,
      description: parent.description ?? "",
      kind: parent.kind,
      subcategories: active
        .filter((c) => c.parentId === parent.id)
        .map((c) => ({
          apiSlug: c.slug,
          slug: c.slug.startsWith(`${parent.slug}-`)
            ? c.slug.slice(parent.slug.length + 1)
            : c.slug,
          name: c.name,
        })),
    }));
}

export type ApiCondition = "new" | "uk_used" | "refurbished" | "open_box";

const CONDITION_TO_UI: Record<string, Product["condition"]> = {
  new: "new",
  uk_used: "uk-used",
  refurbished: "refurbished",
  // Open-box items are shown with the refurbished treatment until the card has its own badge.
  open_box: "refurbished",
};

export const conditionLabel: Record<string, string> = {
  new: "New",
  uk_used: "UK-used",
  refurbished: "Refurbished",
  open_box: "Open box",
};

/** The condition a card shows: the only one when there is one, otherwise "new" (no badge). */
function cardCondition(conditions: string[] | null): Product["condition"] {
  const list = conditions ?? [];
  return list.length === 1 ? (CONDITION_TO_UI[list[0]] ?? "new") : "new";
}

const stockFromBand = (inStock: boolean, band?: string): Product["stock"] =>
  !inStock || band === "out"
    ? "out-of-stock"
    : band === "low"
      ? "low-stock"
      : "in-stock";

/** A search/listing result as a product card. `category` sets the art when the caller knows it. */
export function toCard(
  p: Schema<"ProductSummary">,
  category?: string,
): Product {
  const list = p.price ?? 0;
  const promo = p.promoPrice;
  const firstParty = (p.sellerTypes ?? []).includes("first_party");
  return {
    id: p.id,
    slug: p.slug,
    name: p.name,
    brand: p.brandName ?? "",
    category: artFor(category),
    condition: cardCondition(p.conditions),
    price: { amount: promo ?? list, currency: "NGN" },
    compareAtPrice:
      promo !== undefined && promo < list
        ? { amount: list, currency: "NGN" }
        : p.compareAt && p.compareAt > list
          ? { amount: p.compareAt, currency: "NGN" }
          : undefined,
    rating:
      p.ratingCount > 0 && p.ratingAvg
        ? { average: p.ratingAvg, count: p.ratingCount }
        : undefined,
    seller: firstParty
      ? { id: "techshop", name: "TechShop", type: "techshop" }
      : { id: "marketplace", name: "a verified seller", type: "vendor" },
    offerCount: p.offerCount,
    stock: stockFromBand(p.inStock, p.stockBand),
  };
}
