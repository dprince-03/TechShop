import Link from "next/link";

import { formatKobo, type Schema } from "@techshop/api-client";
import {
  Breadcrumbs,
  CategoryArt,
  Icon,
  Price,
  ProductCard,
  Rating,
  SectionHeader,
} from "@techshop/ui/components";

import { getCategories, listProducts } from "@/features/catalogue/api";
import { artFor, conditionLabel, toCard } from "@/features/catalogue/mappers";
import { savedProductIds } from "@/features/saved/server";

import { BuyBox } from "./components/BuyBox";
import styles from "./ProductDetail.module.css";

type Detail = Schema<"ProductDetail">;
type Attribute = Schema<"CatalogListAttributeValuesRow">;

/** "256 GB", "Yes", or the text value — whichever the attribute carries. */
function attributeValue(a: Attribute): string | null {
  if (a.valueText) return a.valueText;
  if (a.valueNumber !== null)
    return `${a.valueNumber.toLocaleString("en-NG")}${a.unit ? ` ${a.unit}` : ""}`;
  if (a.valueBool !== null) return a.valueBool ? "Yes" : "No";
  return null;
}

/** Finds the top-level category (and the sub, if any) for the product's category slug. */
async function placeInTree(slug: string) {
  const tree = await getCategories();
  for (const parent of tree) {
    if (parent.slug === slug) return { parent, sub: undefined };
    const sub = parent.subcategories.find((s) => s.apiSlug === slug);
    if (sub) return { parent, sub };
  }
  return { parent: undefined, sub: undefined };
}

