import type { Metadata } from "next";

import { queryProducts } from "@techshop/fixtures";
import { Breadcrumbs, Countdown, ProductCard } from "@techshop/ui/components";

import styles from "../curated.module.css";

export const metadata: Metadata = {
  title: "Today’s deals | TechShop Market",
  description:
    "Discounted phones, laptops, gaming and more — prices end at midnight.",
};

export default function DealsPage() {
  const deals = queryProducts({ sort: "discount" }).filter(
    (p) => p.compareAtPrice,
  );
  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <div className={styles.head}>
          <div className="stack stack--sm">
            <Breadcrumbs
              items={[{ label: "Home", href: "/" }, { label: "Deals" }]}
            />
            <h1 className={styles.title}>Today’s deals</h1>
            <p className="text-muted">
              Biggest discounts first. Deal prices end at midnight (WAT).
            </p>
          </div>
          <Countdown label="Ends in" />
        </div>
        <ul role="list" className="grid grid--catalog">
          {deals.map((p) => (
            <li key={p.id}>
              <ProductCard product={p} href={`/p/${p.slug}`} />
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
