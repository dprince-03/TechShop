import { useState } from "react";
import { Text, View } from "react-native";

import { Body, Button, Card, Notice, tokens, useColors } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { api } from "@/lib/api";

import { problem } from "./problem";

const REASONS = [
  { value: "fake", label: "Looks fake" },
  { value: "abuse", label: "Abusive" },
  { value: "personal_data", label: "Personal details" },
  { value: "off_topic", label: "Not about the product" },
] as const;

const date = new Intl.DateTimeFormat("en-NG", { dateStyle: "medium", timeZone: "Africa/Lagos" });

/** Rating summary and the latest reviews (verified purchases first); signed-in shoppers can report one. */
export function ProductReviews({ slug }: { slug: string }) {
  const c = useColors();
  const { status, session } = useSession();
  const r = useLoad(() => api.get("/api/v1/products/{slug}/reviews", { path: { slug }, query: { limit: 10 } }), slug);
  const [reporting, setReporting] = useState<string | null>(null);
  const [notes, setNotes] = useState<Record<string, { ok: boolean; text: string }>>({});
  if (!r.data) return null;
  const { summary } = r.data;
  const reviews = r.data.reviews.items ?? [];

  const report = async (id: string, reason: (typeof REASONS)[number]["value"]) => {
    try {
      await session.api.post("/api/v1/reviews/{id}/report", { path: { id }, body: { reason } });
      setNotes((n) => ({ ...n, [id]: { ok: true, text: "Thanks. A moderator will look at it." } }));
    } catch (e) {
      setNotes((n) => ({ ...n, [id]: { ok: false, text: problem(e) } }));
    }
    setReporting(null);
  };

  return (
    <View style={{ gap: tokens.space[3] }}>
      <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>
        Reviews
      </Text>
      {summary.total > 0 ? (
        <Body muted small>{`${summary.average.toFixed(1)} out of 5 from ${summary.total} review${summary.total === 1 ? "" : "s"} (${summary.verified} verified purchase${summary.verified === 1 ? "" : "s"})`}</Body>
      ) : (
        <Body muted small>No reviews yet. Bought it? Review it from your order.</Body>
      )}
      {reviews.map((rv) => (
        <Card key={rv.id}>
          <Body>{`${rv.rating} out of 5${rv.verified ? " · verified purchase" : ""}`}</Body>
          {rv.title ? <Body>{rv.title}</Body> : null}
          {rv.body ? <Body muted>{rv.body}</Body> : null}
          <Body small muted>{`${rv.reviewer} · ${date.format(new Date(rv.createdAt))}`}</Body>
          {rv.sellerResponse ? <Body small>{`Seller’s answer: ${rv.sellerResponse}`}</Body> : null}
          {notes[rv.id] ? <Notice tone={notes[rv.id].ok ? "success" : "danger"}>{notes[rv.id].text}</Notice> : null}
          {status === "signed-in" && !notes[rv.id]?.ok ? (
            reporting === rv.id ? (
              <View style={{ gap: tokens.space[2] }}>
                {REASONS.map((x) => (
                  <Button key={x.value} label={x.label} onPress={() => report(rv.id, x.value)} />
                ))}
                <Button label="Cancel" onPress={() => setReporting(null)} />
              </View>
            ) : (
              <Button label="Report this review" onPress={() => setReporting(rv.id)} />
            )
          ) : null}
        </Card>
      ))}
    </View>
  );
}
