import type { Metadata } from "next";

import {
  EmptyState,
  Icon,
  PageHeader,
  PreviewForm,
} from "@techshop/ui/components";

import styles from "../inner.module.css";

export const metadata: Metadata = { title: "Careers | TechShop" };

const teams = [
  {
    icon: "settings",
    name: "Technology",
    text: "Web, mobile, backend and data.",
  },
  {
    icon: "box",
    name: "Operations & warehouse",
    text: "Inventory, quality checks and fulfilment.",
  },
  {
    icon: "truck",
    name: "Logistics & dispatch",
    text: "Riders, dispatch control and delivery partners.",
  },
  {
    icon: "support",
    name: "Customer experience",
    text: "Support, returns and warranty.",
  },
  {
    icon: "briefcase",
    name: "Sales & partnerships",
    text: "Business sales, vendors and suppliers.",
  },
  {
    icon: "wallet",
    name: "Finance & admin",
    text: "Payments, reconciliation and HR.",
  },
] as const;

export default function CareersPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          eyebrow="Careers"
          title="Build the future of tech retail in Nigeria"
          lead="We’re building teams across technology, operations, logistics, customer experience and sales."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Careers" }]}
        />
        <section aria-labelledby="teams-title" className="stack">
          <h2 id="teams-title" className={styles.h2}>
            Our teams
          </h2>
          <ul role="list" className={`${styles.cards} ${styles.cards3}`}>
            {teams.map((t) => (
              <li key={t.name} className="card">
                <Icon name={t.icon} size={24} className={styles.icon} />
                <h3 className={styles.cardTitle}>{t.name}</h3>
                <p className="text-sm text-muted">{t.text}</p>
              </li>
            ))}
          </ul>
        </section>
        <section aria-labelledby="roles-title" className="stack">
          <h2 id="roles-title" className={styles.h2}>
            Open roles
          </h2>
          <EmptyState icon="briefcase" title="No open roles listed right now">
            New roles will be posted here. You can join our talent pool below
            and we’ll contact you when a matching role opens.
          </EmptyState>
        </section>
        <section aria-labelledby="pool-title" className={styles.twoCol}>
          <div className="stack stack--sm">
            <h2 id="pool-title" className={styles.h2}>
              Join the talent pool
            </h2>
            <p className="text-muted">
              Share your details and the team you’d like to join.
            </p>
          </div>
          <div className="card card--roomy">
            <PreviewForm
              id="talent"
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
                      name: "email",
                      label: "Email address",
                      type: "email",
                      required: true,
                      autoComplete: "email",
                      half: true,
                    },
                    {
                      name: "team",
                      label: "Team",
                      type: "select",
                      required: true,
                      options: teams.map((t) => t.name),
                      half: true,
                    },
                    {
                      name: "location",
                      label: "Location",
                      autoComplete: "address-level2",
                      half: true,
                    },
                    {
                      name: "cv",
                      label: "CV (PDF)",
                      type: "file",
                      accept: ".pdf",
                      required: true,
                    },
                    {
                      name: "consent",
                      label:
                        "I agree that TechShop can keep my details for up to 12 months to contact me about roles",
                      type: "checkbox",
                      required: true,
                    },
                  ],
                },
              ]}
              submitLabel="Join the talent pool"
              successTitle="Application ready"
              successMessage="In the live site your details would be added to our talent pool."
            />
          </div>
        </section>
      </div>
    </main>
  );
}
