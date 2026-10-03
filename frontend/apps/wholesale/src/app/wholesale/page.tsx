import type { Metadata } from "next";
import Link from "next/link";

import { wholesaleProducts } from "@techshop/fixtures";
import { PageHeader } from "@techshop/ui/components";

import { WholesaleCard } from "@/components/WholesaleCard";

import styles from "../inner.module.css";

export const metadata: Metadata = {
  title: "Trade prices | TechShop Wholesale & Retail",
};

export default function TradePricesPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--wide">
        <PageHeader
          eyebrow="Wholesale"
          title="Trade prices"
          lead="Unit prices drop at each quantity break. Business account required to check out at trade prices."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Trade prices" },
          ]}
          actions={
            <Link href="/quote" className="btn btn--inverse">
              Request a custom quote
            </Link>
          }
        />
        <ul role="list" className={`grid ${styles.bulkGrid}`}>
          {wholesaleProducts.map((p) => (
            <li key={p.id}>
              <WholesaleCard product={p} />
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
