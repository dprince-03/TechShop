"use client";

import { useActionState } from "react";

import { scanItem, type ScanState } from "../actions";

/** One line of a pick list: scan the serial (for devices) and confirm the quantity. */
export function ScanForm({
  pickListId,
  orderLineId,
  remaining,
  serialised,
}: {
  pickListId: string;
  orderLineId: string;
  remaining: number;
  serialised: boolean;
}) {
  const [state, action, pending] = useActionState(
    scanItem.bind(null, pickListId),
    {} as ScanState,
  );
  return (
    <form action={action} className="cluster">
      <input type="hidden" name="orderLineId" value={orderLineId} />
      {serialised ? (
        <input
          name="serial"
          className="input"
          aria-label="IMEI or serial number"
          placeholder="Scan IMEI / serial"
          required
          style={{ maxWidth: "14rem" }}
        />
      ) : null}
      <input
        name="quantity"
        type="number"
        className="input"
        aria-label="Quantity"
        min={1}
        max={remaining}
        defaultValue={serialised ? 1 : remaining}
        readOnly={serialised}
        style={{ width: "5rem" }}
      />
      <button className="btn btn--inverse btn--sm" disabled={pending}>
        {pending ? "Saving…" : "Picked"}
      </button>
      {state.message ? (
        <span
          className={state.ok ? "text-xs text-muted" : "field__error"}
          role={state.ok ? "status" : "alert"}
        >
          {state.message}
        </span>
      ) : null}
    </form>
  );
}
