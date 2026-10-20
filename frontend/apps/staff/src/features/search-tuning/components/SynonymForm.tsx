"use client";

import { useActionState } from "react";

import { createSynonym, type SynonymState } from "../actions";

export function SynonymForm() {
  const [state, action, pending] = useActionState(
    createSynonym,
    {} as SynonymState,
  );
  return (
    <form action={action} className="stack" noValidate>
      {state.message ? (
        <p
          className={`alert ${state.ok ? "alert--success" : "alert--error"}`}
          role="alert"
        >
          {state.message}
        </p>
      ) : null}
      <div className="grid" style={{ ["--grid-min" as string]: "12rem" }}>
        <label className="field">
          <span className="field__label">Word</span>
          <input
            name="target"
            className="input"
            placeholder="e.g. phone"
            required
          />
        </label>
        <label className="field">
          <span className="field__label">Other terms (comma separated)</span>
          <input
            name="terms"
            className="input"
            placeholder="e.g. mobile, handset"
            required
          />
        </label>
        <label className="field">
          <span className="field__label">Type</span>
          <select name="kind" className="input">
            <option value="two_way">Same meaning both ways</option>
            <option value="one_way">Terms also find the word</option>
          </select>
        </label>
      </div>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Adding…" : "Add synonym"}
      </button>
    </form>
  );
}
