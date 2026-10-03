import type { Metadata } from "next";
import { notFound } from "next/navigation";

import type { CategorySlug } from "@techshop/api-client";
import {
  categoryBySlug,
  getSubcategory,
  queryProducts,
} from "@techshop/fixtures";

import { ProductListing } from "@/components/ProductListing";
import { SubcategoryNav } from "@/components/SubcategoryNav";
import { parseQuery } from "@/lib/query";

const resolve = (categorySlug: string, subSlug: string) => {
  const category = categoryBySlug[categorySlug as CategorySlug];
  const sub = category ? getSubcategory(category.slug, subSlug) : undefined;
  return category && sub ? { category, sub } : undefined;
};

export async function generateMetadata(
  props: PageProps<"/c/[category]/[sub]">,
): Promise<Metadata> {
  const { category, sub } = await props.params;
  const r = resolve(category, sub);
  return r
    ? {
        title: `${r.sub.name} ${r.category.name.toLowerCase()} | TechShop Market`,
      }
    : {};
}

export default async function SubcategoryPage(
  props: PageProps<"/c/[category]/[sub]">,
) {
  const params = await props.params;
  const r = resolve(params.category, params.sub);
  if (!r || r.category.slug === "cars") notFound();
  const { category, sub } = r;

  const query = {
    ...parseQuery(await props.searchParams),
    category: category.slug,
    subcategory: sub.slug,
  };
  return (
    <ProductListing
      title={`${sub.name} ${category.name.toLowerCase()}`}
      breadcrumbs={[
        { label: "Home", href: "/" },
        { label: category.name, href: `/c/${category.slug}` },
        { label: sub.name },
      ]}
      action={`/c/${category.slug}/${sub.slug}`}
      query={query}
      products={queryProducts(query)}
      browse={<SubcategoryNav category={category} current={sub.slug} />}
    />
  );
}
