"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import { formatMoney, naira, type Money } from "@techshop/api-client";
import { deliveryEstimate, type CartLine } from "@techshop/fixtures";
import {
  CategoryArt,
  EmptyState,
  Icon,
  QuantityStepper,
} from "@techshop/ui/components";

import styles from "./CartView.module.css";

const sum = (values: Money[]) =>
  naira(values.reduce((t, m) => t + m.amount, 0) / 100);

/**
 * Cart grouped by seller (each seller ships separately). INTERFACE ONLY:
 * changes live in this page's state and are lost on reload.
 */
export function CartView({ initial }: { initial: CartLine[] }) {
  const [lines, setLines] = useState(initial);
  const [announcement, setAnnouncement] = useState("");

  const groups = useMemo(() => {
    const bySeller = new Map<string, CartLine[]>();
    for (const line of lines) {
      const key = line.product.seller.id;
      bySeller.set(key, [...(bySeller.get(key) ?? []), line]);
    }
    return [...bySeller.values()];
  }, [lines]);

  const subtotal = sum(
    lines.map((l) => ({
      ...l.product.price,
      amount: l.product.price.amount * l.quantity,
    })),
  );
  const delivery = sum(groups.map((g) => deliveryEstimate(g[0].product).fee));
  const total = sum([subtotal, delivery]);
  const itemCount = lines.reduce((n, l) => n + l.quantity, 0);

  const setQuantity = (id: string, quantity: number) =>
    setLines((ls) =>
      ls.map((l) => (l.product.id === id ? { ...l, quantity } : l)),
    );

  const remove = (line: CartLine) => {
    setLines((ls) => ls.filter((l) => l.product.id !== line.product.id));
    setAnnouncement(`Removed ${line.product.name} from your cart.`);
  };

  if (!lines.length) {
    return (
      <>
        <p className="visually-hidden" role="status">
          {announcement}
        </p>
        <EmptyState
          icon="cart"
          title="Your cart is empty"
          action={
            <Link href="/" className="btn btn--primary">
              Continue shopping
            </Link>
          }
        >
          Items you add will appear here, grouped by seller.
        </EmptyState>
      </>
    );
  }

  return (
    <div className={styles.layout}>
      <p className="visually-hidden" role="status">
        {announcement}
      </p>

      <div className="stack stack--lg">
        {groups.map((group) => {
          const seller = group[0].product.seller;
          const est = deliveryEstimate(group[0].product);
          return (
            <section
              key={seller.id}
              className={`card ${styles.group}`}
              aria-label={`Items from ${seller.name}`}
            >
              <div className={styles.groupHead}>
                <p className={styles.seller}>
                  <Icon
                    name={seller.type === "techshop" ? "verified" : "store"}
                    size={18}
                  />
                  Sold by {seller.name}
                </p>
                <p className="text-xs text-muted">
                  Delivery {formatMoney(est.fee)} · Lagos{" "}
                  {est.lagos.toLowerCase()}
                </p>
              </div>
              <ul role="list">
                {group.map((line) => (
                  <li key={line.product.id} className={styles.line}>
                    <div className={`frame ${styles.thumb}`}>
                      <CategoryArt category={line.product.category} size="sm" />
                    </div>
                    <div className={styles.lineInfo}>
                      <Link
                        href={`/p/${line.product.slug}`}
                        className={styles.name}
                      >
                        {line.product.name}
                      </Link>
                      {line.product.keySpec ? (
                        <p className="text-xs text-muted">
                          {line.product.keySpec}
                        </p>
                      ) : null}
                      <p className="text-sm tabular-nums">
                        {formatMoney(line.product.price)} each
                      </p>
                      <div className={styles.lineActions}>
                        <QuantityStepper
                          value={line.quantity}
                          onChange={(q) => setQuantity(line.product.id, q)}
                          max={line.product.stock === "low-stock" ? 3 : 10}
                          label={`Quantity of ${line.product.name}`}
                          labelHidden
                        />
                        <button
                          type="button"
                          className="btn btn--ghost btn--sm"
                          onClick={() => remove(line)}
                        >
                          <Icon name="trash" size={16} />
                          Remove
                        </button>
                      </div>
                    </div>
                    <p className={styles.lineTotal}>
                      {formatMoney({
                        ...line.product.price,
                        amount: line.product.price.amount * line.quantity,
                      })}
                    </p>
                  </li>
                ))}
              </ul>
            </section>
          );
        })}
      </div>

      <aside
        className={`card ${styles.summary}`}
        aria-labelledby="summary-title"
      >
        <h2 id="summary-title" className={styles.summaryTitle}>
          Order summary
        </h2>
        <dl className={styles.rows}>
          <div className={styles.row}>
            <dt>Items ({itemCount})</dt>
            <dd>{formatMoney(subtotal)}</dd>
          </div>
          <div className={styles.row}>
            <dt>
              Delivery{" "}
              <span className="text-xs text-muted">
                (estimate, {groups.length}{" "}
                {groups.length === 1 ? "seller" : "sellers"})
              </span>
            </dt>
            <dd>{formatMoney(delivery)}</dd>
          </div>
          <div className={`${styles.row} ${styles.total}`}>
            <dt>Total</dt>
            <dd>{formatMoney(total)}</dd>
          </div>
        </dl>
        <button
          type="button"
          className="btn btn--primary btn--lg btn--block"
          disabled
          aria-describedby="checkout-note"
        >
          Checkout
        </button>
        <p id="checkout-note" className="text-xs text-muted">
          Checkout isn’t available in this preview. Payments with Paystack, OPay
          and Moniepoint come next.
        </p>
        <p className={`text-xs text-muted ${styles.secure}`}>
          <Icon name="lock" size={14} />
          Secure payment · Delivery fee confirmed at checkout
        </p>
      </aside>
    </div>
  );
}
