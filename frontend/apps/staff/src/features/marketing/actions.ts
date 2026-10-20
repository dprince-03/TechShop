"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });

type CampaignStep = "submit" | "approve" | "pause" | "resume" | "cancel";
const CAMPAIGN_DONE: Record<CampaignStep, string> = {
  submit: "Sent for approval.",
  approve: "Approved. It sends at its scheduled time.",
  pause: "Paused.",
  resume: "Resumed.",
  cancel: "Cancelled.",
};

/** Campaign workflow. The API requires a different person (with step-up) to approve. */
export async function campaignStep(
  id: string,
  step: CampaignStep,
): Promise<StaffResult> {
  if (!UUID.test(id) || !(step in CAMPAIGN_DONE))
    return { ok: false, message: "Unknown campaign." };
  return runStaffAction(async () => {
    const a = await api();
    const opts = { path: { id }, body: {} };
    if (step === "submit")
      return a.post("/api/v1/staff/marketing/campaigns/{id}/submit", opts);
    if (step === "approve")
      return a.post("/api/v1/staff/marketing/campaigns/{id}/approve", opts);
    if (step === "pause")
      return a.post("/api/v1/staff/marketing/campaigns/{id}/pause", opts);
    if (step === "resume")
      return a.post("/api/v1/staff/marketing/campaigns/{id}/resume", opts);
    return a.post("/api/v1/staff/marketing/campaigns/{id}/cancel", opts);
  }, CAMPAIGN_DONE[step]);
}

export async function promotionStep(
  id: string,
  action: "end" | "cancel",
): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown promotion." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/marketing/promotions/{id}/{action}", {
        path: { id, action },
      }),
    action === "end" ? "Promotion ended." : "Promotion cancelled.",
  );
}

export async function journeyStep(
  id: string,
  action: "activate" | "pause" | "archive",
): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown journey." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/marketing/journeys/{id}/{action}", {
        path: { id, action },
      }),
    `Journey ${action === "activate" ? "activated" : action === "pause" ? "paused" : "archived"}.`,
  );
}
