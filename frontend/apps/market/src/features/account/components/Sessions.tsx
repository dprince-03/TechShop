"use client";

import { useState, useTransition } from "react";

import type { Schema } from "@techshop/api-client";

import { revokeOtherSessions, revokeSession, type FormState } from "../actions";

import styles from "./Account.module.css";
import { FormFeedback } from "./FormFeedback";

const when = (iso: string) =>
  new Intl.DateTimeFormat("en-NG", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Africa/Lagos",
  }).format(new Date(iso));

/** Signed-in devices with remote sign-out (identity-access.md §5). */
export function Sessions({ sessions }: { sessions: Schema<"SessionDTO">[] }) {
  const [status, setStatus] = useState<FormState>({});
  const [pending, startTransition] = useTransition();
  const others = sessions.filter((s) => !s.current);
  return (
    <div className="stack">
      <FormFeedback state={status} />
      <ul role="list" className={styles.list}>
        {sessions.map((s) => (
          <li key={s.id} className={styles.item}>
            <div className="stack stack--sm">
              <p className={styles.itemTitle}>
                {s.deviceName || s.platform || "Unknown device"}
                {s.current ? (
                  <span className="badge badge--accent">This device</span>
                ) : null}
              </p>
              <p className="text-xs text-muted">
                Last used {when(s.lastUsedAt)}
              </p>
            </div>
            {s.current ? null : (
              <button
                type="button"
                className="btn btn--ghost btn--sm"
                disabled={pending}
                onClick={() =>
                  startTransition(async () =>
                    setStatus(await revokeSession(s.id)),
                  )
                }
              >
                Sign out
              </button>
            )}
          </li>
        ))}
      </ul>
      {others.length ? (
        <button
          type="button"
          className="btn btn--secondary"
          disabled={pending}
          onClick={() =>
            startTransition(async () => setStatus(await revokeOtherSessions()))
          }
        >
          Sign out of all other devices
        </button>
      ) : null}
    </div>
  );
}
