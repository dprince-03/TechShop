"use client";

import { useActionState } from "react";

import type { FormResult } from "@techshop/ui/components";

import { setRole } from "../actions";

/** Change one member's role in place, with the result announced next to it. */
export function RoleForm({
  shopId,
  userId,
  role,
  name,
  canGrantOwner,
}: {
  shopId: string;
  userId: string;
  role: string;
  name: string;
  canGrantOwner: boolean;
}) {
  const [state, run, pending] = useActionState(
    setRole.bind(null, shopId, userId),
    {} as FormResult,
  );
  const options = [
    "staff",
    "manager",
    "finance",
    ...(canGrantOwner ? ["owner"] : []),
  ];
  return (
    <form action={run} className="cluster">
      <select
        name="role"
        className="input"
        defaultValue={role}
        aria-label={`Role for ${name}`}
      >
        {options.map((r) => (
          <option key={r} value={r}>
            {r[0].toUpperCase() + r.slice(1)}
          </option>
        ))}
      </select>
      <button className="btn btn--ghost btn--sm" disabled={pending}>
        {pending ? "Saving…" : "Change"}
      </button>
      {state.message ? (
        <span role={state.ok ? "status" : "alert"} className="text-sm">
          {state.message}
        </span>
      ) : null}
    </form>
  );
}
