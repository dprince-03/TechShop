// Customer-facing words for after-sales statuses (returns, warranty, trade-ins, tickets),
// shared by the web apps and the mobile apps so every screen says the same thing.

export type StatusTone = "success" | "warning" | "info" | "danger" | "neutral";
type Label = { label: string; tone: StatusTone };

export const returnStatus: Record<string, Label> = {
  requested: { label: "Waiting for approval", tone: "warning" },
  approved: { label: "Approved: bring it to a store", tone: "info" },
  awaiting_pickup: { label: "Approved: a rider will collect it", tone: "info" },
  received: { label: "Received, being checked", tone: "info" },
  inspected: { label: "Checked", tone: "info" },
  refunded: { label: "Refund on its way", tone: "success" },
  replaced: { label: "Replaced", tone: "success" },
  rejected: { label: "Not accepted", tone: "danger" },
  closed: { label: "Closed", tone: "neutral" },
};

export const returnReason: Record<string, string> = {
  faulty: "It's faulty",
  damaged: "It arrived damaged",
  not_as_described: "Not as described",
  wrong_item: "Wrong item sent",
  change_of_mind: "Changed my mind",
};

export const claimStatus: Record<string, Label> = {
  new: { label: "Bring it to a store", tone: "warning" },
  received: { label: "Received at the store", tone: "info" },
  in_repair: { label: "In repair", tone: "info" },
  ready: { label: "Ready to collect", tone: "success" },
  collected: { label: "Collected", tone: "neutral" },
  rejected: { label: "Not covered", tone: "danger" },
  closed: { label: "Refunded", tone: "neutral" },
};

export const claimResolution: Record<string, string> = {
  repaired: "Repaired",
  replaced: "Replaced",
  refunded: "Refunded",
  not_covered: "Not covered",
};

export const tradeInStatus: Record<string, Label> = {
  submitted: { label: "We're pricing it", tone: "warning" },
  estimated: { label: "Estimate ready: bring it to a store", tone: "info" },
  awaiting_inspection: { label: "Being inspected", tone: "info" },
  inspected: { label: "Inspected", tone: "info" },
  offer_sent: { label: "Offer ready", tone: "warning" },
  accepted: { label: "Accepted: payment on its way", tone: "success" },
  paid: { label: "Paid", tone: "success" },
  declined: { label: "Declined: collect your device", tone: "neutral" },
  cancelled: { label: "Cancelled", tone: "neutral" },
  blocked: { label: "Couldn't be accepted", tone: "danger" },
};

export const deviceCondition: Record<string, string> = {
  like_new: "Like new",
  good: "Good",
  fair: "Fair",
  faulty: "Faulty",
};

export const deviceType: Record<string, string> = {
  phone: "Phone",
  laptop: "Laptop",
  tablet: "Tablet",
  console: "Games console",
  smartwatch: "Smartwatch",
};

export const ticketStatus: Record<string, Label> = {
  open: { label: "With our team", tone: "info" },
  pending_customer: { label: "Waiting for your reply", tone: "warning" },
  resolved: { label: "Resolved", tone: "success" },
  closed: { label: "Closed", tone: "neutral" },
};

export const ticketTopic: Record<string, string> = {
  order: "An order",
  delivery: "Delivery",
  payment: "A payment",
  refund: "A refund",
  return: "A return",
  warranty: "Warranty or repair",
  trade_in: "A trade-in",
  account: "My account",
  selling: "Selling on TechShop",
  business: "Business buying",
  other: "Something else",
};
