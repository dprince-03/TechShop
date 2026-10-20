"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });
const ROLE = /^[a-z][a-z0-9_]{1,40}$/;

export async function grantRole(
  userId: string,
  role: string,
): Promise<StaffResult> {
  if (!UUID.test(userId) || !ROLE.test(role))
    return { ok: false, message: "Choose a role." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/admin/members/{id}/roles", {
        path: { id: userId },
        body: { role },
      }),
    "Role granted. It applies on their next request.",
  );
}

export async function revokeRole(
  userId: string,
  role: string,
): Promise<StaffResult> {
  if (!UUID.test(userId) || !ROLE.test(role))
    return { ok: false, message: "Unknown role." };
  return runStaffAction(
    async () =>
      (await api()).delete("/api/v1/staff/admin/members/{id}/roles/{role}", {
        path: { id: userId, role },
      }),
    "Role removed.",
  );
}

export async function resetMfa(userId: string): Promise<StaffResult> {
  if (!UUID.test(userId)) return { ok: false, message: "Unknown person." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/admin/members/{id}/reset-mfa", {
        path: { id: userId },
      }),
    "Authenticator reset. They set it up again at next sign-in.",
  );
}
