import type { Metadata } from "next";
import Link from "next/link";

import { FaqList, PageHeader, Prose } from "@techshop/ui/components";

export const metadata: Metadata = { title: "Fees | TechShop Seller Centre" };

export default function FeesPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title="Selling fees"
          lead="No fee to register or list. You pay a commission only when you make a sale."
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Fees" },
          ]}
          actions={
            <Link href="/register" className="btn btn--primary">
              Start selling
            </Link>
          }
        />
        <Prose>
          <h2>Commission</h2>
          <p>
            A commission is deducted from each sale. The rate depends on the
            product category. You’ll always see the exact fee and what you’ll
            receive before you publish a listing.
          </p>
          <h2>Delivery</h2>
          <p>
            If you use TechShop logistics, the delivery fee is charged to the
            customer and paid to our logistics team. If you deliver yourself,
            you keep the delivery fee you set.
          </p>
          <h2>Payouts</h2>
          <p>
            Payouts go to your verified bank account after the customer receives
            the order and the return window closes. There’s no fee for standard
            payouts.
          </p>
        </Prose>
        <FaqList
          items={[
            {
              q: "Are there monthly subscription fees?",
              a: "No. You only pay commission when you sell.",
            },
            {
              q: "What happens to fees if an order is returned?",
              a: "If a sale is refunded, the commission on it is refunded to you.",
            },
          ]}
        />
        <p className="text-xs text-muted">
          Category commission rates will be published here once set by TechShop.
        </p>
      </div>
    </main>
  );
}
