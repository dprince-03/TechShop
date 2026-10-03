import { discountPercent, formatMoney, type Money } from "@techshop/api-client";

import styles from "./Price.module.css";

/** Current price, previous price (struck through), and discount badge. */
export function Price({
  price,
  compareAt,
  size = "md",
}: {
  price: Money;
  compareAt?: Money;
  size?: "md" | "lg";
}) {
  const off = discountPercent(price, compareAt);
  return (
    <div className={`${styles.price} ${styles[size]}`}>
      <span className={styles.current}>{formatMoney(price)}</span>
      {off ? (
        <>
          <span className="strike text-sm tabular-nums">
            <span className="visually-hidden">Was </span>
            {formatMoney(compareAt!)}
          </span>
          <span className="badge badge--sale">−{off}%</span>
        </>
      ) : null}
    </div>
  );
}
