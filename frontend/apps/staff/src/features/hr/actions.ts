"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

export type InviteState = StaffResult & { inviteUrl?: string };

/** Adds a staff member; they get an invitation to set a password and an authenticator. */
export async function inviteStaff(
  _prev: InviteState,
  form: FormData,
): Promise<InviteState> {
  const get = (k: string, max = 100) =>
    String(form.get(k) ?? "")
      .trim()
      .slice(0, max);
  const body = {
    firstName: get("firstName", 60),
    lastName: get("lastName", 60),
    email: get("email", 120).toLowerCase(),
    employeeNo: get("employeeNo", 30),
    jobTitle: get("jobTitle"),
    department: get("department"),
  };
  const errors: Record<string, string> = {};
  for (const [k, v] of Object.entries(body)) if (!v) errors[k] = "Required.";
  if (body.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(body.email))
    errors.email = "Enter a valid email.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };
  let inviteUrl: string | undefined;
  const r = await runStaffAction(async () => {
    inviteUrl = (
      await (
        await bff.api({ refresh: true })
      ).post("/api/v1/staff/hr/members", { body })
    ).inviteUrl;
  }, `Invitation sent to ${body.email}.`);
  // The API only returns the link in development (no email is sent there).
  return { ...r, inviteUrl };
}

export async function exitStaff(userId: string): Promise<StaffResult> {
  if (!UUID.test(userId))
    return { ok: false, message: "Unknown staff member." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/hr/members/{id}/exit",
        { path: { id: userId } },
      ),
    "Exited. Their sessions and roles are revoked.",
  );
}
