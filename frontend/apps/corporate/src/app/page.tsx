import Link from "next/link";

import { categories } from "@techshop/fixtures";
import { CategoryArt, Icon, SectionHeader } from "@techshop/ui/components";
import { sites } from "@techshop/ui/sites";

import styles from "./page.module.css";

const businesses = [
  {
    icon: "store",
    title: "TechShop Market",
    text: "Our online marketplace. Shop from TechShop and verified local sellers, with secure payment and tracked delivery.",
    href: sites.market,
    cta: "Visit the Market",
  },
  {
    icon: "box",
    title: "Wholesale & Retail",
    text: "TechShop’s own stock, at retail or in bulk — with trade pricing, VAT invoices and credit terms for businesses.",
    href: sites.wholesale,
    cta: "Buy for your business",
  },
  {
    icon: "users",
    title: "Seller marketplace",
    text: "Local vendors and individuals sell to customers nationwide, with TechShop handling payments and delivery.",
    href: sites.seller,
    cta: "Become a seller",
  },
  {
    icon: "truck",
    title: "Supply",
    text: "We supply retailers, resellers and institutions with genuine devices, sourced and warrantied by TechShop.",
    href: `${sites.wholesale}/quote`,
    cta: "Request supply",
  },
] as const;

const principles = [
  {
    icon: "verified",
    title: "Genuine, always",
    text: "Every product is sourced from trusted channels and every seller is verified before they can list.",
  },
  {
    icon: "tag",
    title: "Honest prices",
    text: "Clear prices in naira, real discounts, and delivery costs shown before you pay.",
  },
  {
    icon: "support",
    title: "Service that follows through",
    text: "Warranty, repairs, returns and support that stay with you long after delivery.",
  },
] as const;

export default function Home() {
  return (
    <main id="main">
      {/* ---------- Hero ---------- */}
      <section className={styles.hero} aria-labelledby="hero-title">
        <div className="container stack stack--lg">
          <p className="eyebrow">About TechShop</p>
          <h1 id="hero-title" className={`text-display ${styles.heroTitle}`}>
            Technology for every Nigerian home and business.
          </h1>
          <p className={`text-xl text-muted ${styles.heroText}`}>
            TechShop sells, supplies and supports the devices people depend on —
            from phones and laptops to workstations, smart homes and cars.
          </p>
          <div className="cluster">
            <a href={sites.market} className="btn btn--primary btn--lg">
              Shop TechShop Market
            </a>
            <Link href="/partners" className="btn btn--secondary btn--lg">
              Partner with us
            </Link>
          </div>
        </div>
      </section>

      {/* ---------- Businesses ---------- */}
      <section
        className="section section--surface"
        aria-labelledby="businesses-title"
      >
        <div className="container">
          <SectionHeader
            id="businesses-title"
            eyebrow="Our businesses"
            title="One company, four ways to work with us"
          />
          <ul role="list" className={styles.businesses}>
            {businesses.map((b) => (
              <li key={b.title}>
                <article className={`card card--flat ${styles.business}`}>
                  <Icon name={b.icon} size={32} className={styles.accentIcon} />
                  <h3 className={styles.businessTitle}>{b.title}</h3>
                  <p className="text-muted">{b.text}</p>
                  <a
                    href={b.href}
                    className={`link link--arrow ${styles.businessLink}`}
                  >
                    {b.cta}
                  </a>
                </article>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- What we sell ---------- */}
      <section className="section" aria-labelledby="sell-title">
        <div className="container">
          <SectionHeader
            id="sell-title"
            eyebrow="What we sell"
            title="Everything tech, under one roof"
            action={{ label: "Browse the Market", href: sites.market }}
          />
          <ul role="list" className={styles.categories}>
            {categories.map((c) => (
              <li key={c.slug} className={styles.category}>
                <CategoryArt category={c.slug} size="sm" />
                <span className={styles.categoryName}>{c.name}</span>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- Principles ---------- */}
      <section
        className="section section--surface"
        aria-labelledby="principles-title"
      >
        <div className="container">
          <SectionHeader
            id="principles-title"
            eyebrow="How we work"
            title="What you can count on"
          />
          <ul role="list" className={styles.principles}>
            {principles.map((p) => (
              <li key={p.title} className="stack stack--sm">
                <Icon name={p.icon} size={28} className={styles.accentIcon} />
                <h3 className={styles.principleTitle}>{p.title}</h3>
                <p className="text-muted">{p.text}</p>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- Careers ---------- */}
      <section
        className="section section--inverse"
        aria-labelledby="careers-title"
      >
        <div className={`container ${styles.careers}`}>
          <div className="stack">
            <p className={`eyebrow ${styles.inverseEyebrow}`}>Careers</p>
            <h2 id="careers-title" className={styles.careersTitle}>
              Build the future of tech retail in Nigeria
            </h2>
            <p className={styles.inverseText}>
              We’re growing teams across technology, operations, logistics,
              customer service and sales.
            </p>
          </div>
          <Link
            href="/careers"
            className={`btn btn--lg ${styles.inverseButton}`}
          >
            See open roles
          </Link>
        </div>
      </section>

      {/* ---------- Newsroom (empty state) ---------- */}
      <section className="section" aria-labelledby="news-title">
        <div className="container">
          <SectionHeader
            id="news-title"
            eyebrow="Newsroom"
            title="Latest from TechShop"
          />
          <div className={styles.empty}>
            <Icon name="file" size={32} className={styles.accentIcon} />
            <p className={styles.emptyTitle}>No announcements yet</p>
            <p className="text-sm text-muted">
              Press releases and company news will appear here. For media
              enquiries, contact our press team.
            </p>
            <Link href="/contact#press" className="link link--arrow text-sm">
              Contact the press team
            </Link>
          </div>
        </div>
      </section>

      {/* ---------- Contact ---------- */}
      <section
        className="section section--surface"
        aria-labelledby="contact-title"
      >
        <div className="container">
          <SectionHeader id="contact-title" title="Get in touch" />
          <ul role="list" className={styles.contacts}>
            <li className="card card--flat">
              <Icon name="support" size={24} className={styles.accentIcon} />
              <h3 className={styles.contactTitle}>Customer support</h3>
              <p className="text-sm text-muted">
                Orders, delivery, returns and warranty.
              </p>
              <a
                href={`${sites.market}/help`}
                className="link link--arrow text-sm"
              >
                Visit the help centre
              </a>
            </li>
            <li className="card card--flat">
              <Icon name="briefcase" size={24} className={styles.accentIcon} />
              <h3 className={styles.contactTitle}>Business sales</h3>
              <p className="text-sm text-muted">
                Bulk orders, quotes and supply partnerships.
              </p>
              <a
                href={`${sites.wholesale}/quote`}
                className="link link--arrow text-sm"
              >
                Request a quote
              </a>
            </li>
            <li className="card card--flat">
              <Icon name="mail" size={24} className={styles.accentIcon} />
              <h3 className={styles.contactTitle}>Press & partnerships</h3>
              <p className="text-sm text-muted">
                Media, brand partnerships and investor relations.
              </p>
              <Link href="/contact" className="link link--arrow text-sm">
                Contact us
              </Link>
            </li>
          </ul>
        </div>
      </section>
    </main>
  );
}
