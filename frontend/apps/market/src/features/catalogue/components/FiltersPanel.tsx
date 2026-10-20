"use client";

import { useId, useState, type ReactNode } from "react";

import { Icon } from "@techshop/ui/components";

import styles from "./ProductListing.module.css";

/**
 * Collapsible filter panel on mobile; always open from 1024px.
 * Server-rendered contents stay usable without JavaScript (see <noscript> in ProductListing).
 */
export function FiltersPanel({
  activeCount,
  children,
}: {
  activeCount: number;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const id = useId();
  return (
    <div className={styles.filters}>
      <button
        type="button"
        className={`btn btn--secondary btn--block ${styles.filtersToggle}`}
        aria-expanded={open}
        aria-controls={id}
        onClick={() => setOpen((o) => !o)}
      >
        <Icon name="filter" size={18} />
        {open ? "Hide filters" : "Show filters"}
        {activeCount > 0 ? (
          <span className="badge badge--accent">{activeCount}</span>
        ) : null}
      </button>
      <div id={id} className={styles.filtersBody} data-open={open}>
        {children}
      </div>
    </div>
  );
}
