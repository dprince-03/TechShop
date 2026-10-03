import type { CategorySlug } from "@techshop/api-client";

import { Icon, type IconName } from "./Icon";
import styles from "./CategoryArt.module.css";

const art: Record<CategorySlug, IconName> = {
  phones: "phone",
  laptops: "laptop",
  accessories: "headphones",
  gaming: "gamepad",
  "smart-home": "home",
  office: "printer",
  workstations: "monitor",
  cars: "car",
};

/**
 * PLACEHOLDER product/category artwork until real photography exists.
 * Decorative: the surrounding card always names the item.
 */
export function CategoryArt({
  category,
  size = "md",
}: {
  category: CategorySlug;
  size?: "sm" | "md" | "lg";
}) {
  return (
    <span className={`${styles.art} ${styles[size]}`} aria-hidden="true">
      <Icon
        name={art[category]}
        size={size === "lg" ? 96 : size === "md" ? 72 : 36}
        strokeWidth={1.25}
      />
    </span>
  );
}
