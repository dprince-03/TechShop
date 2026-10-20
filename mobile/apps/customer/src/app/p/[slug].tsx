import { Stack, useLocalSearchParams } from "expo-router";
import { ActivityIndicator, Text, View } from "react-native";

import { formatKobo } from "@techshop/api-client";
import { Body, Card, Notice, Screen, Title, tokens, useColors } from "@techshop/mobile-ui";

import { useLoad } from "@/features/catalogue/useLoad";
import { api } from "@/lib/api";

const CONDITION: Record<string, string> = { new: "New", uk_used: "UK-used", refurbished: "Refurbished", open_box: "Open box" };

/** Product detail: every seller offer (best first), warranty and specs. Buying in the app comes next. */
export default function Product() {
  const { slug } = useLocalSearchParams<{ slug: string }>();
  const c = useColors();
  const r = useLoad(() => api.get("/api/v1/products/{slug}", { path: { slug } }), slug);
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading product" />;
  if (!r.data) return <Screen><Notice tone="danger">{r.error ?? "Not found."}</Notice></Screen>;
  const p = r.data;
  const offers = p.offers ?? [];
  const specs = (p.attributes ?? []).map((a) => ({ label: a.label, value: a.valueText ?? (a.valueNumber !== null ? `${a.valueNumber}${a.unit ? ` ${a.unit}` : ""}` : a.valueBool !== null ? (a.valueBool ? "Yes" : "No") : null) })).filter((s) => s.value);
  return (
    <Screen>
      <Stack.Screen options={{ title: p.brand ?? "Product" }} />
      <Title>{p.name}</Title>
      {!p.inStock ? <Notice tone="warning">Out of stock right now.</Notice> : null}
      {offers.map((o) => (
        <Card key={o.listingId}>
          <Text style={{ color: c.text, fontSize: tokens.fontSize["2xl"], fontWeight: "600", fontVariant: ["tabular-nums"] }}>{formatKobo(o.price)}</Text>
          <Body>{o.variantName} · {CONDITION[o.condition] ?? o.condition}</Body>
          <Body muted small>Sold by {o.sellerName}{o.buyBox ? " · best offer" : ""} · {o.warrantyMonths ? `${o.warrantyMonths}-month warranty` : "no warranty"} · delivery from {formatKobo(o.deliveryFee)}</Body>
        </Card>
      ))}
      <Notice>Buying in the app is coming soon. You can order this on the TechShop Market website.</Notice>
      {p.description ? <Body>{p.description}</Body> : null}
      {specs.length ? (
        <View style={{ gap: tokens.space[2] }}>
          <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>Specifications</Text>
          {specs.map((s) => (
            <Body key={s.label}>{s.label}: {s.value}</Body>
          ))}
        </View>
      ) : null}
    </Screen>
  );
}
