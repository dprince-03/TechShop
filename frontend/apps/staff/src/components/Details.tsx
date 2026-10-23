import type { ReactNode } from "react";

/** A record's label/value details; empty values are left out. */
export function Details({ items }: { items: [string, ReactNode][] }) {
  return (
    <dl className="stack stack--sm text-sm">
      {items
        .filter(([, v]) => v !== null && v !== undefined && v !== "")
        .map(([k, v]) => (
          <div key={k} className="cluster">
            <dt className="text-muted">{k}</dt>
            <dd>{v}</dd>
          </div>
        ))}
    </dl>
  );
}
