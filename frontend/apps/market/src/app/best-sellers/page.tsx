import type { Metadata } from "next";
import { connection } from "next/server";

import { ListingPage } from "@/features/catalogue/ListingPage";
import { parseQuery } from "@/features/catalogue/query";

export const metadata: Metadata = {
  title: "Best sellers | TechShop Market",
  description: "The most popular tech on TechShop.",
};

export default async function BestSellersPage(
  props: PageProps<"/best-sellers">,
) {
  await connection();
  return (
    <ListingPage
      kind="best-sellers"
      title="Best sellers"
      description="Ranked by recent sales."
      breadcrumbs={[{ label: "Home", href: "/" }, { label: "Best sellers" }]}
      action="/best-sellers"
      query={parseQuery(await props.searchParams)}
    />
  );
}
