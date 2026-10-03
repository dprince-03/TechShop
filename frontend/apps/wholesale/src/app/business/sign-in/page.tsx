import type { Metadata } from "next";
import Link from "next/link";

import { PageHeader, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Business sign in | TechShop Wholesale & Retail",
  robots: { index: false },
};

export default function BusinessSignInPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Business sign in"
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Business sign in" },
          ]}
        />
        <div className="card card--roomy">
          <PreviewForm
            id="bsignin"
            groups={[
              {
                fields: [
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
                    autoComplete: "current-password",
                  },
                ],
              },
            ]}
            submitLabel="Sign in"
            successTitle="Details accepted"
            successMessage="In the live site you’d now see trade prices, quotes and invoices."
          />
        </div>
        <p className="text-sm text-muted">
          New here?{" "}
          <Link href="/business/register" className="link">
            Open a business account
          </Link>
        </p>
      </div>
    </main>
  );
}
