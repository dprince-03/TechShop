"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { excludePayout } from "../actions";

/** Leave one seller out of a draft batch, with a reason; optionally hold later payouts too. */
export function ExcludePayout({
  batchId,
  payoutId,
}: {
  batchId: string;
  payoutId: string;
}) {
  const [note, setNote] = useState("");
  const [hold, setHold] = useState(false);
  return (
    <div className="cluster">
      <input
        className="input"
        aria-label="Reason for leaving this seller out"
        placeholder="Reason"
        value={note}
        onChange={(e) => setNote(e.target.value)}
      />
      <label className="cluster text-sm">
        <input
          type="checkbox"
          checked={hold}
          onChange={(e) => setHold(e.target.checked)}
        />
        Hold later payouts
      </label>
      <ActionButton
        label="Leave out"
        variant="ghost"
        disabled={note.trim().length < 3}
        action={() => excludePayout(batchId, payoutId, note, hold)}
      />
    </div>
  );
}
