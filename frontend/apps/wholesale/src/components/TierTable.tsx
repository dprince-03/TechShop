import { formatMoney, type PriceTier } from "@techshop/api-client";

import styles from "./TierTable.module.css";

/** Quantity price breaks. The first row is the minimum order quantity. */
export function TierTable({
  tiers,
  caption,
}: {
  tiers: PriceTier[];
  caption: string;
}) {
  return (
    <table className={styles.table}>
      <caption className="visually-hidden">{caption}</caption>
      <thead>
        <tr>
          <th scope="col">Quantity</th>
          <th scope="col">Unit price</th>
        </tr>
      </thead>
      <tbody>
        {tiers.map((t, i) => (
          <tr key={t.minQuantity}>
            <td>
              {i === tiers.length - 1
                ? `${t.minQuantity}+`
                : `${t.minQuantity}–${tiers[i + 1].minQuantity - 1}`}
            </td>
            <td className={styles.price}>{formatMoney(t.unitPrice)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
