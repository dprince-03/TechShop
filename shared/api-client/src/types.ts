// Response shapes returned by the Go API (backend/internal/handler).
// Shared by web and mobile — the API contract.

export type HealthResponse = {
  status: "ok" | "degraded";
  database: "up" | "down";
};

/** Amounts are integers in the currency's minor unit (kobo for NGN). */
export type Money = {
  amount: number;
  currency: "NGN";
};

export type CategorySlug =
  | "phones"
  | "laptops"
  | "accessories"
  | "gaming"
  | "smart-home"
  | "office"
  | "workstations"
  | "cars";

export type Category = {
  slug: CategorySlug;
  name: string;
  description: string;
  subcategories: { slug: string; name: string }[];
};

export type Seller = {
  id: string;
  name: string;
  /** "techshop" = sold by TechShop itself; "vendor" = marketplace seller. */
  type: "techshop" | "vendor";
  rating?: number;
};

export type ProductCondition = "new" | "uk-used" | "refurbished";

export type Product = {
  id: string;
  slug: string;
  name: string;
  brand: string;
  category: CategorySlug;
  subcategory?: string;
  /** Short spec line shown on cards, e.g. "8GB RAM · 256GB". */
  keySpec?: string;
  condition: ProductCondition;
  price: Money;
  /** Previous price, when discounted. */
  compareAtPrice?: Money;
  rating?: { average: number; count: number };
  seller: Seller;
  stock: "in-stock" | "low-stock" | "out-of-stock";
  badges?: ("new" | "deal" | "best-seller")[];
};

/** Wholesale price break: unit price when buying at least `minQuantity`. */
export type PriceTier = {
  minQuantity: number;
  unitPrice: Money;
};

export type WholesaleProduct = Product & {
  minOrderQuantity: number;
  tiers: PriceTier[];
};

/** Cars are listings (inspection, viewing, financing), not add-to-cart items. */
export type CarListing = {
  id: string;
  slug: string;
  title: string;
  year: number;
  mileageKm: number;
  location: string;
  condition: "brand-new" | "foreign-used" | "nigerian-used";
  transmission: "automatic" | "manual";
  price: Money;
  inspected: boolean;
  seller: Seller;
};
