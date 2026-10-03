import type { Metadata } from "next";
import { notFound } from "next/navigation";

import type { CategorySlug } from "@techshop/api-client";
import {
  categories,
  categoryBySlug,
  ownStock,
  wholesaleProducts,
} from "@techshop/fixtures";
import {
  EmptyState,
  PageHeader,
  ProductCard,
  SectionHeader,
} from "@techshop/ui/components";

import { WholesaleCard } from "@/components/WholesaleCard";

import styles from "../../inner.module.css";

const SLUGS = categories.filter((c) => c.slug !== "cars").map((c) => c.slug);

export const dynamicParams = false;

export function generateStaticParams() {
  return SLUGS.map((category) => ({ category }));
}

export async function generateMetadata(
  props: PageProps<"/c/[category]">,
): Promise<Metadata> {
  const c = categoryBySlug[(await props.params).category as CategorySlug];
  return c ? { title: `${c.name} | TechShop Wholesale & Retail` } : {};
}

export default async function CategoryPage(props: PageProps<"/c/[category]">) {
  const slug = (await props.params).category as CategorySlug;
  if (!SLUGS.includes(slug)) notFound();
  const category = categoryBySlug[slug];
  const bulk = wholesaleProducts.filter((p) => p.category === slug);
  const retail = ownStock.filter((p) => p.category === slug);

  return (
    <main id="main" className="section section--page">
      <div className="container container--wide stack stack--lg">
        <PageHeader
          title={category.name}
          lead={category.description}
          breadcrumbs={[
            { label: "Home", href: "/" },
            { label: "Categories", href: "/categories" },
            { label: category.name },
          ]}
        />
        {bulk.length ? (
          <section aria-labelledby="bulk-title">
            <SectionHeader
              id="bulk-title"
              title="Trade prices"
              description="Bulk pricing from the minimum order quantity."
            />
            <ul role="list" className={`grid ${styles.bulkGrid}`}>
              {bulk.map((p) => (
                <li key={p.id}>
                  <WholesaleCard product={p} />
                </li>
              ))}
            </ul>
          </section>
        ) : null}
        <section aria-labelledby="retail-title">
          <SectionHeader
            id="retail-title"
            title="Retail"
            description="Single units at our standard price."
          />
          {retail.length ? (
            <ul role="list" className="grid grid--catalog">
              {retail.map((p) => (
                <li key={p.id}>
                  <ProductCard product={p} href={`/p/${p.slug}`} />
                </li>
              ))}
            </ul>
          ) : (
            <EmptyState
              icon="box"
              title={`No ${category.name.toLowerCase()} in stock right now`}
            >
              Request a quote and we’ll source it for you.
            </EmptyState>
          )}
        </section>
      </div>
    </main>
  );
}
