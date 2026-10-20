import Link from "next/link";
import type { CSSProperties, ReactNode } from "react";

import type { Product } from "@techshop/api-client";
import { categories, type ProductQuery, sortOptions } from "@techshop/fixtures";
import {
  Breadcrumbs,
  EmptyState,
  ProductCard,
  type Crumb,
} from "@techshop/ui/components";

import { activeFilterCount } from "@/lib/query";

import { FiltersPanel } from "./FiltersPanel";
import styles from "./ProductListing.module.css";
import { SortSelect } from "./SortSelect";

type Props = {
  title: string;
  description?: string;
  breadcrumbs: Crumb[];
  /** URL the filter form submits to (the current page path). */
  action: string;
  query: ProductQuery;
  products: Product[];
  /** Extra navigation above the filters, e.g. subcategory links. */
  browse?: ReactNode;
  /** Show the category filter (search results only). */
  showCategoryFilter?: boolean;
};

const conditions = [
  { value: "new", label: "New" },
  { value: "uk-used", label: "UK-used" },
  { value: "refurbished", label: "Refurbished" },
] as const;

const sellers = [
  { value: "techshop", label: "Sold by TechShop" },
  { value: "vendor", label: "Marketplace sellers" },
] as const;

/**
 * Listing template: filters (GET form → URL params), sort, result count, product grid.
 * Works without JavaScript; JS only adds auto-submit on sort and the mobile filter toggle.
 */
export function ProductListing({
  title,
  description,
  breadcrumbs,
  action,
  query,
  products,
  browse,
  showCategoryFilter,
}: Props) {
  const active = activeFilterCount(query);
  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <div className="stack stack--sm">
          <Breadcrumbs items={breadcrumbs} />
          <h1 className={styles.title}>{title}</h1>
          {description ? (
            <p className="text-muted measure">{description}</p>
          ) : null}
        </div>

        <noscript>
          <style>{`.${styles.filtersBody}{display:flex}.${styles.filtersToggle}{display:none}`}</style>
        </noscript>

        <form
          method="get"
          action={action}
          className="with-sidebar"
          style={
            {
              "--sidebar-width": "15rem",
              "--sidebar-content-min": "65%",
            } as CSSProperties
          }
        >
          <aside aria-label="Filters" className={styles.aside}>
            {browse}
            <FiltersPanel activeCount={active}>
              {query.q ? (
                <input type="hidden" name="q" value={query.q} />
              ) : null}

              {showCategoryFilter ? (
                <fieldset className={styles.group}>
                  <legend className={styles.legend}>Category</legend>
                  <select
                    name="category"
                    defaultValue={query.category ?? ""}
                    className={styles.select}
                    aria-label="Category"
                  >
                    <option value="">All categories</option>
                    {categories
                      .filter((c) => c.slug !== "cars")
                      .map((c) => (
                        <option key={c.slug} value={c.slug}>
                          {c.name}
                        </option>
                      ))}
                  </select>
                </fieldset>
              ) : null}

              <fieldset className={styles.group}>
                <legend className={styles.legend}>Condition</legend>
                {conditions.map((c) => (
                  <label key={c.value} className={styles.option}>
                    <input
                      type="checkbox"
                      name="condition"
                      value={c.value}
                      defaultChecked={query.condition?.includes(c.value)}
                    />
                    {c.label}
                  </label>
                ))}
              </fieldset>

              <fieldset className={styles.group}>
                <legend className={styles.legend}>Seller</legend>
                {sellers.map((s) => (
                  <label key={s.value} className={styles.option}>
                    <input
                      type="checkbox"
                      name="seller"
                      value={s.value}
                      defaultChecked={query.seller?.includes(s.value)}
                    />
                    {s.label}
                  </label>
                ))}
              </fieldset>

              <fieldset className={styles.group}>
                <legend className={styles.legend}>Price (₦)</legend>
                <div className={styles.priceRow}>
                  <label className={styles.priceField}>
                    <span className="text-xs text-muted">Min</span>
                    <input
                      name="min"
                      inputMode="numeric"
                      autoComplete="off"
                      placeholder="0"
                      defaultValue={query.min ?? ""}
                      className={styles.input}
                    />
                  </label>
                  <label className={styles.priceField}>
                    <span className="text-xs text-muted">Max</span>
                    <input
                      name="max"
                      inputMode="numeric"
                      autoComplete="off"
                      placeholder="Any"
                      defaultValue={query.max ?? ""}
                      className={styles.input}
                    />
                  </label>
                </div>
              </fieldset>

              <div className={styles.filterActions}>
                <button type="submit" className="btn btn--inverse btn--block">
                  Apply filters
                </button>
                {active > 0 ? (
                  <Link
                    href={
                      query.q
                        ? `${action}?q=${encodeURIComponent(query.q)}`
                        : action
                    }
                    className="btn btn--ghost btn--block"
                  >
                    Clear all
                  </Link>
                ) : null}
              </div>
            </FiltersPanel>
          </aside>

          <div className="stack">
            <div className={styles.toolbar}>
              <p className="text-sm text-muted" role="status">
                {products.length === 1
                  ? "1 product"
                  : `${products.length} products`}
              </p>
              <SortSelect
                value={query.sort ?? "relevance"}
                options={sortOptions}
              />
            </div>

            {products.length ? (
              <ul
                role="list"
                className={`grid grid--catalog ${styles.results}`}
              >
                {products.map((p) => (
                  <li key={p.id}>
                    <ProductCard product={p} href={`/p/${p.slug}`} />
                  </li>
                ))}
              </ul>
            ) : (
              <EmptyState
                icon="search"
                title="No products match"
                action={
                  <Link href={action} className="btn btn--secondary">
                    Clear filters
                  </Link>
                }
              >
                Try removing a filter, widening the price range, or searching
                for something more general.
              </EmptyState>
            )}
          </div>
        </form>
      </div>
    </main>
  );
}
