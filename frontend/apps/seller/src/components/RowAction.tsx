"use client";

import { useActionState } from "react";

import type { FormResult } from "@techshop/ui/components";

/** A one-button action on a row (set default, remove), with its result announced. */
export function RowAction({
  action,
  label,
  pendingLabel,
}: {
  action: (prev: FormResult) => Promise<FormResult>;
  label: string;
  pendingLabel: string;
}) {
  const [state, run, pending] = useActionState(action, {} as FormResult);
  return (
    <form action={run} className="cluster">
      <button className="btn btn--ghost btn--sm" disabled={pending}>
        {pending ? pendingLabel : label}
      </button>
      {state.message ? (
        <span
          role={state.ok ? "status" : "alert"}
          className={`text-sm ${state.ok ? "text-muted" : ""}`}
        >
          {state.message}
        </span>
      ) : null}
    </form>
  );
}
