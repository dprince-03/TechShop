import { Stack, useLocalSearchParams, useRouter } from "expo-router";
import { useState } from "react";

import { Body, Button, Field, Notice, Screen, Title } from "@techshop/mobile-ui";

import { useSession } from "@/features/auth/SessionProvider";
import { problem } from "@/features/reviews/problem";

/** A warranty claim for one covered item. */
export default function NewClaim() {
  const { line, name, serialised } = useLocalSearchParams<{ line: string; name?: string; serialised?: string }>();
  const { session } = useSession();
  const router = useRouter();
  const [serial, setSerial] = useState("");
  const [fault, setFault] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const submit = async () => {
    const imeiOrSerial = serial.replace(/\s+/g, "");
    if (imeiOrSerial.length < 3) return setError(serialised === "1" ? "Enter the IMEI: dial *#06# to see it." : "Enter the serial number on the label or box.");
    if (fault.trim().length < 5) return setError("Tell us what's wrong and when it started.");
    setBusy(true);
    setError(null);
    try {
      const c = await session.api.post("/api/v1/me/warranty-claims", { body: { orderLineId: line, imeiOrSerial, fault: fault.trim() } });
      router.replace({ pathname: "/warranty/[number]", params: { number: c.number } });
    } catch (e) {
      setError(problem(e));
    } finally {
      setBusy(false);
    }
  };
  return (
    <Screen underHeader>
      <Stack.Screen options={{ title: "Warranty claim" }} />
      <Title>{name ?? "Warranty claim"}</Title>
      <Field
        label={serialised === "1" ? "IMEI" : "Serial number"}
        value={serial}
        onChangeText={setSerial}
        autoCapitalize="characters"
        keyboardType={serialised === "1" ? "number-pad" : "default"}
        maxLength={40}
        hint={serialised === "1" ? "Dial *#06# to see it." : "On the label or the box."}
      />
      <Field label="What's wrong?" value={fault} onChangeText={setFault} multiline maxLength={1000} hint="Physical or liquid damage isn't covered." />
      {error ? <Notice tone="danger">{error}</Notice> : null}
      <Button label="Make the claim" variant="primary" busy={busy} onPress={submit} />
      <Body muted small>Then bring the device, its charger and box to a TechShop store.</Body>
    </Screen>
  );
}
