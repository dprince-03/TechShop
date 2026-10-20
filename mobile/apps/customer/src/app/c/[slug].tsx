import { Stack, useLocalSearchParams } from "expo-router";

import { ProductList } from "@/features/catalogue/ProductList";
import { useLoad } from "@/features/catalogue/useLoad";
import { api } from "@/lib/api";

export default function Category() {
  const { slug, name } = useLocalSearchParams<{ slug: string; name?: string }>();
  const r = useLoad(() => api.get("/api/v1/search", { query: { category: slug, limit: 30, sort: "best_selling" } }), slug);
  return (
    <>
      <Stack.Screen options={{ title: name ?? "Category" }} />
      <ProductList items={r.data?.items ?? []} loading={r.loading} error={r.error} onRefresh={r.reload} empty="No products in this category yet." />
    </>
  );
}
