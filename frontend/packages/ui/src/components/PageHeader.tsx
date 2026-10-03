import type { ReactNode } from "react";

import { Breadcrumbs, type Crumb } from "./Breadcrumbs";
import styles from "./PageHeader.module.css";

type Props = {
  title: string;
  eyebrow?: string;
  lead?: ReactNode;
  breadcrumbs?: Crumb[];
  /** Buttons or links beside the title on wide screens. */
  actions?: ReactNode;
};

/** Standard inner-page header: breadcrumbs, eyebrow, the page's single h1, lead, actions. */
export function PageHeader({
  title,
  eyebrow,
  lead,
  breadcrumbs,
  actions,
}: Props) {
  return (
    <header className={styles.header}>
      {breadcrumbs ? <Breadcrumbs items={breadcrumbs} /> : null}
      <div className={styles.row}>
        <div className={styles.text}>
          {eyebrow ? <p className="eyebrow">{eyebrow}</p> : null}
          <h1 className={styles.title}>{title}</h1>
          {lead ? (
            <p className={`text-lg text-muted ${styles.lead}`}>{lead}</p>
          ) : null}
        </div>
        {actions ? <div className={styles.actions}>{actions}</div> : null}
      </div>
    </header>
  );
}
