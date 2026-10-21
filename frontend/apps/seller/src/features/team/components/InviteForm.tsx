"use client";

import { ActionForm, Field } from "@techshop/ui/components";

import { inviteMember } from "../actions";

/** Invite by phone or email. Managers can't offer the owner role. */
export function InviteForm({
  shopId,
  canInviteOwner,
}: {
  shopId: string;
  canInviteOwner: boolean;
}) {
  const roles = [
    { value: "staff", label: "Staff — listings and orders" },
    { value: "manager", label: "Manager — everything but owners and money" },
    { value: "finance", label: "Finance — payouts and statements" },
    ...(canInviteOwner
      ? [{ value: "owner", label: "Owner — everything" }]
      : []),
  ];
  return (
    <ActionForm
      action={inviteMember.bind(null, shopId)}
      submitLabel="Send invite"
      pendingLabel="Sending…"
      resetOnSuccess
    >
      <Field
        name="contact"
        label="Phone number or email"
        maxLength={254}
        autoComplete="off"
        hint="We text or email them a link. They accept it from the account with this phone or email."
      />
      <Field
        name="role"
        label="Role"
        type="select"
        options={roles}
        hint="Finance members must set up an authenticator app before joining."
      />
    </ActionForm>
  );
}
