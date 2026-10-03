import Link from "next/link";
import type { ReactNode } from "react";

import { Icon, Logo, NavDisclosure } from "@techshop/ui/components";

import { SidebarNav } from "./SidebarNav";
import styles from "./StaffShell.module.css";

/** Staff portal app shell: sidebar (desktop) / menu (mobile), top bar, content. */
export function StaffShell({ children }: { children: ReactNode }) {
  return (
    <div className={styles.shell}>
      <aside className={styles.sidebar} aria-label="Staff portal">
        <Link
          href="/"
          className={styles.brand}
          aria-label="TechShop Staff home"
        >
          <Logo product="Staff" />
        </Link>
        <SidebarNav />
      </aside>

      <div className={styles.body}>
        <header className={styles.topbar}>
          <NavDisclosure
            className={styles.menu}
            summary={
              <summary
                className="btn btn--ghost btn--icon"
                aria-label="Open modules menu"
              >
                <Icon name="menu" />
              </summary>
            }
          >
            <div className={styles.menuPanel}>
              <SidebarNav />
            </div>
          </NavDisclosure>

          <Link
            href="/"
            className={styles.mobileBrand}
            aria-label="TechShop Staff home"
          >
            <Logo product="Staff" />
          </Link>

          <form role="search" action="/search" className={styles.search}>
            <label htmlFor="staff-search" className="visually-hidden">
              Search orders, products and customers
            </label>
            <Icon name="search" size={18} className={styles.searchIcon} />
            <input
              id="staff-search"
              name="q"
              type="search"
              placeholder="Search orders, products, customers…"
              className={styles.searchInput}
            />
          </form>

          <div className={styles.actions}>
            <Link
              href="/notifications"
              className="btn btn--ghost btn--icon"
              aria-label="Notifications, 3 unread"
            >
              <Icon name="bell" />
              <span className={styles.dot} aria-hidden="true" />
            </Link>
            <Link
              href="/account"
              className={styles.user}
              aria-label="Your account: Operations Manager (sample)"
            >
              <span className={styles.avatar} aria-hidden="true">
                OM
              </span>
              <span className={styles.userText}>
                <span className={styles.userName}>Operations Manager</span>
                <span className="text-xs text-muted">Sample role</span>
              </span>
            </Link>
          </div>
        </header>

        <main id="main" className={styles.main}>
          {children}
        </main>
      </div>
    </div>
  );
}
