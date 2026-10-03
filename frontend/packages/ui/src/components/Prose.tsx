import type { ReactNode } from "react";

import styles from "./Prose.module.css";

/** Long-form text (help articles, policies): readable measure and vertical rhythm. */
export function Prose({ children }: { children: ReactNode }) {
  return <div className={styles.prose}>{children}</div>;
}
