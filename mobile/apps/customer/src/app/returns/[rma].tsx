import { Stack, useLocalSearchParams } from "expo-router";
import { ActivityIndicator } from "react-native";

import { formatKobo, returnReason, returnStatus } from "@techshop/api-client";
import { Body, Card, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { when } from "@/features/orders/status";

const pickupWords: Record<string, string> = {
  unassigned: "Waiting for a rider",
  assigned: "A rider is on the way",
  picked_up: "Collected by the rider",
  en_route: "On its way to us",
  delivered: "Handed in at our hub",
  failed: "The rider couldn't collect it; we'll try again",
};

export default function ReturnDetail() {
  const { rma } = useLocalSearchParams<{ rma: string }>();
  const { session } = useSession();
  const r = useLoad(() => session.api.get("/api/v1/me/returns/{rma}", { path: { rma } }), rma);
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading return" />;
  if (!r.data) return <Screen underHeader><Notice tone="danger">{r.error ?? "Not found."}</Notice></Screen>;
  const x = r.data;
  const s = returnStatus[x.status] ?? { label: x.status, tone: "neutral" as const };
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Stack.Screen options={{ title: x.rma }} />
      <Title>{`Return ${x.rma}`}</Title>
      <StatusLine label={s.label} tone={s.tone} />
      {x.status === "approved" ? <Notice tone="success">{`Bring the item and its box to any TechShop store and quote ${x.rma}.`}</Notice> : null}
      {x.status === "awaiting_pickup" ? <Notice tone="success">{x.pickup ? "A rider will collect it from your delivery address. Keep it packed." : "We'll arrange a courier and tell you when it's booked."}</Notice> : null}
      {x.status === "rejected" && x.rejectionReason ? <Notice tone="danger">{`Not accepted: ${x.rejectionReason}`}</Notice> : null}
      <Card>
        {(x.items ?? []).map((i) => (
          <Body key={i.id}>{`${i.quantity} × ${i.productName} · ${formatKobo(i.refund)}`}</Body>
        ))}
      </Card>
      <Card>
        <Body small muted>{`Order ${x.orderNumber} · ${returnReason[x.reason] ?? x.reason}`}</Body>
        {x.details ? <Body small>{x.details}</Body> : null}
        {x.pickup ? <Body small>{`Collection: ${pickupWords[x.pickup.status] ?? x.pickup.status}`}</Body> : null}
        <Body>{`Refund: ${formatKobo(x.refund)}${x.status === "refunded" ? " (on its way)" : " once we've checked it"}`}</Body>
        <Body small muted>{`Requested ${when(x.createdAt)}`}</Body>
      </Card>
    </Screen>
  );
}
