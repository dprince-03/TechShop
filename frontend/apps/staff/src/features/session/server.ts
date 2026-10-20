import "server-only";

import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { cache } from "react";

import { signInPath } from "@techshop/bff";

import { ApiError, type Schema } from "@techshop/api-client";

import { bff } from "@/lib/bff";

export type StaffMember = Schema<"Me">;

/**
 * The signed-in staff member with their permissions, read once per request. The API checks
 * every permission itself; the portal only uses them to hide what the person can't use.
 */
export const getStaff = cache(async (): Promise<StaffMember | null> => {
  if (!(await bff.getSession())) return null;
  try {
    const me = await (
      await bff.api()
    ).get("/api/v1/me", { init: { cache: "no-store" } });
    return me.isStaff ? me : null;
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) return null;
    throw e;
  }
});

export const can = (me: StaffMember | null, permission: string) =>
  Boolean(me?.permissions?.includes(permission));

/**
 * The signed-in staff member, or a redirect to sign-in. Every page calls this itself: Next renders
 * layouts and pages in parallel, so the layout's check alone can't stop a page from running.
 */
export async function requireStaff(permission?: string): Promise<StaffMember> {
  const me = await getStaff();
  if (!me) redirect(signInPath((await headers()).get("x-staff-path") ?? "/"));
  if (permission && !can(me, permission)) redirect("/?denied=1");
  return me;
}
