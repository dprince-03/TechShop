import Link from "next/link";
import { connection } from "next/server";

import type { Product } from "@techshop/api-client";
import { cars } from "@techshop/fixtures";
import {
  CategoryArt,
  Icon,
  ProductCard,
  SampleDataNotice,
  SectionHeader,
} from "@techshop/ui/components";
import { sites } from "@techshop/ui/sites";

import { CarCard } from "@/features/cars/components/CarCard";
import { getCategories, listProducts } from "@/features/catalogue/api";
import { artFor, toCard } from "@/features/catalogue/mappers";
import type { ListingQuery } from "@/features/catalogue/query";

import styles from "./page.module.css";

const productHref = (slug: string) => `/p/${slug}`;

const trust = [
  {
    icon: "lock",
    title: "Secure payments",
    text: "Paystack, OPay and Moniepoint",
  },
  {
    icon: "truck",
    title: "Delivery nationwide",
    text: "Track every order to your door",
  },
  {
    icon: "verified",
    title: "Verified sellers",
    text: "Genuine products with warranty",
  },
  {
    icon: "returns",
    title: "Easy returns",
    text: "Hassle-free returns and refunds",
  },
] as const;

const SHELF_SIZE = 10;

const shelfQuery = (category?: string): ListingQuery => ({
  category,
  brands: [],
  conditions: [],
  inStock: true,
  sort: "best_selling",
  limit: SHELF_SIZE,
});

/** One page of cards, or an empty shelf when the API is unavailable (the page still renders). */
async function cards(
  kind: "search" | "deals" | "best-sellers",
  q: ListingQuery,
  art?: string,
): Promise<Product[]> {
  try {
    const r = await listProducts(q, kind);
    return (r.items ?? []).map((p) => toCard(p, art));
  } catch {
    return [];
  }
}

