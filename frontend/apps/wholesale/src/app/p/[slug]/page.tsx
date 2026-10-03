import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { formatMoney } from "@techshop/api-client";
import {
  categoryBySlug,
  getProduct,
  getWholesaleProduct,
  ownStock,
  productSpecs,
  wholesaleProducts,
} from "@techshop/fixtures";
import { Breadcrumbs, CategoryArt, Rating } from "@techshop/ui/components";

import { QuoteBox } from "@/components/QuoteBox";
import { TierTable } from "@/components/TierTable";

import styles from "./product.module.css";

const slugs = [
  ...new Set([...ownStock, ...wholesaleProducts].map((p) => p.slug)),
];

export const dynamicParams = false;

export function generateStaticParams() {
  return slugs.map((slug) => ({ slug }));
}

export async function generateMetadata(
  props: PageProps<"/p/[slug]">,
): Promise<Metadata> {
  const p = getProduct((await props.params).slug);
  return p ? { title: `${p.name} | TechShop Wholesale & Retail` } : {};
}

export default async function ProductPage(props: PageProps<"/p/[slug]">) {
  const { slug } = await props.params;
  if (!slugs.includes(slug)) notFound();
  const product = getProduct(slug)!;
  const bulk = getWholesaleProduct(slug);
  const category = categoryBySlug[product.category];

  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <Breadcrumbs
          items={[
            { label: "Home", href: "/" },
            { label: category.name, href: `/c/${category.slug}` },
            { label: product.name },
          ]}
        />
        <div className={styles.layout}>
          <div className="frame">
            <CategoryArt category={product.category} size="lg" />
          </div>
          <div className="stack stack--lg">
            <div className="stack stack--sm">
              <p className="eyebrow">{product.brand}</p>
              <h1 className={styles.title}>{product.name}</h1>
              {product.keySpec ? (
                <p className="text-muted">{product.keySpec}</p>
              ) : null}
              {product.rating ? (
                <Rating
                  average={product.rating.average}
                  count={product.rating.count}
                />
              ) : null}
              <p className="tabular-nums">
                Retail <strong>{formatMoney(product.price)}</strong> each
                {bulk ? (
                  <> · Minimum trade order {bulk.minOrderQuantity}</>
                ) : null}
              </p>
            </div>
            {bulk ? (
              <section
                aria-labelledby="tiers-title"
                className="stack stack--sm"
              >
                <h2 id="tiers-title" className={styles.h2}>
                  Trade prices
                </h2>
                <TierTable
                  tiers={bulk.tiers}
                  caption={`Trade prices for ${product.name}`}
                />
              </section>
            ) : (
              <p className="text-sm text-muted">
                Need this in bulk? Add it to a quote and we’ll price it for your
                quantity.
              </p>
            )}
            <QuoteBox
              name={product.name}
              retail={product.price}
              tiers={bulk?.tiers}
              minOrderQuantity={bulk ? 1 : undefined}
            />
            <p className="text-xs text-muted">
              Prices exclude delivery. VAT is shown on your invoice. Sample
              prices.
            </p>
          </div>
        </div>
        <section aria-labelledby="specs-title" className={styles.specs}>
          <h2 id="specs-title" className={styles.h2}>
            Specifications
          </h2>
          <dl className={styles.specList}>
            {productSpecs(product).map((s) => (
              <div key={s.label} className={styles.specRow}>
                <dt>{s.label}</dt>
                <dd>{s.value}</dd>
              </div>
            ))}
          </dl>
        </section>
      </div>
    </main>
  );
}
