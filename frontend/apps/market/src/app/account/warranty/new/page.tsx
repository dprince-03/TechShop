import type { Metadata } from "next";
import Link from "next/link";
import { connection } from "next/server";

import { EmptyState, PageHeader } from "@techshop/ui/components";

import { ClaimForm } from "@/features/aftersales/components/WarrantyForms";
import { coveredItems } from "@/features/aftersales/server";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Make a warranty claim | TechShop Market",
  robots: { index: false },
};

export default async function NewClaimPage(
  props: PageProps<"/account/warranty/new">,
) {
  await connection();
  await bff.requireSession("/account/warranty/new");
  const line = (await props.searchParams).line;
  const items = ((await coveredItems()) ?? []).filter((i) => !i.openClaim);
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Make a warranty claim"
          lead="Tell us what's wrong, then bring the device to a TechShop store. We check it there and keep you posted."
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Warranty", href: "/account/warranty" },
            { label: "New claim" },
          ]}
        />
        {items.length ? (
          <section className="card card--roomy" aria-label="Warranty claim">
            <ClaimForm
              selected={typeof line === "string" ? line : undefined}
              items={items.map((i) => ({
                value: i.orderLineId,
                label: `${i.productName} · ${i.variantName} (order ${i.orderNumber})`,
                serialised: i.serialised,
              }))}
            />
          </section>
        ) : (
          <EmptyState icon="shield" title="No items to claim for">
            Nothing you’ve bought is under warranty right now, or it already has
            an open claim.{" "}
            <Link href="/account/support/new" className="link">
              Ask support
            </Link>{" "}
            if you think that’s wrong.
          </EmptyState>
        )}
      </div>
    </main>
  );
}
