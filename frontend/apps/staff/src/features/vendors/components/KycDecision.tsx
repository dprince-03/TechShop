"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { reviewKyc } from "../actions";

/** Approve, ask for more information, or reject, with a note the seller reads. */
export function KycDecision({ submissionId }: { submissionId: string }) {
  const [note, setNote] = useState("");
  const noted = note.trim().length >= 3;
  return (
    <div className="stack">
      <div className="field">
        <label htmlFor="kyc-note" className="field__label">
          Note to the seller{" "}
          <span className="text-muted">(required unless approving)</span>
        </label>
        <textarea
          id="kyc-note"
          className="input"
          rows={3}
          maxLength={500}
          value={note}
          onChange={(e) => setNote(e.target.value)}
        />
      </div>
      <div className="cluster">
        <ActionButton
          label="Approve"
          variant="inverse"
          action={() => reviewKyc(submissionId, "approved", note)}
        />
        <ActionButton
          label="Ask for more information"
          disabled={!noted}
          action={() => reviewKyc(submissionId, "needs_more_info", note)}
        />
        <ActionButton
          label="Reject"
          variant="ghost"
          disabled={!noted}
          confirm="Reject this seller’s verification?"
          action={() => reviewKyc(submissionId, "rejected", note)}
        />
      </div>
    </div>
  );
}
