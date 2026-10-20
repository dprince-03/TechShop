import styles from "./Figures.module.css";

/** A row of headline figures (live values only; never invented). */
export function Figures({
  items,
}: {
  items: { label: string; value: string; note?: string }[];
}) {
  return (
    <ul role="list" className={styles.figures}>
      {items.map((f) => (
        <li key={f.label} className={`card ${styles.figure}`}>
          <p className="text-sm text-muted">{f.label}</p>
          <p className={`${styles.value} tabular-nums`}>{f.value}</p>
          {f.note ? <p className="text-xs text-muted">{f.note}</p> : null}
        </li>
      ))}
    </ul>
  );
}
