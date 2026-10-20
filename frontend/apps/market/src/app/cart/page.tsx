import type { Metadata } from "next";
import { connection } from "next/server";

import { Breadcrumbs } from "@techshop/ui/components";

import { CartView } from "@/features/cart/components/CartView";
import { getCart } from "@/features/cart/server";
import { bff } from "@/lib/bff";

import styles from "../curated.module.css";

export const metadata: Metadata = {
  title: "Cart | TechShop Market",
  robots: { index: false },
};

export default async function CartPage() {
  await connection();
  const [cart, session] = await Promise.all([getCart(), bff.getSession()]);
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <div className="stack stack--sm">
          <Breadcrumbs
            items={[{ label: "Home", href: "/" }, { label: "Cart" }]}
          />
          <h1 className={styles.title}>Your cart</h1>
        </div>
        <CartView cart={cart} signedIn={Boolean(session)} />
      </div>
    </main>
  );
}
