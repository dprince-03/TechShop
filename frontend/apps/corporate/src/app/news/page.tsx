import type { Metadata } from "next";
import Link from "next/link";

import { EmptyState, PageHeader } from "@techshop/ui/components";

export const metadata: Metadata = { title: "Newsroom | TechShop" };

export default function NewsPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          eyebrow="Newsroom"
          title="News and announcements"
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Newsroom" }]}
        />
        <EmptyState
          icon="file"
          title="No announcements yet"
          action={
            <Link href="/contact#press" className="btn btn--secondary">
              Contact the press team
            </Link>
          }
        >
          Press releases, company news and media resources will be published
          here.
        </EmptyState>
      </div>
    </main>
  );
}
