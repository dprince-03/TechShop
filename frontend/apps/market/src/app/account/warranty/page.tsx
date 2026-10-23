import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { claimStatus } from "@techshop/api-client";
import { EmptyState, PageHeader } from "@techshop/ui/components";

import styles from "@/features/aftersales/components/Aftersales.module.css";
import { AccountNav, Records } from "@/features/aftersales/components/Records";
import { coveredItems, myClaims } from "@/features/aftersales/server";
import { formatDate } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Warranty | TechShop Market",
  robots: { index: false },
};

const provider: Record<string, string> = {
  manufacturer: "Manufacturer warranty",
  techshop: "TechShop warranty",
  seller: "Seller warranty",
};

export default async function WarrantyPage() {
  await connection();
  await bff.requireSession("/account/warranty");
  const [items, claims] = await Promise.all([coveredItems(), myClaims()]);
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Warranty and repairs"
          lead="Whoever gives the warranty, claim here: we check the device at a store and handle the repair, replacement or refund."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Account", href: "/account" },
            { label: "Warranty" },
          ]}
        />
        <AccountNav current="/account/warranty" />
        <section
          className="card card--roomy stack"
          aria-labelledby="covered-title"
        >
          <h2 id="covered-title" className={styles.h2}>
            Covered items
          </h2>
          {items?.length ? (
            <ul role="list" className={styles.list}>
              {items.map((i) => (
                <li key={i.orderLineId} className={styles.row}>
                  <span>
                    {i.productName}
                    <span className="text-xs text-muted">
                      {" "}
                      · {i.variantName}
                    </span>
                  </span>
                  {i.openClaim ? (
                    <Link
                      href={`/account/warranty/${i.openClaim}`}
                      className="link text-sm"
                    >
                      Claim {i.openClaim}
                    </Link>
                  ) : (
                    <Link
                      href={`/account/warranty/new?line=${i.orderLineId}`}
                      className="btn btn--secondary btn--sm"
                    >
                      Make a claim
                    </Link>
                  )}
                  <span className={styles.meta}>
                    {provider[i.provider] ?? i.provider} · until{" "}
                    {formatDate(i.endsOn)} · order {i.orderNumber}
                    {i.serialHint ? ` · serial ending ${i.serialHint}` : ""}
                  </span>
                </li>
              ))}
            </ul>
          ) : (
            <EmptyState icon="shield" title="Nothing under warranty">
              Items show here once they’re delivered, for as long as their
              warranty lasts.
            </EmptyState>
          )}
        </section>
        {claims?.length ? (
          <section
            className="card card--roomy stack"
            aria-labelledby="claims-title"
          >
            <h2 id="claims-title" className={styles.h2}>
              Your claims
            </h2>
            <Records
              label="Your warranty claims"
              rows={claims.map((c) => ({
                key: c.number,
                href: `/account/warranty/${c.number}`,
                title: `${c.number} · ${c.productName ?? "Device"}`,
                status: claimStatus[c.status] ?? {
                  label: c.status,
                  tone: "neutral",
                },
                meta: `${c.fault} · ${formatDate(c.createdAt)}`,
              }))}
            />
          </section>
        ) : null}
      </div>
    </main>
  );
}
