"use client";

import { useActionState, useState } from "react";

import type { AuthStep } from "@techshop/bff";

import { staffSignIn } from "../actions";

/** Email + password, then the authenticator code (or a recovery code). */
export function StaffSignInForm({ next }: { next: string }) {
  const [state, action, pending] = useActionState(staffSignIn, {
    step: "start",
  } as AuthStep);
  const [recovery, setRecovery] = useState(false);
  const error = "error" in state ? state.error : undefined;
  const alert = error ? (
    <p id="signin-error" className="alert alert--error" role="alert">
      {error}
    </p>
  ) : null;

  if (state.step === "totp") {
    return (
      <form action={action} className="stack" noValidate>
        <input type="hidden" name="next" value={next} />
        <p>
          {recovery
            ? "Enter one of your recovery codes."
            : "Enter the 6-digit code from your authenticator app."}
        </p>
        {alert}
        <label className="field">
          <span className="field__label">
            {recovery ? "Recovery code" : "Authenticator code"}
          </span>
          <input
            name="code"
            className="input input--lg"
            inputMode={recovery ? "text" : "numeric"}
            autoComplete="one-time-code"
            maxLength={recovery ? 32 : 6}
            required
            autoFocus
            aria-invalid={Boolean(error)}
            aria-describedby={error ? "signin-error" : undefined}
          />
        </label>
        <button
          name="intent"
          value={recovery ? "recovery" : "totp"}
          className="btn btn--inverse btn--lg btn--block"
          disabled={pending}
        >
          {pending ? "Checking…" : "Sign in"}
        </button>
        <div className="cluster cluster--between">
          <button
            type="button"
            className="btn btn--ghost btn--sm"
            onClick={() => setRecovery((r) => !r)}
          >
            {recovery ? "Use the authenticator app" : "Use a recovery code"}
          </button>
          <button
            name="intent"
            value="restart"
            className="btn btn--ghost btn--sm"
            formNoValidate
            disabled={pending}
          >
            Start again
          </button>
        </div>
      </form>
    );
  }

  return (
    <form action={action} className="stack" noValidate>
      <input type="hidden" name="next" value={next} />
      {alert}
      <label className="field">
        <span className="field__label">Work email</span>
        <input
          name="email"
          type="email"
          className="input input--lg"
          autoComplete="username"
          required
          autoFocus
          aria-describedby={error ? "signin-error" : undefined}
        />
      </label>
      <label className="field">
        <span className="field__label">Password</span>
        <input
          name="password"
          type="password"
          className="input input--lg"
          autoComplete="current-password"
          required
        />
      </label>
      <button
        name="intent"
        value="password"
        className="btn btn--inverse btn--lg btn--block"
        disabled={pending}
      >
        {pending ? "Signing in…" : "Continue"}
      </button>
      <p className="text-xs text-muted">
        Staff accounts always need an authenticator app. Lost access? Ask an
        administrator to reset it.
      </p>
    </form>
  );
}
