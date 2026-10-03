import type { Metadata } from "next";
import Link from "next/link";

import { bestSellers } from "@techshop/fixtures";
import {
  Breadcrumbs,
  EmptyState,
  ProductCard,
  SectionHeader,
} from "@techshop/ui/components";

import styles from "../curated.module.css";

export const metadata: Metadata = {
  title: "Saved items | TechShop Market",
  robots: { index: false },
};

export default function SavedPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <div className="stack stack--sm">
          <Breadcrumbs
            items={[{ label: "Home", href: "/" }, { label: "Saved items" }]}
          />
          <h1 className={styles.title}>Saved items</h1>
        </div>

        <EmptyState
          icon="heart"
          title="Nothing saved yet"
          action={
            <div className={`cluster ${styles.center}`}>
              <Link href="/account" className="btn btn--secondary">
                Sign in
              </Link>
              <Link href="/deals" className="btn btn--ghost">
                Browse deals
              </Link>
            </div>
          }
        >
          Tap the heart on any product to save it for later. Sign in to see
          items you saved on other devices.
        </EmptyState>

        <section aria-labelledby="ideas-title">
          <SectionHeader
            id="ideas-title"
            title="Popular right now"
            action={{ label: "See best sellers", href: "/best-sellers" }}
          />
          <ul role="list" className="scroller">
            {bestSellers.map((p) => (
              <li key={p.id}>
                <ProductCard product={p} href={`/p/${p.slug}`} />
              </li>
            ))}
          </ul>
        </section>
      </div>
    </main>
  );
}
