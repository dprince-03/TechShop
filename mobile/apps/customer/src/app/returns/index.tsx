import { useRouter } from "expo-router";
import { ActivityIndicator } from "react-native";

import { returnReason, returnStatus } from "@techshop/api-client";
import { Body, Card, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { when } from "@/features/orders/status";

/** The customer's returns; new ones start from a delivered order. */
export default function Returns() {
  const { session } = useSession();
  const router = useRouter();
  const r = useLoad(() => session.api.get("/api/v1/me/returns"), "returns");
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading returns" />;
  const rows = r.data ?? [];
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Title>Returns</Title>
      {r.error ? <Notice tone="danger">{r.error}</Notice> : null}
      {rows.length === 0 ? <Body muted>To return something, open the order and choose “Return items” (within 7 days of delivery).</Body> : null}
      {rows.map((x) => {
        const s = returnStatus[x.status] ?? { label: x.status, tone: "neutral" as const };
        return (
          <Card key={x.rma} label={`Return ${x.rma}`} onPress={() => router.push({ pathname: "/returns/[rma]", params: { rma: x.rma } })}>
            <Body>{x.rma}</Body>
            <StatusLine label={s.label} tone={s.tone} />
            <Body small muted>{`Order ${x.orderNumber} · ${returnReason[x.reason] ?? x.reason} · ${when(x.createdAt)}`}</Body>
          </Card>
        );
      })}
    </Screen>
  );
}
