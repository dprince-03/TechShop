import type { Metadata } from "next";
import Link from "next/link";

import { EmptyState } from "@techshop/ui/components";

import { modules } from "@/lib/moduleData";

import styles from "./search.module.css";

export const metadata: Metadata = { title: "Search | TechShop Staff" };

export default async function StaffSearch(props: PageProps<"/search">) {
  const sp = await props.searchParams;
  const q = (Array.isArray(sp.q) ? sp.q[0] : sp.q)?.trim() ?? "";
  const needle = q.toLowerCase();
  const results = needle
    ? Object.entries(modules).flatMap(([slug, m]) =>
        m.rows
          .filter((r) => r.cells.join(" ").toLowerCase().includes(needle))
          .map((r) => ({
            slug,
            module: m.title,
            title: r.cells[0],
            detail: r.cells.slice(1, 3).join(" · "),
            href: r.href ?? `/${slug}?q=${encodeURIComponent(r.cells[0])}`,
          })),
      )
    : [];

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1 className={styles.title}>{q ? `Results for “${q}”` : "Search"}</h1>
        <p className="text-sm text-muted" role="status">
          {q
            ? `${results.length} result${results.length === 1 ? "" : "s"} across all modules · Sample data`
            : "Use the search bar to find orders, products, customers and more."}
        </p>
      </div>
      {q && !results.length ? (
        <EmptyState icon="search" title="No matches">
          Try an order number (TS-10482), a customer name or a product name.
        </EmptyState>
      ) : null}
      {results.length ? (
        <ul role="list" className={`card ${styles.list}`}>
          {results.map((r) => (
            <li key={`${r.slug}-${r.title}`}>
              <Link href={r.href} className={styles.item}>
                <span className="badge">{r.module}</span>
                <span className="stack stack--sm">
                  <strong className="text-sm">{r.title}</strong>
                  <span className="text-xs text-muted">{r.detail}</span>
                </span>
              </Link>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
