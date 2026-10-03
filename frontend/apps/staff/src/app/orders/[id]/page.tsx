import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { formatMoney } from "@techshop/api-client";
import { getProductById, sampleOrders } from "@techshop/fixtures";
import { Breadcrumbs, CategoryArt, Icon } from "@techshop/ui/components";

import styles from "./order.module.css";

export const dynamicParams = false;

export function generateStaticParams() {
  return sampleOrders.map((o) => ({ id: o.id }));
}

export async function generateMetadata(
  props: PageProps<"/orders/[id]">,
): Promise<Metadata> {
  return { title: `${(await props.params).id} | TechShop Staff` };
}

const timeline = [
  "Order placed",
  "Payment confirmed",
  "Packed",
  "Out for delivery",
  "Delivered",
];
const reached: Record<string, number> = {
  "Pending payment": 0,
  Paid: 1,
  Packed: 2,
  "Out for delivery": 3,
  Delivered: 4,
  "Refund requested": 4,
};

export default async function OrderDetail(props: PageProps<"/orders/[id]">) {
  const { id } = await props.params;
  const order = sampleOrders.find((o) => o.id === id);
  if (!order) notFound();
  const step = reached[order.status];
  const lines = order.lines.map((l) => ({
    ...l,
    product: getProductById(l.productId)!,
  }));

  return (
    <div className="stack stack--lg">
      <Breadcrumbs
        items={[
          { label: "Dashboard", href: "/" },
          { label: "Orders", href: "/orders" },
          { label: order.id },
        ]}
      />
      <div className={styles.head}>
        <div className="stack stack--sm">
          <h1 className={styles.title}>Order {order.id}</h1>
          <p className="text-sm text-muted">
            {order.customer} · {order.channel} · placed {order.placedAt} ·
            Sample data
          </p>
        </div>
        <div className="cluster">
          <Link
            href={`/dispatch?q=${order.id}`}
            className="btn btn--secondary btn--sm"
          >
            <Icon name="truck" size={16} />
            Dispatch
          </Link>
          <Link
            href={`/support?q=${encodeURIComponent(order.customer)}`}
            className="btn btn--secondary btn--sm"
          >
            <Icon name="support" size={16} />
            Support tickets
          </Link>
        </div>
      </div>

      <div className={styles.layout}>
        <div className="stack">
          <section className="card" aria-labelledby="items-title">
            <h2 id="items-title" className={styles.h2}>
              Items
            </h2>
            <ul role="list" className={styles.lines}>
              {lines.map((l) => (
                <li key={l.product.id} className={styles.line}>
                  <div className={`frame ${styles.thumb}`}>
                    <CategoryArt category={l.product.category} size="sm" />
                  </div>
                  <div className="stack stack--sm">
                    <p className={styles.name}>{l.product.name}</p>
                    <p className="text-xs text-muted">
                      {l.quantity} × {formatMoney(l.unitPrice)}
                      {order.channel === "Wholesale" ? " (trade price)" : ""} ·
                      Sold by {l.product.seller.name}
                    </p>
                  </div>
                </li>
              ))}
            </ul>
            <p className={styles.total}>
              <span>Order total</span>
              <span className="tabular-nums">{formatMoney(order.total)}</span>
            </p>
          </section>
        </div>

        <section className="card" aria-labelledby="progress-title">
          <h2 id="progress-title" className={styles.h2}>
            Progress
          </h2>
          <ol className={styles.timeline}>
            {timeline.map((t, i) => (
              <li key={t} className={styles.stage} data-done={i <= step}>
                <span className={styles.dot} aria-hidden="true">
                  {i <= step ? <Icon name="check" size={12} /> : null}
                </span>
                <span>
                  {t}
                  <span className="visually-hidden">
                    {i <= step ? " — done" : " — not yet"}
                  </span>
                </span>
              </li>
            ))}
          </ol>
          {order.status === "Refund requested" ? (
            <p className={styles.alert}>
              <Icon name="returns" size={16} /> Refund requested by the customer
              — review in Orders → Refunds.
            </p>
          ) : null}
        </section>
      </div>
    </div>
  );
}
