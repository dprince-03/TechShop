import { Stack, useLocalSearchParams } from "expo-router";
import { useState } from "react";
import { ActivityIndicator } from "react-native";

import { ticketStatus } from "@techshop/api-client";
import { Body, Button, Card, Field, Notice, Screen, StatusLine, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { when } from "@/features/orders/status";
import { problem } from "@/features/reviews/problem";

export default function Ticket() {
  const { number } = useLocalSearchParams<{ number: string }>();
  const { session } = useSession();
  const r = useLoad(() => session.api.get("/api/v1/me/tickets/{number}", { path: { number } }), number);
  const [reply, setReply] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  if (r.loading && !r.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading ticket" />;
  if (!r.data) return <Screen underHeader><Notice tone="danger">{r.error ?? "Not found."}</Notice></Screen>;
  const t = r.data;
  const s = ticketStatus[t.status] ?? { label: t.status, tone: "neutral" as const };
  const send = async () => {
    if (!reply.trim()) return setError("Write a message.");
    setBusy(true);
    setError(null);
    try {
      await session.api.post("/api/v1/me/tickets/{number}/messages", { path: { number }, body: { message: reply.trim() } });
      setReply("");
      await r.reload();
    } catch (e) {
      setError(problem(e));
    } finally {
      setBusy(false);
    }
  };
  return (
    <Screen underHeader onRefresh={r.reload} refreshing={r.loading}>
      <Stack.Screen options={{ title: t.number }} />
      <Title>{t.subject}</Title>
      <StatusLine label={s.label} tone={s.tone} />
      {(t.messages ?? []).map((m) => (
        <Card key={m.id}>
          <Body small muted>{`${m.authorKind === "customer" ? "You" : m.author} · ${when(m.createdAt)}`}</Body>
          <Body selectable>{m.body}</Body>
        </Card>
      ))}
      {t.status === "closed" ? (
        <Body muted>{`This ticket is closed. Ask a new question and mention ${t.number}.`}</Body>
      ) : (
        <>
          <Field label="Your reply" value={reply} onChangeText={setReply} multiline maxLength={5000} />
          {error ? <Notice tone="danger">{error}</Notice> : null}
          <Button label="Send reply" variant="primary" busy={busy} onPress={send} />
        </>
      )}
    </Screen>
  );
}
