import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { Breadcrumbs, EmptyState } from "@techshop/ui/components";

import { SavedList } from "@/features/saved/components/SavedList";
import { listSaved } from "@/features/saved/server";
import { bff } from "@/lib/bff";

import styles from "../curated.module.css";

export const metadata: Metadata = {
  title: "Saved items | TechShop Market",
  robots: { index: false },
};

export default async function SavedPage() {
  await connection();
  const session = await bff.getSession();
  const items = session ? await listSaved().catch(() => null) : [];
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <div className="stack stack--sm">
          <Breadcrumbs
            items={[{ label: "Home", href: "/" }, { label: "Saved items" }]}
          />
          <h1 className={styles.title}>Saved items</h1>
        </div>
        {items === null ? (
          <p className="alert alert--error" role="alert">
            We couldn’t load your saved items. Please try again.
          </p>
        ) : items.length ? (
          <SavedList items={items} />
        ) : (
          <EmptyState
            icon="heart"
            title="Nothing saved yet"
            action={
              <div className={`cluster ${styles.center}`}>
                {session ? null : (
                  <Link
                    href="/sign-in?next=%2Fsaved"
                    className="btn btn--secondary"
                  >
                    Sign in
                  </Link>
                )}
                <Link href="/deals" className="btn btn--ghost">
                  Browse deals
                </Link>
              </div>
            }
          >
            Tap the heart on any product to save it for later.
            {session ? null : " Sign in to keep saved items across devices."}
          </EmptyState>
        )}
      </div>
    </main>
  );
}
