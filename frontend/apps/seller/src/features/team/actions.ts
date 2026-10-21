"use server";

import { redirect } from "next/navigation";

import { type ActionState } from "@techshop/bff";

import { act, text } from "@/lib/actions";

const roles = ["owner", "manager", "staff", "finance"] as const;
type Role = (typeof roles)[number];
const isRole = (r: string): r is Role =>
  (roles as readonly string[]).includes(r);

/** Invites someone by phone or email (whichever was typed). */
export async function inviteMember(
  shopId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const contact = text(form, "contact", 254);
  const role = text(form, "role");
  const errors: Record<string, string> = {};
  if (!contact) errors.contact = "Enter their phone number or email address.";
  if (!isRole(role)) errors.role = "Choose a role.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };
  const body = contact.includes("@")
    ? { email: contact, role: role as Role }
    : { phone: contact, role: role as Role };
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/team/invites", {
        path: { shopId },
        body,
      }),
    "Invite sent. The link works for 7 days.",
  );
}

export async function resendInvite(
  shopId: string,
  inviteId: string,
): Promise<ActionState> {
  return act(
    (api) =>
      api.post("/api/v1/seller/shops/{shopId}/team/invites/{inviteId}/resend", {
        path: { shopId, inviteId },
      }),
    "A new link was sent; the old one no longer works.",
  );
}

export async function revokeInvite(
  shopId: string,
  inviteId: string,
): Promise<ActionState> {
  return act(
    (api) =>
      api.delete("/api/v1/seller/shops/{shopId}/team/invites/{inviteId}", {
        path: { shopId, inviteId },
      }),
    "Invite cancelled.",
  );
}

export async function setRole(
  shopId: string,
  userId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const role = text(form, "role");
  if (!isRole(role)) return { ok: false, message: "Choose a role." };
  return act(
    (api) =>
      api.put("/api/v1/seller/shops/{shopId}/team/members/{userId}/role", {
        path: { shopId, userId },
        body: { role },
      }),
    "Role changed.",
  );
}

export async function removeMember(
  shopId: string,
  userId: string,
): Promise<ActionState> {
  return act(
    (api) =>
      api.delete("/api/v1/seller/shops/{shopId}/team/members/{userId}", {
        path: { shopId, userId },
      }),
    "Removed from the team.",
  );
}

/** Joins the shop from the invited account, then opens it. */
export async function acceptInvite(token: string): Promise<ActionState> {
  let shopId = "";
  const r = await act(async (api) => {
    shopId = (
      await api.post("/api/v1/seller/invites/accept", { body: { token } })
    ).id;
  }, "You’ve joined the shop.");
  if (r.ok && shopId) redirect(`/centre/${shopId}`);
  return r;
}
