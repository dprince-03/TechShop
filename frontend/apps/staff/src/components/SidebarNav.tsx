"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { Icon } from "@techshop/ui/components";

import { moduleGroups } from "./modules";
import styles from "./SidebarNav.module.css";

/** Grouped module navigation; marks the current module with aria-current. */
export function SidebarNav() {
  const pathname = usePathname();
  const isCurrent = (href: string) =>
    href === "/"
      ? pathname === "/"
      : pathname === href || pathname.startsWith(`${href}/`);
  return (
    <nav aria-label="Modules" className={styles.nav}>
      <Link
        href="/"
        className={styles.link}
        aria-current={isCurrent("/") ? "page" : undefined}
      >
        <Icon name="grid" size={18} />
        Dashboard
      </Link>
      {moduleGroups.map((group) => (
        <div key={group.title} className={styles.group}>
          <h2 className={styles.groupTitle}>{group.title}</h2>
          <ul role="list">
            {group.modules.map((m) => (
              <li key={m.href}>
                <Link
                  href={m.href}
                  className={styles.link}
                  aria-current={isCurrent(m.href) ? "page" : undefined}
                >
                  <Icon name={m.icon} size={18} />
                  {m.label}
                </Link>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </nav>
  );
}
