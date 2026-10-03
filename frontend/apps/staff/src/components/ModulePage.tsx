import Link from "next/link";

import { EmptyState, Icon } from "@techshop/ui/components";

import type { ModuleConfig } from "@/lib/moduleData";

import { DataTable } from "./DataTable";
import styles from "./ModulePage.module.css";

type Props = { slug: string; config: ModuleConfig; tab?: string; q?: string };

/** Standard staff module screen: header, key figures, tabs, search, sample table. */
export function ModulePage({ slug, config, tab = "all", q = "" }: Props) {
  const activeTab = config.tabs.some((t) => t.key === tab) ? tab : "all";
  const needle = q.trim().toLowerCase();
  const rows = config.rows.filter(
    (r) =>
      (activeTab === "all" || r.tab === activeTab) &&
      (!needle || r.cells.join(" ").toLowerCase().includes(needle)),
  );

  return (
    <div className="stack stack--lg">
      <div className={styles.head}>
        <div className="stack stack--sm">
          <h1 className={styles.title}>{config.title}</h1>
          <p className="text-sm text-muted">
            {config.description} · Sample data
          </p>
        </div>
        {config.primary ? (
          <Link href={config.primary.href} className="btn btn--primary">
            <Icon name="plus" size={16} />
            {config.primary.label}
          </Link>
        ) : null}
      </div>

      <section aria-label="Key figures">
        <ul role="list" className={styles.kpis}>
          {config.kpis.map((k) => (
            <li key={k.label} className={`card ${styles.kpi}`}>
              <p className="text-sm text-muted">{k.label}</p>
              <p className={styles.kpiValue}>{k.value}</p>
              {k.note ? <p className="text-xs text-muted">{k.note}</p> : null}
            </li>
          ))}
        </ul>
      </section>

      <section
        className={`card ${styles.panel}`}
        aria-labelledby={`${slug}-list`}
      >
        <h2 id={`${slug}-list`} className="visually-hidden">
          {config.title} list
        </h2>
        <div className={styles.toolbar}>
          {config.tabs.length > 1 ? (
            <nav aria-label={`${config.title} filters`} className={styles.tabs}>
              {config.tabs.map((t) => (
                <Link
                  key={t.key}
                  href={t.key === "all" ? `/${slug}` : `/${slug}?tab=${t.key}`}
                  className={styles.tab}
                  aria-current={t.key === activeTab ? "page" : undefined}
                >
                  {t.label}
                </Link>
              ))}
            </nav>
          ) : (
            <span />
          )}
          <form
            method="get"
            action={`/${slug}`}
            role="search"
            className={styles.search}
          >
            {activeTab !== "all" ? (
              <input type="hidden" name="tab" value={activeTab} />
            ) : null}
            <label htmlFor={`${slug}-q`} className="visually-hidden">
              Search {config.title.toLowerCase()}
            </label>
            <Icon name="search" size={16} className={styles.searchIcon} />
            <input
              id={`${slug}-q`}
              name="q"
              type="search"
              defaultValue={q}
              placeholder="Search…"
              className={styles.searchInput}
            />
          </form>
        </div>
        {rows.length ? (
          <DataTable
            caption={`${config.title}: ${config.tabs.find((t) => t.key === activeTab)?.label ?? "All"}`}
            columns={config.columns}
            rows={rows}
            statusColumn={config.statusColumn}
            tones={config.tones}
          />
        ) : (
          <div className={styles.empty}>
            <EmptyState
              icon="box"
              title={
                config.rows.length
                  ? "Nothing matches"
                  : `No ${config.title.toLowerCase()} data yet`
              }
            >
              {config.rows.length
                ? "Try another tab or clear your search."
                : "This module will show live data once it’s connected to the TechShop API."}
            </EmptyState>
          </div>
        )}
      </section>
    </div>
  );
}
