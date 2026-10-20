import { useRouter } from "expo-router";
import { useState } from "react";
import { Pressable, Text, View } from "react-native";

import { Button, Field, tokens, useColors } from "@techshop/mobile-ui";

import { ProductList } from "@/features/catalogue/ProductList";
import { useLoad } from "@/features/catalogue/useLoad";
import { useSession } from "@/features/auth/SessionProvider";
import { useCart } from "@/features/cart/CartProvider";
import { api } from "@/lib/api";

/** Home: search, top-level categories, then today's deals (biggest discounts first). */
export default function Home() {
  const router = useRouter();
  const c = useColors();
  const [q, setQ] = useState("");
  const { status, me } = useSession();
  const { count } = useCart();
  const cats = useLoad(async () => ((await api.get("/api/v1/categories")) ?? []).filter((x) => !x.parentId && x.isActive), "categories");
  const deals = useLoad(() => api.get("/api/v1/deals", { query: { limit: 20, sort: "discount" } }), "deals");

  return (
    <ProductList
      items={deals.data?.items ?? []}
      loading={deals.loading}
      error={deals.error ?? cats.error}
      onRefresh={() => {
        void cats.reload();
        void deals.reload();
      }}
      empty="No deals right now."
      header={
        <View style={{ gap: tokens.space[3] }}>
          <Button
            label={status === "signed-in" ? `Account${me ? ` · ${me.firstName}` : ""}` : "Sign in"}
            onPress={() => router.push(status === "signed-in" ? "/account" : "/sign-in")}
          />
          <Button label={count ? `Cart · ${count} item${count === 1 ? "" : "s"}` : "Cart"} onPress={() => router.push("/cart")} />
          <Field label="Search" value={q} onChangeText={setQ} placeholder="Phones, laptops, consoles…" returnKeyType="search" onSubmitEditing={() => q.trim() && router.push({ pathname: "/search", params: { q: q.trim() } })} />
          <Button label="Search" variant="primary" disabled={!q.trim()} onPress={() => router.push({ pathname: "/search", params: { q: q.trim() } })} />
          <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>Shop by category</Text>
          <View style={{ flexDirection: "row", flexWrap: "wrap", gap: tokens.space[2] }}>
            {(cats.data ?? []).map((cat) => (
              <Pressable
                key={cat.id}
                accessibilityRole="button"
                onPress={() => router.push({ pathname: "/c/[slug]", params: { slug: cat.slug, name: cat.name } })}
                style={{ minHeight: tokens.control.md, paddingHorizontal: tokens.space[4], justifyContent: "center", borderRadius: tokens.radius.pill, borderWidth: 1, borderColor: c.borderStrong }}
              >
                <Text style={{ color: c.text, fontSize: tokens.fontSize.base }}>{cat.name}</Text>
              </Pressable>
            ))}
          </View>
          <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>Deals</Text>
        </View>
      }
    />
  );
}
