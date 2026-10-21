/** Plain-language order statuses for sellers. */
export const orderStatus = {
  pending: { label: "New — accept it", tone: "warning" },
  accepted: { label: "Accepted — pack it", tone: "info" },
  picking: { label: "Being picked", tone: "info" },
  packed: { label: "Packed — drop it at a hub", tone: "warning" },
  handed_over: { label: "At the TechShop hub", tone: "success" },
  in_transit: { label: "On its way", tone: "success" },
  delivered: { label: "Delivered", tone: "success" },
  failed: { label: "Delivery failed", tone: "danger" },
  returned: { label: "Returned", tone: "neutral" },
  cancelled: { label: "Cancelled", tone: "neutral" },
} as const;
