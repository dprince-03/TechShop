import { useState } from "react";

import { ApiError, normaliseNigerianPhone, type Schema } from "@techshop/api-client";
import { Body, Button, Card, Field, Notice, Screen, Title } from "@techshop/mobile-ui";

import { orderStatus, parcelStatus, when } from "@/features/orders/status";
import { api } from "@/lib/api";

/** Guest tracking: order number + the phone used at checkout. */
export default function Track() {
  const [number, setNumber] = useState("");
  const [phone, setPhone] = useState("");
  const [result, setResult] = useState<Schema<"Tracking"> | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const look = async () => {
    const n = number.trim().toUpperCase();
    const p = normaliseNigerianPhone(phone);
    if (!/^TS-\d{1,12}$/.test(n)) return setError("Enter your order number, e.g. TS-10482.");
    if (!p) return setError("Enter the mobile number you used at checkout.");
    setBusy(true);
    try {
      setResult(await api.get("/api/v1/track/{number}", { path: { number: n }, query: { phone: p } }));
      setError(null);
    } catch (e) {
      setResult(null);
      setError(e instanceof ApiError && e.code === "rate_limited" ? "Too many lookups. Wait a few minutes." : e instanceof ApiError && e.isUnavailable ? "No connection. Try again." : "We couldn’t find an order with that number and phone.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <Screen>
      <Title>Track an order</Title>
      <Field label="Order number" value={number} onChangeText={setNumber} autoCapitalize="characters" placeholder="TS-10482" />
      <Field label="Phone used at checkout" value={phone} onChangeText={setPhone} keyboardType="phone-pad" placeholder="0803 123 4567" error={error} />
      <Button label="Track order" variant="primary" busy={busy} onPress={look} />
      {result ? (
        <Card>
          <Body>{`${result.orderNumber}: ${orderStatus[result.status] ?? result.status}`}</Body>
          {(result.parcels ?? []).map((p, i) => (
            <Body key={i} small muted>{`${p.seller} · ${p.items} item${p.items === 1 ? "" : "s"} · ${parcelStatus[p.status] ?? p.status}`}</Body>
          ))}
          {(result.history ?? []).map((h, i) => (
            <Body key={`h${i}`} small muted>{`${orderStatus[h.status] ?? parcelStatus[h.status] ?? h.status} · ${when(h.at)}`}</Body>
          ))}
        </Card>
      ) : null}
      {!result && !error ? <Notice>No account needed.</Notice> : null}
    </Screen>
  );
}
