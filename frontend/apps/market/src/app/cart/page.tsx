import type { Metadata } from "next";

import { sampleCart } from "@techshop/fixtures";
import { Breadcrumbs } from "@techshop/ui/components";

import { CartView } from "@/components/CartView";

import styles from "../curated.module.css";

export const metadata: Metadata = {
  title: "Cart | TechShop Market",
  robots: { index: false },
};

export default function CartPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <div className="stack stack--sm">
          <Breadcrumbs
            items={[{ label: "Home", href: "/" }, { label: "Cart" }]}
          />
          <h1 className={styles.title}>Your cart</h1>
          <p className="text-sm text-muted">
            Preview with sample items — changes aren’t saved.
          </p>
        </div>
        <CartView initial={sampleCart} />
      </div>
    </main>
  );
}
