import Link from "next/link";

import { Icon } from "./Icon";
import styles from "./Breadcrumbs.module.css";

export type Crumb = { label: string; href?: string };

/** Breadcrumb trail. The last item is the current page and is not a link. */
export function Breadcrumbs({ items }: { items: Crumb[] }) {
  return (
    <nav aria-label="Breadcrumb" className={styles.nav}>
      <ol className={styles.list}>
        {items.map((c, i) => {
          const last = i === items.length - 1;
          return (
            <li key={`${c.label}-${i}`} className={styles.item}>
              {last || !c.href ? (
                <span
                  aria-current={last ? "page" : undefined}
                  className={last ? styles.current : undefined}
                >
                  {c.label}
                </span>
              ) : (
                <Link href={c.href} className="link link--subtle">
                  {c.label}
                </Link>
              )}
              {last ? null : (
                <Icon name="chevronRight" size={14} className={styles.sep} />
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
