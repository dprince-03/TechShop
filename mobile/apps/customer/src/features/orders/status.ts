// Customer-facing words for the API's statuses (same wording as the market website).
export const orderStatus: Record<string, string> = {
  pending_payment: "Awaiting payment",
  paid: "Payment confirmed",
  processing: "Preparing your order",
  partially_shipped: "Partly on its way",
  shipped: "On its way",
  delivered: "Delivered",
  cancelled: "Cancelled",
  refunded: "Refunded",
  partially_refunded: "Partly refunded",
};

export const parcelStatus: Record<string, string> = {
  pending: "Being prepared",
  accepted: "Accepted by the seller",
  picking: "Being picked",
  packed: "Packed",
  handed_over: "Handed to the courier",
  in_transit: "On its way",
  delivered: "Delivered",
  failed: "Delivery attempt failed",
  returned: "Returned",
  cancelled: "Cancelled",
};

export const when = (iso: string) => new Date(iso).toLocaleString("en-NG", { dateStyle: "medium", timeStyle: "short", timeZone: "Africa/Lagos" });
