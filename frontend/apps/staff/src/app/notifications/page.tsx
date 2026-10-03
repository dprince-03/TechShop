import type { Metadata } from "next";
import Link from "next/link";

import { sampleAttention } from "@techshop/fixtures";
import { Icon } from "@techshop/ui/components";

import styles from "./notifications.module.css";

export const metadata: Metadata = { title: "Notifications | TechShop Staff" };

export default function NotificationsPage() {
  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1 className={styles.title}>Notifications</h1>
        <p className="text-sm text-muted">3 unread · Sample data</p>
      </div>
      <ul role="list" className={`card ${styles.list}`}>
        {sampleAttention.map((a, i) => (
          <li key={a.label}>
            <Link
              href={`/${a.module.toLowerCase()}`}
              className={styles.item}
              data-unread={i < 3}
            >
              <Icon name="bell" size={18} className={styles.icon} />
              <span className="stack stack--sm">
                <span className="text-sm">
                  {a.count} {a.label.charAt(0).toLowerCase() + a.label.slice(1)}
                  {i < 3 ? (
                    <span className="visually-hidden"> (unread)</span>
                  ) : null}
                </span>
                <span className="text-xs text-muted">{a.module}</span>
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
