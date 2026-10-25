import { Stack, useLocalSearchParams } from "expo-router";
import { ActivityIndicator } from "react-native";

import { claimResolution, claimStatus } from "@techshop/api-client";
import { Body, Card, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";

const repairWords: Record<string, string> = {
  queued: "Waiting for a technician",
  diagnosing: "Being diagnosed",
  awaiting_parts: "Waiting for parts",
  repairing: "Being repaired",
  done: "Repaired",
  unrepairable: "Can't be repaired",
};

export default function Claim() {
  const { number } = useLocalSearchParams<{ number: string }>();
  const { session } = useSession();
  const r = useLoad(() => session.api.get("/api/v1/me/warranty-claims/{number}", { path: { number } }), number);
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading claim" />;
  if (!r.data) return <Screen underHeader><Notice tone="danger">{r.error ?? "Not found."}</Notice></Screen>;
  const c = r.data;
  const s = claimStatus[c.status] ?? { label: c.status, tone: "neutral" as const };
  const latest = (c.repairs ?? []).at(-1);
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Stack.Screen options={{ title: c.number }} />
      <Title>{`Claim ${c.number}`}</Title>
      <StatusLine label={s.label} tone={s.tone} />
      {c.status === "ready" ? <Notice tone="success">{`${c.resolution === "replaced" ? "Your replacement" : "Your device"} is ready to collect${c.store ? ` at ${c.store}` : ""}. Bring an ID.`}</Notice> : null}
      {c.status === "rejected" && c.rejectionReason ? <Notice tone="danger">{`Not covered: ${c.rejectionReason}`}</Notice> : null}
      <Card>
        <Body>{`${c.productName ?? ""} ${c.variantName ?? ""}`.trim()}</Body>
        <Body small muted>{`IMEI or serial ${c.imeiOrSerial}`}</Body>
        <Body small>{c.fault}</Body>
        {latest ? <Body small>{`Workshop: ${repairWords[latest.status] ?? latest.status}`}</Body> : null}
        {c.resolution ? <Body small>{`Outcome: ${claimResolution[c.resolution] ?? c.resolution}`}</Body> : null}
      </Card>
      {(c.dropOff ?? []).length ? <Body>Bring it to:</Body> : null}
      {(c.dropOff ?? []).map((st) => (
        <Card key={st.id}>
          <Body>{st.name}</Body>
          <Body small muted>{`${st.address}, ${st.city}`}</Body>
        </Card>
      ))}
    </Screen>
  );
}
