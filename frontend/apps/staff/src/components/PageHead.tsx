import type { ReactNode } from "react";

import styles from "./PageHead.module.css";

/** A module page's single h1, a short description and its main actions. */
export function PageHead({
  title,
  description,
  actions,
}: {
  title: string;
  description?: ReactNode;
  actions?: ReactNode;
}) {
  return (
    <div className={styles.head}>
      <div className="stack stack--sm">
        <h1 className={styles.title}>{title}</h1>
        {description ? (
          <p className="text-sm text-muted">{description}</p>
        ) : null}
      </div>
      {actions ? <div className="cluster">{actions}</div> : null}
    </div>
  );
}
