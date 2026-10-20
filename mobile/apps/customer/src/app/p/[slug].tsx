import { Stack, useLocalSearchParams, useRouter } from "expo-router";
import { useState } from "react";
import { ActivityIndicator, Text, View } from "react-native";

import { formatKobo } from "@techshop/api-client";
import { Body, Button, Card, Notice, Screen, Title, tokens, useColors } from "@techshop/mobile-ui";

import { useCart } from "@/features/cart/CartProvider";

import { useLoad } from "@/features/catalogue/useLoad";
import { api } from "@/lib/api";

const CONDITION: Record<string, string> = { new: "New", uk_used: "UK-used", refurbished: "Refurbished", open_box: "Open box" };

/** Product detail: every seller offer (best first), warranty and specs. Buying in the app comes next. */
export default function Product() {
  const { slug } = useLocalSearchParams<{ slug: string }>();
  const c = useColors();
  const r = useLoad(() => api.get("/api/v1/products/{slug}", { path: { slug } }), slug);
  const { add } = useCart();
  const router = useRouter();
  const [status, setStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [adding, setAdding] = useState<string | null>(null);
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading product" />;
  if (!r.data) return <Screen underHeader><Notice tone="danger">{r.error ?? "Not found."}</Notice></Screen>;
  const p = r.data;
  const offers = p.offers ?? [];
  const specs = (p.attributes ?? []).map((a) => ({ label: a.label, value: a.valueText ?? (a.valueNumber !== null ? `${a.valueNumber}${a.unit ? ` ${a.unit}` : ""}` : a.valueBool !== null ? (a.valueBool ? "Yes" : "No") : null) })).filter((s) => s.value);
  return (
    <Screen underHeader>
      <Stack.Screen options={{ title: p.brand ?? "Product" }} />
      <Title>{p.name}</Title>
      {!p.inStock ? <Notice tone="warning">Out of stock right now.</Notice> : null}
      {offers.map((o) => (
        <Card key={o.listingId}>
          <Text style={{ color: c.text, fontSize: tokens.fontSize["2xl"], fontWeight: "600", fontVariant: ["tabular-nums"] }}>{formatKobo(o.price)}</Text>
          <Body>{o.variantName} · {CONDITION[o.condition] ?? o.condition}</Body>
          <Body muted small>Sold by {o.sellerName}{o.buyBox ? " · best offer" : ""} · {o.warrantyMonths ? `${o.warrantyMonths}-month warranty` : "no warranty"} · delivery from {formatKobo(o.deliveryFee)}</Body>
          <Button
            label="Add to cart"
            variant={o.buyBox ? "primary" : "secondary"}
            disabled={!p.inStock}
            busy={adding === o.listingId}
            onPress={async () => {
              setAdding(o.listingId);
              const e = await add(o.listingId, 1);
              setAdding(null);
              setStatus(e ? { ok: false, text: e } : { ok: true, text: `Added ${p.name} to your cart.` });
            }}
          />
        </Card>
      ))}
      {status ? <Notice tone={status.ok ? "success" : "danger"}>{status.text}</Notice> : null}
      {status?.ok ? <Button label="View cart" onPress={() => router.push("/cart")} /> : null}
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
