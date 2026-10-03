import Link from "next/link";

import type { Tone } from "@/lib/moduleData";

import styles from "./DataTable.module.css";

type Props = {
  caption: string;
  columns: string[];
  rows: { href?: string; cells: string[] }[];
  statusColumn?: number;
  tones?: Record<string, Tone>;
};

const numeric = /^(₦|[\d.,]+%?$|[▲▼])/;

/** Scrollable data table. First cell is the row header (linked when `href` is set). Status shows dot + text. */
export function DataTable({
  caption,
  columns,
  rows,
  statusColumn,
  tones = {},
}: Props) {
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
              <th key={c} scope="col">
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.cells.join("|")}>
              {r.cells.map((cell, i) => {
                if (i === 0) {
                  return (
                    <th key={i} scope="row">
                      {r.href ? (
                        <Link href={r.href} className="link">
                          {cell}
                        </Link>
                      ) : (
                        cell
                      )}
                    </th>
                  );
                }
                if (i === statusColumn) {
                  return (
                    <td key={i}>
                      <span
                        className={`${styles.status} ${styles[tones[cell] ?? "neutral"]}`}
                      >
                        {cell}
                      </span>
                    </td>
                  );
                }
                return (
                  <td
                    key={i}
                    className={numeric.test(cell) ? styles.num : undefined}
                  >
                    {cell}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
