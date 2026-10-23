import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { formatKobo, returnReason, returnStatus } from "@techshop/api-client";
import { PageHeader, Status } from "@techshop/ui/components";

import styles from "@/features/aftersales/components/Aftersales.module.css";
import { Facts } from "@/features/aftersales/components/Records";
import { myReturn } from "@/features/aftersales/server";
import { formatDateTime } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Return | TechShop Market",
  robots: { index: false },
};

const pickupWords: Record<string, string> = {
  unassigned: "Waiting for a rider",
  assigned: "A rider is on the way",
  picked_up: "Collected by the rider",
  en_route: "On its way to us",
  delivered: "Handed in at our hub",
  failed: "The rider couldn't collect it; we'll try again",
};

export default async function ReturnPage(
  props: PageProps<"/account/returns/[rma]">,
) {
  await connection();
  const { rma } = await props.params;
  await bff.requireSession(`/account/returns/${rma}`);
  const r = await myReturn(rma);
  if (!r) notFound();
  const isNew = (await props.searchParams).new === "1";
  const status = returnStatus[r.status] ?? {
    label: r.status,
    tone: "neutral" as const,
  };

  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title={`Return ${r.rma}`}
          lead={<Status label={status.label} tone={status.tone} />}
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Returns", href: "/account/returns" },
            { label: r.rma },
          ]}
        />
        {isNew ? (
          <p className="alert alert--success" role="status">
            {r.autoApproved
              ? "Your return is approved."
              : "We've got your request. Support will check it, usually within a working day."}
            {r.status === "approved"
              ? " Bring the item and its box to any TechShop store and quote " +
                r.rma +
                "."
              : ""}
            {r.status === "awaiting_pickup"
              ? r.pickup
                ? " A rider will collect it from your delivery address; keep it packed."
                : " We'll arrange a courier and tell you when it's booked."
              : ""}
          </p>
        ) : null}
        {r.status === "rejected" && r.rejectionReason ? (
          <p className="alert alert--error" role="status">
            Not accepted: {r.rejectionReason}{" "}
            <Link
              href={`/account/support/new?order=${r.orderNumber}`}
              className="link"
            >
              Contact support
            </Link>
          </p>
        ) : null}
        <section className="card card--roomy stack" aria-labelledby="ret-items">
          <h2 id="ret-items" className={styles.h2}>
            Items
          </h2>
          <ul role="list" className={styles.list}>
            {(r.items ?? []).map((i) => (
              <li key={i.id} className={styles.row}>
                <span>
                  {i.quantity} × {i.productName}
                  <span className="text-xs text-muted"> · {i.variantName}</span>
                </span>
                <span className="tabular-nums">{formatKobo(i.refund)}</span>
              </li>
            ))}
          </ul>
        </section>
        <section
          className="card card--roomy stack"
          aria-labelledby="ret-details"
        >
          <h2 id="ret-details" className={styles.h2}>
            Details
          </h2>
          <Facts
            items={[
              [
                "Order",
                <Link
                  key="o"
                  href={`/orders/${r.orderNumber}`}
                  className="link"
                >
                  {r.orderNumber}
                </Link>,
              ],
              ["Reason", returnReason[r.reason] ?? r.reason],
              ["What happened", r.details],
              [
                "How it comes back",
                r.method === "pickup"
                  ? "Collected from your address"
                  : "Brought to a TechShop store",
              ],
              [
                "Collection",
                r.pickup
                  ? (pickupWords[r.pickup.status] ?? r.pickup.status)
                  : null,
              ],
              [
                "Refund",
                <span key="f" className="tabular-nums">
                  {formatKobo(r.refund)}
                  {r.status === "refunded"
                    ? " (on its way to your original payment method)"
                    : " once we've checked it"}
                </span>,
              ],
              ["Requested", formatDateTime(r.createdAt)],
            ]}
          />
        </section>
      </div>
    </main>
  );
}
