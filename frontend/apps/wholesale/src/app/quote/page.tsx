import type { Metadata } from "next";

import { nigerianStates } from "@techshop/fixtures";
import { PageHeader, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Request a quote | TechShop Wholesale & Retail",
};

export default function QuotePage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--md">
        <PageHeader
          title="Request a quote"
          lead="Tell us what you need. Our B2B team replies with prices, availability and delivery timelines."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Request a quote" },
          ]}
        />
        <div className="card card--roomy">
          <PreviewForm
            id="quote"
            groups={[
              {
                legend: "Your business",
                fields: [
                  {
                    name: "company",
                    label: "Business name",
                    required: true,
                    autoComplete: "organization",
                    half: true,
                  },
                  {
                    name: "rc",
                    label: "CAC RC number",
                    half: true,
                    hint: "Helps us offer trade pricing and credit terms.",
                  },
                  {
                    name: "contact",
                    label: "Contact name",
                    required: true,
                    autoComplete: "name",
                    half: true,
                  },
                  {
                    name: "role",
                    label: "Job title",
                    autoComplete: "organization-title",
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
                    required: true,
                    autoComplete: "tel",
                    half: true,
                  },
                ],
              },
              {
                legend: "What you need",
                fields: [
                  {
                    name: "items",
                    label: "Products and quantities",
                    type: "textarea",
                    required: true,
                    placeholder:
                      "e.g. 40 × business laptops (Core i5, 16GB, 512GB), 40 × wireless mice",
                    hint: "Model names, specs and quantities. Add “or similar” if you’re flexible.",
                  },
                  {
                    name: "state",
                    label: "Delivery state",
                    type: "select",
                    required: true,
                    options: nigerianStates,
                    half: true,
                  },
                  {
                    name: "date",
                    label: "Needed by",
                    type: "date",
                    half: true,
                  },
                  {
                    name: "vat",
                    label: "We need a VAT invoice",
                    type: "checkbox",
                  },
                  {
                    name: "credit",
                    label:
                      "We’d like to pay on credit terms (subject to approval)",
                    type: "checkbox",
                  },
                ],
              },
            ]}
            submitLabel="Send quote request"
            successTitle="Quote request ready"
            successMessage="In the live site our B2B team would receive this and reply by email."
          />
        </div>
      </div>
    </main>
  );
}
