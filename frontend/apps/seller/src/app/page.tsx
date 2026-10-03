import Link from "next/link";

import { Icon, SectionHeader } from "@techshop/ui/components";

import styles from "./page.module.css";

const included = [
  "Your own storefront page on TechShop Market",
  "Payments collected securely through Paystack, OPay and Moniepoint",
  "Payouts to your bank account",
  "Order, stock and payout tools in one dashboard",
  "Optional delivery handled by TechShop logistics",
];

const benefits = [
  {
    icon: "users",
    title: "Reach buyers nationwide",
    text: "List once and sell to customers across every state, on the web and the TechShop app.",
  },
  {
    icon: "lock",
    title: "Get paid safely",
    text: "Customers pay TechShop; we pay you after delivery. No chasing transfers.",
  },
  {
    icon: "truck",
    title: "Delivery handled",
    text: "Use TechShop dispatch riders and partners, or deliver yourself.",
  },
  {
    icon: "chart",
    title: "Tools that help you grow",
    text: "Track sales, stock and reviews, and see what customers are searching for.",
  },
] as const;

const steps = [
  {
    title: "Register",
    text: "Create your seller account with your phone number and email.",
  },
  {
    title: "Verify",
    text: "Upload your ID (and CAC details for businesses). We review within a few working days.",
  },
  {
    title: "List products",
    text: "Add photos, specs, condition and price. Used devices need proof of ownership.",
  },
  {
    title: "Sell & get paid",
    text: "Ship or hand over orders, then receive payouts to your bank account.",
  },
];

const requirements = [
  {
    icon: "user",
    text: "A valid government ID (NIN slip, international passport, driver’s licence or voter’s card)",
  },
  {
    icon: "wallet",
    text: "A Nigerian bank account in your name or your business’s name",
  },
  {
    icon: "building",
    text: "CAC registration documents — registered businesses only",
  },
  { icon: "phone", text: "A phone number and email you check regularly" },
  {
    icon: "shield",
    text: "For used phones and laptops: proof of ownership and a clean IMEI/serial check",
  },
] as const;

const faqs = [
  {
    q: "Can I sell as an individual?",
    a: "Yes. Individuals can sell new or used items with a valid ID and a bank account in their name. Registered businesses also provide CAC documents.",
  },
  {
    q: "How much does it cost to sell?",
    a: "There’s no fee to register. We charge a commission on each sale, which varies by category. You’ll see the exact fee for a product before you list it.",
  },
  {
    q: "When do I get paid?",
    a: "Payouts are made to your bank account after the customer receives their order and the return window closes.",
  },
  {
    q: "Do I have to handle delivery?",
    a: "No. You can drop orders at a TechShop hub or have our riders pick them up. You can also deliver yourself in your area.",
  },
  {
    q: "What can’t I sell?",
    a: "Counterfeit or stolen goods, devices that fail IMEI or serial checks, and anything prohibited by Nigerian law. Listings are reviewed and sellers who break the rules are removed.",
  },
];

