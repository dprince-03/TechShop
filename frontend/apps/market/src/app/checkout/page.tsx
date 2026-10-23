import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { EmptyState, PageHeader } from "@techshop/ui/components";

import { listAddresses } from "@/features/account/server";
import { getCart } from "@/features/cart/server";
import { getStates } from "@/features/catalogue/api";
import { storeCredit } from "@/features/aftersales/server";
import { CheckoutForm } from "@/features/checkout/components/CheckoutForm";
import { initialQuote, paymentMethods } from "@/features/checkout/server";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Checkout | TechShop Market",
  robots: { index: false },
};

export default async function CheckoutPage() {
  await connection();
  await bff.requireSession("/checkout");
  const [cart, addresses, states, methods] = await Promise.all([
    getCart(),
    listAddresses(),
    getStates(),
    paymentMethods(),
  ]);
  const crumbs = [
    { label: "Home", href: "/" },
    { label: "Cart", href: "/cart" },
    { label: "Checkout" },
  ];
  if (!cart?.lines?.length) {
    return (
      <main id="main" className="section section--page">
        <div className="container container--narrow stack stack--lg">
          <PageHeader title="Checkout" breadcrumbs={crumbs} />
          <EmptyState
            icon="cart"
            title="Your cart is empty"
            action={
              <Link href="/" className="btn btn--primary">
                Continue shopping
              </Link>
            }
          >
            Add something to your cart to check out.
          </EmptyState>
        </div>
      </main>
    );
  }
  const first = addresses.find((a) => a.isDefault) ?? addresses[0];
  const [quote, credit] = await Promise.all([
    initialQuote(first?.id),
    storeCredit().catch(() => null),
  ]);
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader title="Checkout" breadcrumbs={crumbs} />
        <CheckoutForm
          addresses={addresses}
          states={states}
          methods={methods}
          initial={quote}
          storeCredit={credit?.balanceKobo ?? 0}
        />
      </div>
    </main>
  );
}
