import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import {
  ApiError,
  formatKobo,
  returnReason,
  returnStatus,
} from "@techshop/api-client";
import { Status, humanise } from "@techshop/ui/components";

import { DisputeForm } from "@/features/aftersales/components/Forms";
import { requireShop, sellerApi } from "@/features/session/server";
import { ShopPage } from "@/features/shops/components/ShopPage";
import { dateTime } from "@/lib/format";

export const metadata: Metadata = { title: "Return | TechShop Seller Centre" };

const grades: Record<string, string> = {
  A: "A (as new)",
  B: "B (light wear)",
  C: "C (heavy wear)",
  faulty: "Faulty",
};

export default async function ReturnPage(
  props: PageProps<"/centre/[shopId]/returns/[rma]">,
) {
  await connection();
  const { shopId, rma } = await props.params;
  const here = `/centre/${shopId}/returns/${rma}`;
  const shop = await requireShop(shopId, here);
  const api = await sellerApi(here);
  const r = await api
    .get("/api/v1/seller/shops/{shopId}/returns/{rma}", {
      path: { shopId, rma },
      init: { cache: "no-store" },
    })
    .catch((e) => {
      if (e instanceof ApiError && (e.status === 404 || e.status === 422))
        notFound();
      throw e;
    });
  const s = returnStatus[r.status] ?? {
    label: r.status,
    tone: "neutral" as const,
  };
  const canAct = shop.role === "owner" || shop.role === "manager";
  return (
    <ShopPage shop={shop} section="returns" title={`Return ${r.rma}`}>
      <section className="card stack" aria-labelledby="sr-sum">
        <h2 id="sr-sum" className="text-xl">
          <Status label={s.label} tone={s.tone} />
        </h2>
        <dl className="stack stack--sm text-sm">
          <div className="cluster">
            <dt className="text-muted">Order</dt>
            <dd>{r.orderNumber}</dd>
          </div>
          <div className="cluster">
            <dt className="text-muted">Customer’s reason</dt>
            <dd>{returnReason[r.reason] ?? r.reason}</dd>
          </div>
          {r.details ? (
            <div className="cluster">
              <dt className="text-muted">Customer says</dt>
              <dd>{r.details}</dd>
            </div>
          ) : null}
          <div className="cluster">
            <dt className="text-muted">Charged to your payable</dt>
            <dd className="tabular-nums">
              {formatKobo(r.charge)} (your commission on it is given back)
            </dd>
          </div>
          <div className="cluster">
            <dt className="text-muted">Requested</dt>
            <dd>{dateTime.format(new Date(r.createdAt))}</dd>
          </div>
        </dl>
      </section>
      <section className="card stack" aria-labelledby="sr-items">
        <h2 id="sr-items" className="text-xl">
          Your items
        </h2>
        <ul role="list" className="stack stack--sm text-sm">
          {(r.items ?? []).map((i) => (
            <li key={i.id}>
              {i.quantity} × {i.productName} · {i.variantName}
              {i.grade
                ? ` · inspected: ${grades[i.grade] ?? i.grade}, ${humanise(i.disposition ?? "")}`
                : " · not inspected yet"}
            </li>
          ))}
        </ul>
        <p className="text-xs text-muted">
          Items marked “return to seller” wait for you at the hub you drop
          parcels at.
        </p>
      </section>
      {r.dispute ? (
        <section className="card stack" aria-labelledby="sr-dispute">
          <h2 id="sr-dispute" className="text-xl">
            Your dispute
          </h2>
          <p className="text-sm">{r.dispute.reason}</p>
          <p className="text-sm text-muted">
            {r.dispute.decision
              ? `${r.dispute.decision === "upheld" ? "Upheld: the charge was given back to you." : "Not upheld."}${r.dispute.note ? ` ${r.dispute.note}` : ""}`
              : "Waiting for a decision."}
          </p>
        </section>
      ) : r.canDispute && canAct ? (
        <section className="card stack" aria-labelledby="sr-new-dispute">
          <h2 id="sr-new-dispute" className="text-xl">
            Disagree with the inspection?
          </h2>
          {r.disputeUntil ? (
            <p className="text-sm text-muted">
              You can dispute until {dateTime.format(new Date(r.disputeUntil))}.
            </p>
          ) : null}
          <DisputeForm shopId={shop.id} rma={r.rma} />
        </section>
      ) : null}
    </ShopPage>
  );
}
