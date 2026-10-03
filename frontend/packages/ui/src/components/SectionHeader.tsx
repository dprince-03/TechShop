import Link from "next/link";
import type { ReactNode } from "react";

import styles from "./SectionHeader.module.css";

type Props = {
  id: string;
  title: string;
  eyebrow?: string;
  description?: ReactNode;
  action?: { label: string; href: string };
  /** Extra content beside the title, e.g. a countdown. */
  aside?: ReactNode;
};

/** Section heading row. `id` labels the parent <section aria-labelledby>. */
export function SectionHeader({
  id,
  title,
  eyebrow,
  description,
  action,
  aside,
}: Props) {
  return (
    <div className={styles.header}>
      <div className={styles.text}>
        {eyebrow ? <p className="eyebrow">{eyebrow}</p> : null}
        <div className={styles.titleRow}>
          <h2 id={id} className={styles.title}>
            {title}
          </h2>
          {aside}
        </div>
        {description ? (
          <p className="text-muted measure">{description}</p>
        ) : null}
      </div>
      {action ? (
        <Link
          href={action.href}
          className={`link link--arrow ${styles.action}`}
        >
          {action.label}
        </Link>
      ) : null}
    </div>
  );
}
