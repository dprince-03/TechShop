import { Icon } from "./Icon";
import styles from "./Rating.module.css";

/** Star rating with review count. Announced as a sentence for screen readers. */
export function Rating({ average, count }: { average: number; count: number }) {
  const label = `Rated ${average.toFixed(1)} out of 5 from ${count.toLocaleString("en-NG")} reviews`;
  return (
    <span className={styles.rating} role="img" aria-label={label}>
      <Icon name="star" size={14} className={styles.star} />
      <span className={styles.value}>{average.toFixed(1)}</span>
      <span className={styles.count}>({count.toLocaleString("en-NG")})</span>
    </span>
  );
}
