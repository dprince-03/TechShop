/**
 * Public URLs of every TechShop web app, for cross-app links.
 * Override per environment with NEXT_PUBLIC_*_URL; defaults are the local nginx hostnames.
 */
export const sites = {
  corporate:
    process.env.NEXT_PUBLIC_CORPORATE_URL ?? "http://techshop.localhost",
  market:
    process.env.NEXT_PUBLIC_MARKET_URL ?? "http://market.techshop.localhost",
  wholesale:
    process.env.NEXT_PUBLIC_WHOLESALE_URL ??
    "http://wholesale.techshop.localhost",
  seller:
    process.env.NEXT_PUBLIC_SELLER_URL ?? "http://seller.techshop.localhost",
  staff: process.env.NEXT_PUBLIC_STAFF_URL ?? "http://staff.techshop.localhost",
} as const;
