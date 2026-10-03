import type { Metadata } from "next";
import Link from "next/link";

import { categories, ownStock } from "@techshop/fixtures";
import { PageHeader, ProductCard } from "@techshop/ui/components";

import styles from "../inner.module.css";

export const metadata: Metadata = {
  title: "Retail | TechShop Wholesale & Retail",
};

export default function RetailPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--wide">
        <PageHeader
          eyebrow="Retail"
          title="Shop TechShop stock"
          lead="Buy one or a few items at our standard price — sold, shipped and warrantied by TechShop."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Retail" }]}
        />
        <ul role="list" className={`cluster ${styles.chips}`}>
          {categories
            .filter((c) => c.slug !== "cars")
            .map((c) => (
              <li key={c.slug}>
                <Link
                  href={`/c/${c.slug}`}
                  className="btn btn--sm btn--secondary"
                >
                  {c.name}
                </Link>
              </li>
            ))}
        </ul>
        <ul role="list" className="grid grid--catalog">
          {ownStock.map((p) => (
            <li key={p.id}>
              <ProductCard product={p} href={`/p/${p.slug}`} />
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
