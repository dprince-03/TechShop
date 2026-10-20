"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { Icon } from "@techshop/ui/components";

import type { ModuleGroup } from "./modules";
import styles from "./SidebarNav.module.css";

/** Grouped module navigation (only modules the person may open); marks the current one. */
export function SidebarNav({ groups }: { groups: ModuleGroup[] }) {
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
      {groups.map((group) => (
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
                  {m.live ? null : (
                    <span className={styles.preview}>Preview</span>
                  )}
                </Link>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </nav>
  );
}
