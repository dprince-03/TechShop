"use client";

import { useState, useTransition } from "react";

import { cancelOrder, retryPayment, type ActionResult } from "../actions";
import { providerLabel } from "../status";

import styles from "./Orders.module.css";

/** Pay (or pay again) and cancel, for an order that is still awaiting payment. */
export function OrderActions({
  number,
  methods,
  canCancel,
  canPay,
}: {
  number: string;
  methods: string[];
  canCancel: boolean;
  canPay: boolean;
}) {
  const [provider, setProvider] = useState(methods[0] ?? "");
  const [result, setResult] = useState<ActionResult | null>(null);
  const [pending, startTransition] = useTransition();
  return (
    <div className="stack">
      {result?.message ? (
        <p
          className={`alert ${result.ok ? "alert--success" : "alert--error"}`}
          role={result.ok ? "status" : "alert"}
        >
          {result.message}
        </p>
      ) : null}
      {canPay && methods.length ? (
        <div className={styles.payRow}>
          <label className="field">
            <span className="field__label">Pay with</span>
            <select
              className="input"
              value={provider}
              onChange={(e) => setProvider(e.target.value)}
            >
              {methods.map((m) => (
                <option key={m} value={m}>
                  {providerLabel[m] ?? m}
                </option>
              ))}
            </select>
          </label>
          <button
            type="button"
            className="btn btn--primary"
            disabled={pending}
            onClick={() =>
              startTransition(async () =>
                setResult(await retryPayment(number, provider)),
              )
            }
          >
            {pending ? "Please wait…" : "Pay now"}
          </button>
        </div>
      ) : null}
      {canCancel ? (
        <button
          type="button"
          className="btn btn--ghost btn--sm"
          disabled={pending}
          onClick={() => {
            if (
              !window.confirm(`Cancel order ${number}? This can’t be undone.`)
            )
              return;
            startTransition(async () =>
              setResult(await cancelOrder(number, "Cancelled by customer")),
            );
          }}
        >
          Cancel order
        </button>
      ) : null}
    </div>
  );
}
