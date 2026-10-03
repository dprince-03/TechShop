import Link from "next/link";

import { categories, wholesaleProducts } from "@techshop/fixtures";
import { CategoryArt, Icon, SectionHeader } from "@techshop/ui/components";

import { TierTable } from "@/components/TierTable";
import { WholesaleCard } from "@/components/WholesaleCard";

import styles from "./page.module.css";

const featured = wholesaleProducts[1];

const benefits = [
  {
    icon: "tag",
    title: "Tiered trade pricing",
    text: "The more you buy, the lower the unit price — shown upfront on every product.",
  },
  {
    icon: "receipt",
    title: "VAT invoices",
    text: "Proper invoices for every order, ready for your accounts team.",
  },
  {
    icon: "wallet",
    title: "Credit terms",
    text: "Approved businesses can pay on invoice instead of upfront.",
  },
  {
    icon: "users",
    title: "Account manager",
    text: "One contact for quotes, large orders and after-sales support.",
  },
] as const;

const steps = [
  {
    title: "Open a business account",
    text: "Register your business and verify your CAC details.",
  },
  {
    title: "Build a quote or order",
    text: "Add products at trade prices, or request a custom quote.",
  },
  {
    title: "Pay your way",
    text: "Paystack, OPay, Moniepoint, bank transfer, or approved credit terms.",
  },
  {
    title: "Receive and track",
    text: "Bulk delivery to your office or warehouse, tracked end to end.",
  },
];

const audiences = [
  { icon: "building", label: "Offices & SMEs" },
  { icon: "briefcase", label: "Schools & universities" },
  { icon: "store", label: "Resellers & retailers" },
  { icon: "shield", label: "Government & NGOs" },
  { icon: "support", label: "Hospitals & clinics" },
  { icon: "users", label: "Churches & event centres" },
] as const;

