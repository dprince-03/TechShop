import { Stack, useLocalSearchParams, useRouter } from "expo-router";
import { useState } from "react";

import { Body, Button, Choice, Field, Notice, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { problem } from "@/features/reviews/problem";

type Topic = "order" | "delivery" | "payment" | "refund" | "return" | "warranty" | "trade_in" | "account" | "other";

export default function NewTicket() {
  const { order } = useLocalSearchParams<{ order?: string }>();
  const { session } = useSession();
  const router = useRouter();
  const [topic, setTopic] = useState<Topic | null>(order ? "order" : null);
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const submit = async () => {
    if (!topic) return setError("Choose what it's about.");
    if (subject.trim().length < 3) return setError("Give it a short subject.");
    if (!message.trim()) return setError("Tell us what's happening.");
    setBusy(true);
    setError(null);
    try {
      const t = await session.api.post("/api/v1/me/tickets", { body: { topic, subject: subject.trim(), message: message.trim(), orderNumber: order || undefined } });
      router.replace({ pathname: "/support/[number]", params: { number: t.number } });
    } catch (e) {
      setError(problem(e));
    } finally {
      setBusy(false);
    }
  };
  return (
    <Screen underHeader>
      <Stack.Screen options={{ title: "Ask support" }} />
      <Title>Ask support</Title>
      {order ? <Body muted small>{`About order ${order}`}</Body> : null}
      <Choice<Topic>
        label="What's it about?"
        value={topic}
        onChange={setTopic}
        options={[
          { value: "order", label: "An order" },
          { value: "delivery", label: "Delivery" },
          { value: "payment", label: "A payment" },
          { value: "refund", label: "A refund" },
          { value: "return", label: "A return" },
          { value: "warranty", label: "Warranty or repair" },
          { value: "trade_in", label: "A trade-in" },
          { value: "account", label: "My account" },
          { value: "other", label: "Something else" },
        ]}
      />
      <Field label="Subject" value={subject} onChangeText={setSubject} maxLength={140} />
      <Field label="How can we help?" value={message} onChangeText={setMessage} multiline maxLength={5000} hint="Payment and refund questions: answered within 4 hours. Others within 24." />
      {error ? <Notice tone="danger">{error}</Notice> : null}
      <Button label="Send" variant="primary" busy={busy} onPress={submit} />
    </Screen>
  );
}
