"use client";

import { useActionState } from "react";

import { inviteStaff, type InviteState } from "../actions";

const FIELDS = [
  ["firstName", "First name", "given-name"],
  ["lastName", "Last name", "family-name"],
  ["email", "Work email", "email"],
  ["employeeNo", "Employee number", "off"],
  ["jobTitle", "Job title", "organization-title"],
  ["department", "Department", "off"],
] as const;

export function InviteForm() {
  const [state, action, pending] = useActionState(
    inviteStaff,
    {} as InviteState,
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
      {state.inviteUrl ? (
        <p className="text-sm">
          Development only, invitation link:{" "}
          <a className="link" href={state.inviteUrl}>
            {state.inviteUrl}
          </a>
        </p>
      ) : null}
      <div className="grid" style={{ ["--grid-min" as string]: "14rem" }}>
        {FIELDS.map(([name, label, auto]) => (
          <label key={name} className="field">
            <span className="field__label">{label}</span>
            <input
              name={name}
              className="input"
              type={name === "email" ? "email" : "text"}
              autoComplete={auto}
              required
              aria-invalid={state.errors?.[name] ? true : undefined}
            />
            {state.errors?.[name] ? (
              <span className="field__error">{state.errors[name]}</span>
            ) : null}
          </label>
        ))}
      </div>
      <button className="btn btn--inverse" disabled={pending}>
        {pending ? "Inviting…" : "Send invitation"}
      </button>
    </form>
  );
}
