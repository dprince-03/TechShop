import type { Metadata } from "next";

import { PageHeader, PreviewForm } from "@techshop/ui/components";

import styles from "./trade-in.module.css";

export const metadata: Metadata = {
  title: "Trade in your device | TechShop Market",
  description:
    "Get an estimate for your old phone, laptop or console and put it towards something new.",
};

const steps = [
  {
    title: "Tell us about your device",
    text: "Model, storage and condition. It takes a minute.",
  },
  {
    title: "Get an estimate",
    text: "We’ll send a provisional value based on what you tell us.",
  },
  {
    title: "Inspection",
    text: "Drop it off or have it picked up. We check the condition, IMEI/serial and ownership.",
  },
  {
    title: "Get paid or upgrade",
    text: "Take the confirmed value to your bank account, or put it towards a new device.",
  },
];

export default function TradeInPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container">
        <PageHeader
          eyebrow="Trade-in"
          title="Turn your old device into credit"
          lead="Phones, laptops, tablets and consoles. Final value is confirmed after inspection."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Trade-in" }]}
        />
        <div className={styles.layout}>
          <div className="card card--roomy">
            <PreviewForm
              id="tradein"
              groups={[
                {
                  legend: "Your device",
                  fields: [
                    {
                      name: "type",
                      label: "Device type",
                      type: "select",
                      required: true,
                      options: [
                        "Phone",
                        "Laptop",
                        "Tablet",
                        "Games console",
                        "Smartwatch",
                      ],
                      half: true,
                    },
                    {
                      name: "brand",
                      label: "Brand",
                      required: true,
                      half: true,
                    },
                    {
                      name: "model",
                      label: "Model",
                      required: true,
                      placeholder: "e.g. Nova X5 Pro",
                      half: true,
                    },
                    {
                      name: "storage",
                      label: "Storage",
                      type: "select",
                      options: [
                        "64GB",
                        "128GB",
                        "256GB",
                        "512GB",
                        "1TB or more",
                      ],
                      half: true,
                    },
                    {
                      name: "condition",
                      label: "Condition",
                      type: "select",
                      required: true,
                      options: [
                        "Like new — no marks",
                        "Good — light scratches",
                        "Fair — visible wear",
                        "Faulty — screen, battery or other issue",
                      ],
                    },
                    {
                      name: "imei",
                      label: "IMEI or serial number",
                      hint: "Dial *#06# on a phone to see its IMEI. We check it isn’t reported stolen.",
                      inputMode: "numeric",
                    },
                  ],
                },
                {
                  legend: "Your details",
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
                      name: "owner",
                      label:
                        "I own this device and can show proof of purchase or ownership",
                      type: "checkbox",
                      required: true,
                    },
                  ],
                },
              ]}
              submitLabel="Get my estimate"
              successTitle="Trade-in details ready"
              successMessage="In the live store you’d receive a provisional estimate by SMS and email."
            />
          </div>
          <aside aria-labelledby="how-title" className="stack">
            <h2 id="how-title" className="text-xl">
              How trade-in works
            </h2>
            <ol className={styles.steps}>
              {steps.map((s, i) => (
                <li key={s.title} className={styles.step}>
                  <span className={styles.num} aria-hidden="true">
                    {i + 1}
                  </span>
                  <div>
                    <p className={styles.stepTitle}>{s.title}</p>
                    <p className="text-sm text-muted">{s.text}</p>
                  </div>
                </li>
              ))}
            </ol>
          </aside>
        </div>
      </div>
    </main>
  );
}
