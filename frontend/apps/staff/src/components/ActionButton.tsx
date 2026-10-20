"use client";

import { useState, useTransition } from "react";

import { confirmStepUp } from "@/features/auth/actions";
import type { StaffResult } from "@/lib/actions";

import styles from "./ActionButton.module.css";

type Props = {
  label: string;
  /** The Server Action to run (already bound to its arguments). */
  action: () => Promise<StaffResult>;
  /** Asked before running, for anything hard to undo. */
  confirm?: string;
  variant?: "primary" | "secondary" | "inverse" | "ghost";
  size?: "sm" | "md";
  disabled?: boolean;
};

/**
 * Runs one staff action and shows its result next to the button. When the API asks for a
 * recent authenticator check (step-up), it asks for the code and then retries the action.
 */
export function ActionButton({
  label,
  action,
  confirm,
  variant = "secondary",
  size = "sm",
  disabled,
}: Props) {
  const [result, setResult] = useState<StaffResult | null>(null);
  const [code, setCode] = useState("");
  const [pending, startTransition] = useTransition();

  const run = () =>
    startTransition(async () => {
      setResult(await action());
    });

  const confirmAndRetry = () =>
    startTransition(async () => {
      const up = await confirmStepUp(code);
      setCode("");
      if (!up.ok) {
        setResult({ ok: false, stepUp: true, message: up.message });
        return;
      }
      setResult(await action());
    });

  return (
    <div className={styles.wrap}>
      <button
        type="button"
        className={`btn btn--${variant} ${size === "sm" ? "btn--sm" : ""}`}
        disabled={disabled || pending}
        onClick={() => {
          if (confirm && !window.confirm(confirm)) return;
          run();
        }}
      >
        {pending ? "Working…" : label}
      </button>
      {result?.stepUp ? (
        <form
          className={styles.stepUp}
          onSubmit={(e) => {
            e.preventDefault();
            confirmAndRetry();
          }}
        >
          <label className="field">
            <span className="field__label">Authenticator code</span>
            <input
              className="input"
              inputMode="numeric"
              autoComplete="one-time-code"
              maxLength={6}
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
              autoFocus
              required
            />
          </label>
          <button
            className="btn btn--inverse btn--sm"
            disabled={pending || code.length !== 6}
          >
            Confirm and continue
          </button>
        </form>
      ) : null}
      {result?.message ? (
        <p
          className={result.ok ? styles.ok : styles.error}
          role={result.ok ? "status" : "alert"}
        >
          {result.message}
        </p>
      ) : null}
    </div>
  );
}
