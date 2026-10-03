import type { Money } from "./types";

const formatters = new Map<string, Intl.NumberFormat>();

function formatter(currency: string, fractionDigits: number) {
  const key = `${currency}:${fractionDigits}`;
  let f = formatters.get(key);
  if (!f) {
    f = new Intl.NumberFormat("en-NG", {
      style: "currency",
      currency,
      currencyDisplay: "narrowSymbol",
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    });
    formatters.set(key, f);
  }
  return f;
}

/** Formats minor units as currency, e.g. 125000000 kobo → "₦1,250,000". Kobo shown only when non-zero. */
export function formatMoney({ amount, currency }: Money): string {
  const major = amount / 100;
  return formatter(currency, amount % 100 === 0 ? 0 : 2).format(major);
}

/** Whole-number discount percentage between a previous and current price. */
export function discountPercent(price: Money, compareAt?: Money): number | undefined {
  if (!compareAt || compareAt.amount <= price.amount) return undefined;
  return Math.round(((compareAt.amount - price.amount) / compareAt.amount) * 100);
}

/** Builds a Money value from naira (major units). */
export const naira = (major: number): Money => ({ amount: Math.round(major * 100), currency: "NGN" });
