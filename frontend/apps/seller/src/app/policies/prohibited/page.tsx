import type { Metadata } from "next";

import { Icon, PageHeader } from "@techshop/ui/components";

import styles from "./prohibited.module.css";

export const metadata: Metadata = {
  title: "Prohibited items | TechShop Seller Centre",
};

const prohibited = [
  "Counterfeit, replica or “master copy” devices and accessories",
  "Stolen goods, or devices that fail IMEI or serial checks",
  "Devices locked to someone else’s account (iCloud, Google or carrier locks)",
  "Items you can’t prove you own",
  "Software licences, accounts or game keys sold without authorisation",
  "Signal jammers and other equipment illegal to sell in Nigeria",
  "Recalled products and devices with known safety defects",
  "Cars without valid documents, or with tampered VIN or odometer",
];

export default function ProhibitedPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container container--md stack stack--lg">
        <PageHeader
          title="Prohibited items"
          lead="Listings of these items are removed. Repeated or serious breaches lead to account suspension."
          breadcrumbs={[
            { label: "Seller Centre", href: "/" },
            { label: "Policies", href: "/policies" },
            { label: "Prohibited items" },
          ]}
        />
        <ul role="list" className={styles.list}>
          {prohibited.map((p) => (
            <li key={p} className={styles.item}>
              <Icon name="close" size={18} className={styles.icon} />
              {p}
            </li>
          ))}
        </ul>
        <p className="text-xs text-muted">
          Draft list — to be confirmed by TechShop legal and compliance.
        </p>
      </div>
    </main>
  );
}
