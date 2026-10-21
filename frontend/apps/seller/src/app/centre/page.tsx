import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";
import { connection } from "next/server";

import { EmptyState, PageHeader, Status } from "@techshop/ui/components";

import { myShops } from "@/features/session/server";
import { shopStatus } from "@/features/shops/server";

export const metadata: Metadata = {
  title: "Your shops | TechShop Seller Centre",
};

export default async function CentrePage() {
  await connection();
  const shops = await myShops("/centre");
  if (shops.length === 1) redirect(`/centre/${shops[0].id}`);
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Your shops"
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Your shops" },
          ]}
        />
        {shops.length === 0 ? (
          <EmptyState
            icon="box"
            title="You don’t have a shop yet"
            action={
              <Link href="/centre/open-shop" className="btn btn--primary">
                Open a shop
              </Link>
            }
          >
            Open a shop, verify your identity and add a payout account. Then you
            can list products.
          </EmptyState>
        ) : (
          <ul role="list" className="stack">
            {shops.map((s) => (
              <li key={s.id}>
                <Link
                  href={`/centre/${s.id}`}
                  className="card card--interactive cluster cluster--between"
                >
                  <span className="stack stack--sm">
                    <strong>{s.displayName}</strong>
                    <span className="text-sm text-muted">
                      Your role: {s.role}
                    </span>
                  </span>
                  <Status
                    label={shopStatus[s.status].label}
                    tone={shopStatus[s.status].tone}
                  />
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </main>
  );
}
