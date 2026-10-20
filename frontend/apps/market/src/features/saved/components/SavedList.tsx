"use client";

import Link from "next/link";
import { useState, useTransition } from "react";

import type { Schema } from "@techshop/api-client";
import { CategoryArt } from "@techshop/ui/components";

import { toggleSaved } from "../actions";

import styles from "./SavedList.module.css";

const when = (iso: string) =>
  new Intl.DateTimeFormat("en-NG", {
    dateStyle: "medium",
    timeZone: "Africa/Lagos",
  }).format(new Date(iso));

export function SavedList({
  items,
}: {
  items: Schema<"CatalogListSavedRow">[];
}) {
  const [message, setMessage] = useState("");
  const [pending, startTransition] = useTransition();
  return (
    <>
      <p role="status" className="visually-hidden">
        {message}
      </p>
      <ul role="list" className={styles.list}>
        {items.map((p) => (
          <li key={p.id} className={`card ${styles.item}`}>
            <div className={`frame ${styles.thumb}`}>
              <CategoryArt size="sm" />
            </div>
            <div className="stack stack--sm">
              <Link href={`/p/${p.slug}`} className={styles.name}>
                {p.name}
              </Link>
              <p className="text-xs text-muted">Saved {when(p.savedAt)}</p>
            </div>
            <button
              type="button"
              className="btn btn--ghost btn--sm"
              disabled={pending}
              onClick={() =>
                startTransition(async () => {
                  const r = await toggleSaved(p.id);
                  setMessage(
                    r.ok ? `Removed ${p.name} from saved items.` : r.message,
                  );
                })
              }
            >
              Remove
            </button>
          </li>
        ))}
      </ul>
    </>
  );
}
