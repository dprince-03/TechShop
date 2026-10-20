// Customer-facing words for the API's order, parcel and payment statuses (no system jargon).
export const orderStatus: Record<
  string,
  { label: string; tone: "neutral" | "success" | "warning" | "danger" }
> = {
  pending_payment: { label: "Awaiting payment", tone: "warning" },
  paid: { label: "Payment confirmed", tone: "success" },
  processing: { label: "Preparing your order", tone: "neutral" },
  partially_shipped: { label: "Partly on its way", tone: "neutral" },
  shipped: { label: "On its way", tone: "neutral" },
  delivered: { label: "Delivered", tone: "success" },
  cancelled: { label: "Cancelled", tone: "danger" },
  refunded: { label: "Refunded", tone: "neutral" },
  partially_refunded: { label: "Partly refunded", tone: "neutral" },
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

export const paymentStatus: Record<string, string> = {
  initiated: "Started",
  pending: "Waiting for confirmation",
  pending_review: "Being reviewed",
  succeeded: "Paid",
  failed: "Failed",
  cancelled: "Cancelled",
  reversed: "Reversed",
};

export const providerLabel: Record<string, string> = {
  paystack: "Card (Paystack)",
  opay: "OPay",
  moniepoint: "Bank transfer (Moniepoint)",
};

export const formatDateTime = (iso: string) =>
  new Intl.DateTimeFormat("en-NG", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Africa/Lagos",
  }).format(new Date(iso));

export const formatDate = (iso: string) =>
  new Intl.DateTimeFormat("en-NG", {
    dateStyle: "medium",
    timeZone: "Africa/Lagos",
  }).format(new Date(iso));
