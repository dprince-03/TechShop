import * as Crypto from "expo-crypto";
import { Redirect, useRouter } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import { useEffect, useRef, useState } from "react";
import { ActivityIndicator, Pressable, Text } from "react-native";

import { ApiError, formatKobo, normaliseNigerianPhone, type Schema } from "@techshop/api-client";
import { Body, Button, Card, Field, Notice, Screen, Title, tokens, useColors } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useCart } from "@/features/cart/CartProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { api } from "@/lib/api";

type Quote = Schema<"Quote">;
type Provider = "paystack" | "opay" | "moniepoint";
const PROVIDER: Record<string, string> = { paystack: "Card (Paystack)", opay: "OPay", moniepoint: "Bank transfer (Moniepoint)" };
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

/**
 * Checkout: address → payment method → the API's quote → pay. The order is placed with the
 * quote's price hash and one idempotency key per attempt, then the provider's page opens in an
 * in-app browser (no card details ever touch the app) and the payment status is checked.
 */
export default function Checkout() {
  const { status, session, me } = useSession();
  const { reload: reloadCart } = useCart();
  const router = useRouter();
  const c = useColors();
  const data = useLoad(async () => {
    const [addresses, methods, states] = await Promise.all([
      session.api.get("/api/v1/me/addresses"),
      session.api.get("/api/v1/checkout/payment-methods"),
      api.get("/api/v1/states"),
    ]);
    return { addresses: addresses ?? [], methods: (methods ?? []) as Provider[], states: states ?? [] };
  }, status);
  const [addressId, setAddressId] = useState<string | null>(null);
  const [provider, setProvider] = useState<Provider | null>(null);
  const [quote, setQuote] = useState<Quote | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [transfer, setTransfer] = useState<Schema<"TransferAccount"> | null>(null);
  const key = useRef(Crypto.randomUUID());
  // New address form
  const [adding, setAdding] = useState(false);
  const [form, setForm] = useState({ recipientName: "", phone: "", line1: "", city: "", state: "" });

  const loaded = data.data;
  const addresses = loaded?.addresses ?? [];
  // Pre-select the default address and the first payment method once they load.
  useEffect(() => {
    const t = setTimeout(() => {
      const list = loaded?.addresses ?? [];
      if (!addressId && list.length) setAddressId((list.find((a) => a.isDefault) ?? list[0]).id);
      if (!provider && loaded?.methods.length) setProvider(loaded.methods[0]);
    }, 0);
    return () => clearTimeout(t);
  }, [loaded, addressId, provider]);

  useEffect(() => {
    if (!addressId) return;
    let live = true;
    void session.api
      .post("/api/v1/checkout/preview", { body: { addressId } })
      .then((q) => live && (setQuote(q), setError(null)))
      .catch(() => live && setError("We couldn’t price your cart for that address."));
    return () => {
      live = false;
    };
  }, [addressId, session]);

  if (status === "signed-out") return <Redirect href="/sign-in" />;
  if (data.loading && !data.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading checkout" />;

  const saveAddress = async () => {
    const phone = normaliseNigerianPhone(form.phone);
    const state = data.data?.states.find((s) => s.name.toLowerCase() === form.state.trim().toLowerCase() || s.code === form.state.trim().toUpperCase());
    if (!form.recipientName.trim() || !form.line1.trim() || !form.city.trim()) return setError("Fill in the name, street and city.");
    if (!phone) return setError("Enter a Nigerian mobile number for the rider.");
    if (!state) return setError("Type a state name, e.g. Lagos.");
    setBusy(true);
    try {
      const a = await session.api.post("/api/v1/me/addresses", { body: { recipientName: form.recipientName.trim(), phone, line1: form.line1.trim(), city: form.city.trim(), stateCode: state.code, isDefault: !addresses.length } });
      await data.reload();
      setAddressId(a.id);
      setAdding(false);
      setError(null);
    } catch {
      setError("We couldn’t save the address.");
    } finally {
      setBusy(false);
    }
  };

  const pay = async () => {
    if (!quote || !addressId || !provider) return;
    setBusy(true);
    setError(null);
    try {
      const placed = await session.api.post("/api/v1/orders", { idempotencyKey: key.current, body: { addressId, priceHash: quote.priceHash, provider } });
      await reloadCart();
      if (placed.payment.transfer) {
        setTransfer(placed.payment.transfer);
        return;
      }
      if (placed.payment.checkoutUrl) await WebBrowser.openBrowserAsync(placed.payment.checkoutUrl);
      // Back from the provider: the webhook marks the order paid; check for a few seconds.
      for (let i = 0; i < 10; i++) {
        const s = await session.api.get("/api/v1/me/payments/{id}", { path: { id: placed.payment.paymentId } }).catch(() => null);
        if (s?.orderStatus === "paid") break;
        await sleep(1500);
      }
      router.replace({ pathname: "/orders/[number]", params: { number: placed.orderNumber } });
    } catch (e) {
      if (e instanceof ApiError && e.code === "price_changed") {
        key.current = Crypto.randomUUID();
        setQuote(await session.api.post("/api/v1/checkout/preview", { body: { addressId } }).catch(() => null));
        setError("Prices or stock changed. Check the new total, then pay.");
      } else if (e instanceof ApiError && e.isUnavailable) {
        setError("No connection. You haven’t been charged; try again.");
      } else {
        setError(e instanceof ApiError && e.detail ? e.detail : "We couldn’t place your order. You haven’t been charged.");
      }
    } finally {
      setBusy(false);
    }
  };

  if (transfer) {
    return (
      <Screen underHeader>
        <Title>Pay by bank transfer</Title>
        <Card>
          <Body>{`Transfer exactly ${formatKobo(transfer.amountKobo)} to:`}</Body>
          <Text style={{ color: c.text, fontSize: tokens.fontSize["2xl"], fontWeight: "700" }} selectable>{transfer.accountNumber}</Text>
          <Body>{transfer.bankName}</Body>
          <Body muted small>{`This account is for this order only. Pay before ${new Date(transfer.expiresAt).toLocaleString("en-NG", { timeZone: "Africa/Lagos" })}.`}</Body>
        </Card>
        <Button label="Your orders" variant="primary" onPress={() => router.replace("/orders")} />
      </Screen>
    );
  }

  const choice = (selected: boolean) => ({ minHeight: tokens.control.lg, borderWidth: 1, borderRadius: tokens.radius.md, padding: tokens.space[3], justifyContent: "center" as const, borderColor: selected ? c.accent : c.borderStrong, backgroundColor: selected ? c.accentSubtle : c.bg });

  return (
    <Screen underHeader>
      <Title>Checkout</Title>
      {error ? <Notice tone="danger">{error}</Notice> : null}
      <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>1. Delivery address</Text>
      {addresses.map((a) => (
        <Pressable key={a.id} accessibilityRole="radio" accessibilityState={{ selected: a.id === addressId }} onPress={() => setAddressId(a.id)} style={choice(a.id === addressId)}>
          <Text style={{ color: c.text, fontWeight: "600" }}>{a.recipientName}</Text>
          <Text style={{ color: c.textMuted }}>{[a.line1, a.city, data.data?.states.find((s) => s.code === a.stateCode)?.name].filter(Boolean).join(", ")}</Text>
        </Pressable>
      ))}
      {adding || !addresses.length ? (
        <Card>
          <Field label="Recipient’s name" value={form.recipientName} onChangeText={(v) => setForm({ ...form, recipientName: v })} placeholder={me ? `${me.firstName} ${me.lastName}` : undefined} />
          <Field label="Phone for the rider" value={form.phone} onChangeText={(v) => setForm({ ...form, phone: v })} keyboardType="phone-pad" />
          <Field label="Street address" value={form.line1} onChangeText={(v) => setForm({ ...form, line1: v })} />
          <Field label="City or town" value={form.city} onChangeText={(v) => setForm({ ...form, city: v })} />
          <Field label="State" value={form.state} onChangeText={(v) => setForm({ ...form, state: v })} placeholder="e.g. Lagos" hint={form.state.length >= 2 ? (data.data?.states.filter((s) => s.name.toLowerCase().startsWith(form.state.toLowerCase())).map((s) => s.name).slice(0, 4).join(" · ") ?? "") : undefined} />
          <Button label="Save address" busy={busy} onPress={saveAddress} />
        </Card>
      ) : (
        <Button label="Add a new address" onPress={() => setAdding(true)} />
      )}
      <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>2. Payment method</Text>
      {(data.data?.methods ?? []).map((m) => (
        <Pressable key={m} accessibilityRole="radio" accessibilityState={{ selected: m === provider }} onPress={() => setProvider(m)} style={choice(m === provider)}>
          <Text style={{ color: c.text }}>{PROVIDER[m] ?? m}</Text>
        </Pressable>
      ))}
      <Body muted small>You pay on the provider’s secure page. TechShop never sees your card details.</Body>
      {quote ? (
        <Card>
          <Body>{`Items: ${formatKobo(quote.itemsKobo)}`}</Body>
          {quote.discountKobo ? <Body>{`Discounts: −${formatKobo(quote.discountKobo)}`}</Body> : null}
          <Body>{`Delivery: ${quote.freeDelivery ? "Free" : formatKobo(quote.deliveryKobo)}`}</Body>
          <Body muted small>{`Includes VAT of ${formatKobo(quote.vatKobo)}`}</Body>
          <Text style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "700", fontVariant: ["tabular-nums"] }}>{`Total ${formatKobo(quote.totalKobo)}`}</Text>
          {(quote.problems ?? []).map((p) => <Notice key={p} tone="warning">{p}</Notice>)}
        </Card>
      ) : null}
      <Button label={quote ? `Pay ${formatKobo(quote.totalKobo)}` : "Pay"} variant="primary" busy={busy} disabled={!quote?.purchasable || !provider} onPress={pay} />
    </Screen>
  );
}
