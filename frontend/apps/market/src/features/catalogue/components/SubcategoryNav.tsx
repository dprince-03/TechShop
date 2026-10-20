import Link from "next/link";

import type { Category } from "@techshop/api-client";

import styles from "./ProductListing.module.css";

/** "Browse" links to a category's subcategories, with the current one marked. */
export function SubcategoryNav({
  category,
  current,
}: {
  category: Category;
  current?: string;
}) {
  return (
    <nav aria-label={`${category.name} types`} className={styles.browse}>
      <p className={styles.legend}>Browse</p>
      <Link
        href={`/c/${category.slug}`}
        className={styles.browseLink}
        aria-current={current ? undefined : "page"}
      >
        All {category.name.toLowerCase()}
      </Link>
      {category.subcategories.map((s) => (
        <Link
          key={s.slug}
          href={`/c/${category.slug}/${s.slug}`}
          className={styles.browseLink}
          aria-current={current === s.slug ? "page" : undefined}
        >
          {s.name}
        </Link>
      ))}
    </nav>
  );
}
