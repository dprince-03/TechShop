import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ApiError } from "@techshop/api-client";
import { Status, Table, humanise } from "@techshop/ui/components";

import { PageHead } from "@/components/PageHead";
import { can, requireStaff } from "@/features/session/server";
import { KycDecision } from "@/features/vendors/components/KycDecision";
import { RevealId } from "@/features/vendors/components/RevealId";
import { vendor } from "@/features/vendors/server";
import { formatDateTime } from "@/lib/format";

export const metadata = { title: "Vendor | TechShop Staff" };

const match = {
  match: { label: "Matches the holder", tone: "success" },
  partial: { label: "Partly matches", tone: "warning" },
  mismatch: { label: "Doesn’t match", tone: "danger" },
  unknown: { label: "Not checked", tone: "neutral" },
} as const;

export default async function VendorPage(props: PageProps<"/vendors/[id]">) {
  await connection();
  const me = await requireStaff("vendors.view");
  const { id } = await props.params;
  const v = await vendor(id).catch((e) => {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      notFound();
    throw e;
  });
  const k = v.kyc;
  const reviewable =
    k && (k.status === "submitted" || k.status === "in_review");
  return (
    <div className="stack stack--lg">
      <PageHead
        title={v.shop.displayName}
        description={`${humanise(v.shop.type)} seller · ${humanise(v.shop.status)} · ${[v.shop.city, v.shop.stateCode].filter(Boolean).join(", ")}`}
      />

      <section aria-labelledby="kyc-title" className="card stack">
        <h2 id="kyc-title" className="text-xl">
          Verification
        </h2>
        {!k ? (
          <p className="text-muted">Not submitted yet.</p>
        ) : (
          <>
            <dl className="stack stack--sm text-sm">
              <div className="cluster">
                <dt className="text-muted">Status</dt>
                <dd>
                  <Status
                    label={humanise(k.status)}
                    tone={
                      k.status === "approved"
                        ? "success"
                        : k.status === "rejected"
                          ? "danger"
                          : "info"
                    }
                  />
                </dd>
              </div>
              <div className="cluster">
                <dt className="text-muted">ID</dt>
                <dd>
                  {humanise(k.idType)} ending {k.idNumberLast4}
                </dd>
              </div>
              {k.cacRcNumber ? (
                <div className="cluster">
                  <dt className="text-muted">CAC number</dt>
                  <dd>{k.cacRcNumber}</dd>
                </div>
              ) : null}
              <div className="cluster">
                <dt className="text-muted">Submitted</dt>
                <dd>{formatDateTime(k.submittedAt)}</dd>
              </div>
              <div className="cluster">
                <dt className="text-muted">Name on bank accounts must match</dt>
                <dd>{v.legalHolder}</dd>
              </div>
            </dl>
            {v.otherSellersWithIdNumber > 0 ? (
              <p className="alert alert--warning">
                This ID number was also submitted by{" "}
                {v.otherSellersWithIdNumber} other shop
                {v.otherSellersWithIdNumber === 1 ? "" : "s"}. Check for a
                banned seller returning.
              </p>
            ) : null}
            {k.note ? (
              <p className="alert alert--warning">Previous note: {k.note}</p>
            ) : null}
            <div className="stack stack--sm">
              <h3 className="text-lg">Documents</h3>
              <ul role="list" className="cluster">
                {(k.documents ?? []).map((d) => (
                  <li key={d.fileId}>
                    <a
                      href={`/vendors/${v.shop.id}/documents/${d.fileId}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="btn btn--secondary btn--sm"
                    >
                      {humanise(d.kind)}
                    </a>
                  </li>
                ))}
              </ul>
              <p className="text-xs text-muted">
                Each opening is logged. Links last 5 minutes.
              </p>
            </div>
            {reviewable && can(me, "vendors.kyc_review") ? (
              <KycDecision submissionId={k.id} />
            ) : null}
            {can(me, "vendors.reveal_id") ? (
              <RevealId submissionId={k.id} />
            ) : null}
          </>
        )}
      </section>

      <section aria-labelledby="bank-title" className="stack">
        <h2 id="bank-title" className="text-xl">
          Payout accounts
        </h2>
        <Table
          caption="Payout accounts"
          rows={v.bankAccounts ?? []}
          rowKey={(b) => b.id}
          empty={{
            title: "No payout account",
            body: "The seller hasn’t added one. Approval needs a verified account.",
          }}
          columns={[
            {
              label: "Account",
              cell: (b) => `${b.bankCode} ••••${b.accountNumberLast4}`,
            },
            { label: "Typed name", cell: (b) => b.accountName },
            { label: "Bank’s name", cell: (b) => b.resolvedAccountName ?? "—" },
            {
              label: "Check",
              cell: (b) => (
                <Status
                  label={match[b.nameMatch].label}
                  tone={match[b.nameMatch].tone}
                />
              ),
            },
            {
              label: "Also used by",
              numeric: true,
              cell: (b) =>
                b.otherSellersWithAccount
                  ? `${b.otherSellersWithAccount} other shop(s)`
                  : "—",
            },
            { label: "Default", cell: (b) => (b.isDefault ? "Yes" : "") },
          ]}
        />
      </section>

      <section aria-labelledby="members-title" className="stack">
        <h2 id="members-title" className="text-xl">
          People
        </h2>
        <Table
          caption="Shop members"
          rows={v.members ?? []}
          rowKey={(m) => m.userId}
          columns={[
            { label: "Name", cell: (m) => m.name },
            { label: "Role", cell: (m) => humanise(m.role) },
            { label: "Since", cell: (m) => formatDateTime(m.createdAt) },
          ]}
        />
      </section>
    </div>
  );
}
