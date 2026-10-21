import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { ApiError } from "@techshop/api-client";
import { PageHeader } from "@techshop/ui/components";

import { RowAction } from "@/components/RowAction";
import { sellerApi } from "@/features/session/server";
import { acceptInvite } from "@/features/team/actions";
import { dateOnly } from "@/lib/format";

export const metadata: Metadata = {
  title: "Join a shop | TechShop Seller Centre",
};

const roleText: Record<string, string> = {
  owner: "an owner (everything)",
  manager: "a manager (listings, orders and the team)",
  staff: "staff (listings and orders)",
  finance: "finance (payouts and statements)",
};

export default async function InvitePage(props: PageProps<"/invite/[token]">) {
  await connection();
  const { token } = await props.params;
  const here = `/invite/${token}`;
  const api = await sellerApi(here);
  let preview;
  let problem = "";
  try {
    preview = await api.post("/api/v1/seller/invites/preview", {
      body: { token },
    });
  } catch (e) {
    problem =
      e instanceof ApiError && e.detail
        ? e.detail
        : "We couldn’t open this invite. Try the link again.";
  }
  const me = preview
    ? await api.get("/api/v1/me", { init: { cache: "no-store" } })
    : null;
  const needsAuthenticator = preview?.role === "finance" && !me?.mfaEnabled;
  return (
    <main className="section section--page">
      <div className="container container--narrow stack">
        <PageHeader title="Join a shop" />
        {preview ? (
          <section className="card card--roomy stack">
            <p>
              You’re invited to help run <strong>{preview.shop}</strong> as{" "}
              {roleText[preview.role] ?? preview.role}.
            </p>
            <p className="text-sm text-muted">
              Sent to {preview.sentTo}. Accept from that account by{" "}
              {dateOnly.format(new Date(preview.expiresAt))}.
            </p>
            {needsAuthenticator ? (
              <p className="alert alert--warning">
                Finance members need an authenticator app first.{" "}
                <Link
                  className="link"
                  href={`/account/security?next=${encodeURIComponent(here)}`}
                >
                  Set it up
                </Link>
                , then come back to this page.
              </p>
            ) : (
              <RowAction
                action={acceptInvite.bind(null, token)}
                label="Accept and join"
                pendingLabel="Joining…"
              />
            )}
          </section>
        ) : (
          <p className="alert alert--error">{problem}</p>
        )}
      </div>
    </main>
  );
}
