import { orderStatus } from "../status";

import styles from "./Orders.module.css";

export function StatusBadge({ status }: { status: string }) {
  const s = orderStatus[status] ?? { label: status, tone: "neutral" as const };
  const tone = s.tone === "neutral" ? "" : styles[s.tone];
  return <span className={`${styles.status} ${tone}`}>{s.label}</span>;
}
