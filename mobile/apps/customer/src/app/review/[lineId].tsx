import { Stack, useLocalSearchParams, useRouter } from "expo-router";
import { useState } from "react";

import { Body, Button, Field, Notice, Screen, StarPicker, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { problem } from "@/features/reviews/problem";

/** Review an item from a delivered order (a verified purchase; the API finds the product). */
export default function ReviewItem() {
  const { lineId, name } = useLocalSearchParams<{ lineId: string; name?: string }>();
  const { session } = useSession();
  const router = useRouter();
  const [rating, setRating] = useState(0);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<{ ok: boolean; text: string } | null>(null);

  const submit = async () => {
    if (!rating) return setResult({ ok: false, text: "Choose 1 to 5 stars." });
    setBusy(true);
    try {
      const r = await session.api.post("/api/v1/me/order-lines/{lineId}/review", {
        path: { lineId },
        body: { rating, title: title.trim() || undefined, body: body.trim() || undefined },
      });
      setResult({ ok: true, text: r.status === "published" ? "Thanks! Your review is live." : "Thanks! Your review will appear after a quick check." });
    } catch (e) {
      setResult({ ok: false, text: problem(e) });
    }
    setBusy(false);
  };

  return (
    <Screen underHeader>
      <Stack.Screen options={{ title: "Write a review" }} />
      <Title>{name ?? "Your review"}</Title>
      {result?.ok ? (
        <>
          <Notice tone="success">{result.text}</Notice>
          <Button label="Back to your order" onPress={() => router.back()} />
        </>
      ) : (
        <>
          <StarPicker label="Your rating" value={rating} onChange={setRating} />
          <Field label="Headline (optional)" value={title} onChangeText={setTitle} maxLength={120} />
          <Field label="Your review (optional)" value={body} onChangeText={setBody} maxLength={3000} multiline hint="Please don’t include phone numbers, emails or links." />
          {result ? <Notice tone="danger">{result.text}</Notice> : null}
          <Button label="Post review" variant="primary" busy={busy} onPress={submit} />
          <Body muted small>Your first name and last initial are shown with the review.</Body>
        </>
      )}
    </Screen>
  );
}
