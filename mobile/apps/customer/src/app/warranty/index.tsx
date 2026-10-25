import { useRouter } from "expo-router";
import { ActivityIndicator } from "react-native";

import { claimStatus } from "@techshop/api-client";
import { Body, Button, Card, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { when } from "@/features/orders/status";

/** Items under warranty and the customer's claims. */
export default function Warranty() {
  const { session } = useSession();
  const router = useRouter();
  const r = useLoad(async () => {
    const [items, claims] = await Promise.all([session.api.get("/api/v1/me/warranty/items"), session.api.get("/api/v1/me/warranty-claims")]);
    return { items: items ?? [], claims: claims ?? [] };
  }, "warranty");
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading warranty" />;
  const { items = [], claims = [] } = r.data ?? {};
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Title>Warranty and repairs</Title>
      {r.error ? <Notice tone="danger">{r.error}</Notice> : null}
      <Body muted small>Whoever gives the warranty, claim here: we check the device at a store and handle the repair, replacement or refund.</Body>
      {items.length === 0 ? <Body muted>Nothing you’ve bought is under warranty right now.</Body> : null}
      {items.map((i) => (
        <Card key={i.orderLineId}>
          <Body>{i.productName}</Body>
          <Body small muted>{`${i.variantName} · until ${when(i.endsOn)}`}</Body>
          {i.openClaim ? (
            <Button label={`Open claim ${i.openClaim}`} onPress={() => router.push({ pathname: "/warranty/[number]", params: { number: i.openClaim! } })} />
          ) : (
            <Button
              label="Make a claim"
              onPress={() => router.push({ pathname: "/warranty/new", params: { line: i.orderLineId, name: i.productName, serialised: i.serialised ? "1" : "0" } })}
            />
          )}
        </Card>
      ))}
      {claims.map((c) => {
        const s = claimStatus[c.status] ?? { label: c.status, tone: "neutral" as const };
        return (
          <Card key={c.number} label={`Claim ${c.number}`} onPress={() => router.push({ pathname: "/warranty/[number]", params: { number: c.number } })}>
            <Body>{`${c.number} · ${c.productName ?? "Device"}`}</Body>
            <StatusLine label={s.label} tone={s.tone} />
            <Body small muted>{c.fault}</Body>
          </Card>
        );
      })}
    </Screen>
  );
}
