"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { resolveException } from "../actions";

export function ExceptionActions({ id }: { id: string }) {
  const [text, setText] = useState("");
  return (
    <div className="cluster">
      <input
        className="input"
        aria-label="What was done"
        placeholder="What was done"
        value={text}
        onChange={(e) => setText(e.target.value)}
        style={{ maxWidth: "16rem" }}
      />
      <ActionButton
        label="Resolve"
        variant="inverse"
        disabled={text.trim().length < 3}
        action={() => resolveException(id, "resolved", text)}
      />
      <ActionButton
        label="Escalate"
        variant="ghost"
        disabled={text.trim().length < 3}
        action={() => resolveException(id, "escalated", text)}
      />
    </div>
  );
}
