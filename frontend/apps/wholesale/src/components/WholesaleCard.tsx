import Link from "next/link";

import { formatMoney, type WholesaleProduct } from "@techshop/api-client";
import { CategoryArt } from "@techshop/ui/components";

import { TierTable } from "./TierTable";
import styles from "./WholesaleCard.module.css";

/** Bulk product: retail price for single units, tiered trade pricing from the minimum order quantity. */
export function WholesaleCard({ product }: { product: WholesaleProduct }) {
  return (
    <article className={`card ${styles.card}`}>
      <div className={styles.head}>
        <div className={`frame ${styles.thumb}`}>
          <CategoryArt category={product.category} size="sm" />
        </div>
        <div className="stack stack--sm">
          <h3 className={styles.title}>
            <Link href={`/p/${product.slug}`} className="link--subtle">
              {product.name}
            </Link>
          </h3>
          {product.keySpec ? (
            <p className="text-xs text-muted">{product.keySpec}</p>
          ) : null}
          <p className="text-xs">
            Retail{" "}
            <span className="tabular-nums">{formatMoney(product.price)}</span>{" "}
            each
          </p>
        </div>
      </div>
      <TierTable
        tiers={product.tiers}
        caption={`Trade prices for ${product.name}`}
      />
      <div className={styles.foot}>
        <span className="badge badge--outline">
          Min. order {product.minOrderQuantity}
        </span>
        <Link
          href={`/quote?product=${product.slug}`}
          className="btn btn--sm btn--secondary"
        >
          Add to quote
        </Link>
      </div>
    </article>
  );
}
