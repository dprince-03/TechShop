"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { revealId } from "../actions";

/** Reveals the full ID number after a reason is given (audited; needs a recent authenticator check). */
export function RevealId({ submissionId }: { submissionId: string }) {
  const [reason, setReason] = useState("");
  return (
    <div className="stack">
      <div className="field">
        <label htmlFor="reveal-reason" className="field__label">
          Why do you need the full number?
        </label>
        <input
          id="reveal-reason"
          className="input"
          maxLength={300}
          value={reason}
          onChange={(e) => setReason(e.target.value)}
        />
        <span className="field__hint">
          Recorded in the audit log with your name.
        </span>
      </div>
      <ActionButton
        label="Show full ID number"
        disabled={reason.trim().length < 10}
        action={() => revealId(submissionId, reason)}
      />
    </div>
  );
}
