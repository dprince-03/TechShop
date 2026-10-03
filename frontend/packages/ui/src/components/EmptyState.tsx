import type { ReactNode } from "react";

import { Icon, type IconName } from "./Icon";
import styles from "./EmptyState.module.css";

/** Explains what belongs here and offers a next action. Never a dead end. */
export function EmptyState({
  icon,
  title,
  children,
  action,
}: {
  icon: IconName;
  title: string;
  children?: ReactNode;
  action?: ReactNode;
}) {
  return (
    <div className={styles.empty}>
      <span className={styles.icon}>
        <Icon name={icon} size={28} />
      </span>
      <h2 className={styles.title}>{title}</h2>
      {children ? (
        <div className={`text-muted ${styles.text}`}>{children}</div>
      ) : null}
      {action}
    </div>
  );
}
