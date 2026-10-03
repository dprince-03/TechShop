import type { Metadata } from "next";

import { Icon, PageHeader, PreviewForm } from "@techshop/ui/components";

import styles from "../inner.module.css";

export const metadata: Metadata = { title: "Partners | TechShop" };

const kinds = [
  {
    icon: "box",
    title: "Brands & manufacturers",
    text: "Become an authorised channel for your devices and accessories across Nigeria.",
  },
  {
    icon: "truck",
    title: "Distributors & suppliers",
    text: "Supply stock to TechShop retail and wholesale with reliable volumes and payment.",
  },
  {
    icon: "pin",
    title: "Logistics partners",
    text: "Deliver TechShop orders in your region alongside our own dispatch riders.",
  },
  {
    icon: "wrench",
    title: "Repair & service centres",
    text: "Handle warranty repairs and trade-in inspections for TechShop customers.",
  },
] as const;

export default function PartnersPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          eyebrow="Partners"
          title="Grow with TechShop"
          lead="We work with brands, suppliers, logistics and service partners to get genuine tech to customers across Nigeria."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Partners" }]}
        />
        <ul role="list" className={styles.cards}>
          {kinds.map((k) => (
            <li
              key={k.title}
              id={k.title.startsWith("Logistics") ? "logistics" : undefined}
              className="card"
            >
              <Icon name={k.icon} size={28} className={styles.icon} />
              <h2 className={styles.cardTitle}>{k.title}</h2>
              <p className="text-muted">{k.text}</p>
            </li>
          ))}
        </ul>
        <section aria-labelledby="enquire-title" className={styles.twoCol}>
          <div className="stack stack--sm">
            <h2 id="enquire-title" className={styles.h2}>
              Start a conversation
            </h2>
            <p className="text-muted">
              Tell us about your business and the partnership you have in mind.
              The right team will get back to you.
            </p>
          </div>
          <div className="card card--roomy">
            <PreviewForm
              id="partner"
              groups={[
                {
                  fields: [
                    {
                      name: "company",
                      label: "Company name",
                      required: true,
                      autoComplete: "organization",
                      half: true,
                    },
                    {
                      name: "type",
                      label: "Partnership type",
                      type: "select",
                      required: true,
                      options: kinds.map((k) => k.title),
                      half: true,
                    },
                    {
                      name: "name",
                      label: "Your name",
                      required: true,
                      autoComplete: "name",
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
                      name: "message",
                      label: "Tell us about the partnership",
                      type: "textarea",
                      required: true,
                    },
                  ],
                },
              ]}
              submitLabel="Send enquiry"
              successTitle="Enquiry ready"
              successMessage="In the live site this would reach our partnerships team."
            />
          </div>
        </section>
      </div>
    </main>
  );
}
