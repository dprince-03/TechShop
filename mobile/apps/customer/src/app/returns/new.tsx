import { Stack, useLocalSearchParams, useRouter } from "expo-router";
import { useState } from "react";
import { ActivityIndicator } from "react-native";

import { Body, Button, Choice, Field, Notice, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { problem } from "@/features/reviews/problem";

const WINDOW_MS = 7 * 24 * 60 * 60 * 1000;
type Reason = "faulty" | "damaged" | "not_as_described" | "wrong_item" | "change_of_mind";

/** Return items from a delivered order: how many, why, and how they come back. */
export default function NewReturn() {
  const { number } = useLocalSearchParams<{ number: string }>();
  const { session } = useSession();
  const router = useRouter();
  const order = useLoad(() => session.api.get("/api/v1/me/orders/{number}", { path: { number } }), number);
  const [qty, setQty] = useState<Record<string, string>>({});
  const [reason, setReason] = useState<Reason | null>(null);
  const [method, setMethod] = useState<"store_dropoff" | "pickup">("store_dropoff");
  const [unopened, setUnopened] = useState<"yes" | "no">("no");
  const [details, setDetails] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [now] = useState(() => Date.now());

  if (order.loading && !order.data) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading order" />;
  const lines = (order.data?.fulfilments ?? [])
    .filter((f) => f.status === "delivered" && f.deliveredAt && now - new Date(f.deliveredAt).getTime() <= WINDOW_MS)
    .flatMap((f) => f.lines ?? [])
    .filter((l) => !l.cancelledAt);

  const submit = async () => {
    const items = lines.map((l) => ({ orderLineId: l.id, quantity: Number.parseInt(qty[l.id] ?? (lines.length === 1 ? String(l.quantity) : "0"), 10) || 0 })).filter((i) => i.quantity > 0);
    if (!items.length) return setError("Set how many of each item you're returning.");
    if (!reason) return setError("Choose why you're returning it.");
    if (reason === "change_of_mind" && unopened !== "yes") return setError("Change-of-mind returns must be unopened with the seals intact.");
    setBusy(true);
    setError(null);
    try {
      const r = await session.api.post("/api/v1/me/orders/{number}/returns", {
        path: { number },
        body: { items, reason, method: reason === "change_of_mind" ? "store_dropoff" : method, details: details.trim() || undefined, unopened: unopened === "yes" || undefined },
      });
      router.replace({ pathname: "/returns/[rma]", params: { rma: r.rma } });
    } catch (e) {
      setError(problem(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Screen underHeader>
      <Stack.Screen options={{ title: "Return items" }} />
      <Title>Return items</Title>
      {lines.length === 0 ? (
        <Notice>Returns open when an item is delivered and stay open for 7 days. Faulty after that? Make a warranty claim from your account.</Notice>
      ) : (
        <>
          {lines.map((l) => (
            <Field
              key={l.id}
              label={`${l.productName} (bought ${l.quantity})`}
              value={qty[l.id] ?? (lines.length === 1 ? String(l.quantity) : "0")}
              onChangeText={(v) => setQty({ ...qty, [l.id]: v.replace(/\D/g, "") })}
              keyboardType="number-pad"
              maxLength={3}
            />
          ))}
          <Choice<Reason>
            label="Why are you returning it?"
            value={reason}
            onChange={setReason}
            options={[
              { value: "faulty", label: "It's faulty" },
              { value: "damaged", label: "It arrived damaged" },
              { value: "not_as_described", label: "It's not as described" },
              { value: "wrong_item", label: "You sent the wrong item" },
              { value: "change_of_mind", label: "I changed my mind", hint: "Unopened only, brought to a store" },
            ]}
          />
          {reason === "change_of_mind" ? (
            <Choice<"yes" | "no">
              label="Is it unopened, with the seals intact?"
              value={unopened}
              onChange={setUnopened}
              options={[
                { value: "yes", label: "Yes, unopened" },
                { value: "no", label: "No, it's been opened" },
              ]}
            />
          ) : (
            <Choice<"store_dropoff" | "pickup">
              label="How will it come back?"
              value={method}
              onChange={setMethod}
              options={[
                { value: "store_dropoff", label: "I'll bring it to a store" },
                { value: "pickup", label: "Collect it from my address", hint: "Free in Lagos and Abuja" },
              ]}
            />
          )}
          <Field label="What happened? (optional)" value={details} onChangeText={setDetails} multiline maxLength={1000} />
          {error ? <Notice tone="danger">{error}</Notice> : null}
          <Button label="Request the return" variant="primary" busy={busy} onPress={submit} />
          <Body muted small>Faulty, damaged, wrong or not-as-described items are usually approved straight away.</Body>
        </>
      )}
    </Screen>
  );
}
