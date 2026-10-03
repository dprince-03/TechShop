import type { Metadata } from "next";
import Link from "next/link";

import { nigerianStates } from "@techshop/fixtures";
import { PageHeader, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Open a business account | TechShop Wholesale & Retail",
  robots: { index: false },
};

export default function BusinessRegisterPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--md stack stack--lg">
        <PageHeader
          title="Open a business account"
          lead="Unlock trade prices, VAT invoices and credit terms. We verify business details before activating trade pricing."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Business account" },
          ]}
        />
        <div className="card card--roomy">
          <PreviewForm
            id="business"
            groups={[
              {
                legend: "Business details",
                fields: [
                  {
                    name: "company",
                    label: "Registered business name",
                    required: true,
                    autoComplete: "organization",
                  },
                  {
                    name: "rc",
                    label: "CAC RC / BN number",
                    required: true,
                    half: true,
                  },
                  { name: "tin", label: "Tax ID (TIN)", half: true },
                  {
                    name: "type",
                    label: "Business type",
                    type: "select",
                    required: true,
                    options: [
                      "Office / SME",
                      "School or university",
                      "Reseller or retailer",
                      "Government agency",
                      "NGO",
                      "Healthcare",
                      "Other",
                    ],
                    half: true,
                  },
                  {
                    name: "size",
                    label: "Number of staff",
                    type: "select",
                    options: [
                      "1–10",
                      "11–50",
                      "51–200",
                      "201–1,000",
                      "More than 1,000",
                    ],
                    half: true,
                  },
                  {
                    name: "address",
                    label: "Business address",
                    required: true,
                    autoComplete: "street-address",
                  },
                  {
                    name: "state",
                    label: "State",
                    type: "select",
                    required: true,
                    options: nigerianStates,
                    half: true,
                  },
                  {
                    name: "city",
                    label: "City or LGA",
                    required: true,
                    autoComplete: "address-level2",
                    half: true,
                  },
                ],
              },
              {
                legend: "Account owner",
                fields: [
                  {
                    name: "name",
                    label: "Full name",
                    required: true,
                    autoComplete: "name",
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
                  {
                    name: "email",
                    label: "Work email",
                    type: "email",
                    required: true,
                    autoComplete: "email",
                  },
                  {
                    name: "password",
                    label: "Password",
                    type: "password",
                    required: true,
                    autoComplete: "new-password",
                    minLength: 8,
                    half: true,
                    hint: "At least 8 characters.",
                  },
                  {
                    name: "confirm",
                    label: "Confirm password",
                    type: "password",
                    required: true,
                    autoComplete: "new-password",
                    matches: "password",
                    half: true,
                  },
                  {
                    name: "terms",
                    label: "I agree to the Terms of use and Privacy policy",
                    type: "checkbox",
                    required: true,
                  },
                ],
              },
            ]}
            submitLabel="Create business account"
            successTitle="Business details ready"
            successMessage="In the live site we’d verify your CAC details and email you when trade pricing is active."
          />
        </div>
        <p className="text-sm text-muted">
          Already have an account?{" "}
          <Link href="/business/sign-in" className="link">
            Sign in
          </Link>
        </p>
      </div>
    </main>
  );
}
