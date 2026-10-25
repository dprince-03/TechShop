import { Stack, useRouter } from "expo-router";
import { useState } from "react";
import { ActivityIndicator } from "react-native";

import { formatKobo } from "@techshop/api-client";
import { Body, Button, Choice, Field, Notice, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { useLoad } from "@/features/catalogue/useLoad";
import { problem } from "@/features/reviews/problem";

type DeviceType = "phone" | "laptop" | "tablet" | "console" | "smartwatch";
type Condition = "like_new" | "good" | "fair" | "faulty";
const OTHER = "other";

/** Start a trade-in: type, model (or typed in), condition, IMEI, and the owner's confirmation. */
export default function NewTradeIn() {
  const { session } = useSession();
  const router = useRouter();
  const [type, setType] = useState<DeviceType | null>(null);
  const [model, setModel] = useState<string | null>(null);
  const [brand, setBrand] = useState("");
  const [typed, setTyped] = useState("");
  const [condition, setCondition] = useState<Condition | null>(null);
  const [imei, setImei] = useState("");
  const [owner, setOwner] = useState<"yes" | "no">("no");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const models = useLoad(async () => (type ? ((await session.api.get("/api/v1/trade-ins/devices", { query: { type } })) ?? []) : []), type ?? "none");

  const chosen = (models.data ?? []).find((m) => m.variantId === model);
  const submit = async () => {
    if (!type || !condition) return setError("Choose the device type and its condition.");
    if (!model || (model === OTHER && (!brand.trim() || !typed.trim()))) return setError("Choose your model, or type the brand and model.");
    const cleanImei = imei.replace(/\s+/g, "");
    if (cleanImei && !/^\d{15}$/.test(cleanImei)) return setError("An IMEI is 15 digits. Dial *#06# to see it.");
    if (owner !== "yes") return setError("Only the owner can trade a device in.");
    setBusy(true);
    setError(null);
    try {
      const t = await session.api.post("/api/v1/me/trade-ins", {
        body: {
          deviceType: type,
          condition,
          variantId: model !== OTHER ? model : undefined,
          brand: model === OTHER ? brand.trim() : undefined,
          model: model === OTHER ? typed.trim() : undefined,
          imei: cleanImei || undefined,
        },
      });
      router.replace({ pathname: "/trade-in/[ref]", params: { ref: t.reference } });
    } catch (e) {
      setError(problem(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Screen underHeader>
      <Stack.Screen options={{ title: "Trade in" }} />
      <Title>Trade in a device</Title>
      <Choice<DeviceType>
        label="What are you trading in?"
        value={type}
        onChange={(t) => {
          setType(t);
          setModel(null);
        }}
        options={[
          { value: "phone", label: "Phone" },
          { value: "laptop", label: "Laptop" },
          { value: "tablet", label: "Tablet" },
          { value: "console", label: "Games console" },
          { value: "smartwatch", label: "Smartwatch" },
        ]}
      />
      {type && models.loading ? <ActivityIndicator accessibilityLabel="Loading models" /> : null}
      {type && !models.loading ? (
        <Choice<string>
          label="Model"
          value={model}
          onChange={setModel}
          options={[
            ...(models.data ?? []).map((m) => ({ value: m.variantId, label: `${m.brand} ${m.productName} ${m.variantName}`, hint: `Up to ${formatKobo(m.estimates.like_new ?? 0)}` })),
            { value: OTHER, label: "My model isn't listed", hint: "We'll price it within a working day" },
          ]}
        />
      ) : null}
      {model === OTHER ? (
        <>
          <Field label="Brand" value={brand} onChangeText={setBrand} maxLength={60} />
          <Field label="Model" value={typed} onChangeText={setTyped} maxLength={80} />
        </>
      ) : null}
      {model ? (
        <Choice<Condition>
          label="Condition"
          value={condition}
          onChange={setCondition}
          options={[
            { value: "like_new", label: "Like new", hint: chosen ? formatKobo(chosen.estimates.like_new ?? 0) : "No marks" },
            { value: "good", label: "Good", hint: chosen ? formatKobo(chosen.estimates.good ?? 0) : "Light scratches" },
            { value: "fair", label: "Fair", hint: chosen ? formatKobo(chosen.estimates.fair ?? 0) : "Visible wear" },
            { value: "faulty", label: "Faulty", hint: chosen ? formatKobo(chosen.estimates.faulty ?? 0) : "Screen, battery or other issue" },
          ]}
        />
      ) : null}
      {type === "phone" || type === "tablet" ? <Field label="IMEI (optional)" value={imei} onChangeText={setImei} keyboardType="number-pad" maxLength={20} hint="Dial *#06#. We check it isn't reported stolen." /> : null}
      <Choice<"yes" | "no">
        label="Do you own this device?"
        value={owner}
        onChange={setOwner}
        options={[
          { value: "yes", label: "Yes, and I can show proof" },
          { value: "no", label: "No" },
        ]}
      />
      {error ? <Notice tone="danger">{error}</Notice> : null}
      <Button label="Get my estimate" variant="primary" busy={busy} onPress={submit} />
      <Body muted small>The final value is confirmed when we inspect it at a store.</Body>
    </Screen>
  );
}
