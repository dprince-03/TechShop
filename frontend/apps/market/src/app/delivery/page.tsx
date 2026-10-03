import type { Metadata } from "next";
import Link from "next/link";

import { nigerianStates } from "@techshop/fixtures";
import { PageHeader, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Delivery location | TechShop Market",
  robots: { index: false },
};

export default function DeliveryPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Where should we deliver?"
          lead="We’ll show delivery estimates and fees for your area on every product."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Delivery location" },
          ]}
        />
        <div className="card card--roomy">
          <PreviewForm
            id="location"
            groups={[
              {
                fields: [
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
            ]}
            submitLabel="Save location"
            successTitle="Location set"
            successMessage="In the live store, estimates and fees across the site would update for this area."
          />
        </div>
        <p className="text-sm text-muted">
          <Link href="/help/delivery" className="link">
            How delivery works
          </Link>
        </p>
      </div>
    </main>
  );
}
