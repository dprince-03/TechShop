"use client";

import type { SortKey } from "@techshop/fixtures";

import styles from "./ProductListing.module.css";

/** Sort dropdown that submits its form on change. Without JS, the form's Apply button does the same. */
export function SortSelect({
  value,
  options,
}: {
  value: SortKey;
  options: { value: SortKey; label: string }[];
}) {
  return (
    <div className={styles.sort}>
      <label htmlFor="sort" className="text-sm text-muted">
        Sort by
      </label>
      <select
        id="sort"
        name="sort"
        defaultValue={value}
        className={styles.select}
        onChange={(e) => e.currentTarget.form?.requestSubmit()}
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </div>
  );
}
