import type { Metadata } from "next";

import { queryProducts } from "@techshop/fixtures";

import { ProductListing } from "@/components/ProductListing";
import { parseQuery } from "@/lib/query";

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
  const query = parseQuery(await props.searchParams);
  const title = query.q ? `Results for “${query.q}”` : "All products";
  return (
    <ProductListing
      title={title}
      breadcrumbs={[{ label: "Home", href: "/" }, { label: "Search" }]}
      action="/search"
      query={query}
      products={queryProducts(query)}
      showCategoryFilter
    />
  );
}
