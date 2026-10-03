"use client";

import { useState } from "react";

import { formatMoney, type Product } from "@techshop/api-client";
import { Icon, QuantityStepper } from "@techshop/ui/components";

import styles from "./BuyBox.module.css";

/**
 * Quantity + add to cart + save. INTERFACE ONLY: nothing is stored until the
 * cart API exists; the confirmation says so honestly.
 */
export function BuyBox({ product }: { product: Product }) {
  const [quantity, setQuantity] = useState(1);
  const [message, setMessage] = useState("");
  const outOfStock = product.stock === "out-of-stock";
  const max = product.stock === "low-stock" ? 3 : 10;

  const add = () =>
    setMessage(
      `Added ${quantity} × ${product.name} to your cart. (Preview: the cart isn’t saved yet.)`,
    );
  const save = () =>
    setMessage(
      `Saved ${product.name}. (Preview: saved items aren’t stored yet.)`,
    );

  return (
    <div className={styles.box}>
      {outOfStock ? (
        <p className={styles.out}>Out of stock</p>
      ) : (
        <QuantityStepper value={quantity} onChange={setQuantity} max={max} />
      )}
      <div className={styles.actions}>
        <button
          type="button"
          className="btn btn--primary btn--lg btn--block"
          onClick={add}
          disabled={outOfStock}
        >
          <Icon name="cart" size={20} />
          Add to cart
        </button>
        <button
          type="button"
          className="btn btn--secondary btn--lg btn--icon"
          onClick={save}
          aria-label={`Save ${product.name}`}
        >
          <Icon name="heart" />
        </button>
      </div>
      <p className={styles.message} role="status" aria-live="polite">
        {message}
      </p>

      {/* Mobile: price + add to cart stay in the thumb zone while scrolling */}
      <div className={styles.sticky} aria-hidden="true">
        <div className={styles.stickyPrice}>
          <span className="text-xs text-muted">Total</span>
          <span className="tabular-nums">
            {formatMoney({
              ...product.price,
              amount: product.price.amount * quantity,
            })}
          </span>
        </div>
        <button
          type="button"
          className="btn btn--primary"
          onClick={add}
          disabled={outOfStock}
          tabIndex={-1}
        >
          Add to cart
        </button>
      </div>
    </div>
  );
}
