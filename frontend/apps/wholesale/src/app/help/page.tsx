import type { Metadata } from "next";
import Link from "next/link";

import { FaqList, PageHeader } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Help | TechShop Wholesale & Retail",
};

export default function HelpPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title="Wholesale & retail help"
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Help" }]}
          actions={
            <Link href="/contact" className="btn btn--inverse">
              Contact sales
            </Link>
          }
        />
        <FaqList
          items={[
            {
              q: "What’s the difference between retail and wholesale?",
              a: "Retail is single units at our standard price, no account needed. Wholesale is bulk buying from a minimum order quantity at lower unit prices, with a business account.",
            },
            {
              q: "How do quantity breaks work?",
              a: "Each product shows its price tiers. Your unit price drops when your quantity reaches the next tier, and the cart applies it automatically.",
            },
            {
              q: "Can I get a VAT invoice?",
              a: "Yes. Business accounts get a VAT invoice for every order, downloadable from Invoices.",
            },
            {
              q: "Do you deliver bulk orders outside Lagos?",
              a: "Yes, to all 36 states and the FCT. Large orders get a delivery timeline in the quote.",
            },
            {
              q: "Can you source items you don’t list?",
              a: "Often, yes. Request a quote with the model and quantity, and we’ll check with our suppliers.",
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
