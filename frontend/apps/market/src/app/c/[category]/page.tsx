import type { Metadata } from "next";
import { notFound } from "next/navigation";

import type { CategorySlug } from "@techshop/api-client";
import { categoryBySlug, cars, queryProducts } from "@techshop/fixtures";
import { Breadcrumbs } from "@techshop/ui/components";

import { CarCard } from "@/components/CarCard";
import { ProductListing } from "@/components/ProductListing";
import { SubcategoryNav } from "@/components/SubcategoryNav";
import { parseQuery } from "@/lib/query";

import styles from "./category.module.css";

const getCategory = (slug: string) => categoryBySlug[slug as CategorySlug];

export async function generateMetadata(
  props: PageProps<"/c/[category]">,
): Promise<Metadata> {
  const { category } = await props.params;
  const c = getCategory(category);
  return c
    ? { title: `${c.name} | TechShop Market`, description: c.description }
    : {};
}

export default async function CategoryPage(props: PageProps<"/c/[category]">) {
  const { category: slug } = await props.params;
  const category = getCategory(slug);
  if (!category) notFound();

  const crumbs = [{ label: "Home", href: "/" }, { label: category.name }];

  // Cars are listings with a viewing flow, not a product grid.
  if (category.slug === "cars") {
    return (
      <main id="main" className="section section--page">
        <div className="container container--wide stack stack--lg">
          <div className="stack stack--sm">
            <Breadcrumbs items={crumbs} />
            <h1 className={styles.title}>Cars</h1>
            <p className="text-muted measure">
              {category.description} Book a viewing before you pay.
            </p>
          </div>
          <p className="text-sm text-muted" role="status">
            {cars.length} cars
          </p>
          <ul role="list" className={`grid ${styles.carGrid}`}>
            {cars.map((car) => (
              <li key={car.id}>
                <CarCard car={car} />
              </li>
            ))}
          </ul>
        </div>
      </main>
    );
  }

  const query = {
    ...parseQuery(await props.searchParams),
    category: category.slug,
  };
  return (
    <ProductListing
      title={category.name}
      description={category.description}
      breadcrumbs={crumbs}
      action={`/c/${category.slug}`}
      query={query}
      products={queryProducts(query)}
      browse={<SubcategoryNav category={category} />}
    />
  );
}
