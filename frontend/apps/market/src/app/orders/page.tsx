import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { formatKobo } from "@techshop/api-client";
import { EmptyState, PageHeader } from "@techshop/ui/components";

import { StatusBadge } from "@/features/orders/components/StatusBadge";
import { TrackForm } from "@/features/orders/components/TrackForm";
import styles from "@/features/orders/components/Orders.module.css";
import { listMyOrders } from "@/features/orders/server";
import { formatDate } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Orders | TechShop Market",
  robots: { index: false },
};

export default async function OrdersPage(props: PageProps<"/orders">) {
  await connection();
  const session = await bff.getSession();
  const cursor = (await props.searchParams).cursor;

  if (!session) {
    return (
      <main id="main" className="section section--page">
        <div className="container container--narrow stack stack--lg">
          <PageHeader
            title="Track an order"
            lead="Enter your order number and the phone number used at checkout. No account needed."
            breadcrumbs={[
              { label: "Home", href: "/" },
              { label: "Track an order" },
            ]}
            actions={
              <Link
                href="/sign-in?next=%2Forders"
                className="btn btn--secondary"
              >
                Sign in to see all orders
              </Link>
            }
          />
          <div className="card card--roomy">
            <TrackForm />
          </div>
        </div>
      </main>
    );
  }

  const page = await listMyOrders(
    typeof cursor === "string" ? cursor : undefined,
  ).catch(() => null);
  const orders = page?.items ?? [];
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Your orders"
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Account", href: "/account" },
            { label: "Orders" },
          ]}
        />
        {page === null ? (
          <p className="alert alert--error" role="alert">
            We couldn’t load your orders. Please try again.
          </p>
        ) : orders.length ? (
          <ul role="list" className={`card ${styles.list}`}>
            {orders.map((o) => (
              <li key={o.id} className={styles.row}>
                <div className="stack stack--sm">
                  <Link
                    href={`/orders/${o.orderNumber}`}
                    className={styles.number}
                  >
                    {o.orderNumber}
                  </Link>
                  <span className="text-xs text-muted">
                    Placed {formatDate(o.placedAt)}
                  </span>
                </div>
                <div
                  className="stack stack--sm"
                  style={{ alignItems: "flex-end" }}
                >
                  <span className="tabular-nums">
                    {formatKobo(o.totalKobo)}
                  </span>
                  <StatusBadge status={o.status} />
                </div>
              </li>
            ))}
          </ul>
        ) : (
          <EmptyState
            icon="box"
            title="No orders yet"
            action={
              <Link href="/" className="btn btn--primary">
                Start shopping
              </Link>
            }
          >
            When you place an order, you’ll see it here with its delivery
            status.
          </EmptyState>
        )}
        {page?.page.hasMore && page.page.nextCursor ? (
          <Link
            href={`/orders?cursor=${encodeURIComponent(page.page.nextCursor)}`}
            className="btn btn--secondary"
          >
            Older orders
          </Link>
        ) : null}
      </div>
    </main>
  );
}
