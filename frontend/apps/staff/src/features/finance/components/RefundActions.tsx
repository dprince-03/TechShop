"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { approveRefund, rejectRefund } from "../actions";

/** Approve (step-up may be asked) or reject with a note. The API blocks self-approval. */
export function RefundActions({ id }: { id: string }) {
  const [note, setNote] = useState("");
  return (
    <div className="cluster">
      <ActionButton
        label="Approve"
        variant="inverse"
        confirm="Approve this refund? The money goes back to the customer."
        action={approveRefund.bind(null, id)}
      />
      <input
        className="input"
        aria-label="Reason for rejecting"
        placeholder="Reason to reject"
        value={note}
        onChange={(e) => setNote(e.target.value)}
        style={{ maxWidth: "14rem" }}
      />
      <ActionButton
        label="Reject"
        variant="ghost"
        disabled={note.trim().length < 3}
        action={() => rejectRefund(id, note)}
      />
    </div>
  );
}
