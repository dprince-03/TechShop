import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { PageHeader } from "@techshop/ui/components";

import { getMyOrder } from "@/features/orders/server";
import { ReviewItemForm } from "@/features/reviews/components/ReviewForms";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Write a review | TechShop Market",
  robots: { index: false },
};

export default async function ReviewItemPage(
  props: PageProps<"/orders/[number]/review/[lineId]">,
) {
  await connection();
  const { number, lineId } = await props.params;
  await bff.requireSession(`/orders/${number}/review/${lineId}`);
  const order = await getMyOrder(number);
  const f = (order?.fulfilments ?? []).find((x) =>
    (x.lines ?? []).some((l) => l.id === lineId),
  );
  const line = f?.lines?.find((l) => l.id === lineId);
  if (!order || !f || !line) notFound();
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Write a review"
          lead={`${line.productName} · ${line.variantName}`}
          breadcrumbs={[
            { label: "Orders", href: "/orders" },
            { label: order.orderNumber, href: `/orders/${order.orderNumber}` },
            { label: "Review" },
          ]}
        />
        <div className="card card--roomy">
          {f.status === "delivered" && !line.cancelledAt ? (
            <ReviewItemForm lineId={line.id} />
          ) : (
            <p>You can review this item once it has been delivered.</p>
          )}
        </div>
      </div>
    </main>
  );
}
