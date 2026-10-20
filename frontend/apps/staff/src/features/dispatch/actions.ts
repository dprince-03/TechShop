"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });

export async function assignJob(
  jobId: string,
  riderId: string,
): Promise<StaffResult> {
  if (!UUID.test(jobId) || !UUID.test(riderId))
    return { ok: false, message: "Choose a rider." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/dispatch/jobs/{id}/assign", {
        path: { id: jobId },
        body: { riderId },
      }),
    "Assigned. The rider sees it in their app.",
  );
}

export async function unassignJob(jobId: string): Promise<StaffResult> {
  if (!UUID.test(jobId)) return { ok: false, message: "Unknown job." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/dispatch/jobs/{id}/unassign", {
        path: { id: jobId },
      }),
    "Unassigned.",
  );
}

export async function returnToHub(
  jobId: string,
  note: string,
): Promise<StaffResult> {
  if (!UUID.test(jobId)) return { ok: false, message: "Unknown job." };
  if (note.trim().length < 3)
    return { ok: false, message: "Say why it’s coming back." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/dispatch/jobs/{id}/return", {
        path: { id: jobId },
        body: { note: note.trim().slice(0, 300) },
      }),
    "Marked to return to the hub.",
  );
}

export async function setRiderStatus(
  riderId: string,
  action: "suspend" | "reinstate",
): Promise<StaffResult> {
  if (!UUID.test(riderId)) return { ok: false, message: "Unknown rider." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/dispatch/riders/{id}/{action}", {
        path: { id: riderId, action },
      }),
    action === "suspend" ? "Rider suspended." : "Rider reinstated.",
  );
}

export async function decideDevice(
  deviceId: string,
  action: "approve" | "revoke",
): Promise<StaffResult> {
  if (!UUID.test(deviceId)) return { ok: false, message: "Unknown phone." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/dispatch/devices/{id}/{action}", {
        path: { id: deviceId, action },
      }),
    action === "approve" ? "Phone approved." : "Phone revoked.",
  );
}
