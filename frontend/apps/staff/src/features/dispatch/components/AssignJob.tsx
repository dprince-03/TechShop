"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { assignJob } from "../actions";

/** Pick a rider on shift for this job. Riders off shift are listed but marked. */
export function AssignJob({
  jobId,
  riders,
}: {
  jobId: string;
  riders: { id: string; label: string; onShift: boolean }[];
}) {
  const [riderId, setRiderId] = useState("");
  return (
    <div className="cluster">
      <select
        className="input"
        aria-label="Rider"
        value={riderId}
        onChange={(e) => setRiderId(e.target.value)}
        style={{ maxWidth: "14rem" }}
      >
        <option value="">Choose a rider</option>
        {riders.map((r) => (
          <option key={r.id} value={r.id}>
            {r.label}
            {r.onShift ? "" : " (off shift)"}
          </option>
        ))}
      </select>
      <ActionButton
        label="Assign"
        variant="inverse"
        disabled={!riderId}
        action={() => assignJob(jobId, riderId)}
      />
    </div>
  );
}
