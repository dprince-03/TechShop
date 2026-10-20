"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { returnToHub } from "../actions";

export function ReturnJob({ jobId }: { jobId: string }) {
  const [note, setNote] = useState("");
  return (
    <div className="cluster">
      <input
        className="input"
        aria-label="Why it returns"
        placeholder="Why it returns"
        value={note}
        onChange={(e) => setNote(e.target.value)}
        style={{ maxWidth: "12rem" }}
      />
      <ActionButton
        label="Return to hub"
        variant="ghost"
        disabled={note.trim().length < 3}
        action={() => returnToHub(jobId, note)}
      />
    </div>
  );
}
