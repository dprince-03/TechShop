import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { connection } from "next/server";

import { formatKobo } from "@techshop/api-client";

import { getProduct } from "@/features/catalogue/api";
import { ProductDetail } from "@/features/product/ProductDetail";

export async function generateMetadata(
  props: PageProps<"/p/[slug]">,
): Promise<Metadata> {
  const product = await getProduct((await props.params).slug).catch(() => null);
  if (!product) return {};
  const best = product.offers?.find((o) => o.buyBox) ?? product.offers?.[0];
  return {
    title: `${product.name} | TechShop Market`,
    description: best
      ? `${product.name}: ${formatKobo(best.price)} from ${best.sellerName}.`
      : product.name,
  };
}

export default async function ProductPage(props: PageProps<"/p/[slug]">) {
  await connection();
  const product = await getProduct((await props.params).slug);
  if (!product) notFound();
  return <ProductDetail product={product} />;
}
