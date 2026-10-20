import type { Metadata } from "next";
import { connection } from "next/server";

import { ListingPage } from "@/features/catalogue/ListingPage";
import { parseQuery } from "@/features/catalogue/query";

export const metadata: Metadata = {
  title: "Deals | TechShop Market",
  description: "Discounted phones, laptops, gaming and more.",
};

export default async function DealsPage(props: PageProps<"/deals">) {
  await connection();
  return (
    <ListingPage
      kind="deals"
      title="Deals"
      description="Products with a lower price right now, biggest discounts first."
      breadcrumbs={[{ label: "Home", href: "/" }, { label: "Deals" }]}
      action="/deals"
      query={parseQuery(await props.searchParams)}
    />
  );
}
