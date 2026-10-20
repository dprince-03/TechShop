// Display helpers shared by every staff module.
export { formatKobo } from "@techshop/api-client";

const dateTime = new Intl.DateTimeFormat("en-NG", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "Africa/Lagos",
});
const date = new Intl.DateTimeFormat("en-NG", {
  dateStyle: "medium",
  timeZone: "Africa/Lagos",
});

export const formatDateTime = (iso?: string | null) =>
  iso ? dateTime.format(new Date(iso)) : "—";
export const formatDate = (iso?: string | null) =>
  iso ? date.format(new Date(iso)) : "—";

/** Reads one search param as a string (Next gives string | string[] | undefined). */
export const param = (v: string | string[] | undefined) =>
  typeof v === "string" ? v : undefined;
