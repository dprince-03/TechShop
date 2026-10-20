import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { connection } from "next/server";

import { getCategories } from "@/features/catalogue/api";
import { ListingPage } from "@/features/catalogue/ListingPage";
import { parseQuery } from "@/features/catalogue/query";

export async function generateMetadata(
  props: PageProps<"/search">,
): Promise<Metadata> {
  const { q } = parseQuery(await props.searchParams);
  return {
    title: q ? `“${q}” | TechShop Market` : "Search | TechShop Market",
    robots: { index: false },
  };
}

export default async function SearchPage(props: PageProps<"/search">) {
  await connection();
  const query = parseQuery(await props.searchParams);
  const categories = await getCategories();
  // Searching a whole category is browsing it: send the shopper to the category page.
  if (
    !query.q &&
    query.category &&
    categories.some((c) => c.slug === query.category)
  ) {
    redirect(`/c/${query.category}`);
  }
  return (
    <ListingPage
      title={query.q ? `Results for “${query.q}”` : "All products"}
      breadcrumbs={[{ label: "Home", href: "/" }, { label: "Search" }]}
      action="/search"
      query={query}
      artCategory={query.category}
      categories={categories}
    />
  );
}
