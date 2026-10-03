import type { Metadata } from "next";
import Link from "next/link";

import { categories } from "@techshop/fixtures";
import { CategoryArt, PageHeader } from "@techshop/ui/components";

import styles from "../inner.module.css";

export const metadata: Metadata = {
  title: "Categories | TechShop Wholesale & Retail",
};

export default function CategoriesPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container">
        <PageHeader
          title="Categories"
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Categories" }]}
        />
        <ul role="list" className={`grid ${styles.categories}`}>
          {categories
            .filter((c) => c.slug !== "cars")
            .map((c) => (
              <li key={c.slug}>
                <Link href={`/c/${c.slug}`} className={styles.categoryTile}>
                  <CategoryArt category={c.slug} size="sm" />
                  <span className="stack stack--sm">
                    <strong>{c.name}</strong>
                    <span className="text-xs text-muted">{c.description}</span>
                  </span>
                </Link>
              </li>
            ))}
        </ul>
      </div>
    </main>
  );
}
