"use client";

import { useActionState } from "react";

import { packFulfilment, type PackState } from "../actions";

/** Pack one fulfilment: weigh it; the API compares with the catalogue and books the handover. */
export function PackForm({ fulfilmentId }: { fulfilmentId: string }) {
  const [state, action, pending] = useActionState(
    packFulfilment,
    {} as PackState,
  );
  if (state.ok)
    return (
      <p className="alert alert--success" role="status">
        {state.message}
      </p>
    );
  return (
    <form action={action} className="cluster">
      <input type="hidden" name="fulfilmentId" value={fulfilmentId} />
      <input
        name="weightGrams"
        className="input"
        inputMode="numeric"
        aria-label="Parcel weight in grams"
        placeholder="Weight (g)"
        required
        style={{ width: "8rem" }}
      />
      <input
        name="boxCode"
        className="input"
        aria-label="Box code (optional)"
        placeholder="Box (optional)"
        style={{ width: "8rem" }}
      />
      <button className="btn btn--inverse btn--sm" disabled={pending}>
        {pending ? "Packing…" : "Pack"}
      </button>
      {state.message ? (
        <span className="field__error" role="alert">
          {state.message}
        </span>
      ) : null}
    </form>
  );
}
