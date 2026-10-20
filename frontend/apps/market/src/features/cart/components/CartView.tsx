"use client";

import Link from "next/link";
import { useState, useTransition } from "react";

import { formatKobo } from "@techshop/api-client";
import {
  CategoryArt,
  EmptyState,
  Icon,
  QuantityStepper,
} from "@techshop/ui/components";

import { removeItem, setQuantity, type CartActionResult } from "../actions";
import type { Cart } from "../server";

import styles from "./CartView.module.css";

type Line = NonNullable<Cart["lines"]>[number];

const lineProblem: Record<string, string> = {
  unavailable: "No longer available — remove it to check out.",
  sold_out: "Sold out — remove it to check out.",
  not_enough_stock: "Not enough stock for this quantity.",
  min_quantity: "Below the minimum quantity for this item.",
};

const MAX_QUANTITY = 10;

/**
 * The live cart, grouped by seller (each seller ships separately). Totals always come from the
 * API; this component only sends changes and shows their result.
 */
export function CartView({
  cart,
  signedIn,
}: {
  cart: Cart | null;
  signedIn: boolean;
}) {
  const [pending, startTransition] = useTransition();
  const [announcement, setAnnouncement] = useState("");
  const [error, setError] = useState("");

  const run = (action: () => Promise<CartActionResult>) =>
    startTransition(async () => {
      const r = await action();
      setAnnouncement(r.ok ? r.message : "");
      setError(r.ok ? "" : r.message);
    });

  const lines = cart?.lines ?? [];
  if (!cart || !lines.length) {
    return (
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
    );
  }

  const groups = (cart.groups ?? []).map((g) => ({
    ...g,
    lines: lines.filter(
      (l) => l.sellerId === g.sellerId && l.fulfilledBy === g.fulfilledBy,
    ),
  }));
  const itemCount = lines.reduce((n, l) => n + l.quantity, 0);
  const checkoutHref = signedIn
    ? "/checkout"
    : `/sign-in?next=${encodeURIComponent("/checkout")}`;

  return (
    <div className={styles.layout} aria-busy={pending}>
      <p className="visually-hidden" role="status">
        {announcement}
      </p>

      <div className="stack stack--lg">
        {error ? (
          <p className={styles.error} role="alert">
            {error}
          </p>
        ) : null}
        {groups.map((group) => (
          <section
            key={`${group.sellerId}:${group.fulfilledBy}`}
            className={`card ${styles.group}`}
            aria-label={`Items from ${group.sellerName}`}
          >
            <div className={styles.groupHead}>
              <p className={styles.seller}>
                <Icon
                  name={group.fulfilledBy === "techshop" ? "verified" : "store"}
                  size={18}
                />
                Sold by {group.sellerName}
              </p>
              <p className="text-xs text-muted">
                Delivery {formatKobo(group.deliveryFeeKobo)}
                {cart.shipState ? ` to ${cart.shipState}` : ""}
              </p>
            </div>
            <ul role="list">
              {group.lines.map((line) => (
                <CartLine
                  key={line.cartItemId}
                  line={line}
                  disabled={pending}
                  onQuantity={(q) => run(() => setQuantity(line.cartItemId, q))}
                  onRemove={() => run(() => removeItem(line.cartItemId))}
                />
              ))}
            </ul>
          </section>
        ))}
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
            <dd>{formatKobo(cart.itemsKobo)}</dd>
          </div>
          {cart.discountKobo > 0 ? (
            <div className={styles.row}>
              <dt>{cart.promotion ?? "Discount"}</dt>
              <dd className={styles.saving}>
                −{formatKobo(cart.discountKobo)}
              </dd>
            </div>
          ) : null}
          <div className={styles.row}>
            <dt>
              Delivery{" "}
              <span className="text-xs text-muted">
                (estimate, {groups.length}{" "}
                {groups.length === 1 ? "seller" : "sellers"})
              </span>
            </dt>
            <dd>
              {cart.freeDelivery ? "Free" : formatKobo(cart.deliveryKobo)}
            </dd>
          </div>
          <div className={`${styles.row} ${styles.total}`}>
            <dt>Total</dt>
            <dd>{formatKobo(cart.totalKobo)}</dd>
          </div>
        </dl>
        {(cart.problems ?? []).length ? (
          <ul role="list" className={styles.problems}>
            {cart.problems!.map((p) => (
              <li key={p}>{p}</li>
            ))}
          </ul>
        ) : null}
        {cart.purchasable ? (
          <Link
            href={checkoutHref}
            className="btn btn--primary btn--lg btn--block"
          >
            {signedIn ? "Checkout" : "Sign in to check out"}
          </Link>
        ) : (
          <button
            type="button"
            className="btn btn--primary btn--lg btn--block"
            disabled
            aria-describedby="checkout-note"
          >
            Checkout
          </button>
        )}
        <p id="checkout-note" className="text-xs text-muted">
          {cart.purchasable
            ? "Delivery is confirmed for your address at checkout."
            : "Fix the items marked above to check out."}
        </p>
        <p className={`text-xs text-muted ${styles.secure}`}>
          <Icon name="lock" size={14} />
          Secure payment with Paystack, OPay or Moniepoint
        </p>
      </aside>
    </div>
  );
}

function CartLine({
  line,
  disabled,
  onQuantity,
  onRemove,
}: {
  line: Line;
  disabled: boolean;
  onQuantity: (q: number) => void;
  onRemove: () => void;
}) {
  const max = Math.max(
    1,
    Math.min(MAX_QUANTITY, line.available || MAX_QUANTITY),
  );
  return (
    <li className={styles.line}>
      <div className={`frame ${styles.thumb}`}>
        <CategoryArt size="sm" />
      </div>
      <div className={styles.lineInfo}>
        <Link href={`/p/${line.productSlug}`} className={styles.name}>
          {line.productName}
        </Link>
        <p className="text-xs text-muted">{line.variantName}</p>
        <p className="text-sm tabular-nums">
          {formatKobo(line.unitPriceKobo)} each
          {line.priceSource === "flash" ? (
            <span className="badge badge--promo">Flash price</span>
          ) : null}
        </p>
        {line.problem ? (
          <p className={styles.lineProblem}>
            {lineProblem[line.problem] ?? "This item needs attention."}
          </p>
        ) : null}
        <div className={styles.lineActions}>
          <QuantityStepper
            value={line.quantity}
            onChange={onQuantity}
            max={max}
            label={`Quantity of ${line.productName}`}
            labelHidden
          />
          <button
            type="button"
            className="btn btn--ghost btn--sm"
            onClick={onRemove}
            disabled={disabled}
          >
            <Icon name="trash" size={16} />
            Remove
          </button>
        </div>
      </div>
      <p className={styles.lineTotal}>{formatKobo(line.lineTotalKobo)}</p>
    </li>
  );
}
