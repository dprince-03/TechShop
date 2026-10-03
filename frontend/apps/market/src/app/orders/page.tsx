import type { Metadata } from "next";
import Link from "next/link";

import { PageHeader, PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Track an order | TechShop Market",
  robots: { index: false },
};

const stages = [
  "Order placed",
  "Payment confirmed",
  "Packed",
  "Out for delivery",
  "Delivered",
];

export default function TrackOrderPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Track an order"
          lead="Enter your order number and the phone number used at checkout. No account needed."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Help", href: "/help" },
            { label: "Track an order" },
          ]}
        />
        <div className="card card--roomy">
          <PreviewForm
            id="track"
            groups={[
              {
                fields: [
                  {
                    name: "order",
                    label: "Order number",
                    required: true,
                    placeholder: "TS-10482",
                    hint: "It’s in your confirmation SMS and email.",
                    autoComplete: "off",
                  },
                  {
                    name: "phone",
                    label: "Phone number",
                    type: "tel",
                    required: true,
                    autoComplete: "tel",
                    placeholder: "0803 123 4567",
                  },
                ],
              },
            ]}
            submitLabel="Track order"
            successTitle="Order lookup ready"
            successMessage="In the live store you’d now see where your order is, the rider’s details on delivery day, and a delivery window."
          />
        </div>
        <section aria-labelledby="stages-title" className="stack">
          <h2 id="stages-title" className="text-xl">
            What each status means
          </h2>
          <ol className="stack stack--sm">
            {stages.map((s, i) => (
              <li key={s} className="text-sm">
                <strong>
                  {i + 1}. {s}
                </strong>
              </li>
            ))}
          </ol>
          <p className="text-sm text-muted">
            Problem with an order?{" "}
            <Link href="/help/contact" className="link">
              Contact support
            </Link>
          </p>
        </section>
      </div>
    </main>
  );
}
