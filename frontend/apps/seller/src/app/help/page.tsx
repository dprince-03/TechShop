import type { Metadata } from "next";
import Link from "next/link";

import { FaqList, PageHeader } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Seller help | TechShop Seller Centre",
};

export default function SellerHelpPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title="Seller help"
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Help" },
          ]}
          actions={
            <Link href="/help/contact" className="btn btn--inverse">
              Contact seller support
            </Link>
          }
        />
        <FaqList
          items={[
            {
              q: "How long does verification take?",
              a: "Usually a few working days. We’ll email you if we need anything else.",
            },
            {
              q: "How do I list a used phone?",
              a: "Choose the condition, add your own photos, enter the IMEI and battery health, and upload proof of ownership.",
            },
            {
              q: "How do I ship an order?",
              a: "Drop it at a TechShop hub or request a pickup from Orders. You can also deliver yourself if you set that up.",
            },
            {
              q: "When will I get paid?",
              a: "After the customer receives the order and the return window closes. You’ll get an SMS and email for every payout.",
            },
            {
              q: "A customer wants to return an item. What happens?",
              a: "We review the request. If it’s faulty or not as described, the item comes back to you and the customer is refunded.",
            },
          ]}
        />
        <p className="text-xs text-muted">
          Sample help content — to be confirmed before launch.
        </p>
      </div>
    </main>
  );
}
