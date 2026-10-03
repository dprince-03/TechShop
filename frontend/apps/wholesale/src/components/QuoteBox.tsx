"use client";

import { useState } from "react";

import { formatMoney, type Money, type PriceTier } from "@techshop/api-client";
import { QuantityStepper } from "@techshop/ui/components";

import styles from "./QuoteBox.module.css";

type Props = {
  name: string;
  retail: Money;
  tiers?: PriceTier[];
  minOrderQuantity?: number;
};

/** Unit price for a quantity: the best tier reached, otherwise retail. */
function unitPrice(
  quantity: number,
  retail: Money,
  tiers?: PriceTier[],
): { price: Money; tier?: PriceTier } {
  const tier = [...(tiers ?? [])]
    .reverse()
    .find((t) => quantity >= t.minQuantity);
  return tier ? { price: tier.unitPrice, tier } : { price: retail };
}

/** Quantity + live unit price + add to quote / buy at retail. INTERFACE ONLY. */
export function QuoteBox({ name, retail, tiers, minOrderQuantity }: Props) {
  const [quantity, setQuantity] = useState(minOrderQuantity ?? 1);
  const [message, setMessage] = useState("");
  const { price, tier } = unitPrice(quantity, retail, tiers);
  const total: Money = { ...price, amount: price.amount * quantity };
  const next = tiers?.find((t) => t.minQuantity > quantity);

  return (
    <div className={styles.box}>
      <QuantityStepper value={quantity} onChange={setQuantity} max={1000} />
      <dl className={styles.rows}>
        <div className={styles.row}>
          <dt>Unit price</dt>
          <dd>
            {formatMoney(price)}{" "}
            <span className="text-xs text-muted">
              {tier ? "trade" : "retail"}
            </span>
          </dd>
        </div>
        <div className={`${styles.row} ${styles.total}`}>
          <dt>Total</dt>
          <dd>{formatMoney(total)}</dd>
        </div>
      </dl>
      <p className="text-xs text-muted" aria-live="polite">
        {next
          ? `Order ${next.minQuantity} or more to pay ${formatMoney(next.unitPrice)} each.`
          : tier
            ? "You’re at the best trade price."
            : ""}
      </p>
      <div className={styles.actions}>
        <button
          type="button"
          className="btn btn--primary btn--lg"
          onClick={() =>
            setMessage(
              `Added ${quantity} × ${name} to your quote. (Preview: quotes aren’t saved yet.)`,
            )
          }
        >
          Add to quote
        </button>
        <button
          type="button"
          className="btn btn--secondary btn--lg"
          onClick={() =>
            setMessage(
              `Added ${quantity} × ${name} to your cart at ${tier ? "trade" : "retail"} price. (Preview: the cart isn’t saved yet.)`,
            )
          }
        >
          Buy now
        </button>
      </div>
      <p className={styles.message} role="status" aria-live="polite">
        {message}
      </p>
    </div>
  );
}
