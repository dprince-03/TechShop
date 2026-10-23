import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { returnReason, returnStatus } from "@techshop/api-client";
import { EmptyState, PageHeader } from "@techshop/ui/components";

import { AccountNav, Records } from "@/features/aftersales/components/Records";
import { myReturns } from "@/features/aftersales/server";
import { formatDate } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Returns | TechShop Market",
  robots: { index: false },
};

export default async function ReturnsPage() {
  await connection();
  await bff.requireSession("/account/returns");
  const returns = (await myReturns()) ?? [];
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Returns"
          lead="Start a return from the order, within 7 days of delivery."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Account", href: "/account" },
            { label: "Returns" },
          ]}
        />
        <AccountNav current="/account/returns" />
        <section className="card card--roomy">
          {returns.length ? (
            <Records
              label="Your returns"
              rows={returns.map((r) => ({
                key: r.rma,
                href: `/account/returns/${r.rma}`,
                title: r.rma,
                status: returnStatus[r.status] ?? {
                  label: r.status,
                  tone: "neutral",
                },
                meta: `Order ${r.orderNumber} · ${returnReason[r.reason] ?? r.reason} · ${formatDate(r.createdAt)}`,
              }))}
            />
          ) : (
            <EmptyState icon="returns" title="No returns">
              To return something, open the{" "}
              <Link href="/orders" className="link">
                order
              </Link>{" "}
              and choose “Return items”.
            </EmptyState>
          )}
        </section>
      </div>
    </main>
  );
}