export async function ProductDetail({ product }: { product: Detail }) {
  const offers = product.offers ?? [];
  const best = offers.find((o) => o.buyBox) ?? offers[0];
  const { parent, sub } = await placeInTree(product.category.Slug);
  const art = artFor(parent?.slug);
  const [saved, related] = await Promise.all([
    savedProductIds(),
    parent
      ? listProducts({
          category: parent.slug,
          brands: [],
          conditions: [],
          inStock: true,
          sort: "best_selling",
          limit: 9,
        }).catch(() => null)
      : Promise.resolve(null),
  ]);
  const relatedCards = (related?.items ?? [])
    .filter((p) => p.id !== product.id)
    .slice(0, 8)
    .map((p) => toCard(p, parent?.slug));
  const specs = (product.attributes ?? [])
    .map((a) => ({ label: a.label, value: attributeValue(a) }))
    .filter((s): s is { label: string; value: string } => Boolean(s.value));
  const conditions = [...new Set(offers.map((o) => o.condition))];

  return (
    <main id="main" className={styles.page}>
      <div className="container container--wide stack stack--lg">
        <Breadcrumbs
          items={[
            { label: "Home", href: "/" },
            ...(parent
              ? [{ label: parent.name, href: `/c/${parent.slug}` }]
              : []),
            ...(parent && sub
              ? [{ label: sub.name, href: `/c/${parent.slug}/${sub.slug}` }]
              : []),
            { label: product.name },
          ]}
        />

        <div className={styles.layout}>
          {/* ---------- Gallery ---------- */}
          <div className={styles.gallery}>
            {(product.images ?? []).length ? (
              <div className="frame">
                {/* eslint-disable-next-line @next/next/no-img-element -- asset host not configured for next/image yet */}
                <img
                  src={product.images![0].url}
                  alt={product.images![0].alt || product.name}
                />
              </div>
            ) : (
              <>
                <div className="frame">
                  <CategoryArt category={art} size="lg" />
                </div>
                <p className="text-xs text-muted">
                  Product photos coming soon.
                </p>
              </>
            )}
          </div>

          {/* ---------- Buying information ---------- */}
          <div className={styles.info}>
            <div className="stack stack--sm">
              {product.brand ? (
                <p className="eyebrow">{product.brand}</p>
              ) : null}
              <h1 className={styles.title}>{product.name}</h1>
              <div className="cluster">
                {product.ratingCount > 0 && product.ratingAvg ? (
                  <Rating
                    average={product.ratingAvg}
                    count={product.ratingCount}
                  />
                ) : (
                  <span className="text-sm text-muted">No reviews yet</span>
                )}
                {conditions.map((c) => (
                  <span key={c} className="badge badge--outline">
                    {conditionLabel[c] ?? c}
                  </span>
                ))}
              </div>
            </div>

            {best ? (
              <div className={styles.priceBlock}>
                <Price
                  price={{ amount: best.price, currency: "NGN" }}
                  compareAt={
                    best.compareAt && best.compareAt > best.price
                      ? { amount: best.compareAt, currency: "NGN" }
                      : undefined
                  }
                  size="lg"
                />
                {product.inStock ? (
                  <p className={styles.inStock}>In stock</p>
                ) : null}
              </div>
            ) : null}

            <BuyBox
              productId={product.id}
              productName={product.name}
              offers={offers}
              inStock={product.inStock}
              saved={saved.has(product.id)}
            />

            {best ? (
              <ul role="list" className={styles.facts}>
                <li className={styles.fact}>
                  <Icon
                    name={
                      best.sellerType === "first_party" ? "verified" : "store"
                    }
                    size={22}
                    className={styles.factIcon}
                  />
                  <div>
                    <p className={styles.factTitle}>
                      Sold by {best.sellerName}
                    </p>
                    <p className="text-xs text-muted">
                      {best.fulfilledBy === "techshop"
                        ? "Shipped by TechShop"
                        : `Shipped by the seller${best.sellerRating ? ` · rated ${best.sellerRating.toFixed(1)}/5` : ""}`}
                    </p>
                  </div>
                </li>
                <li className={styles.fact}>
                  <Icon name="truck" size={22} className={styles.factIcon} />
                  <div>
                    <p className={styles.factTitle}>
                      Delivery from {formatKobo(best.deliveryFee)}
                    </p>
                    <p className="text-xs text-muted">
                      The exact fee for your address is shown at checkout.{" "}
                      <Link href="/delivery" className="link">
                        Delivery areas
                      </Link>
                    </p>
                  </div>
                </li>
                <li className={styles.fact}>
                  <Icon name="shield" size={22} className={styles.factIcon} />
                  <div>
                    <p className={styles.factTitle}>
                      {best.warrantyMonths > 0
                        ? `${best.warrantyMonths}-month warranty`
                        : "No warranty"}
                    </p>
                    <p className="text-xs text-muted">
                      From the seller, on this offer.
                    </p>
                  </div>
                </li>
                <li className={styles.fact}>
                  <Icon name="lock" size={22} className={styles.factIcon} />
                  <div>
                    <p className={styles.factTitle}>Secure payment</p>
                    <p className="text-xs text-muted">
                      Paystack, OPay or Moniepoint. Card details never reach the
                      seller.
                    </p>
                  </div>
                </li>
              </ul>
            ) : null}
          </div>
        </div>

        {offers.length > 1 ? (
          <section aria-labelledby="offers-title" className={styles.specs}>
            <h2 id="offers-title" className={styles.sectionTitle}>
              All offers ({offers.length})
            </h2>
            <div
              role="region"
              aria-labelledby="offers-title"
              tabIndex={0}
              className={styles.tableWrap}
            >
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th scope="col">Seller</th>
                    <th scope="col">Option</th>
                    <th scope="col">Condition</th>
                    <th scope="col">Warranty</th>
                    <th scope="col" className={styles.num}>
                      Price
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {offers.map((o) => (
                    <tr key={o.listingId}>
                      <td>
                        {o.sellerName}
                        {o.buyBox ? (
                          <span className="badge badge--accent">
                            Best offer
                          </span>
                        ) : null}
                      </td>
                      <td>{o.variantName}</td>
                      <td>
                        {conditionLabel[o.condition] ?? o.condition}
                        {o.grade ? ` (grade ${o.grade})` : ""}
                      </td>
                      <td>
                        {o.warrantyMonths ? `${o.warrantyMonths} months` : "—"}
                      </td>
                      <td className={`${styles.num} tabular-nums`}>
                        {formatKobo(o.price)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        ) : null}

        {product.description ? (
          <section aria-labelledby="about-title" className={styles.specs}>
            <h2 id="about-title" className={styles.sectionTitle}>
              About this product
            </h2>
            <p className="measure">{product.description}</p>
          </section>
        ) : null}

        {specs.length ? (
          <section aria-labelledby="specs-title" className={styles.specs}>
            <h2 id="specs-title" className={styles.sectionTitle}>
              Specifications
            </h2>
            <dl className={styles.specList}>
              {specs.map((s) => (
                <div key={s.label} className={styles.specRow}>
                  <dt>{s.label}</dt>
                  <dd>{s.value}</dd>
                </div>
              ))}
            </dl>
          </section>
        ) : null}

        {(product.accessoriesThatFit ?? []).length ? (
          <section aria-labelledby="fits-title" className={styles.specs}>
            <h2 id="fits-title" className={styles.sectionTitle}>
              Accessories that fit
            </h2>
            <ul role="list" className="cluster">
              {product.accessoriesThatFit!.map((a) => (
                <li key={a.id}>
                  <Link
                    href={`/p/${a.slug}`}
                    className="btn btn--sm btn--secondary"
                  >
                    {a.name}
                  </Link>
                </li>
              ))}
            </ul>
          </section>
        ) : null}

        {relatedCards.length && parent ? (
          <section aria-labelledby="related-title">
            <SectionHeader
              id="related-title"
              title={`More ${parent.name.toLowerCase()}`}
              action={{
                label: `Shop all ${parent.name.toLowerCase()}`,
                href: `/c/${parent.slug}`,
              }}
            />
            <ul role="list" className="scroller">
              {relatedCards.map((p) => (
                <li key={p.id}>
                  <ProductCard product={p} href={`/p/${p.slug}`} />
                </li>
              ))}
            </ul>
          </section>
        ) : null}
      </div>
    </main>
  );
}