export default function Home() {
  return (
    <main id="main">
      {/* ---------- Hero ---------- */}
      <section className="section" aria-labelledby="hero-title">
        <div className={`container ${styles.hero}`}>
          <div className="stack stack--lg">
            <div className="stack">
              <p className="eyebrow">TechShop Seller Centre</p>
              <h1 id="hero-title" className={styles.heroTitle}>
                Sell your tech to customers across Nigeria
              </h1>
              <p className="text-lg text-muted measure">
                Whether you run a gadget store or have a few devices to sell,
                list on TechShop Market and reach buyers nationwide.
              </p>
            </div>
            <div className="cluster">
              <Link href="/register" className="btn btn--primary btn--lg">
                Start selling
              </Link>
              <Link href="#how-title" className="btn btn--secondary btn--lg">
                See how it works
              </Link>
            </div>
            <p className="text-sm text-muted">
              Already a seller?{" "}
              <Link href="/sign-in" className="link">
                Sign in to Seller Centre
              </Link>
            </p>
          </div>

          <aside
            className={`card ${styles.included}`}
            aria-labelledby="included-title"
          >
            <h2 id="included-title" className={styles.includedTitle}>
              Every seller account includes
            </h2>
            <ul role="list" className="stack">
              {included.map((item) => (
                <li key={item} className={styles.check}>
                  <Icon name="check" size={18} className={styles.checkIcon} />
                  {item}
                </li>
              ))}
            </ul>
          </aside>
        </div>
      </section>

      {/* ---------- Benefits ---------- */}
      <section
        className="section section--surface"
        aria-labelledby="benefits-title"
      >
        <div className="container">
          <SectionHeader
            id="benefits-title"
            eyebrow="Why sell with us"
            title="We handle the hard parts"
          />
          <ul role="list" className={styles.benefits}>
            {benefits.map((b) => (
              <li key={b.title} className="card card--flat">
                <Icon name={b.icon} size={28} className={styles.accentIcon} />
                <h3 className={styles.itemTitle}>{b.title}</h3>
                <p className="text-sm text-muted">{b.text}</p>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- Who can sell ---------- */}
      <section className="section" aria-labelledby="who-title">
        <div className="container">
          <SectionHeader id="who-title" title="Who can sell" />
          <div className={styles.who}>
            <article className="card">
              <Icon name="store" size={28} className={styles.accentIcon} />
              <h3 className={styles.whoTitle}>Businesses</h3>
              <p className="text-muted">
                Gadget stores, phone shops, laptop dealers and electronics
                retailers with CAC registration. Sell new, UK-used and
                refurbished stock at scale.
              </p>
            </article>
            <article className="card">
              <Icon name="user" size={28} className={styles.accentIcon} />
              <h3 className={styles.whoTitle}>Individuals</h3>
              <p className="text-muted">
                Selling a phone, laptop or console you no longer need? Verify
                your ID and list it — we help with payment and delivery.
              </p>
            </article>
          </div>
        </div>
      </section>

      {/* ---------- How it works ---------- */}
      <section className="section section--surface" aria-labelledby="how-title">
        <div className="container">
          <SectionHeader
            id="how-title"
            eyebrow="How it works"
            title="Start selling in four steps"
          />
          <ol className={styles.steps}>
            {steps.map((s, i) => (
              <li key={s.title} className={styles.step}>
                <span className={styles.stepNumber} aria-hidden="true">
                  {i + 1}
                </span>
                <h3 className={styles.itemTitle}>{s.title}</h3>
                <p className="text-sm text-muted">{s.text}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      {/* ---------- Requirements ---------- */}
      <section className="section" aria-labelledby="req-title">
        <div className={`container ${styles.split}`}>
          <SectionHeader
            id="req-title"
            title="What you’ll need"
            description="Have these ready and verification goes much faster."
          />
          <ul role="list" className={styles.requirements}>
            {requirements.map((r) => (
              <li key={r.text} className={styles.requirement}>
                <Icon name={r.icon} size={22} className={styles.accentIcon} />
                <span>{r.text}</span>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- FAQ ---------- */}
      <section className="section section--surface" aria-labelledby="faq-title">
        <div className="container container--md">
          <SectionHeader id="faq-title" title="Questions sellers ask" />
          <div className={styles.faqs}>
            {faqs.map((f) => (
              <details key={f.q} className={styles.faq}>
                <summary className={styles.question}>
                  {f.q}
                  <Icon
                    name="chevronDown"
                    size={18}
                    className={styles.chevron}
                  />
                </summary>
                <p className={`text-muted ${styles.answer}`}>{f.a}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      {/* ---------- CTA ---------- */}
      <section className="section section--inverse" aria-labelledby="cta-title">
        <div className={`container ${styles.cta}`}>
          <h2 id="cta-title" className={styles.ctaTitle}>
            Ready to reach more customers?
          </h2>
          <Link href="/register" className={`btn btn--lg ${styles.ctaButton}`}>
            Create your seller account
          </Link>
        </div>
      </section>
    </main>
  );
}
