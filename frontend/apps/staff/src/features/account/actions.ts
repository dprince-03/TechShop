"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

export async function revokeDevice(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown device." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).delete("/api/v1/me/sessions/{id}", {
        path: { id },
      }),
    "Signed out of that device.",
  );
}

export async function revokeOtherDevices(): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/me/sessions/revoke-others",
      ),
    "Signed out of all other devices.",
  );
}
