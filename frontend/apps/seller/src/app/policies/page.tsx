import type { Metadata } from "next";
import Link from "next/link";

import { PageHeader, Prose } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Seller policies | TechShop Seller Centre",
};

export default function PoliciesPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title="Seller policies"
          lead="The standards every TechShop seller agrees to."
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Policies" },
          ]}
        />
        <Prose>
          <h2>Honest listings</h2>
          <ul>
            <li>
              Use your own photos of the actual item for used and refurbished
              devices.
            </li>
            <li>
              Grade condition accurately: new, UK-used or refurbished. Disclose
              faults, battery health and repairs.
            </li>
            <li>
              Show the real price. No fake “was” prices or inflated discounts.
            </li>
          </ul>
          <h2>Genuine, legal products only</h2>
          <p>
            Every item must be genuine and yours to sell. Used devices need
            proof of ownership and must pass IMEI or serial checks. See the full{" "}
            <Link href="/policies/prohibited">prohibited items list</Link>.
          </p>
          <h2>Orders and service</h2>
          <ul>
            <li>
              Confirm and dispatch orders within the time shown on your listing.
            </li>
            <li>Respond to customer messages promptly and politely.</li>
            <li>Accept returns for faulty or not-as-described items.</li>
          </ul>
          <h2>What happens if a policy is broken</h2>
          <p>
            Depending on how serious it is: a warning, listing removal, payout
            holds, or account suspension.
          </p>
          <p>
            Full terms are in the{" "}
            <Link href="/legal/seller-agreement">Seller agreement</Link>.
          </p>
        </Prose>
        <p className="text-xs text-muted">
          Draft policy summary — to be confirmed before launch.
        </p>
      </div>
    </main>
  );
}
