import { formatMoney } from "@techshop/api-client";

/** ₦ from kobo, for tables and summaries. */
export const naira = (kobo: number) =>
  formatMoney({ amount: kobo, currency: "NGN" });

export const dateTime = new Intl.DateTimeFormat("en-NG", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "Africa/Lagos",
});
