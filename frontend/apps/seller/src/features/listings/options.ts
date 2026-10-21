/** Choices shared by the listing forms (values match the API). */
export const conditions = [
  { value: "new", label: "New" },
  { value: "open_box", label: "Open box" },
  { value: "uk_used", label: "UK used" },
  { value: "refurbished", label: "Refurbished" },
];
export const grades = [
  { value: "A", label: "A — like new" },
  { value: "B", label: "B — light marks" },
  { value: "C", label: "C — visible wear" },
];
export const warranties = [
  { value: "manufacturer", label: "Manufacturer" },
  { value: "seller", label: "Your shop" },
  { value: "none", label: "No warranty" },
];

/** Plain-language listing statuses. */
export const listingStatus = {
  draft: { label: "Draft", tone: "neutral" },
  auto_check: { label: "Checking", tone: "info" },
  in_review: { label: "In review", tone: "info" },
  changes_requested: { label: "Changes requested", tone: "warning" },
  active: { label: "On sale", tone: "success" },
  paused: { label: "Paused", tone: "neutral" },
  suspended: { label: "Suspended", tone: "danger" },
  rejected: { label: "Rejected", tone: "danger" },
} as const;
