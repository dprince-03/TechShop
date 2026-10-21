import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { formatKobo } from "@techshop/api-client";
import { PageHeader } from "@techshop/ui/components";

import { paymentMethods } from "@/features/checkout/server";
import { OrderActions } from "@/features/orders/components/OrderActions";
import styles from "@/features/orders/components/Orders.module.css";
import { StatusBadge } from "@/features/orders/components/StatusBadge";
import { getMyOrder } from "@/features/orders/server";
import { RateSellerForm } from "@/features/reviews/components/ReviewForms";
import {
  formatDateTime,
  orderStatus,
  parcelStatus,
  paymentStatus,
  providerLabel,
} from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Order | TechShop Market",
  robots: { index: false },
};

export default async function OrderPage(props: PageProps<"/orders/[number]">) {
  await connection();
  const { number } = await props.params;
  await bff.requireSession(`/orders/${number}`);
  const [order, methods] = await Promise.all([
    getMyOrder(number),
    paymentMethods(),
  ]);
  if (!order) notFound();
  const placed = (await props.searchParams).placed === "1";
  const awaiting = order.status === "pending_payment";
  const items = (order.fulfilments ?? []).flatMap((f) => f.lines ?? []);
  const itemsKobo = items.reduce((t, l) => t + l.lineTotalKobo, 0);

  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title={`Order ${order.orderNumber}`}
          lead={
            <>
              Placed {formatDateTime(order.placedAt)} ·{" "}
              <StatusBadge status={order.status} />
            </>
          }
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Orders", href: "/orders" },
            { label: order.orderNumber },
          ]}
          actions={
            <Link
              href={`/orders/${order.orderNumber}`}
              className="btn btn--ghost btn--sm"
            >
              Refresh status
            </Link>
          }
        />
        {placed && awaiting ? (
          <p className="alert alert--warning" role="status">
            We’re waiting for your payment to be confirmed. This page updates
            when you refresh it.
            {order.paymentDueAt
              ? ` Pay by ${formatDateTime(order.paymentDueAt)} or the items are released.`
              : ""}
          </p>
        ) : null}
        {placed && order.status === "paid" ? (
          <p className="alert alert--success" role="status">
            Thank you! Your payment is confirmed and your order is being
            prepared.
          </p>
        ) : null}

        <div className={styles.layout}>
          <div className="stack stack--lg">
            {(order.fulfilments ?? []).map((f) => (
              <section
                key={f.id}
                className="card card--roomy stack"
                aria-label={`Parcel from ${f.sellerName}`}
              >
                <div className="cluster cluster--between">
                  <h2 className={styles.h2}>From {f.sellerName}</h2>
                  <span className="text-sm">
                    {parcelStatus[f.status] ?? f.status}
                    {f.deliveredAt ? ` · ${formatDateTime(f.deliveredAt)}` : ""}
                  </span>
                </div>
                <ul role="list" className={styles.lines}>
                  {(f.lines ?? []).map((l) => (
                    <li key={l.id} className={styles.line}>
                      <span>
                        {l.quantity} × {l.productName}
                        <span className="text-xs text-muted">
                          {" "}
                          · {l.variantName}
                        </span>
                        {l.cancelledAt ? (
                          <span className="badge">Cancelled</span>
                        ) : null}
                      </span>
                      <span className="cluster">
                        {f.status === "delivered" && !l.cancelledAt ? (
                          <Link
                            href={`/orders/${order.orderNumber}/review/${l.id}`}
                            className="link text-sm"
                          >
                            Review
                          </Link>
                        ) : null}
                        <span className="tabular-nums">
                          {formatKobo(l.lineTotalKobo)}
                        </span>
                      </span>
                    </li>
                  ))}
                </ul>
                <p className="text-xs text-muted">
                  Delivery {formatKobo(f.deliveryFeeKobo)} ·{" "}
                  {f.fulfilledBy === "techshop"
                    ? "Shipped by TechShop"
                    : "Shipped by the seller"}
                </p>
                {f.status === "delivered" ? (
                  <details>
                    <summary className="link text-sm">
                      Rate {f.sellerName}
                    </summary>
                    <RateSellerForm fulfilmentId={f.id} seller={f.sellerName} />
                  </details>
                ) : null}
              </section>
            ))}

            {(order.history ?? []).length ? (
              <section
                className="card card--roomy stack"
                aria-labelledby="history-title"
              >
                <h2 id="history-title" className={styles.h2}>
                  History
                </h2>
                <ol className={styles.timeline}>
                  {order.history!.map((h) => (
                    <li key={h.id}>
                      <span>
                        {orderStatus[h.toStatus]?.label ??
                          parcelStatus[h.toStatus] ??
                          h.toStatus}
                      </span>
                      <span className="text-xs text-muted">
                        {formatDateTime(h.createdAt)}
                      </span>
                    </li>
                  ))}
                </ol>
              </section>
            ) : null}
          </div>

          <aside
            className="card card--roomy stack"
            aria-labelledby="summary-title"
          >
            <h2 id="summary-title" className={styles.h2}>
              Summary
            </h2>
            <dl className={styles.totals}>
              <div>
                <dt>Items</dt>
                <dd>{formatKobo(itemsKobo)}</dd>
              </div>
              {order.discountKobo > 0 ? (
                <div>
                  <dt>Discounts</dt>
                  <dd>−{formatKobo(order.discountKobo)}</dd>
                </div>
              ) : null}
              <div>
                <dt>Delivery</dt>
                <dd>{formatKobo(order.deliveryFeeKobo)}</dd>
              </div>
              <div>
                <dt>VAT included</dt>
                <dd>{formatKobo(order.vatKobo)}</dd>
              </div>
              <div className={styles.grand}>
                <dt>Total</dt>
                <dd>{formatKobo(order.totalKobo)}</dd>
              </div>
            </dl>
            {(order.payments ?? []).length ? (
              <div className="stack stack--sm">
                <h3 className="text-sm">Payments</h3>
                <ul role="list" className="stack stack--sm text-sm">
                  {order.payments!.map((p) => (
                    <li key={p.id}>
                      {providerLabel[p.provider] ?? p.provider}:{" "}
                      {paymentStatus[p.status] ?? p.status} ·{" "}
                      {formatKobo(p.amountKobo)}
                    </li>
                  ))}
                </ul>
              </div>
            ) : null}
            {(order.refunds ?? []).length ? (
              <p className="text-sm">
                Refunds:{" "}
                {order
                  .refunds!.map(
                    (r) => `${formatKobo(r.amountKobo)} (${r.status})`,
                  )
                  .join(", ")}
              </p>
            ) : null}
            <OrderActions
              number={order.orderNumber}
              methods={methods}
              canPay={awaiting}
              canCancel={awaiting}
            />
          </aside>
        </div>
      </div>
    </main>
  );
}
