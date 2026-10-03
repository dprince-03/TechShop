import type { Metadata } from "next";
import Link from "next/link";

import { FaqList, Icon, PageHeader } from "@techshop/ui/components";

import { helpTopics } from "@/lib/help";

import styles from "./help.module.css";

export const metadata: Metadata = { title: "Help centre | TechShop Market" };

const commonQuestions = [
  {
    q: "How do I track my order?",
    a: "Use Track an order with your order number and phone number — no account needed.",
  },
  {
    q: "Is it safe to buy from marketplace sellers?",
    a: "Sellers are verified before they can list, and you pay TechShop, not the seller. Sellers are paid after delivery.",
  },
  {
    q: "Can I pay on delivery?",
    a: "Payment on delivery isn’t available yet. You can pay with Paystack, OPay or Moniepoint at checkout.",
  },
];

export default function HelpPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title="How can we help?"
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Help" }]}
          actions={
            <Link href="/orders" className="btn btn--inverse">
              Track an order
            </Link>
          }
        />
        <ul role="list" className={`grid ${styles.topics}`}>
          {helpTopics.map((t) => (
            <li key={t.slug}>
              <Link
                href={`/help/${t.slug}`}
                className={`card card--interactive ${styles.topic}`}
              >
                <Icon name={t.icon} size={28} className={styles.icon} />
                <span className={styles.topicTitle}>{t.title}</span>
                <span className="text-sm text-muted">{t.summary}</span>
              </Link>
            </li>
          ))}
          <li>
            <Link
              href="/help/contact"
              className={`card card--interactive ${styles.topic}`}
            >
              <Icon name="support" size={28} className={styles.icon} />
              <span className={styles.topicTitle}>Contact us</span>
              <span className="text-sm text-muted">
                Can’t find an answer? Send us a message.
              </span>
            </Link>
          </li>
        </ul>
        <section aria-labelledby="common-title" className="stack">
          <h2 id="common-title" className={styles.h2}>
            Common questions
          </h2>
          <FaqList items={commonQuestions} />
        </section>
        <p className="text-xs text-muted">
          Help content is sample text until TechShop’s policies are finalised.
        </p>
      </div>
    </main>
  );
}
