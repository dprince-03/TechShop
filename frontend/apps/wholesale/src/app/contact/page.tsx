import type { Metadata } from "next";

import { PageHeader, PreviewForm } from "@techshop/ui/components";

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
          <PreviewForm
            id="sales"
            groups={[
              {
                fields: [
                  {
                    name: "name",
                    label: "Full name",
                    required: true,
                    autoComplete: "name",
                    half: true,
                  },
                  {
                    name: "company",
                    label: "Business name",
                    autoComplete: "organization",
                    half: true,
                  },
                  {
                    name: "email",
                    label: "Work email",
                    type: "email",
                    required: true,
                    autoComplete: "email",
                    half: true,
                  },
                  {
                    name: "phone",
                    label: "Phone number",
                    type: "tel",
                    autoComplete: "tel",
                    half: true,
                  },
                  {
                    name: "topic",
                    label: "Topic",
                    type: "select",
                    required: true,
                    options: [
                      "Bulk order",
                      "Pricing",
                      "Business account",
                      "Credit terms",
                      "Invoice or payment",
                      "Something else",
                    ],
                  },
                  {
                    name: "message",
                    label: "Message",
                    type: "textarea",
                    required: true,
                  },
                ],
              },
            ]}
            submitLabel="Send message"
            successTitle="Message ready"
            successMessage="In the live site our sales team would reply within working hours."
          />
        </div>
      </div>
    </main>
  );
}
