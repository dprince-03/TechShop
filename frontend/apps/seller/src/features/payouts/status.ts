/** Plain-language payout statuses for sellers. */
export const payoutStatus = {
  scheduled: { label: "Scheduled", tone: "info" },
  processing: { label: "On its way to your bank", tone: "info" },
  paid: { label: "Paid", tone: "success" },
  failed: { label: "Failed — check your account", tone: "danger" },
  on_hold: { label: "On hold", tone: "warning" },
} as const;
