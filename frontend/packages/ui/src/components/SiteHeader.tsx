import Link from "next/link";
import type { ReactNode } from "react";

import { Icon } from "./Icon";
import { Logo } from "./Logo";
import { NavDisclosure } from "./NavDisclosure";
import styles from "./SiteHeader.module.css";

export type NavItem = { label: string; href: string };

type Props = {
  product?: string;
  nav: NavItem[];
  /** Right-hand actions (sign in, CTA). Shown at every width. */
  actions?: ReactNode;
};

/**
 * Sticky site header for marketing-style apps. Navigation collapses into a
 * native <details> menu below 1024px, so it works without JavaScript.
 */
export function SiteHeader({ product, nav, actions }: Props) {
  return (
    <header className={styles.header}>
      <div className={`container ${styles.inner}`}>
        <Link
          href="/"
          className={styles.home}
          aria-label={`TechShop ${product ?? ""} home`.replace("  ", " ")}
        >
          <Logo product={product} />
        </Link>

        <nav aria-label="Main" className={styles.nav}>
          <ul role="list" className={styles.navList}>
            {nav.map((item) => (
              <li key={item.href}>
                <Link href={item.href} className={styles.navLink}>
                  {item.label}
                </Link>
              </li>
            ))}
          </ul>
        </nav>

        <div className={styles.actions}>
          {actions}
          <NavDisclosure
            className={styles.menu}
            summary={
              <summary className="btn btn--ghost btn--icon" aria-label="Menu">
                <Icon name="menu" />
              </summary>
            }
          >
            <nav aria-label="Main" className={styles.menuPanel}>
              <ul role="list" className="container stack stack--sm">
                {nav.map((item) => (
                  <li key={item.href}>
                    <Link href={item.href} className={styles.menuLink}>
                      {item.label}
                      <Icon name="chevronRight" size={16} />
                    </Link>
                  </li>
                ))}
              </ul>
            </nav>
          </NavDisclosure>
        </div>
      </div>
    </header>
  );
}
