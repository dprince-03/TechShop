import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { getCategory } from "@/features/catalogue/api";
import { SubcategoryNav } from "@/features/catalogue/components/SubcategoryNav";
import { ListingPage } from "@/features/catalogue/ListingPage";
import { parseQuery } from "@/features/catalogue/query";

async function resolve(params: Promise<{ category: string; sub: string }>) {
  const { category: c, sub: s } = await params;
  const category = await getCategory(c);
  const sub = category?.subcategories.find((x) => x.slug === s);
  return category && sub ? { category, sub } : undefined;
}

export async function generateMetadata(
  props: PageProps<"/c/[category]/[sub]">,
): Promise<Metadata> {
  const r = await resolve(props.params);
  return r ? { title: `${r.sub.name} | TechShop Market` } : {};
}

export default async function SubcategoryPage(
  props: PageProps<"/c/[category]/[sub]">,
) {
  await connection();
  const r = await resolve(props.params);
  if (!r || r.category.kind === "vehicle") notFound();
  const { category, sub } = r;
  return (
    <ListingPage
      title={sub.name}
      breadcrumbs={[
        { label: "Home", href: "/" },
        { label: category.name, href: `/c/${category.slug}` },
        { label: sub.name },
      ]}
      action={`/c/${category.slug}/${sub.slug}`}
      query={{ ...parseQuery(await props.searchParams), category: sub.apiSlug }}
      artCategory={category.slug}
      browse={<SubcategoryNav category={category} current={sub.slug} />}
    />
  );
}
