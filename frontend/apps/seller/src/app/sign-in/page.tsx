import type { Metadata } from "next";
import Link from "next/link";

import { PageHeader, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = { title: "Sign in | TechShop Seller Centre" };

export default function SignInPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Sign in to Seller Centre"
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Sign in" },
          ]}
        />
        <div className="card card--roomy">
          <PreviewForm
            id="ssignin"
            groups={[
              {
                fields: [
                  {
                    name: "email",
                    label: "Email or phone number",
                    required: true,
                    autoComplete: "username",
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
            successMessage="In the live Seller Centre you’d now see your orders, listings and payouts."
          />
        </div>
        <p className="text-sm text-muted">
          New seller?{" "}
          <Link href="/register" className="link">
            Create a seller account
          </Link>
        </p>
      </div>
    </main>
  );
}
