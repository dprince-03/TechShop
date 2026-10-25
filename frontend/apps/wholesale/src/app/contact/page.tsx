import type { Metadata } from "next";

import { PageHeader, ContactForm } from "@techshop/ui/components";

import { contact } from "@/features/contact/actions";

export const metadata: Metadata = {
  title: "Contact sales | TechShop Wholesale & Retail",
};

export default function ContactPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--md">
        <PageHeader
          title="Contact sales"
          lead="Questions about bulk orders, pricing or your business account? Send us a message."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Contact sales" },
          ]}
        />
        <div className="card card--roomy">
          <ContactForm
            action={contact}
            topics={[
              { value: "business", label: "Business buying and quotes" },
              { value: "order", label: "An order" },
              { value: "payment", label: "Payments and invoices" },
              { value: "other", label: "Something else" },
            ]}
            askOrder={true}
          />
        </div>
      </div>
    </main>
  );
}
