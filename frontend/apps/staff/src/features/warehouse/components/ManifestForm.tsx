"use client";

import { useActionState } from "react";

import { createManifest, type ManifestState } from "../actions";

export function ManifestForm({
  warehouseId,
  riders,
  carriers,
}: {
  warehouseId: string;
  riders: { id: string; name: string }[];
  carriers: { id: string; name: string }[];
}) {
  const [state, action, pending] = useActionState(
    createManifest.bind(null, warehouseId),
    {} as ManifestState,
  );
  return (
    <form action={action} className="stack">
      {state.message ? (
        <p
          className={`alert ${state.ok ? "alert--success" : "alert--error"}`}
          role="alert"
        >
          {state.message}
        </p>
      ) : null}
      <label className="field">
        <span className="field__label">Collected by</span>
        <select name="target" className="input" defaultValue="" required>
          <option value="" disabled>
            Choose a rider or carrier
          </option>
          <optgroup label="Riders">
            {riders.map((r) => (
              <option key={r.id} value={`rider:${r.id}`}>
                {r.name}
              </option>
            ))}
          </optgroup>
          <optgroup label="Carriers">
            {carriers.map((c) => (
              <option key={c.id} value={`carrier:${c.id}`}>
                {c.name}
              </option>
            ))}
          </optgroup>
        </select>
      </label>
      <label className="field">
        <span className="field__label">Parcel labels</span>
        <textarea
          name="labels"
          className="input"
          rows={4}
          placeholder="Scan each label, one per line"
          required
        />
        <span className="field__hint">
          Each parcel can be on one manifest only.
        </span>
      </label>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Creating…" : "Create manifest"}
      </button>
    </form>
  );
}
