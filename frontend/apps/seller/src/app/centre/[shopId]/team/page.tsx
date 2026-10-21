import type { Metadata } from "next";
import { connection } from "next/server";

import { Status, Table } from "@techshop/ui/components";

import { RowAction } from "@/components/RowAction";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import {
  removeMember,
  resendInvite,
  revokeInvite,
} from "@/features/team/actions";
import { InviteForm } from "@/features/team/components/InviteForm";
import { RoleForm } from "@/features/team/components/RoleForm";
import { dateOnly } from "@/lib/format";

export const metadata: Metadata = { title: "Team | TechShop Seller Centre" };

const roleLabel: Record<string, string> = {
  owner: "Owner",
  manager: "Manager",
  staff: "Staff",
  finance: "Finance",
};

export default async function TeamPage(
  props: PageProps<"/centre/[shopId]/team">,
) {
  await connection();
  const { shopId } = await props.params;
  const here = `/centre/${shopId}/team`;
  const shop = await requireShop(shopId, here);
  const manages = shop.role === "owner" || shop.role === "manager";
  if (!manages)
    return (
      <ShopPage shop={shop} section="team" title="Team">
        <p className="text-sm text-muted">
          Only owners and managers can see and change the team. You’re{" "}
          {roleLabel[shop.role ?? ""] ?? shop.role}.
        </p>
      </ShopPage>
    );
  const api = await sellerApi(here);
  const team = await api.get("/api/v1/seller/shops/{shopId}/team", {
    path: { shopId },
    init: { cache: "no-store" },
  });
  const isOwner = shop.role === "owner";
  return (
    <ShopPage
      shop={shop}
      section="team"
      title="Team"
      lead="People who help run your shop, and what each can do."
    >
      <Table
        caption="Team members"
        rows={team.members ?? []}
        rowKey={(m) => m.userId}
        columns={[
          { label: "Name", cell: (m) => `${m.firstName} ${m.lastName}` },
          { label: "Contact", cell: (m) => m.contact },
          {
            label: "Role",
            cell: (m) =>
              m.role === "owner" && !isOwner ? (
                roleLabel.owner
              ) : (
                <RoleForm
                  shopId={shop.id}
                  userId={m.userId}
                  role={m.role}
                  name={m.firstName}
                  canGrantOwner={isOwner}
                />
              ),
          },
          {
            label: "Authenticator",
            cell: (m) =>
              m.hasAuthenticator ? (
                <Status label="On" tone="success" />
              ) : (
                <Status
                  label={m.role === "finance" ? "Needed" : "Off"}
                  tone={m.role === "finance" ? "danger" : "neutral"}
                />
              ),
          },
          {
            label: "Joined",
            cell: (m) => dateOnly.format(new Date(m.joinedAt)),
          },
          {
            label: "Remove",
            cell: (m) =>
              m.role === "owner" && !isOwner ? null : (
                <RowAction
                  action={removeMember.bind(null, shop.id, m.userId)}
                  label="Remove"
                  pendingLabel="Removing…"
                />
              ),
          },
        ]}
      />
      <section aria-labelledby="invites-title" className="stack">
        <h2 id="invites-title" className="text-xl">
          Invites waiting
        </h2>
        <Table
          caption="Open invites"
          rows={team.invites ?? []}
          rowKey={(i) => i.id}
          empty={{
            title: "No open invites",
            body: "Invite someone below. Each link works once, for 7 days.",
          }}
          columns={[
            { label: "Sent to", cell: (i) => i.contact },
            { label: "Role", cell: (i) => roleLabel[i.role] ?? i.role },
            { label: "Invited by", cell: (i) => i.invitedBy },
            {
              label: "Link",
              cell: (i) =>
                i.expired ? (
                  <Status label="Expired" tone="warning" />
                ) : (
                  `Works until ${dateOnly.format(new Date(i.expiresAt))}`
                ),
            },
            {
              label: "Actions",
              cell: (i) =>
                i.role === "owner" && !isOwner ? null : (
                  <div className="cluster">
                    <RowAction
                      action={resendInvite.bind(null, shop.id, i.id)}
                      label="Send a new link"
                      pendingLabel="Sending…"
                    />
                    <RowAction
                      action={revokeInvite.bind(null, shop.id, i.id)}
                      label="Cancel"
                      pendingLabel="Cancelling…"
                    />
                  </div>
                ),
            },
          ]}
        />
      </section>
      <section
        aria-labelledby="invite-title"
        className="card card--roomy stack"
      >
        <h2 id="invite-title" className="text-xl">
          Invite someone
        </h2>
        <InviteForm shopId={shop.id} canInviteOwner={isOwner} />
      </section>
    </ShopPage>
  );
}
