import Link from "next/link";
import type { CSSProperties, ReactNode } from "react";

import type { Product, Schema } from "@techshop/api-client";
import {
  Breadcrumbs,
  EmptyState,
  ProductCard,
  type Crumb,
} from "@techshop/ui/components";

import type { MarketCategory } from "../mappers";
import {
  activeFilterCount,
  conditionOptions,
  listingHref,
  sellerOptions,
  sortOptions,
  type ListingQuery,
} from "../query";

import { FiltersPanel } from "./FiltersPanel";
import styles from "./ProductListing.module.css";
import { SortSelect } from "./SortSelect";

type Props = {
  title: string;
  description?: string;
  breadcrumbs: Crumb[];
  /** URL the filter form submits to (the current page path). */
  action: string;
  query: ListingQuery;
  products: Product[];
  total: number;
  /** Brand facet from the API (value = brand slug). */
  brands?: Schema<"Facet">[];
  nextCursor?: string;
  /** Extra navigation above the filters, e.g. subcategory links. */
  browse?: ReactNode;
  /** Category options for the search page's category filter. */
  categories?: MarketCategory[];
  /** Shown above the grid, e.g. "Showing results for …". */
  notice?: ReactNode;
};

/**
 * Listing template: filters (GET form → URL params), sort, result count, product grid, next page.
 * Works without JavaScript; JS only adds auto-submit on sort and the mobile filter toggle.
 */
export function ProductListing({
  title,
  description,
  breadcrumbs,
  action,
  query,
  products,
  total,
  brands = [],
  nextCursor,
  browse,
  categories,
  notice,
}: Props) {
  const active = activeFilterCount(query);
  const clearHref = query.q
    ? `${action}?q=${encodeURIComponent(query.q)}`
    : action;
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

              {categories?.length ? (
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
                      .filter((c) => c.kind !== "vehicle")
                      .map((c) => (
                        <option key={c.slug} value={c.slug}>
                          {c.name}
                        </option>
                      ))}
                  </select>
                </fieldset>
              ) : null}

              {brands.length ? (
                <fieldset className={styles.group}>
                  <legend className={styles.legend}>Brand</legend>
                  {brands.map((b) => (
                    <label key={b.value} className={styles.option}>
                      <input
                        type="checkbox"
                        name="brand"
                        value={b.value}
                        defaultChecked={query.brands.includes(b.value)}
                      />
                      {b.label}
                      <span className="text-xs text-subtle">({b.count})</span>
                    </label>
                  ))}
                </fieldset>
              ) : null}

              <fieldset className={styles.group}>
                <legend className={styles.legend}>Condition</legend>
                {conditionOptions.map((c) => (
                  <label key={c.value} className={styles.option}>
                    <input
                      type="checkbox"
                      name="condition"
                      value={c.value}
                      defaultChecked={query.conditions.includes(c.value)}
                    />
                    {c.label}
                  </label>
                ))}
              </fieldset>

              <fieldset className={styles.group}>
                <legend className={styles.legend}>Seller</legend>
                <label className={styles.option}>
                  <input
                    type="radio"
                    name="seller"
                    value=""
                    defaultChecked={!query.sellerType}
                  />
                  All sellers
                </label>
                {sellerOptions.map((s) => (
                  <label key={s.value} className={styles.option}>
                    <input
                      type="radio"
                      name="seller"
                      value={s.value}
                      defaultChecked={query.sellerType === s.value}
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

              <label className={styles.option}>
                <input
                  type="checkbox"
                  name="inStock"
                  value="1"
                  defaultChecked={query.inStock}
                />
                In stock only
              </label>

              <div className={styles.filterActions}>
                <button type="submit" className="btn btn--inverse btn--block">
                  Apply filters
                </button>
                {active > 0 ? (
                  <Link href={clearHref} className="btn btn--ghost btn--block">
                    Clear all
                  </Link>
                ) : null}
              </div>
            </FiltersPanel>
          </aside>

          <div className="stack">
            <div className={styles.toolbar}>
              <p className="text-sm text-muted" role="status">
                {total === 1 ? "1 product" : `${total} products`}
              </p>
              <SortSelect value={query.sort} options={sortOptions} />
            </div>

            {notice}

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
                  <Link href={clearHref} className="btn btn--secondary">
                    Clear filters
                  </Link>
                }
              >
                Try removing a filter, widening the price range, or searching
                for something more general.
              </EmptyState>
            )}

            {nextCursor || query.cursor ? (
              <nav aria-label="Pages" className="cluster cluster--between">
                {query.cursor ? (
                  <Link href={listingHref(action, query)} className="link">
                    Back to first page
                  </Link>
                ) : (
                  <span />
                )}
                {nextCursor ? (
                  <Link
                    href={listingHref(action, query, nextCursor)}
                    className="btn btn--secondary"
                    rel="next"
                  >
                    Next page
                  </Link>
                ) : null}
              </nav>
            ) : null}
          </div>
        </form>
      </div>
    </main>
  );
}
