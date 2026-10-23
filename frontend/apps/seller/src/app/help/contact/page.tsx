import type { Metadata } from "next";

import { PageHeader, ContactForm } from "@techshop/ui/components";

import { contact } from "@/features/contact/actions";

export const metadata: Metadata = {
  title: "Contact seller support | TechShop Seller Centre",
};

export default function SellerContactPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--md">
        <PageHeader
          title="Contact seller support"
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Help", href: "/help" },
            { label: "Contact" },
          ]}
        />
        <div className="card card--roomy">
          <ContactForm
            action={contact}
            topics={[
              {
                value: "selling",
                label: "Registration, verification or listings",
              },
              { value: "order", label: "Orders and delivery" },
              { value: "payment", label: "Payouts" },
              { value: "return", label: "Returns" },
              { value: "other", label: "Something else" },
            ]}
            askOrder={false}
          />
        </div>
      </div>
    </main>
  );
}
