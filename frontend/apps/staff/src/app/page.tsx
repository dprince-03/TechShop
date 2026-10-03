import Link from "next/link";
import { connection } from "next/server";

import { formatMoney } from "@techshop/api-client";
import {
  sampleAttention,
  sampleKpis,
  sampleOrders,
  type SampleOrder,
} from "@techshop/fixtures";
import { Icon } from "@techshop/ui/components";

import styles from "./page.module.css";

const statusTone: Record<SampleOrder["status"], string> = {
  "Pending payment": styles.warning,
  Paid: styles.success,
  Packed: styles.info,
  "Out for delivery": styles.info,
  Delivered: styles.neutral,
  "Refund requested": styles.danger,
};

const quickActions = [
  { label: "Add a product", href: "/catalogue/new", icon: "plus" },
  { label: "Assign riders", href: "/dispatch", icon: "truck" },
  {
    label: "Review refunds",
    href: "/orders?tab=refund-requested",
    icon: "returns",
  },
  {
    label: "Reconcile payments",
    href: "/finance?tab=reconciliation",
    icon: "wallet",
  },
] as const;

const today = new Intl.DateTimeFormat("en-NG", {
  weekday: "long",
  day: "numeric",
  month: "long",
  year: "numeric",
  timeZone: "Africa/Lagos",
});

export default async function Dashboard() {
  // Dashboard figures are per-request (live data later), so never prerender.
  await connection();

  return (
    <div className="stack stack--lg">
      <div className={styles.pageHead}>
        <div className="stack stack--sm">
          <h1 className={styles.title}>Dashboard</h1>
          <p className="text-sm text-muted">
            {today.format(new Date())} · Sample data
          </p>
        </div>
        <ul role="list" className={`cluster ${styles.quick}`}>
          {quickActions.map((a) => (
            <li key={a.href}>
              <Link href={a.href} className="btn btn--secondary btn--sm">
                <Icon name={a.icon} size={16} />
                {a.label}
              </Link>
            </li>
          ))}
        </ul>
      </div>

      <section aria-labelledby="kpi-title">
        <h2 id="kpi-title" className="visually-hidden">
          Key figures
        </h2>
        <ul role="list" className={styles.kpis}>
          {sampleKpis.map((k) => (
            <li key={k.label} className={`card ${styles.kpi}`}>
              <p className="text-sm text-muted">{k.label}</p>
              <p className={styles.kpiValue}>
                {typeof k.value === "number"
                  ? k.value.toLocaleString("en-NG")
                  : formatMoney(k.value)}
              </p>
              <p className="text-xs text-muted">{k.change}</p>
            </li>
          ))}
        </ul>
      </section>

      <div className={styles.columns}>
        <section
          className={`card ${styles.panel}`}
          aria-labelledby="orders-title"
        >
          <div className={styles.panelHead}>
            <h2 id="orders-title" className={styles.panelTitle}>
              Recent orders
            </h2>
            <Link href="/orders" className="link link--arrow text-sm">
              All orders
            </Link>
          </div>
          <div
            className={styles.tableWrap}
            role="region"
            aria-labelledby="orders-title"
            tabIndex={0}
          >
            <table className={styles.table}>
              <thead>
                <tr>
                  <th scope="col">Order</th>
                  <th scope="col">Customer</th>
                  <th scope="col">Channel</th>
                  <th scope="col" className={styles.num}>
                    Items
                  </th>
                  <th scope="col" className={styles.num}>
                    Total
                  </th>
                  <th scope="col">Status</th>
                </tr>
              </thead>
              <tbody>
                {sampleOrders.map((o) => (
                  <tr key={o.id}>
                    <th scope="row">
                      <Link href={`/orders/${o.id}`} className="link">
                        {o.id}
                      </Link>
                      <span className={styles.placed}>{o.placedAt}</span>
                    </th>
                    <td>{o.customer}</td>
                    <td className="text-muted">{o.channel}</td>
                    <td className={styles.num}>{o.items}</td>
                    <td className={styles.num}>{formatMoney(o.total)}</td>
                    <td>
                      <span
                        className={`${styles.status} ${statusTone[o.status]}`}
                      >
                        {o.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <section
          className={`card ${styles.panel}`}
          aria-labelledby="attention-title"
        >
          <div className={styles.panelHead}>
            <h2 id="attention-title" className={styles.panelTitle}>
              Needs attention
            </h2>
          </div>
          <ul role="list" className={styles.attention}>
            {sampleAttention.map((a) => (
              <li key={a.label}>
                <Link
                  href={`/${a.module.toLowerCase()}`}
                  className={styles.attentionItem}
                >
                  <span className={styles.count}>{a.count}</span>
                  <span className="stack stack--sm">
                    <span className="text-sm">{a.label}</span>
                    <span className="text-xs text-muted">{a.module}</span>
                  </span>
                  <Icon
                    name="chevronRight"
                    size={16}
                    className={styles.chevron}
                  />
                </Link>
              </li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  );
}
