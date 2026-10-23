import Link from "next/link";
import type { ReactNode } from "react";

import type { StatusTone } from "@techshop/api-client";
import { Status } from "@techshop/ui/components";

import styles from "./Aftersales.module.css";

export type RecordRow = {
  key: string;
  href: string;
  title: string;
  status: { label: string; tone: StatusTone };
  meta: ReactNode;
};

/** A list of returns, claims, trade-ins or tickets: linked title, status (dot + words), details. */
export function Records({ label, rows }: { label: string; rows: RecordRow[] }) {
  return (
    <ul role="list" className={styles.list} aria-label={label}>
      {rows.map((r) => (
        <li key={r.key} className={styles.row}>
          <Link href={r.href} className={styles.title}>
            {r.title}
          </Link>
          <Status label={r.status.label} tone={r.status.tone} />
          <span className={styles.meta}>{r.meta}</span>
        </li>
      ))}
    </ul>
  );
}

const SECTIONS = [
  { href: "/orders", label: "Orders" },
  { href: "/account/returns", label: "Returns" },
  { href: "/account/warranty", label: "Warranty" },
  { href: "/account/trade-ins", label: "Trade-ins" },
  { href: "/account/support", label: "Support" },
  { href: "/account", label: "Account" },
];

/** Links between the account's after-sales pages. */
export function AccountNav({ current }: { current: string }) {
  return (
    <nav aria-label="Your account" className={styles.nav}>
      {SECTIONS.map((s) => (
        <Link
          key={s.href}
          href={s.href}
          className={styles.navLink}
          aria-current={s.href === current ? "page" : undefined}
        >
          {s.label}
        </Link>
      ))}
    </nav>
  );
}

/** A label/value list (dl) for a record's details. */
export function Facts({ items }: { items: [string, ReactNode][] }) {
  return (
    <dl className={styles.facts}>
      {items
        .filter(([, v]) => v !== null && v !== undefined && v !== "")
        .map(([k, v]) => (
          <div key={k}>
            <dt>{k}</dt>
            <dd>{v}</dd>
          </div>
        ))}
    </dl>
  );
}
