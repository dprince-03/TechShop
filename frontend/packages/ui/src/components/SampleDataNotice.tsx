import styles from "./SampleDataNotice.module.css";

/** Development-only notice: everything on the page is sample data. Remove when real data is wired up. */
export function SampleDataNotice() {
  return (
    <p className={styles.notice} role="note">
      Preview: products, sellers, prices and figures on this page are sample
      data.
    </p>
  );
}
