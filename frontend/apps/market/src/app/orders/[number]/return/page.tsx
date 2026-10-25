import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { EmptyState, PageHeader } from "@techshop/ui/components";

import { ReturnForm } from "@/features/aftersales/components/ReturnForm";
import { getMyOrder } from "@/features/orders/server";
import { inReturnWindow } from "@/features/aftersales/window";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Return items | TechShop Market",
  robots: { index: false },
};

export default async function ReturnPage(
  props: PageProps<"/orders/[number]/return">,
) {
  await connection();
  const { number } = await props.params;
  await bff.requireSession(`/orders/${number}/return`);
  const order = await getMyOrder(number);
  if (!order) notFound();
  const lines = (order.fulfilments ?? [])
    .filter((f) => f.status === "delivered" && inReturnWindow(f.deliveredAt))
    .flatMap((f) => f.lines ?? [])
    .filter((l) => !l.cancelledAt)
    .map((l) => ({
      id: l.id,
      name: l.productName,
      variant: l.variantName,
      quantity: l.quantity,
    }));

  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Return items"
          lead={`From order ${order.orderNumber}. Faulty, damaged, wrong or not-as-described items within 7 days of delivery are usually approved straight away.`}
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Orders", href: "/orders" },
            { label: order.orderNumber, href: `/orders/${order.orderNumber}` },
            { label: "Return" },
          ]}
        />
        {lines.length ? (
          <section className="card card--roomy" aria-label="Return request">
            <ReturnForm number={order.orderNumber} lines={lines} />
          </section>
        ) : (
          <EmptyState icon="box" title="Nothing to return right now">
            Returns open when an item is delivered and stay open for 7 days.
            Faulty after that?{" "}
            <Link href="/account/warranty" className="link">
              Make a warranty claim
            </Link>
            .
          </EmptyState>
        )}
        <p className="text-sm text-muted">
          Read the{" "}
          <Link href="/help/returns" className="link">
            returns guide
          </Link>{" "}
          for what can come back and how refunds work.
        </p>
      </div>
    </main>
  );
}
