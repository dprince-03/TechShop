import { Stack, useLocalSearchParams } from "expo-router";
import { ActivityIndicator } from "react-native";

import { formatKobo } from "@techshop/api-client";
import { Body, Card, Notice, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { orderStatus, parcelStatus, when } from "@/features/orders/status";

/** One order: status, parcels by seller, totals and history. */
export default function Order() {
  const { number } = useLocalSearchParams<{ number: string }>();
  const { session } = useSession();
  const r = useLoad(() => session.api.get("/api/v1/me/orders/{number}", { path: { number } }), number);
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading order" />;
  if (!r.data) return <Screen underHeader><Notice tone="danger">{r.error ?? "Not found."}</Notice></Screen>;
  const o = r.data;
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Stack.Screen options={{ title: o.orderNumber }} />
      <Title>{orderStatus[o.status] ?? o.status}</Title>
      <Body muted>{`Placed ${when(o.placedAt)}`}</Body>
      {o.status === "pending_payment" ? <Notice tone="warning">We’re waiting for your payment. Pull down to refresh once you’ve paid.</Notice> : null}
      {(o.fulfilments ?? []).map((f) => (
        <Card key={f.id}>
          <Body>{`From ${f.sellerName} · ${parcelStatus[f.status] ?? f.status}`}</Body>
          {(f.lines ?? []).map((l) => (
            <Body key={l.id} small muted>{`${l.quantity} × ${l.productName} · ${formatKobo(l.lineTotalKobo)}`}</Body>
          ))}
        </Card>
      ))}
      <Card>
        <Body>{`Delivery: ${formatKobo(o.deliveryFeeKobo)}`}</Body>
        {o.discountKobo ? <Body>{`Discounts: −${formatKobo(o.discountKobo)}`}</Body> : null}
        <Body>{`Total: ${formatKobo(o.totalKobo)}`}</Body>
      </Card>
      {(o.history ?? []).map((h) => (
        <Body key={h.id} small muted>{`${orderStatus[h.toStatus] ?? parcelStatus[h.toStatus] ?? h.toStatus} · ${when(h.createdAt)}`}</Body>
      ))}
    </Screen>
  );
}
