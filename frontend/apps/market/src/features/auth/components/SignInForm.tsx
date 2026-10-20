"use client";

import { useActionState, useState } from "react";

import type { AuthStep } from "@techshop/bff";

import { signInStep } from "../actions";

import styles from "./SignInForm.module.css";

/** Phone-first sign-in (OTP), with email + password as the alternative. */
export function SignInForm({ next }: { next: string }) {
  const [state, action, pending] = useActionState(signInStep, {
    step: "start",
  } as AuthStep);
  const [method, setMethod] = useState<"phone" | "password">("phone");
  const error = "error" in state ? state.error : undefined;

  const hidden = <input type="hidden" name="next" value={next} />;
  const errorBox = error ? (
    <p id="signin-error" className="alert alert--error" role="alert">
      {error}
    </p>
  ) : null;
  const described = error ? "signin-error" : undefined;

  if (state.step === "code" || state.step === "device") {
    const intent = state.step === "code" ? "verify-code" : "device-code";
    return (
      <form action={action} className={styles.form} noValidate>
        {hidden}
        <p>
          Enter the 6-digit code we sent to <strong>{state.hint}</strong>.
        </p>
        {errorBox}
        <label className="field">
          <span className="field__label">Code</span>
          <input
            name="code"
            inputMode="numeric"
            autoComplete="one-time-code"
            pattern="\d{6}"
            maxLength={6}
            required
            autoFocus
            aria-invalid={Boolean(error)}
            aria-describedby={described}
            className="input input--lg"
          />
        </label>
        <button
          name="intent"
          value={intent}
          className="btn btn--primary btn--lg btn--block"
          disabled={pending}
        >
          {pending ? "Checking…" : "Continue"}
        </button>
        <div className="cluster cluster--between">
          {state.step === "code" ? (
            <button
              name="intent"
              value="resend-code"
              className="btn btn--ghost btn--sm"
              disabled={pending}
              formNoValidate
            >
              Send a new code
            </button>
          ) : (
            <span />
          )}
          <button
            name="intent"
            value="restart"
            className="btn btn--ghost btn--sm"
            disabled={pending}
            formNoValidate
          >
            Use a different number
          </button>
        </div>
      </form>
    );
  }

  if (state.step === "profile") {
    return (
      <form action={action} className={styles.form}>
        {hidden}
        <p>Welcome to TechShop! What should we call you?</p>
        {errorBox}
        <label className="field">
          <span className="field__label">First name</span>
          <input
            name="firstName"
            autoComplete="given-name"
            required
            maxLength={60}
            autoFocus
            className="input input--lg"
          />
        </label>
        <label className="field">
          <span className="field__label">Last name</span>
          <input
            name="lastName"
            autoComplete="family-name"
            required
            maxLength={60}
            className="input input--lg"
          />
        </label>
        <button
          name="intent"
          value="profile"
          className="btn btn--primary btn--lg btn--block"
          disabled={pending}
        >
          {pending ? "Creating your account…" : "Create account"}
        </button>
      </form>
    );
  }

  return (
    <div className={styles.form}>
      <div role="group" aria-label="Sign-in method" className={styles.tabs}>
        <button
          type="button"
          aria-pressed={method === "phone"}
          className={styles.tab}
          onClick={() => setMethod("phone")}
        >
          Phone number
        </button>
        <button
          type="button"
          aria-pressed={method === "password"}
          className={styles.tab}
          onClick={() => setMethod("password")}
        >
          Email and password
        </button>
      </div>
      {method === "phone" ? (
        <form action={action} className={styles.form} noValidate>
          {hidden}
          {errorBox}
          <label className="field">
            <span className="field__label">Mobile number</span>
            <input
              name="phone"
              type="tel"
              inputMode="tel"
              autoComplete="tel-national"
              placeholder="0803 123 4567"
              required
              aria-invalid={Boolean(error)}
              aria-describedby={described ?? "phone-hint"}
              className="input input--lg"
            />
            <span id="phone-hint" className="field__hint">
              We’ll text you a code. New here? The same code creates your
              account.
            </span>
          </label>
          <button
            name="intent"
            value="request-code"
            className="btn btn--primary btn--lg btn--block"
            disabled={pending}
          >
            {pending ? "Sending…" : "Send code"}
          </button>
        </form>
      ) : (
        <form action={action} className={styles.form} noValidate>
          {hidden}
          {errorBox}
          <label className="field">
            <span className="field__label">Email or phone</span>
            <input
              name="identifier"
              autoComplete="username"
              required
              aria-describedby={described}
              className="input input--lg"
            />
          </label>
          <label className="field">
            <span className="field__label">Password</span>
            <input
              name="password"
              type="password"
              autoComplete="current-password"
              required
              className="input input--lg"
            />
          </label>
          <button
            name="intent"
            value="password"
            className="btn btn--primary btn--lg btn--block"
            disabled={pending}
          >
            {pending ? "Signing in…" : "Sign in"}
          </button>
        </form>
      )}
    </div>
  );
}
