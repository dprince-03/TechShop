import styles from "./Logo.module.css";

/** TechShop wordmark. `product` adds the app name, e.g. "Market". */
export function Logo({ product }: { product?: string }) {
  return (
    <span className={styles.logo}>
      <svg
        className={styles.mark}
        viewBox="0 0 32 32"
        aria-hidden="true"
        focusable="false"
      >
        <rect width="32" height="32" rx="9" />
        <path d="M9 10.5h14M16 10.5V23" />
      </svg>
      <span className={styles.word}>TechShop</span>
      {product ? <span className={styles.product}>{product}</span> : null}
    </span>
  );
}
