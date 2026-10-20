import Link from "next/link";

import styles from "./Tabs.module.css";

/** Filter tabs as links (shareable URLs, no JavaScript needed). */
export function Tabs({
  items,
  current,
  label,
}: {
  items: { key: string; label: string; href: string }[];
  current: string;
  label: string;
}) {
  return (
    <nav aria-label={label} className={styles.tabs}>
      {items.map((t) => (
        <Link
          key={t.key}
          href={t.href}
          className={styles.tab}
          aria-current={t.key === current ? "page" : undefined}
        >
          {t.label}
        </Link>
      ))}
    </nav>
  );
}
