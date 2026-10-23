import { notFound } from "next/navigation";
import { connection } from "next/server";

import { ApiError, returnReason, returnStatus } from "@techshop/api-client";
import { Status, humanise } from "@techshop/ui/components";

import { Details } from "@/components/Details";
import { PageHead } from "@/components/PageHead";
import {
  DisputeDecision,
  InspectItemForm,
  ReceiveReturnButton,
  RefundReturnButton,
  ReturnDecision,
} from "@/features/aftersales/components/Controls";
import { locations, returnDetail } from "@/features/aftersales/server";
import { requireStaff } from "@/features/session/server";
import { formatDateTime, formatKobo } from "@/lib/format";

export const metadata = { title: "Return | TechShop Staff" };

export default async function ReturnPage(props: PageProps<"/returns/[rma]">) {
  await connection();
  await requireStaff("returns.manage");
  const { rma } = await props.params;
  const r = await returnDetail(rma).catch((e) => {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      notFound();
    throw e;
  });
  const places = ((await locations().catch(() => [])) ?? []).map((l) => ({
    value: l.id,
    label: `${l.name}, ${l.city}`,
  }));
  const status = returnStatus[r.status] ?? {
    label: r.status,
    tone: "neutral" as const,
  };
  const items = r.items ?? [];
  return (
    <div className="stack stack--lg">
      <PageHead
        title={`Return ${r.rma}`}
        description={<Status label={status.label} tone={status.tone} />}
      />
      <section className="card stack" aria-labelledby="ret-sum">
        <h2 id="ret-sum" className="text-xl">
          Request
        </h2>
        <Details
          items={[
            ["Order", r.orderNumber],
            ["Reason", returnReason[r.reason] ?? r.reason],
            ["Customer says", r.details],
            [
              "Comes back by",
              r.method === "pickup" ? "Rider pickup" : "Store drop-off",
            ],
            [
              "Pickup job",
              r.pickup
                ? humanise(r.pickup.status)
                : r.method === "pickup"
                  ? "None (outside delivery zones: arrange a courier)"
                  : null,
            ],
            ["Auto-approved", r.autoApproved ? "Yes" : "No"],
            ["Rejected because", r.rejectionReason],
            ["Refund", formatKobo(r.refund)],
            ["Requested", formatDateTime(r.createdAt)],
          ]}
        />
        {r.status === "requested" ? <ReturnDecision rma={r.rma} /> : null}
        {r.status === "approved" || r.status === "awaiting_pickup" ? (
          <ReceiveReturnButton rma={r.rma} />
        ) : null}
        {r.status === "inspected" ? <RefundReturnButton rma={r.rma} /> : null}
        {r.status === "received" && r.reason === "change_of_mind" ? (
          <ReturnDecision rma={r.rma} />
        ) : null}
      </section>
      <section className="card stack" aria-labelledby="ret-items">
        <h2 id="ret-items" className="text-xl">
          Items
        </h2>
        <ul role="list" className="stack stack--lg">
          {items.map((i) => (
            <li key={i.id} className="stack">
              <p>
                <strong>
                  {i.quantity} × {i.productName}
                </strong>{" "}
                <span className="text-sm text-muted">
                  {i.variantName} · {formatKobo(i.refund)}
                  {i.serialised ? " · tracked device" : ""}
                </span>
              </p>
              {i.grade ? (
                <p className="text-sm">
                  Grade {i.grade} · {humanise(i.disposition ?? "")}
                </p>
              ) : r.status === "received" ? (
                <InspectItemForm
                  rma={r.rma}
                  itemId={i.id}
                  serialised={i.serialised}
                  sellerItem={i.sellerShipped}
                  locations={places}
                />
              ) : null}
            </li>
          ))}
        </ul>
        <p className="text-xs text-muted">
          Seller-shipped items can only go back to the seller; the API refuses
          any other outcome.
        </p>
      </section>
      {r.dispute ? (
        <section className="card stack" aria-labelledby="ret-dispute">
          <h2 id="ret-dispute" className="text-xl">
            Seller dispute
          </h2>
          <Details
            items={[
              ["Seller says", r.dispute.reason],
              ["Raised", formatDateTime(r.dispute.at)],
              [
                "Decision",
                r.dispute.decision ? humanise(r.dispute.decision) : null,
              ],
              ["Note", r.dispute.note],
            ]}
          />
          {!r.dispute.decision ? <DisputeDecision rma={r.rma} /> : null}
        </section>
      ) : null}
    </div>
  );
}
