import type { Metadata } from "next";

import { PageHeader, PreviewForm } from "@techshop/ui/components";

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
          <PreviewForm
            id="sellersupport"
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
                  { name: "shop", label: "Shop name", half: true },
                  {
                    name: "email",
                    label: "Email address",
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
                      "Registration & verification",
                      "Listings",
                      "Orders & delivery",
                      "Payouts",
                      "Returns",
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
            successMessage="In the live Seller Centre this would open a support ticket."
          />
        </div>
      </div>
    </main>
  );
}
