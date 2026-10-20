import Link from "next/link";

import { EmptyState } from "@techshop/ui/components";

import { getCategories } from "@/features/catalogue/api";

import styles from "./not-found.module.css";

export default async function NotFound() {
  const categories = await getCategories();
  return (
    <main id="main" className="section">
      <div className="container container--narrow stack stack--lg">
        <EmptyState
          icon="search"
          title="We couldn’t find that page"
          action={
            <form
              role="search"
              action="/search"
              className={`cluster ${styles.center}`}
            >
              <label htmlFor="nf-q" className="visually-hidden">
                Search products
              </label>
              <input
                id="nf-q"
                name="q"
                type="search"
                placeholder="Search products"
                className={styles.input}
              />
              <button type="submit" className="btn btn--primary">
                Search
              </button>
            </form>
          }
        >
          The link may be old, or this page may not exist yet. Search for what
          you need, or pick a category below.
        </EmptyState>
        <nav aria-label="Categories">
          <ul role="list" className={`cluster ${styles.center}`}>
            {categories.map((c) => (
              <li key={c.slug}>
                <Link
                  href={`/c/${c.slug}`}
                  className="btn btn--sm btn--secondary"
                >
                  {c.name}
                </Link>
              </li>
            ))}
          </ul>
        </nav>
      </div>
    </main>
  );
}
