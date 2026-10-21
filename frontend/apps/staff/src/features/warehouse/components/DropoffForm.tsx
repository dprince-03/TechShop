"use client";

import { useActionState } from "react";

import { receiveDropoff, type DropoffState } from "../actions";

/** Hub desk: scan a seller's parcel label to take it in. */
export function DropoffForm({ warehouseId }: { warehouseId: string }) {
  const [state, action, pending] = useActionState(
    receiveDropoff.bind(null, warehouseId),
    {} as DropoffState,
  );
  return (
    <form action={action} className="stack">
      {state.message ? (
        <p
          className={`alert ${state.ok ? "alert--success" : "alert--error"}`}
          role={state.ok ? "status" : "alert"}
        >
          {state.message}
        </p>
      ) : null}
      <label className="field">
        <span className="field__label">Seller parcel label</span>
        <input
          name="label"
          className="input"
          autoComplete="off"
          autoCapitalize="characters"
          spellCheck={false}
          placeholder="TSS-…"
          required
        />
        <span className="field__hint">
          Check the parcel is sealed before you scan it. Once scanned, TechShop
          is responsible for it.
        </span>
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Receiving…" : "Receive parcel"}
      </button>
    </form>
  );
}
