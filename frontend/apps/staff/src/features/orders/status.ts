import type { Tone } from "@techshop/ui/components";

export const orderTone: Record<string, Tone> = {
  pending_payment: "warning",
  paid: "success",
  processing: "info",
  partially_shipped: "info",
  shipped: "info",
  delivered: "neutral",
  cancelled: "danger",
  refunded: "neutral",
  partially_refunded: "neutral",
};
