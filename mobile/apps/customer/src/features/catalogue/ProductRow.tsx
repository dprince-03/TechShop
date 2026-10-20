import { useRouter } from "expo-router";
import { Text, View } from "react-native";

import { formatKobo, type Schema } from "@techshop/api-client";
import { Card, tokens, useColors } from "@techshop/mobile-ui";

type Summary = Schema<"ProductSummary">;

/** One search result: name, brand, price (with the old price and discount read out in full). */
export function ProductRow({ p }: { p: Summary }) {
  const router = useRouter();
  const c = useColors();
  const price = p.promoPrice ?? p.price ?? 0;
  const was = p.promoPrice !== undefined && p.price && p.promoPrice < p.price ? p.price : p.compareAt && p.compareAt > price ? p.compareAt : undefined;
  const off = was ? Math.round(((was - price) / was) * 100) : 0;
  const spoken = `${p.name}, ${formatKobo(price)}${was ? `, was ${formatKobo(was)}, ${off}% off` : ""}${p.inStock ? "" : ", out of stock"}`;
  return (
    <Card label={spoken} onPress={() => router.push({ pathname: "/p/[slug]", params: { slug: p.slug } })}>
      <Text style={{ color: c.text, fontSize: tokens.fontSize.lg, fontWeight: "600" }} numberOfLines={2}>
        {p.name}
      </Text>
      {p.brandName ? <Text style={{ color: c.textMuted, fontSize: tokens.fontSize.sm }}>{p.brandName}</Text> : null}
      <View style={{ flexDirection: "row", gap: tokens.space[2], alignItems: "baseline", flexWrap: "wrap" }}>
        <Text style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600", fontVariant: ["tabular-nums"] }}>{formatKobo(price)}</Text>
        {was ? <Text style={{ color: c.textMuted, textDecorationLine: "line-through", fontVariant: ["tabular-nums"] }}>{formatKobo(was)}</Text> : null}
        {off ? <Text style={{ color: c.sale, fontWeight: "600" }}>{off}% off</Text> : null}
      </View>
      <Text style={{ color: p.inStock ? c.textMuted : c.danger, fontSize: tokens.fontSize.sm }}>
        {!p.inStock ? "Out of stock" : p.offerCount > 1 ? `From ${p.offerCount} sellers` : (p.sellerTypes ?? []).includes("first_party") ? "Sold by TechShop" : "Sold by a verified seller"}
      </Text>
    </Card>
  );
}
