import { Stack, useLocalSearchParams } from "expo-router";

import { Body } from "@techshop/mobile-ui";

import { ProductList } from "@/features/catalogue/ProductList";
import { useLoad } from "@/features/catalogue/useLoad";
import { api } from "@/lib/api";

export default function Search() {
  const { q = "" } = useLocalSearchParams<{ q?: string }>();
  const r = useLoad(() => api.get("/api/v1/search", { query: { q: q.slice(0, 100), limit: 30 } }), q);
  const suggestion = r.data?.didYouMean && r.data.didYouMean.toLowerCase() !== q.toLowerCase() ? r.data.didYouMean : null;
  return (
    <>
      <Stack.Screen options={{ title: q ? `“${q}”` : "Search" }} />
      <ProductList
        items={r.data?.items ?? []}
        loading={r.loading}
        error={r.error}
        onRefresh={r.reload}
        header={r.data ? <Body muted>{`${r.data.total} products${suggestion ? ` · did you mean “${suggestion}”?` : ""}`}</Body> : undefined}
        empty="Nothing found. Try a shorter or more general search."
      />
    </>
  );
}
