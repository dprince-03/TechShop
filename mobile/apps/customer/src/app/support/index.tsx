import { useRouter } from "expo-router";
import { ActivityIndicator } from "react-native";

import { ticketStatus, ticketTopic } from "@techshop/api-client";
import { Body, Button, Card, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { when } from "@/features/orders/status";

/** Conversations with support; replies arrive here with a notification. */
export default function Support() {
  const { session } = useSession();
  const router = useRouter();
  const r = useLoad(() => session.api.get("/api/v1/me/tickets"), "tickets");
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading support" />;
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Title>Support</Title>
      {r.error ? <Notice tone="danger">{r.error}</Notice> : null}
      <Button label="Ask a question" variant="primary" onPress={() => router.push("/support/new")} />
      {(r.data ?? []).map((t) => {
        const s = ticketStatus[t.status] ?? { label: t.status, tone: "neutral" as const };
        return (
          <Card key={t.number} label={t.subject} onPress={() => router.push({ pathname: "/support/[number]", params: { number: t.number } })}>
            <Body>{`${t.unread ? "● " : ""}${t.subject}`}</Body>
            <StatusLine label={s.label} tone={s.tone} />
            <Body small muted>{`${t.number} · ${ticketTopic[t.topic] ?? t.topic}${t.unread ? " · new reply" : ""} · ${when(t.lastMessageAt)}`}</Body>
          </Card>
        );
      })}
    </Screen>
  );
}
