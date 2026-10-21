import Link from "next/link";

import styles from "./ShopNav.module.css";

const sections = [
  { key: "overview", label: "Overview", path: "" },
  { key: "listings", label: "Listings", path: "/listings" },
  { key: "orders", label: "Orders", path: "/orders" },
  { key: "reviews", label: "Reviews", path: "/reviews" },
  { key: "verification", label: "Verification", path: "/verification" },
  { key: "payouts", label: "Payout accounts", path: "/payouts" },
] as const;

export type Section = (typeof sections)[number]["key"];

/** The shop's sections; scrolls sideways on small screens. */
export function ShopNav({
  shopId,
  current,
}: {
  shopId: string;
  current: Section;
}) {
  return (
    <nav aria-label="Shop" className={styles.nav}>
      <ul role="list" className={styles.list}>
        {sections.map((s) => (
          <li key={s.key}>
            <Link
              href={`/centre/${shopId}${s.path}`}
              aria-current={s.key === current ? "page" : undefined}
              className={styles.link}
            >
              {s.label}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
}
