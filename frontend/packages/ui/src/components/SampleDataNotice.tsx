import type { ReactNode } from "react";

import styles from "./SampleDataNotice.module.css";

/**
 * Marks sample data. Required on any page (or section) whose content is not yet served by the
 * API; pass `children` to say exactly what is sample.
 */
export function SampleDataNotice({ children }: { children?: ReactNode }) {
  return (
    <p className={styles.notice} role="note">
      {children ??
        "Preview: products, sellers, prices and figures on this page are sample data."}
    </p>
  );
}
