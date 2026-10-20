"use client";

import { useState } from "react";

import { ActionButton } from "@/components/ActionButton";

import { grantRole, revokeRole } from "../actions";

/** A staff member's roles: remove one, or grant another. */
export function RoleEditor({
  userId,
  current,
  roles,
}: {
  userId: string;
  current: string[];
  roles: { key: string; name: string }[];
}) {
  const [role, setRole] = useState("");
  const addable = roles.filter((r) => !current.includes(r.key));
  return (
    <div className="stack stack--sm">
      <div className="cluster">
        {current.map((r) => (
          <span key={r} className="cluster" style={{ gap: "var(--space-1)" }}>
            <span className="badge badge--outline">
              {roles.find((x) => x.key === r)?.name ?? r}
            </span>
            <ActionButton
              label={`Remove`}
              variant="ghost"
              confirm={`Remove the ${r} role?`}
              action={() => revokeRole(userId, r)}
            />
          </span>
        ))}
      </div>
      {addable.length ? (
        <div className="cluster">
          <select
            className="input"
            aria-label="Role to grant"
            value={role}
            onChange={(e) => setRole(e.target.value)}
            style={{ maxWidth: "14rem" }}
          >
            <option value="">Add a role</option>
            {addable.map((r) => (
              <option key={r.key} value={r.key}>
                {r.name}
              </option>
            ))}
          </select>
          <ActionButton
            label="Grant"
            disabled={!role}
            action={() => grantRole(userId, role)}
          />
        </div>
      ) : null}
    </div>
  );
}
