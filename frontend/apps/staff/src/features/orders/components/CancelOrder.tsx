"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { cancelOrder } from "../actions";

/** Cancel with a required reason (it goes into the order history the customer sees). */
export function CancelOrder({ number }: { number: string }) {
  const [reason, setReason] = useState("");
  return (
    <div className="stack stack--sm">
      <label className="field">
        <span className="field__label">Reason for cancelling</span>
        <input
          className="input"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          maxLength={200}
          placeholder="e.g. Customer asked by phone"
        />
      </label>
      <ActionButton
        label="Cancel order"
        variant="ghost"
        disabled={reason.trim().length < 3}
        confirm={`Cancel order ${number}? Stock is released and any payment goes to refunds.`}
        action={() => cancelOrder(number, reason)}
      />
    </div>
  );
}
