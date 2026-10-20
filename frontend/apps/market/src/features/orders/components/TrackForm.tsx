"use client";

import { useActionState } from "react";

import { trackOrder, type TrackResult } from "../actions";
import { formatDateTime, orderStatus, parcelStatus } from "../status";

import styles from "./Orders.module.css";

/** Guest order lookup (order number + checkout phone). The phone is posted, never put in a URL. */
export function TrackForm() {
  const [state, action, pending] = useActionState(trackOrder, {
    state: "idle",
  } as TrackResult);
  return (
    <div className="stack stack--lg">
      <form action={action} className={styles.form} noValidate>
        {state.state === "error" ? (
          <p className="alert alert--error" role="alert">
            {state.message}
          </p>
        ) : null}
        <label className="field">
          <span className="field__label">Order number</span>
          <input
            name="number"
            className="input"
            placeholder="TS-10482"
            autoComplete="off"
            autoCapitalize="characters"
            required
          />
          <span className="field__hint">
            It’s in your confirmation SMS and email.
          </span>
        </label>
        <label className="field">
          <span className="field__label">Phone number used at checkout</span>
          <input
            name="phone"
            type="tel"
            inputMode="tel"
            className="input"
            placeholder="0803 123 4567"
            autoComplete="tel-national"
            required
          />
        </label>
        <button className="btn btn--inverse" disabled={pending}>
          {pending ? "Looking up…" : "Track order"}
        </button>
      </form>
      {state.state === "found" ? (
        <section
          className="card card--roomy stack"
          aria-live="polite"
          aria-labelledby="track-result"
        >
          <h2 id="track-result" className={styles.h2}>
            {state.tracking.orderNumber}:{" "}
            {orderStatus[state.tracking.status]?.label ?? state.tracking.status}
          </h2>
          <p className="text-sm text-muted">
            Placed {formatDateTime(state.tracking.placedAt)}
          </p>
          {(state.tracking.parcels ?? []).length ? (
            <ul role="list" className={styles.parcels}>
              {state.tracking.parcels!.map((p, i) => (
                <li key={i}>
                  <strong>{p.seller}</strong> · {p.items} item
                  {p.items === 1 ? "" : "s"} ·{" "}
                  {parcelStatus[p.status] ?? p.status}
                  {p.deliveredAt ? ` on ${formatDateTime(p.deliveredAt)}` : ""}
                </li>
              ))}
            </ul>
          ) : null}
          {(state.tracking.history ?? []).length ? (
            <ol className={styles.timeline}>
              {state.tracking.history!.map((h, i) => (
                <li key={i}>
                  <span>
                    {orderStatus[h.status]?.label ??
                      parcelStatus[h.status] ??
                      h.status}
                  </span>
                  <span className="text-xs text-muted">
                    {formatDateTime(h.at)}
                  </span>
                </li>
              ))}
            </ol>
          ) : null}
        </section>
      ) : null}
    </div>
  );
}
