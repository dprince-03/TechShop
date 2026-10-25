import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { formatKobo, tradeInStatus } from "@techshop/api-client";
import { EmptyState, PageHeader } from "@techshop/ui/components";

import { AccountNav, Records } from "@/features/aftersales/components/Records";
import { myTradeIns } from "@/features/aftersales/server";
import { formatDate } from "@/features/orders/status";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Trade-ins | TechShop Market",
  robots: { index: false },
};

export default async function TradeInsPage() {
  await connection();
  await bff.requireSession("/account/trade-ins");
  const rows = (await myTradeIns()) ?? [];
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Trade-ins"
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Account", href: "/account" },
            { label: "Trade-ins" },
          ]}
          actions={
            <Link href="/trade-in" className="btn btn--secondary">
              Trade in a device
            </Link>
          }
        />
        <AccountNav current="/account/trade-ins" />
        <section className="card card--roomy">
          {rows.length ? (
            <Records
              label="Your trade-ins"
              rows={rows.map((t) => ({
                key: t.reference,
                href: `/account/trade-ins/${t.reference}`,
                title: `${t.reference} · ${t.device}`,
                status: tradeInStatus[t.status] ?? {
                  label: t.status,
                  tone: "neutral",
                },
                meta: `${t.offer != null ? `Offer ${formatKobo(t.offer)}` : t.estimate != null ? `Estimate ${formatKobo(t.estimate)}` : "Being priced"} · ${formatDate(t.createdAt)}`,
              }))}
            />
          ) : (
            <EmptyState icon="swap" title="No trade-ins yet">
              <Link href="/trade-in" className="link">
                Get an estimate
              </Link>{" "}
              for your old phone, laptop or console.
            </EmptyState>
          )}
        </section>
      </div>
    </main>
  );
}