export default function Home() {
  return (
    <main id="main">
      {/* ---------- Hero ---------- */}
      <section className="section" aria-labelledby="hero-title">
        <div className={`container ${styles.hero}`}>
          <div className="stack stack--lg">
            <div className="stack">
              <p className="eyebrow">Wholesale & retail</p>
              <h1 id="hero-title" className={styles.heroTitle}>
                Tech for your business, at trade prices
              </h1>
              <p className="text-lg text-muted measure">
                Buy single units at retail, or order in bulk with tiered
                pricing, VAT invoices and credit terms — straight from
                TechShop’s own stock.
              </p>
            </div>
            <div className="cluster">
              <Link
                href="/business/register"
                className="btn btn--primary btn--lg"
              >
                Open a business account
              </Link>
              <Link href="/quote" className="btn btn--secondary btn--lg">
                Request a quote
              </Link>
            </div>
          </div>

          <aside
            className={`card ${styles.heroCard}`}
            aria-labelledby="hero-card-title"
          >
            <div className={styles.heroCardHead}>
              <div className={`frame ${styles.heroThumb}`}>
                <CategoryArt category={featured.category} size="sm" />
              </div>
              <div>
                <p className="eyebrow">Example price breaks</p>
                <h2 id="hero-card-title" className={styles.heroCardTitle}>
                  {featured.name}
                </h2>
              </div>
            </div>
            <TierTable
              tiers={featured.tiers}
              caption={`Trade prices for ${featured.name}`}
            />
            <p className="text-xs text-muted">
              Prices exclude delivery. VAT shown on your invoice.
            </p>
          </aside>
        </div>
      </section>

      {/* ---------- Two ways to buy ---------- */}
      <section
        className="section section--surface"
        aria-labelledby="ways-title"
      >
        <div className="container">
          <SectionHeader id="ways-title" title="Two ways to buy" />
          <div className={styles.ways}>
            <article className="card card--flat">
              <Icon name="cart" size={28} className={styles.accentIcon} />
              <h3 className={styles.cardTitle}>Retail</h3>
              <p className="text-muted">
                Buy one or a few items at our standard price, with warranty and
                tracked delivery. No account needed.
              </p>
              <Link href="/retail" className="link link--arrow">
                Shop retail
              </Link>
            </article>
            <article className="card card--flat">
              <Icon name="box" size={28} className={styles.accentIcon} />
              <h3 className={styles.cardTitle}>Wholesale</h3>
              <p className="text-muted">
                Buy from the minimum order quantity and unlock lower unit prices
                as your order grows. Business account required.
              </p>
              <Link href="/wholesale" className="link link--arrow">
                See trade prices
              </Link>
            </article>
          </div>
        </div>
      </section>

      {/* ---------- Benefits ---------- */}
      <section className="section" aria-labelledby="benefits-title">
        <div className="container">
          <SectionHeader
            id="benefits-title"
            eyebrow="Why buy here"
            title="Built for business buyers"
          />
          <ul role="list" className={styles.benefits}>
            {benefits.map((b) => (
              <li key={b.title} className="stack stack--sm">
                <Icon name={b.icon} size={28} className={styles.accentIcon} />
                <h3 className={styles.benefitTitle}>{b.title}</h3>
                <p className="text-sm text-muted">{b.text}</p>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- Bulk offers ---------- */}
      <section
        className="section section--surface"
        aria-labelledby="bulk-title"
      >
        <div className="container">
          <SectionHeader
            id="bulk-title"
            title="Popular for bulk orders"
            description="Unit prices drop at each quantity break."
            action={{ label: "All trade prices", href: "/wholesale" }}
          />
          <ul role="list" className={`grid ${styles.bulkGrid}`}>
            {wholesaleProducts.map((p) => (
              <li key={p.id}>
                <WholesaleCard product={p} />
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- Categories ---------- */}
      <section className="section" aria-labelledby="categories-title">
        <div className="container">
          <SectionHeader id="categories-title" title="Shop by category" />
          <ul role="list" className={`grid ${styles.categories}`}>
            {categories
              .filter((c) => c.slug !== "cars")
              .map((c) => (
                <li key={c.slug}>
                  <Link href={`/c/${c.slug}`} className={styles.categoryTile}>
                    <CategoryArt category={c.slug} size="sm" />
                    <span className="stack stack--sm">
                      <span className={styles.categoryName}>{c.name}</span>
                      <span className="text-xs text-muted">
                        {c.subcategories.map((s) => s.name).join(" · ")}
                      </span>
                    </span>
                  </Link>
                </li>
              ))}
          </ul>
        </div>
      </section>

      {/* ---------- How it works ---------- */}
      <section className="section section--surface" aria-labelledby="how-title">
        <div className="container">
          <SectionHeader
            id="how-title"
            eyebrow="How it works"
            title="From quote to delivery in four steps"
          />
          <ol className={styles.steps}>
            {steps.map((s, i) => (
              <li key={s.title} className={styles.step}>
                <span className={styles.stepNumber} aria-hidden="true">
                  {i + 1}
                </span>
                <h3 className={styles.benefitTitle}>{s.title}</h3>
                <p className="text-sm text-muted">{s.text}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      {/* ---------- Who we serve ---------- */}
      <section className="section" aria-labelledby="serve-title">
        <div className="container">
          <SectionHeader id="serve-title" title="Who we supply" />
          <ul role="list" className={styles.audiences}>
            {audiences.map((a) => (
              <li key={a.label} className={styles.audience}>
                <Icon name={a.icon} size={22} className={styles.accentIcon} />
                {a.label}
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- CTA ---------- */}
      <section className="section section--inverse" aria-labelledby="cta-title">
        <div className={`container ${styles.cta}`}>
          <div className="stack">
            <h2 id="cta-title" className={styles.ctaTitle}>
              Planning a large order?
            </h2>
            <p className={styles.ctaText}>
              Tell us what you need and our B2B team will send a quote with
              delivery timelines.
            </p>
          </div>
          <Link href="/quote" className={`btn btn--lg ${styles.ctaButton}`}>
            Request a quote
          </Link>
        </div>
      </section>
    </main>
  );
}
