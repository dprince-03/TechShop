import Link from "next/link";

import { Icon, Logo, type IconName } from "@techshop/ui/components";
import { sites } from "@techshop/ui/sites";

import { getCategories } from "@/features/catalogue/api";

import styles from "./MarketHeader.module.css";

const categoryIcon: Record<string, IconName> = {
  phones: "phone",
  laptops: "laptop",
  accessories: "headphones",
  gaming: "gamepad",
  "smart-home": "home",
  office: "printer",
  workstations: "monitor",
  cars: "car",
};

type Props = {
  signedIn: boolean;
  /** Total quantity in the cart (0 when empty or unknown). */
  cartCount: number;
};

/** Search-first marketplace header: utility bar, logo + search + account/cart, category row. */
export async function MarketHeader({ signedIn, cartCount }: Props) {
  const categories = await getCategories();
  return (
    <>
      <div className={styles.utility}>
        <div className={`container ${styles.utilityInner}`}>
          <Link href="/delivery" className={styles.utilityLink}>
            <Icon name="truck" size={16} />
            Delivery & fees
          </Link>
          <ul role="list" className={styles.utilityLinks}>
            <li>
              <a href={sites.seller} className={styles.utilityLink}>
                Sell on TechShop
              </a>
            </li>
            <li className="hide-mobile">
              <a href={sites.wholesale} className={styles.utilityLink}>
                Wholesale & business
              </a>
            </li>
            <li className="hide-mobile">
              <Link href="/help" className={styles.utilityLink}>
                Help
              </Link>
            </li>
          </ul>
        </div>
      </div>

      <header className={styles.header}>
        <div className={styles.main}>
          <div className={`container ${styles.mainInner}`}>
            <Link
              href="/"
              className={styles.home}
              aria-label="TechShop Market home"
            >
              <Logo product="Market" />
            </Link>

            <form role="search" action="/search" className={styles.search}>
              <label htmlFor="search-category" className="visually-hidden">
                Search in
              </label>
              <select
                id="search-category"
                name="category"
                className={styles.scope}
                defaultValue=""
              >
                <option value="">All</option>
                {categories.map((c) => (
                  <option key={c.slug} value={c.slug}>
                    {c.name}
                  </option>
                ))}
              </select>
              <label htmlFor="search-q" className="visually-hidden">
                Search products
              </label>
              <input
                id="search-q"
                name="q"
                type="search"
                placeholder="Search phones, laptops, consoles…"
                autoComplete="off"
                enterKeyHint="search"
                className={styles.input}
              />
              <button
                type="submit"
                className={styles.submit}
                aria-label="Search"
              >
                <Icon name="search" />
              </button>
            </form>

            <nav aria-label="Account" className={styles.account}>
              <Link
                href={signedIn ? "/account" : "/sign-in"}
                className={styles.action}
              >
                <Icon name="user" />
                <span className={styles.actionLabel}>
                  {signedIn ? "Account" : "Sign in"}
                </span>
              </Link>
              <Link href="/saved" className={`${styles.action} hide-mobile`}>
                <Icon name="heart" />
                <span className={styles.actionLabel}>Saved</span>
              </Link>
              <Link
                href="/cart"
                className={styles.action}
                aria-label={
                  cartCount
                    ? `Cart, ${cartCount} item${cartCount === 1 ? "" : "s"}`
                    : "Cart, empty"
                }
              >
                <Icon name="cart" />
                <span className={styles.actionLabel}>Cart</span>
                {cartCount ? (
                  <span className={styles.count} aria-hidden="true">
                    {cartCount > 99 ? "99+" : cartCount}
                  </span>
                ) : null}
              </Link>
            </nav>
          </div>
        </div>

        <nav aria-label="Categories" className={styles.categories}>
          <ul role="list" className={`container ${styles.categoryList}`}>
            <li>
              <Link
                href="/deals"
                className={`${styles.category} ${styles.deals}`}
              >
                <Icon name="bolt" size={18} />
                Deals
              </Link>
            </li>
            {categories.map((c) => (
              <li key={c.slug}>
                <Link href={`/c/${c.slug}`} className={styles.category}>
                  <Icon name={categoryIcon[c.slug] ?? "grid"} size={18} />
                  {c.name}
                </Link>
              </li>
            ))}
          </ul>
        </nav>
      </header>
    </>
  );
}
