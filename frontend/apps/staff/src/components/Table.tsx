import Link from "next/link";
import type { ReactNode } from "react";

import { EmptyState } from "@techshop/ui/components";

import styles from "./DataTable.module.css";

export type Column<T> = {
  label: string;
  /** Cell content for a row. */
  cell: (row: T) => ReactNode;
  /** Right-aligned tabular numbers (money, counts). */
  numeric?: boolean;
};

type Props<T> = {
  caption: string;
  columns: Column<T>[];
  rows: T[];
  rowKey: (row: T) => string;
  /** Links the first cell (the row header). */
  href?: (row: T) => string | undefined;
  empty?: { title: string; body: string };
};

/** Live data table: typed rows, first cell is the row header, scrolls inside a focusable region. */
export function Table<T>({
  caption,
  columns,
  rows,
  rowKey,
  href,
  empty,
}: Props<T>) {
  if (!rows.length && empty) {
    return (
      <EmptyState icon="box" title={empty.title}>
        {empty.body}
      </EmptyState>
    );
  }
  return (
    <div
      className={styles.wrap}
      role="region"
      aria-label={caption}
      tabIndex={0}
    >
      <table className={styles.table}>
        <caption className="visually-hidden">{caption}</caption>
        <thead>
          <tr>
            {columns.map((c) => (
              <th
                key={c.label}
                scope="col"
                className={c.numeric ? styles.num : undefined}
              >
                {c.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => {
            const link = href?.(r);
            return (
              <tr key={rowKey(r)}>
                {columns.map((c, i) =>
                  i === 0 ? (
                    <th key={c.label} scope="row">
                      {link ? (
                        <Link href={link} className="link">
                          {c.cell(r)}
                        </Link>
                      ) : (
                        c.cell(r)
                      )}
                    </th>
                  ) : (
                    <td
                      key={c.label}
                      className={
                        c.numeric ? `${styles.num} tabular-nums` : undefined
                      }
                    >
                      {c.cell(r)}
                    </td>
                  ),
                )}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
