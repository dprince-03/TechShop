import { Icon } from "./Icon";
import styles from "./FaqList.module.css";

export type Faq = { q: string; a: string };

/** Questions as native disclosures: keyboard and screen-reader friendly, no JS needed. */
export function FaqList({ items }: { items: Faq[] }) {
  return (
    <div className={styles.list}>
      {items.map((f) => (
        <details key={f.q} className={styles.item}>
          <summary className={styles.question}>
            {f.q}
            <Icon name="chevronDown" size={18} className={styles.chevron} />
          </summary>
          <p className={styles.answer}>{f.a}</p>
        </details>
      ))}
    </div>
  );
}