export default async function Home() {
  await connection();
  const categories = await getCategories();
  const productCategories = categories.filter((c) => c.kind !== "vehicle");
  const [deals, bestSellers, ...shelfItems] = await Promise.all([
    cards("deals", { ...shelfQuery(), sort: "discount" }),
    cards("best-sellers", shelfQuery()),
    ...productCategories.map((c) =>
      cards("search", shelfQuery(c.slug), c.slug),
    ),
  ]);
  const shelves = productCategories
    .map((category, i) => ({ category, items: shelfItems[i] }))
    .filter((s) => s.items.length > 0);
  const promoCategories = productCategories.slice(1, 3);
  return (
    <main id="main">
      {/* ---------- Hero ---------- */}
      <section className={styles.heroBand} aria-labelledby="hero-title">
        <div className={`container container--wide ${styles.hero}`}>
          <div className={styles.heroMain}>
            <p className="eyebrow">New season, new tech</p>
            <h1 id="hero-title" className={styles.heroTitle}>
              The tech you want, delivered across Nigeria
            </h1>
            <p className={styles.heroText}>
              Phones, laptops, gaming, smart home and more — from TechShop and
              verified sellers, with secure payment and tracked delivery.
            </p>
            <div className="cluster">
              <Link href="/deals" className="btn btn--primary btn--lg">
                Shop today’s deals
              </Link>
              <Link
                href="/c/phones"
                className={`btn btn--lg ${styles.heroSecondary}`}
              >
                Browse phones
              </Link>
            </div>
          </div>

          <div className={styles.heroSide}>
            {promoCategories.map((c) => (
              <Link key={c.slug} href={`/c/${c.slug}`} className={styles.promo}>
                <div className={styles.promoText}>
                  <p className="eyebrow">Shop by category</p>
                  <p className={styles.promoTitle}>{c.name}</p>
                  <span className="link link--arrow text-sm">Shop now</span>
                </div>
                <CategoryArt category={artFor(c.slug)} size="sm" />
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* ---------- Trust strip ---------- */}
      <section aria-label="Why shop with TechShop" className={styles.trustBand}>
        <ul role="list" className={`container container--wide ${styles.trust}`}>
          {trust.map((t) => (
            <li key={t.title} className={styles.trustItem}>
              <Icon name={t.icon} size={24} className={styles.trustIcon} />
              <div>
                <p className={styles.trustTitle}>{t.title}</p>
                <p className="text-xs text-muted">{t.text}</p>
              </div>
            </li>
          ))}
        </ul>
      </section>

      {/* ---------- Shop by category ---------- */}
      <section className="section" aria-labelledby="categories-title">
        <div className="container container--wide">
          <SectionHeader id="categories-title" title="Shop by category" />
          <ul role="list" className={styles.categoryGrid}>
            {categories.map((c) => (
              <li key={c.slug}>
                <Link href={`/c/${c.slug}`} className={styles.categoryTile}>
                  <CategoryArt category={artFor(c.slug)} size="sm" />
                  <span className={styles.categoryName}>{c.name}</span>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- Deals ---------- */}
      {deals.length ? (
        <section
          className="section section--surface"
          aria-labelledby="deals-title"
        >
          <div className="container container--wide">
            <SectionHeader
              id="deals-title"
              title="Deals"
              action={{ label: "See all deals", href: "/deals" }}
            />
            <ul role="list" className="scroller">
              {deals.map((p) => (
                <li key={p.id}>
                  <ProductCard product={p} href={productHref(p.slug)} />
                </li>
              ))}
            </ul>
          </div>
        </section>
      ) : null}

      {/* ---------- Category shelves ---------- */}
      {shelves.map(({ category, items }, i) => {
        const slug = category.slug;
        const titleId = `shelf-${slug}`;
        return (
          <section
            key={slug}
            className={`section ${i % 2 ? "section--surface" : ""}`}
            aria-labelledby={titleId}
          >
            <div className="container container--wide">
              <SectionHeader
                id={titleId}
                title={category.name}
                description={category.description || undefined}
                action={{
                  label: `Shop all ${category.name.toLowerCase()}`,
                  href: `/c/${slug}`,
                }}
              />
              <ul role="list" className={`cluster ${styles.chips}`}>
                {category.subcategories.map((sub) => (
                  <li key={sub.slug}>
                    <Link
                      href={`/c/${slug}/${sub.slug}`}
                      className={`btn btn--sm btn--secondary`}
                    >
                      {sub.name}
                    </Link>
                  </li>
                ))}
              </ul>
              <ul role="list" className="scroller">
                {items.map((p) => (
                  <li key={p.id}>
                    <ProductCard product={p} href={productHref(p.slug)} />
                  </li>
                ))}
              </ul>
            </div>
          </section>
        );
      })}

      {/* ---------- Cars (listings, not add-to-cart; sample until backend M7) ---------- */}
      <section
        className="section section--surface"
        aria-labelledby="cars-title"
      >
        <div className="container container--wide">
          <SectionHeader
            id="cars-title"
            eyebrow="Cars"
            title="Inspected cars from verified dealers"
            description="Book a viewing, see the inspection report, and ask about financing before you pay."
            action={{ label: "Browse all cars", href: "/c/cars" }}
          />
          <SampleDataNotice>
            Preview: car listings are sample data until the cars service is
            live.
          </SampleDataNotice>
          <ul role="list" className={`grid ${styles.carGrid}`}>
            {cars.map((car) => (
              <li key={car.id}>
                <CarCard car={car} />
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* ---------- Best sellers ---------- */}
      {bestSellers.length ? (
        <section className="section" aria-labelledby="best-title">
          <div className="container container--wide">
            <SectionHeader
              id="best-title"
              title="Best sellers"
              action={{ label: "See all", href: "/best-sellers" }}
            />
            <ul role="list" className="grid grid--catalog">
              {bestSellers.map((p) => (
                <li key={p.id}>
                  <ProductCard product={p} href={productHref(p.slug)} />
                </li>
              ))}
            </ul>
          </div>
        </section>
      ) : null}

      {/* ---------- Marketplace & business ---------- */}
      <section
        className="section section--inverse"
        aria-labelledby="sell-title"
      >
        <div className={`container container--wide ${styles.sell}`}>
          <div className="stack">
            <p className={`eyebrow ${styles.inverseEyebrow}`}>
              For sellers & businesses
            </p>
            <h2 id="sell-title" className={styles.sellTitle}>
              Grow with TechShop
            </h2>
            <p className={styles.inverseText}>
              Reach customers across Nigeria by selling on the marketplace, or
              buy in bulk at trade prices for your business.
            </p>
          </div>
          <div className={styles.sellCards}>
            <a href={sites.seller} className={styles.sellCard}>
              <Icon name="store" size={28} />
              <span className={styles.sellCardTitle}>Sell on TechShop</span>
              <span className="text-sm">
                List your products and get paid securely.
              </span>
              <span className="link--arrow text-sm">Start selling</span>
            </a>
            <a href={sites.wholesale} className={styles.sellCard}>
              <Icon name="briefcase" size={28} />
              <span className={styles.sellCardTitle}>
                Buy for your business
              </span>
              <span className="text-sm">
                Tiered pricing, invoices and quotes.
              </span>
              <span className="link--arrow text-sm">
                Open a business account
              </span>
            </a>
          </div>
        </div>
      </section>
    </main>
  );
}
