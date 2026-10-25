import { useRouter } from "expo-router";
import { ActivityIndicator } from "react-native";

import { formatKobo, tradeInStatus } from "@techshop/api-client";
import { Body, Button, Card, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";

/** The customer's trade-ins, and a way to start one. */
export default function TradeIns() {
  const { session } = useSession();
  const router = useRouter();
  const r = useLoad(() => session.api.get("/api/v1/me/trade-ins"), "trade-ins");
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading trade-ins" />;
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Title>Trade-ins</Title>
      {r.error ? <Notice tone="danger">{r.error}</Notice> : null}
      <Button label="Trade in a device" variant="primary" onPress={() => router.push("/trade-in/new")} />
      <Body muted small>Get an estimate, bring it to a store for inspection, then take store credit at once or a bank transfer.</Body>
      {(r.data ?? []).map((t) => {
        const s = tradeInStatus[t.status] ?? { label: t.status, tone: "neutral" as const };
        const amount = t.offer ?? t.estimate;
        return (
          <Card key={t.reference} label={`Trade-in ${t.reference}`} onPress={() => router.push({ pathname: "/trade-in/[ref]", params: { ref: t.reference } })}>
            <Body>{t.device}</Body>
            <StatusLine label={s.label} tone={s.tone} />
            <Body small muted>{`${t.reference}${amount != null ? ` · ${t.offer != null ? "Offer" : "Estimate"} ${formatKobo(amount)}` : ""}`}</Body>
          </Card>
        );
      })}
    </Screen>
  );
}
