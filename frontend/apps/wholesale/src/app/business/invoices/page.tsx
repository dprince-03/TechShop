import type { Metadata } from "next";
import Link from "next/link";

import { EmptyState, PageHeader } from "@techshop/ui/components";

export const metadata: Metadata = {
  title: "Invoices | TechShop Wholesale & Retail",
  robots: { index: false },
};

export default function InvoicesPage() {
  return (
    <main id="main" className="section section--page">
      <div className="container stack stack--lg">
        <PageHeader
          title="Invoices"
          breadcrumbs={[{ label: "Home", href: "/" }, { label: "Invoices" }]}
        />
        <EmptyState
          icon="receipt"
          title="Sign in to see your invoices"
          action={
            <div className="cluster">
              <Link href="/business/sign-in" className="btn btn--primary">
                Business sign in
              </Link>
              <Link href="/business/register" className="btn btn--ghost">
                Open an account
              </Link>
            </div>
          }
        >
          Download VAT invoices, see what’s due and track payments for every
          order on your business account.
        </EmptyState>
      </div>
    </main>
  );
}
