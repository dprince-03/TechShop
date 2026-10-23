"use client";

import Link from "next/link";
import { useEffect, useRef, useState, useTransition } from "react";

import { formatKobo, type Schema } from "@techshop/api-client";
import { Icon } from "@techshop/ui/components";

import { AddressForm } from "@/features/account/components/AddressForm";
import { formatAddress } from "@/features/account/components/AddressBook";
import { providerLabel } from "@/features/orders/status";

import {
  placeOrder,
  previewCheckout,
  type PlaceResult,
  type PreviewResult,
} from "../actions";

import styles from "./CheckoutForm.module.css";

type Quote = Schema<"Quote">;
type Address = Schema<"IdentityAddress">;

const STEPS = ["Delivery", "Payment", "Review"] as const;

/**
 * Checkout: address → payment method → review and pay. Every total shown is the API's quote;
 * the order is placed with the quote's price hash, so the customer pays exactly what they saw.
 */
export function CheckoutForm({
  addresses,
  states,
  methods,
  initial,
  storeCredit = 0,
}: {
  addresses: Address[];
  states: Schema<"IdentityNigerianState">[];
  methods: string[];
  initial: Quote | null;
  /** Kobo available; it can pay an order only in full. */
  storeCredit?: number;
}) {
  const defaultAddress = addresses.find((a) => a.isDefault) ?? addresses[0];
  const [addressId, setAddressId] = useState(defaultAddress?.id ?? "");
  const [adding, setAdding] = useState(addresses.length === 0);
  const [coupon, setCoupon] = useState("");
  const [provider, setProvider] = useState(methods[0] ?? "");
  const [quote, setQuote] = useState<Quote | null>(initial);
  const [error, setError] = useState<{
    message: string;
    field?: string;
  } | null>(null);
  const [transfer, setTransfer] = useState<Extract<
    PlaceResult,
    { ok: true }
  > | null>(null);
  const [pending, startTransition] = useTransition();
  // One key per checkout attempt: retrying after a timeout can't create a second order.
  const idempotencyKey = useRef<string>(crypto.randomUUID());

  const reprice = (nextAddress = addressId, nextCoupon = coupon) =>
    startTransition(async () => {
      const r: PreviewResult = await previewCheckout(nextAddress, nextCoupon);
      if (r.ok) {
        setQuote(r.quote);
        setError(null);
      } else {
        setError({ message: r.message, field: r.field });
      }
    });

  // A newly saved address (the list refreshes from the server) becomes the selected one.
  const known = useRef(new Set(addresses.map((a) => a.id)));
  useEffect(() => {
    const added = addresses.find((a) => !known.current.has(a.id));
    known.current = new Set(addresses.map((a) => a.id));
    if (added) {
      setAddressId(added.id);
      reprice(added.id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- react to a new address only
  }, [addresses]);

  const pay = () =>
    startTransition(async () => {
      if (!quote) return;
      const r = await placeOrder({
        addressId,
        coupon,
        priceHash: quote.priceHash,
        provider,
        idempotencyKey: idempotencyKey.current,
      });
      if (r.ok) {
        setTransfer(r);
        return;
      }
      setError({ message: r.message });
      if (r.quote) {
        setQuote(r.quote);
        // New prices mean a new decision: the next attempt is a new order.
        idempotencyKey.current = crypto.randomUUID();
      }
    });

  if (transfer) {
    const t = transfer.payment.transfer;
    return (
      <section
        className="card card--roomy stack"
        aria-labelledby="transfer-title"
      >
        <h2 id="transfer-title" className={styles.h2}>
          Order {transfer.orderNumber} placed
        </h2>
        {t ? (
          <>
            <p>
              Transfer exactly{" "}
              <strong className="tabular-nums">
                {formatKobo(t.amountKobo)}
              </strong>{" "}
              to:
            </p>
            <dl className={styles.transfer}>
              <div>
                <dt>Bank</dt>
                <dd>{t.bankName}</dd>
              </div>
              <div>
                <dt>Account number</dt>
                <dd className="tabular-nums">{t.accountNumber}</dd>
              </div>
            </dl>
            <p className="text-sm text-muted">
              This account is for this order only and expires{" "}
              {new Date(t.expiresAt).toLocaleString("en-NG", {
                timeZone: "Africa/Lagos",
              })}
              . We’ll confirm as soon as the money arrives.
            </p>
          </>
        ) : (
          <p>
            {transfer.payment.message ??
              "Follow the payment instructions from your provider."}
          </p>
        )}
        <Link
          href={`/orders/${transfer.orderNumber}`}
          className="btn btn--inverse"
        >
          View order
        </Link>
      </section>
    );
  }

  const creditCovers = Boolean(quote && storeCredit >= quote.totalKobo);
  const canPay = Boolean(
    quote?.purchasable && addressId && provider && !adding,
  );

  return (
    <div className={styles.layout}>
      <div className="stack stack--lg">
        <ol className={styles.steps} aria-label="Checkout steps">
          {STEPS.map((s, i) => (
            <li key={s} className={styles.step}>
              <span className={styles.stepNum} aria-hidden="true">
                {i + 1}
              </span>
              {s}
            </li>
          ))}
        </ol>

        {error && !error.field ? (
          <p className="alert alert--error" role="alert">
            {error.message}
          </p>
        ) : null}

        <section
          className="card card--roomy stack"
          aria-labelledby="address-title"
        >
          <h2 id="address-title" className={styles.h2}>
            1. Delivery address
          </h2>
          {error?.field === "address" ? (
            <p className="alert alert--error" role="alert">
              {error.message}
            </p>
          ) : null}
          {addresses.length ? (
            <fieldset className={styles.choices}>
              <legend className="visually-hidden">Choose an address</legend>
              {addresses.map((a) => (
                <label key={a.id} className={styles.choice}>
                  <input
                    type="radio"
                    name="address"
                    value={a.id}
                    checked={addressId === a.id && !adding}
                    onChange={() => {
                      setAdding(false);
                      setAddressId(a.id);
                      reprice(a.id);
                    }}
                  />
                  <span>
                    <strong>{a.recipientName}</strong>
                    <span className="text-sm text-muted">
                      {" "}
                      ·{" "}
                      {formatAddress(
                        a,
                        states.find((s) => s.code === a.stateCode)?.name,
                      )}
                    </span>
                  </span>
                </label>
              ))}
            </fieldset>
          ) : null}
          {adding ? (
            <AddressForm
              states={states}
              submitLabel="Save and use this address"
              onDone={() => setAdding(false)}
            />
          ) : (
            <button
              type="button"
              className="btn btn--ghost btn--sm"
              onClick={() => setAdding(true)}
            >
              <Icon name="plus" size={16} /> Add a new address
            </button>
          )}
        </section>

        <section
          className="card card--roomy stack"
          aria-labelledby="payment-title"
        >
          <h2 id="payment-title" className={styles.h2}>
            2. Payment method
          </h2>
          {methods.length || creditCovers ? (
            <fieldset className={styles.choices}>
              <legend className="visually-hidden">Choose how to pay</legend>
              {creditCovers ? (
                <label className={styles.choice}>
                  <input
                    type="radio"
                    name="provider"
                    value="store_credit"
                    checked={provider === "store_credit"}
                    onChange={() => setProvider("store_credit")}
                  />
                  <span>
                    Store credit ({formatKobo(storeCredit)} available)
                  </span>
                </label>
              ) : null}
              {methods.map((m) => (
                <label key={m} className={styles.choice}>
                  <input
                    type="radio"
                    name="provider"
                    value={m}
                    checked={provider === m}
                    onChange={() => setProvider(m)}
                  />
                  <span>{providerLabel[m] ?? m}</span>
                </label>
              ))}
            </fieldset>
          ) : (
            <p className="alert alert--warning">
              Payments are unavailable right now. Please try again shortly.
            </p>
          )}
          {storeCredit > 0 && !creditCovers ? (
            <p className="text-xs text-muted">
              You have {formatKobo(storeCredit)} store credit. It can pay an
              order only when it covers the whole total.
            </p>
          ) : null}
          <p className="text-xs text-muted">
            <Icon name="lock" size={14} /> You’ll pay on the provider’s secure
            page. TechShop never sees your card details.
          </p>
        </section>

        <section
          className="card card--roomy stack"
          aria-labelledby="coupon-title"
        >
          <h2 id="coupon-title" className={styles.h2}>
            Discount code
          </h2>
          <form
            className={styles.coupon}
            onSubmit={(e) => {
              e.preventDefault();
              reprice();
            }}
          >
            <label className="field">
              <span className="field__label">
                Code <span className="text-muted">(optional)</span>
              </span>
              <input
                className="input"
                value={coupon}
                onChange={(e) => setCoupon(e.target.value)}
                autoComplete="off"
                autoCapitalize="characters"
                aria-invalid={error?.field === "coupon" || undefined}
                aria-describedby={
                  error?.field === "coupon" ? "coupon-error" : undefined
                }
              />
              {error?.field === "coupon" ? (
                <span id="coupon-error" className="field__error">
                  {error.message}
                </span>
              ) : null}
            </label>
            <button
              className="btn btn--secondary"
              disabled={pending || !addressId}
            >
              Apply
            </button>
          </form>
        </section>
      </div>

      <aside
        className={`card ${styles.summary}`}
        aria-labelledby="summary-title"
        aria-busy={pending}
      >
        <h2 id="summary-title" className={styles.h2}>
          3. Review and pay
        </h2>
        {quote ? (
          <>
            <ul role="list" className={styles.lines}>
              {(quote.lines ?? []).map((l) => (
                <li key={l.cartItemId} className={styles.line}>
                  <span>
                    {l.quantity} × {l.productName}
                    <span className="text-xs text-muted">
                      {" "}
                      · {l.variantName} · {l.sellerName}
                    </span>
                  </span>
                  <span className="tabular-nums">
                    {formatKobo(l.lineTotalKobo)}
                  </span>
                </li>
              ))}
            </ul>
            <dl className={styles.rows}>
              <div className={styles.row}>
                <dt>Items</dt>
                <dd>{formatKobo(quote.itemsKobo)}</dd>
              </div>
              {quote.discountKobo > 0 ? (
                <div className={styles.row}>
                  <dt>{quote.promotion ?? "Discount"}</dt>
                  <dd className={styles.saving}>
                    −{formatKobo(quote.discountKobo)}
                  </dd>
                </div>
              ) : null}
              {quote.couponKobo > 0 ? (
                <div className={styles.row}>
                  <dt>Code {quote.couponCode}</dt>
                  <dd className={styles.saving}>
                    −{formatKobo(quote.couponKobo)}
                  </dd>
                </div>
              ) : null}
              <div className={styles.row}>
                <dt>
                  Delivery
                  {quote.shipState
                    ? ` to ${states.find((s) => s.code === quote.shipState)?.name ?? quote.shipState}`
                    : ""}
                </dt>
                <dd>
                  {quote.freeDelivery ? "Free" : formatKobo(quote.deliveryKobo)}
                </dd>
              </div>
              <div className={styles.row}>
                <dt>VAT included</dt>
                <dd>{formatKobo(quote.vatKobo)}</dd>
              </div>
              <div className={`${styles.row} ${styles.total}`}>
                <dt>Total</dt>
                <dd>{formatKobo(quote.totalKobo)}</dd>
              </div>
            </dl>
            {(quote.problems ?? []).length ? (
              <ul role="list" className="alert alert--warning">
                {quote.problems!.map((p) => (
                  <li key={p}>{p}</li>
                ))}
              </ul>
            ) : null}
          </>
        ) : (
          <p className="text-sm text-muted">
            Choose a delivery address to see your total.
          </p>
        )}
        <button
          type="button"
          className="btn btn--primary btn--lg btn--block"
          disabled={!canPay || pending}
          onClick={pay}
        >
          {pending
            ? "Please wait…"
            : quote
              ? `Pay ${formatKobo(quote.totalKobo)}`
              : "Pay"}
        </button>
        <p className="text-xs text-muted">
          By placing your order you agree to the{" "}
          <Link href="/legal/terms" className="link">
            terms of sale
          </Link>
          .
        </p>
      </aside>
    </div>
  );
}
