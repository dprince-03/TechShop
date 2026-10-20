import { useRouter } from "expo-router";
import { useState } from "react";
import { Text, View } from "react-native";

import { formatKobo } from "@techshop/api-client";
import { Body, Button, Card, Notice, Screen, Title, tokens, useColors } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useCart } from "@/features/cart/CartProvider";

/** The cart, grouped as the API prices it: lines, delivery estimate, discounts and total. */
export default function CartScreen() {
  const { cart, setQuantity, remove } = useCart();
  const { status } = useSession();
  const router = useRouter();
  const c = useColors();
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const lines = cart?.lines ?? [];

  const run = async (id: string, fn: () => Promise<string | null>) => {
    setBusy(id);
    setError(await fn());
    setBusy(null);
  };

  if (!lines.length) {
    return (
      <Screen>
        <Title>Your cart is empty</Title>
        <Body muted>Items you add appear here.</Body>
        <Button label="Continue shopping" variant="primary" onPress={() => router.replace("/")} />
      </Screen>
    );
  }

  return (
    <Screen>
      <Title>Your cart</Title>
      {error ? <Notice tone="danger">{error}</Notice> : null}
      {lines.map((l) => (
        <Card key={l.cartItemId}>
          <Text style={{ color: c.text, fontSize: tokens.fontSize.lg, fontWeight: "600" }}>{l.productName}</Text>
          <Body muted small>{l.variantName} · sold by {l.sellerName}</Body>
          <Body>{`${formatKobo(l.unitPriceKobo)} each · ${formatKobo(l.lineTotalKobo)}`}</Body>
          {l.problem ? <Notice tone="warning">{l.problem === "sold_out" ? "Sold out — remove it to check out." : "Not enough stock for this quantity."}</Notice> : null}
          <View style={{ flexDirection: "row", gap: tokens.space[2], alignItems: "center" }}>
            <View style={{ flex: 1 }}>
              <Button label="−" accessibilityHint={`Fewer ${l.productName}`} disabled={l.quantity <= 1} busy={busy === `${l.cartItemId}-`} onPress={() => run(`${l.cartItemId}-`, () => setQuantity(l.cartItemId, l.quantity - 1))} />
            </View>
            <Text accessibilityLabel={`Quantity ${l.quantity}`} style={{ color: c.text, fontSize: tokens.fontSize.xl, minWidth: 32, textAlign: "center" }}>{l.quantity}</Text>
            <View style={{ flex: 1 }}>
              <Button label="+" accessibilityHint={`More ${l.productName}`} busy={busy === `${l.cartItemId}+`} onPress={() => run(`${l.cartItemId}+`, () => setQuantity(l.cartItemId, l.quantity + 1))} />
            </View>
          </View>
          <Button label="Remove" busy={busy === `${l.cartItemId}x`} onPress={() => run(`${l.cartItemId}x`, () => remove(l.cartItemId))} />
        </Card>
      ))}
      <Card>
        <Body>{`Items: ${formatKobo(cart!.itemsKobo)}`}</Body>
        {cart!.discountKobo > 0 ? <Body>{`${cart!.promotion ?? "Discount"}: −${formatKobo(cart!.discountKobo)}`}</Body> : null}
        <Body>{`Delivery (estimate): ${cart!.freeDelivery ? "Free" : formatKobo(cart!.deliveryKobo)}`}</Body>
        <Text style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "700", fontVariant: ["tabular-nums"] }}>{`Total ${formatKobo(cart!.totalKobo)}`}</Text>
      </Card>
      {(cart!.problems ?? []).map((p) => (
        <Notice key={p} tone="warning">{p}</Notice>
      ))}
      <Notice>Checkout in the app is coming soon. {status === "signed-in" ? "Your cart is saved to your account, so you can check out on the TechShop Market website." : "Sign in to keep this cart on your account and check out on the website."}</Notice>
      {status !== "signed-in" ? <Button label="Sign in" onPress={() => router.push("/sign-in")} /> : null}
    </Screen>
  );
}
