import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { formatMoney } from "@techshop/api-client";
import {
  categoryBySlug,
  deliveryEstimate,
  getProduct,
  getSubcategory,
  products,
  productSpecs,
  relatedProducts,
  warrantyFor,
} from "@techshop/fixtures";
import {
  Breadcrumbs,
  CategoryArt,
  Icon,
  Price,
  ProductCard,
  Rating,
  SectionHeader,
} from "@techshop/ui/components";

import { BuyBox } from "@/components/BuyBox";

import styles from "./product.module.css";

export const dynamicParams = false;

export function generateStaticParams() {
  return products.map((p) => ({ slug: p.slug }));
}

export async function generateMetadata(
  props: PageProps<"/p/[slug]">,
): Promise<Metadata> {
  const product = getProduct((await props.params).slug);
  if (!product) return {};
  return {
    title: `${product.name} | TechShop Market`,
    description: `${product.name}${product.keySpec ? ` — ${product.keySpec}` : ""}. ${formatMoney(product.price)} from ${product.seller.name}.`,
  };
}

const conditionLabel = {
  new: "New",
  "uk-used": "UK-used",
  refurbished: "Refurbished",
} as const;

export default async function ProductPage(props: PageProps<"/p/[slug]">) {
  const product = getProduct((await props.params).slug);
  if (!product) notFound();

  const category = categoryBySlug[product.category];
  const sub = product.subcategory
    ? getSubcategory(product.category, product.subcategory)
    : undefined;
  const delivery = deliveryEstimate(product);
  const related = relatedProducts(product);

  return (
    <main id="main" className={styles.page}>
      <div className="container container--wide stack stack--lg">
        <Breadcrumbs
          items={[
            { label: "Home", href: "/" },
            { label: category.name, href: `/c/${category.slug}` },
            ...(sub
              ? [{ label: sub.name, href: `/c/${category.slug}/${sub.slug}` }]
              : []),
            { label: product.name },
          ]}
        />

        <div className={styles.layout}>
          {/* ---------- Gallery (placeholder until product photos exist) ---------- */}
          <div className={styles.gallery}>
            <div className="frame">
              <CategoryArt category={product.category} size="lg" />
            </div>
            <p className="text-xs text-muted">Product photos coming soon.</p>
          </div>

          {/* ---------- Buying information ---------- */}
          <div className={styles.info}>
            <div className="stack stack--sm">
              <p className="eyebrow">{product.brand}</p>
              <h1 className={styles.title}>{product.name}</h1>
              {product.keySpec ? (
                <p className="text-muted">{product.keySpec}</p>
              ) : null}
              <div className="cluster">
                {product.rating ? (
                  <Rating
                    average={product.rating.average}
                    count={product.rating.count}
                  />
                ) : null}
                <span className="badge badge--outline">
                  {conditionLabel[product.condition]}
                </span>
                {product.badges?.includes("best-seller") ? (
                  <span className="badge">Best seller</span>
                ) : null}
              </div>
            </div>

            <div className={styles.priceBlock}>
              <Price
                price={product.price}
                compareAt={product.compareAtPrice}
                size="lg"
              />
              {product.stock === "low-stock" ? (
                <p className={styles.stock}>Only a few left in stock</p>
              ) : null}
              {product.stock === "in-stock" ? (
                <p className={styles.inStock}>In stock</p>
              ) : null}
            </div>

            <BuyBox product={product} />

            <ul role="list" className={styles.facts}>
              <li className={styles.fact}>
                <Icon
                  name={
                    product.seller.type === "techshop" ? "verified" : "store"
                  }
                  size={22}
                  className={styles.factIcon}
                />
                <div>
                  <p className={styles.factTitle}>
                    Sold by {product.seller.name}
                  </p>
                  <p className="text-xs text-muted">
                    {product.seller.type === "techshop"
                      ? "Sold and shipped by TechShop"
                      : `Verified marketplace seller${product.seller.rating ? ` · rated ${product.seller.rating.toFixed(1)}/5` : ""}`}
                  </p>
                </div>
              </li>
              <li className={styles.fact}>
                <Icon name="truck" size={22} className={styles.factIcon} />
                <div>
                  <p className={styles.factTitle}>
                    Delivery from {formatMoney(delivery.fee)}
                  </p>
                  <p className="text-xs text-muted">
                    Lagos: {delivery.lagos} · Other states: {delivery.elsewhere}
                  </p>
                </div>
              </li>
              <li className={styles.fact}>
                <Icon name="shield" size={22} className={styles.factIcon} />
                <div>
                  <p className={styles.factTitle}>{warrantyFor(product)}</p>
                  <p className="text-xs text-muted">
                    Returns accepted if the item is faulty or not as described.
                  </p>
                </div>
              </li>
              <li className={styles.fact}>
                <Icon name="lock" size={22} className={styles.factIcon} />
                <div>
                  <p className={styles.factTitle}>Secure payment</p>
                  <p className="text-xs text-muted">
                    Paystack, OPay or Moniepoint. Your card details are never
                    stored by sellers.
                  </p>
                </div>
              </li>
            </ul>
            <p className="text-xs text-muted">
              Delivery times, fees and warranty shown are sample information.
            </p>
          </div>
        </div>

        {/* ---------- Specifications ---------- */}
        <section aria-labelledby="specs-title" className={styles.specs}>
          <h2 id="specs-title" className={styles.sectionTitle}>
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

        {/* ---------- Related ---------- */}
        {related.length ? (
          <section aria-labelledby="related-title">
            <SectionHeader
              id="related-title"
              title={`More ${category.name.toLowerCase()}`}
              action={{
                label: `Shop all ${category.name.toLowerCase()}`,
                href: `/c/${category.slug}`,
              }}
            />
            <ul role="list" className="scroller">
              {related.map((p) => (
                <li key={p.id}>
                  <ProductCard product={p} href={`/p/${p.slug}`} />
                </li>
              ))}
            </ul>
          </section>
        ) : (
          <p>
            <Link href={`/c/${category.slug}`} className="link link--arrow">
              Shop all {category.name.toLowerCase()}
            </Link>
          </p>
        )}
      </div>
    </main>
  );
}
