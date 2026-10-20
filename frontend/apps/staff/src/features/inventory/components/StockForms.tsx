"use client";

import { useActionState } from "react";

import { adjustStock, receiveUnits, type FormResult } from "../actions";

const conditionOptions = (
  <>
    <option value="new">New</option>
    <option value="uk_used">UK-used</option>
    <option value="refurbished">Refurbished</option>
    <option value="open_box">Open box</option>
  </>
);

function Feedback({ state }: { state: FormResult }) {
  return state.message ? (
    <p
      className={`alert ${state.ok ? "alert--success" : "alert--error"}`}
      role="alert"
    >
      {state.message}
    </p>
  ) : null;
}

export function AdjustForm({
  warehouses,
}: {
  warehouses: { id: string; name: string }[];
}) {
  const [state, action, pending] = useActionState(
    adjustStock,
    {} as FormResult,
  );
  return (
    <form action={action} className="stack" noValidate>
      <Feedback state={state} />
      <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
        <label className="field">
          <span className="field__label">Warehouse</span>
          <select name="warehouseId" className="input">
            {warehouses.map((w) => (
              <option key={w.id} value={w.id}>
                {w.name}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span className="field__label">Variant ID</span>
          <input
            name="variantId"
            className="input"
            placeholder="From the stock table"
            required
          />
        </label>
        <label className="field">
          <span className="field__label">Change</span>
          <input
            name="delta"
            className="input"
            inputMode="numeric"
            placeholder="e.g. 10 or -2"
            required
          />
        </label>
        <label className="field">
          <span className="field__label">Reason</span>
          <select name="reason" className="input">
            <option value="purchase_receipt">Purchase receipt</option>
            <option value="adjustment">Count correction</option>
            <option value="write_off">Write-off (damaged, lost)</option>
          </select>
        </label>
        <label className="field">
          <span className="field__label">Condition</span>
          <select name="condition" className="input">
            {conditionOptions}
          </select>
        </label>
      </div>
      <label className="field">
        <span className="field__label">Note</span>
        <input name="note" className="input" required maxLength={300} />
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Saving…" : "Adjust stock"}
      </button>
    </form>
  );
}

export function ReceiveUnitsForm({
  warehouses,
}: {
  warehouses: { id: string; name: string }[];
}) {
  const [state, action, pending] = useActionState(
    receiveUnits,
    {} as FormResult,
  );
  return (
    <form action={action} className="stack" noValidate>
      <Feedback state={state} />
      <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
        <label className="field">
          <span className="field__label">Warehouse</span>
          <select name="warehouseId" className="input">
            {warehouses.map((w) => (
              <option key={w.id} value={w.id}>
                {w.name}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span className="field__label">Variant ID</span>
          <input name="variantId" className="input" required />
        </label>
        <label className="field">
          <span className="field__label">Condition</span>
          <select name="condition" className="input">
            {conditionOptions}
          </select>
        </label>
      </div>
      <label className="field">
        <span className="field__label">IMEIs or serial numbers</span>
        <textarea
          name="codes"
          className="input"
          rows={5}
          placeholder="Scan one per line"
          required
        />
        <span className="field__hint">
          15-digit numbers are saved as IMEIs; anything else as a serial number.
        </span>
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Receiving…" : "Receive units"}
      </button>
    </form>
  );
}
