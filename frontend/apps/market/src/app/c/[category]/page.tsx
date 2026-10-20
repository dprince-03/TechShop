import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { getCategory } from "@/features/catalogue/api";
import { CarsListing } from "@/features/cars/CarsListing";
import { SubcategoryNav } from "@/features/catalogue/components/SubcategoryNav";
import { ListingPage } from "@/features/catalogue/ListingPage";
import { parseQuery } from "@/features/catalogue/query";

export async function generateMetadata(
  props: PageProps<"/c/[category]">,
): Promise<Metadata> {
  const c = await getCategory((await props.params).category);
  return c
    ? { title: `${c.name} | TechShop Market`, description: c.description }
    : {};
}

export default async function CategoryPage(props: PageProps<"/c/[category]">) {
  await connection();
  const category = await getCategory((await props.params).category);
  if (!category) notFound();

  // Cars are listings with a viewing flow, not a product grid.
  if (category.kind === "vehicle") return <CarsListing category={category} />;

  return (
    <ListingPage
      title={category.name}
      description={category.description}
      breadcrumbs={[{ label: "Home", href: "/" }, { label: category.name }]}
      action={`/c/${category.slug}`}
      query={{
        ...parseQuery(await props.searchParams),
        category: category.slug,
      }}
      artCategory={category.slug}
      browse={<SubcategoryNav category={category} />}
    />
  );
}
