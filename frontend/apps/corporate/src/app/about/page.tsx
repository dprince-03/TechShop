import type { Metadata } from "next";

import { Icon, PageHeader, Prose } from "@techshop/ui/components";
import { sites } from "@techshop/ui/sites";

import styles from "../inner.module.css";

export const metadata: Metadata = {
  title: "About | TechShop",
  description: "Who TechShop is and what we do.",
};

const lines = [
  {
    icon: "store",
    title: "Retail",
    text: "TechShop’s own stock — phones, laptops, gaming, smart home and office tech — sold online with warranty and tracked delivery.",
  },
  {
    icon: "users",
    title: "Marketplace",
    text: "Verified local vendors and individuals sell to customers nationwide, with TechShop handling payment and delivery.",
  },
  {
    icon: "briefcase",
    title: "Business & wholesale",
    text: "Trade pricing, VAT invoices and credit terms for offices, schools, resellers and institutions.",
  },
  {
    icon: "truck",
    title: "Supply & logistics",
    text: "We source genuine devices for partners and run our own dispatch and delivery operation.",
  },
] as const;

export default function AboutPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          eyebrow="About"
          title="Making genuine tech easy to buy in Nigeria"
          lead="TechShop sells, supplies and supports the technology Nigerian homes and businesses depend on."
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "About" }]}
        />
        <section aria-labelledby="what-title" className="stack">
          <h2 id="what-title" className={styles.h2}>
            What we do
          </h2>
          <ul role="list" className={styles.cards}>
            {lines.map((l) => (
              <li key={l.title} className="card">
                <Icon name={l.icon} size={28} className={styles.icon} />
                <h3 className={styles.cardTitle}>{l.title}</h3>
                <p className="text-muted">{l.text}</p>
              </li>
            ))}
          </ul>
        </section>
        <Prose>
          <h2>Why we exist</h2>
          <p>
            Buying tech shouldn’t mean worrying about fakes, hidden fees or what
            happens when something breaks. We verify what we sell and who sells
            with us, show clear prices in naira, and stay with customers after
            delivery through warranty, repairs and support.
          </p>
          <h2>Where to find us</h2>
          <p>
            Shop online at <a href={sites.market}>TechShop Market</a>.
            Businesses can buy through{" "}
            <a href={sites.wholesale}>Wholesale &amp; Retail</a>, and vendors
            can apply at the <a href={sites.seller}>Seller Centre</a>.
          </p>
        </Prose>
        <p className="text-xs text-muted">
          Draft company copy — to be reviewed by the TechShop team.
        </p>
      </div>
    </main>
  );
}
