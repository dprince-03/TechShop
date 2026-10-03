import type { Metadata } from "next";

import { queryProducts } from "@techshop/fixtures";
import { Breadcrumbs, ProductCard } from "@techshop/ui/components";

import styles from "../curated.module.css";

export const metadata: Metadata = {
  title: "Best sellers | TechShop Market",
  description: "The most popular tech on TechShop, ranked by customer ratings.",
};

export default function BestSellersPage() {
  const ranked = queryProducts({ sort: "rating" }).filter(
    (p) => (p.rating?.count ?? 0) >= 50,
  );
  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <div className="stack stack--sm">
          <Breadcrumbs
            items={[{ label: "Home", href: "/" }, { label: "Best sellers" }]}
          />
          <h1 className={styles.title}>Best sellers</h1>
          <p className="text-muted">
            Top-rated products with at least 50 reviews.
          </p>
        </div>
        <ol role="list" className="grid grid--catalog">
          {ranked.map((p) => (
            <li key={p.id}>
              <ProductCard product={p} href={`/p/${p.slug}`} />
            </li>
          ))}
        </ol>
      </div>
    </main>
  );
}
