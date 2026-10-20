"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { signManifest } from "../actions";

export function SignManifest({ id }: { id: string }) {
  const [name, setName] = useState("");
  return (
    <div className="cluster">
      <input
        className="input"
        aria-label="Name of the person collecting"
        placeholder="Name of the person collecting"
        value={name}
        onChange={(e) => setName(e.target.value)}
        style={{ maxWidth: "18rem" }}
      />
      <ActionButton
        label="Sign hand-over"
        variant="inverse"
        disabled={name.trim().length < 2}
        confirm="Confirm the parcels on this manifest have been handed over?"
        action={() => signManifest(id, name)}
      />
    </div>
  );
}
