"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useTransition } from "react";

import { formatKobo, type Schema } from "@techshop/api-client";
import { Icon, QuantityStepper } from "@techshop/ui/components";

import { addToCart } from "@/features/cart/actions";
import { toggleSaved } from "@/features/saved/actions";

import styles from "./BuyBox.module.css";

type Offer = Schema<"Offer">;

const MAX_QUANTITY = 10;

/**
 * Variant + quantity + add to cart + save. The chosen offer is the buy box for that variant
 * (the API sorts offers buy box first); prices shown are the API's.
 */
export function BuyBox({
  productId,
  productName,
  offers,
  inStock,
  saved: initiallySaved,
}: {
  productId: string;
  productName: string;
  offers: Offer[];
  inStock: boolean;
  saved: boolean;
}) {
  const pathname = usePathname();
  const variants = [...new Map(offers.map((o) => [o.variantId, o])).values()];
  const [variantId, setVariantId] = useState(variants[0]?.variantId ?? "");
  const [quantity, setQuantity] = useState(1);
  const [message, setMessage] = useState<{ ok: boolean; text: string } | null>(
    null,
  );
  const [saved, setSaved] = useState(initiallySaved);
  const [needsSignIn, setNeedsSignIn] = useState(false);
  const [pending, startTransition] = useTransition();

  // Best offer for the chosen variant: the buy box if it is for this variant, else the first.
  const offer =
    offers.find((o) => o.variantId === variantId && o.buyBox) ??
    offers.find((o) => o.variantId === variantId);
  const available = inStock && Boolean(offer);

  const add = () =>
    offer &&
    startTransition(async () => {
      const r = await addToCart(offer.listingId, quantity);
      setMessage({
        ok: r.ok,
        text: r.ok
          ? `Added ${quantity} × ${productName} to your cart.`
          : r.message,
      });
    });

  const save = () =>
    startTransition(async () => {
      const r = await toggleSaved(productId);
      if (r.ok) {
        setSaved(r.saved);
        setMessage({
          ok: true,
          text: r.saved ? "Saved for later." : "Removed from saved items.",
        });
      } else {
        setNeedsSignIn(Boolean(r.signIn));
        setMessage({ ok: false, text: r.message });
      }
    });

  return (
    <div className={styles.box}>
      {variants.length > 1 ? (
        <label className={styles.field}>
          <span className="text-sm text-muted">Option</span>
          <select
            className={styles.select}
            value={variantId}
            onChange={(e) => setVariantId(e.target.value)}
          >
            {variants.map((v) => (
              <option key={v.variantId} value={v.variantId}>
                {v.variantName}
              </option>
            ))}
          </select>
        </label>
      ) : null}

      {offer && offer.price !== offers[0]?.price ? (
        <p className="text-sm tabular-nums">
          {offer.variantName}: <strong>{formatKobo(offer.price)}</strong>
        </p>
      ) : null}

      {available ? (
        <QuantityStepper
          value={quantity}
          onChange={setQuantity}
          max={MAX_QUANTITY}
        />
      ) : (
        <p className={styles.out}>Out of stock</p>
      )}
      <div className={styles.actions}>
        <button
          type="button"
          className="btn btn--primary btn--lg btn--block"
          onClick={add}
          disabled={!available || pending}
        >
          <Icon name="cart" size={20} />
          {pending ? "Adding…" : "Add to cart"}
        </button>
        <button
          type="button"
          className="btn btn--secondary btn--lg btn--icon"
          onClick={save}
          disabled={pending}
          aria-pressed={saved}
          aria-label={
            saved
              ? `Remove ${productName} from saved items`
              : `Save ${productName}`
          }
        >
          <Icon name="heart" className={saved ? styles.savedIcon : undefined} />
        </button>
      </div>
      <p
        className={message?.ok === false ? styles.messageError : styles.message}
        role="status"
        aria-live="polite"
      >
        {message?.text}{" "}
        {message?.ok ? (
          <Link href="/cart" className="link">
            View cart
          </Link>
        ) : needsSignIn ? (
          <Link
            href={`/sign-in?next=${encodeURIComponent(pathname)}`}
            className="link"
          >
            Sign in
          </Link>
        ) : null}
      </p>

      {/* Mobile: price + add to cart stay in the thumb zone while scrolling */}
      <div className={styles.sticky} aria-hidden="true">
        <div className={styles.stickyPrice}>
          <span className="text-xs text-muted">Total</span>
          <span className="tabular-nums">
            {offer ? formatKobo(offer.price * quantity) : "—"}
          </span>
        </div>
        <button
          type="button"
          className="btn btn--primary"
          onClick={add}
          disabled={!available || pending}
          tabIndex={-1}
        >
          Add to cart
        </button>
      </div>
    </div>
  );
}
