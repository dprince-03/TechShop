import Link from "next/link";

import type { Product } from "@techshop/api-client";

import { CategoryArt } from "./CategoryArt";
import { Icon } from "./Icon";
import { Price } from "./Price";
import { Rating } from "./Rating";
import styles from "./ProductCard.module.css";

const conditionLabel = {
  new: null,
  "uk-used": "UK-used",
  refurbished: "Refurbished",
} as const;

/**
 * Product card: image → badges → title → key spec → rating → price → seller.
 * The title link is stretched over the whole card, so the card is one click target
 * while screen readers hear only the product name as the link.
 */
export function ProductCard({
  product,
  href,
}: {
  product: Product;
  href: string;
}) {
  const condition = conditionLabel[product.condition];
  return (
    <article className={`card card--interactive ${styles.card}`}>
      <div className="frame">
        <CategoryArt category={product.category} />
      </div>

      {product.badges?.includes("new") ||
      product.badges?.includes("best-seller") ||
      condition ? (
        <div className={styles.badges}>
          {product.badges?.includes("new") ? (
            <span className="badge badge--accent">New</span>
          ) : null}
          {product.badges?.includes("best-seller") ? (
            <span className="badge badge--outline">Best seller</span>
          ) : null}
          {condition ? <span className="badge">{condition}</span> : null}
        </div>
      ) : null}

      <h3 className={`line-clamp ${styles.title}`}>
        <Link href={href} className={styles.link}>
          {product.name}
        </Link>
      </h3>

      {product.keySpec ? (
        <p className={`text-xs text-muted ${styles.spec}`}>{product.keySpec}</p>
      ) : null}
      {product.rating ? (
        <Rating average={product.rating.average} count={product.rating.count} />
      ) : null}

      <div className={styles.footer}>
        <Price price={product.price} compareAt={product.compareAtPrice} />
        {product.stock === "low-stock" ? (
          <p className={styles.stock}>Only a few left</p>
        ) : null}
        <p className={styles.seller}>
          {product.seller.type === "techshop" ? (
            <>
              <Icon name="verified" size={14} className={styles.verified} />
              Sold by TechShop
            </>
          ) : (
            <>Sold by {product.seller.name}</>
          )}
        </p>
      </div>
    </article>
  );
}
