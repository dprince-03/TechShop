import Link from "next/link";
import type { CSSProperties, ReactNode } from "react";

import { Logo } from "./Logo";
import styles from "./SiteFooter.module.css";

export type FooterColumn = {
  title: string;
  links: { label: string; href: string }[];
};

type Props = {
  columns: FooterColumn[];
  /** Short line under the logo. */
  tagline?: string;
  /** Extra row above the legal line, e.g. payment methods. */
  extra?: ReactNode;
};

const PAYMENT_METHODS = ["Paystack", "OPay", "Moniepoint"];

export function PaymentMethods() {
  return (
    <div className={styles.payments}>
      <span className="text-xs text-muted">Secure payments with</span>
      <ul
        role="list"
        className="cluster"
        style={{ "--cluster-gap": "var(--space-2)" } as CSSProperties}
      >
        {PAYMENT_METHODS.map((m) => (
          <li key={m} className="badge badge--outline">
            {m}
          </li>
        ))}
      </ul>
    </div>
  );
}

export function SiteFooter({ columns, tagline, extra }: Props) {
  return (
    <footer className={styles.footer}>
      <div className={`container ${styles.top}`}>
        <div className={styles.brand}>
          <Logo />
          {tagline ? (
            <p className="text-sm text-muted measure">{tagline}</p>
          ) : null}
        </div>
        <div className={styles.columns}>
          {columns.map((col) => (
            <nav key={col.title} aria-label={col.title}>
              <h2 className={styles.colTitle}>{col.title}</h2>
              <ul role="list" className="stack stack--sm">
                {col.links.map((l) => (
                  <li key={l.label}>
                    <Link href={l.href} className="link link--subtle text-sm">
                      {l.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </nav>
          ))}
        </div>
      </div>
      <div className={`container ${styles.bottom}`}>
        {extra}
        <div className={styles.legal}>
          <p className="text-xs text-muted">
            © {new Date().getFullYear()} TechShop. All rights reserved.
          </p>
          <ul role="list" className="cluster text-xs">
            <li>
              <Link href="/legal/terms" className="link link--subtle">
                Terms
              </Link>
            </li>
            <li>
              <Link href="/legal/privacy" className="link link--subtle">
                Privacy
              </Link>
            </li>
            <li>
              <Link href="/legal/cookies" className="link link--subtle">
                Cookies
              </Link>
            </li>
          </ul>
        </div>
      </div>
    </footer>
  );
}
