"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { moderateListing } from "../actions";

export function ModerationActions({ id }: { id: string }) {
  const [note, setNote] = useState("");
  return (
    <div className="cluster">
      <ActionButton
        label="Approve"
        variant="inverse"
        action={() => moderateListing(id, "approved", note)}
      />
      <input
        className="input"
        aria-label="Note to the seller"
        placeholder="Note to the seller"
        value={note}
        onChange={(e) => setNote(e.target.value)}
        style={{ maxWidth: "14rem" }}
      />
      <ActionButton
        label="Ask for changes"
        disabled={note.trim().length < 3}
        action={() => moderateListing(id, "changes_requested", note)}
      />
      <ActionButton
        label="Reject"
        variant="ghost"
        disabled={note.trim().length < 3}
        action={() => moderateListing(id, "rejected", note)}
      />
    </div>
  );
}
