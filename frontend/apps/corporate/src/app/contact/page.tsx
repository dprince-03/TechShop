import type { Metadata } from "next";

import { Icon, PageHeader, PreviewForm } from "@techshop/ui/components";
import { sites } from "@techshop/ui/sites";

import styles from "../inner.module.css";

export const metadata: Metadata = { title: "Contact | TechShop" };

export default function ContactPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title="Contact TechShop"
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Contact" }]}
        />
        <ul role="list" className={`${styles.cards} ${styles.cards3}`}>
          <li className="card">
            <Icon name="support" size={24} className={styles.icon} />
            <h2 className={styles.cardTitle}>Customer support</h2>
            <p className="text-sm text-muted">
              Orders, delivery, returns and warranty.
            </p>
            <a
              href={`${sites.market}/help/contact`}
              className="link link--arrow text-sm"
            >
              Contact support
            </a>
          </li>
          <li className="card">
            <Icon name="briefcase" size={24} className={styles.icon} />
            <h2 className={styles.cardTitle}>Business sales</h2>
            <p className="text-sm text-muted">
              Bulk orders, quotes and supply.
            </p>
            <a
              href={`${sites.wholesale}/contact`}
              className="link link--arrow text-sm"
            >
              Contact sales
            </a>
          </li>
          <li className="card">
            <Icon name="store" size={24} className={styles.icon} />
            <h2 className={styles.cardTitle}>Sellers</h2>
            <p className="text-sm text-muted">Selling on TechShop Market.</p>
            <a
              href={`${sites.seller}/help/contact`}
              className="link link--arrow text-sm"
            >
              Seller support
            </a>
          </li>
        </ul>
        <section
          id="press"
          aria-labelledby="press-title"
          className={styles.twoCol}
        >
          <div className="stack stack--sm">
            <h2 id="press-title" className={styles.h2}>
              Press, partnerships and everything else
            </h2>
            <p className="text-muted">
              For media, investor and general enquiries.
            </p>
          </div>
          <div className="card card--roomy">
            <PreviewForm
              id="general"
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
                      name: "org",
                      label: "Organisation",
                      autoComplete: "organization",
                      half: true,
                    },
                    {
                      name: "topic",
                      label: "Topic",
                      type: "select",
                      required: true,
                      options: [
                        "Press & media",
                        "Partnerships",
                        "Investors",
                        "Something else",
                      ],
                      half: true,
                    },
                    {
                      name: "message",
                      label: "Message",
                      type: "textarea",
                      required: true,
                    },
                  ],
                },
              ]}
              submitLabel="Send message"
              successTitle="Message ready"
              successMessage="In the live site this would reach the right TechShop team."
            />
          </div>
        </section>
      </div>
    </main>
  );
}
