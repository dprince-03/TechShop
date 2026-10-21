"use client";

import { useActionState, useState, useTransition } from "react";

import {
  confirmAuthenticator,
  startAuthenticator,
  type SetupState,
} from "../actions";

/** Set up an authenticator app: add the key, confirm a code, keep the recovery codes. */
export function AuthenticatorSetup({ next }: { next?: string }) {
  const [started, setStarted] = useState<SetupState | null>(null);
  const [starting, start] = useTransition();
  const [state, confirm, confirming] = useActionState(confirmAuthenticator, {});

  if (state.recoveryCodes)
    return (
      <div className="stack">
        <p role="status" className="alert alert--success">
          Your authenticator is on. From now on, signing in asks for its code.
        </p>
        <h2 className="text-xl">Your recovery codes</h2>
        <p>
          Each code works once if you lose your phone. Save them somewhere safe
          now; we won’t show them again.
        </p>
        <ul className="tabular-nums">
          {state.recoveryCodes.map((c) => (
            <li key={c}>
              <code>{c}</code>
            </li>
          ))}
        </ul>
        {next ? (
          <p>
            <a className="btn btn--inverse" href={next}>
              Continue
            </a>
          </p>
        ) : null}
      </div>
    );

  if (!started?.secret)
    return (
      <div className="stack">
        {started?.message ? (
          <p role="alert" className="alert alert--error">
            {started.message}
          </p>
        ) : null}
        <p>
          Use an authenticator app such as Google Authenticator or Microsoft
          Authenticator. You’ll add a key to it, then type the code it shows.
        </p>
        <div>
          <button
            className="btn btn--inverse"
            disabled={starting}
            onClick={() =>
              start(async () => setStarted(await startAuthenticator()))
            }
          >
            {starting ? "Starting…" : "Set up an authenticator"}
          </button>
        </div>
      </div>
    );

  return (
    <form action={confirm} className="stack">
      <ol className="stack">
        <li>
          In your authenticator app, choose <strong>Enter a setup key</strong>{" "}
          and type this key (time-based):
          <p>
            <code className="tabular-nums">
              {started.secret.match(/.{1,4}/g)?.join(" ")}
            </code>
          </p>
          <p className="text-sm text-muted">
            On this phone?{" "}
            <a className="link" href={started.otpauthUrl}>
              Open it in your authenticator app
            </a>
            .
          </p>
        </li>
        <li>Type the 6-digit code the app now shows.</li>
      </ol>
      {state.message ? (
        <p role="alert" className="alert alert--error">
          {state.message}
        </p>
      ) : null}
      <label className="field">
        <span className="field__label">Code from the app</span>
        <input
          name="code"
          className="input"
          inputMode="numeric"
          autoComplete="one-time-code"
          maxLength={6}
          required
        />
      </label>
      <div>
        <button className="btn btn--inverse" disabled={confirming}>
          {confirming ? "Checking…" : "Turn on"}
        </button>
      </div>
    </form>
  );
}
