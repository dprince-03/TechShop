import type { Metadata } from "next";
import { connection } from "next/server";

import { PageHeader } from "@techshop/ui/components";

import { NewTicketForm } from "@/features/aftersales/components/SupportForms";
import { listMyOrders } from "@/features/orders/server";
import { bff } from "@/lib/bff";

export const metadata: Metadata = {
  title: "Ask support | TechShop Market",
  robots: { index: false },
};

export default async function NewTicketPage(
  props: PageProps<"/account/support/new">,
) {
  await connection();
  await bff.requireSession("/account/support/new");
  const order = (await props.searchParams).order;
  const orders = ((await listMyOrders().catch(() => null))?.items ?? []).map(
    (o) => o.orderNumber,
  );
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow stack stack--lg">
        <PageHeader
          title="Ask support"
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Support", href: "/account/support" },
            { label: "New question" },
          ]}
        />
        <section className="card card--roomy" aria-label="Your question">
          <NewTicketForm
            orders={orders}
            orderNumber={typeof order === "string" ? order : undefined}
          />
        </section>
      </div>
    </main>
  );
}
