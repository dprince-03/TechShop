import * as Location from "expo-location";
import { useLocalSearchParams, useRouter } from "expo-router";
import { useState } from "react";
import { ActivityIndicator, Linking, Platform, Pressable, StyleSheet, Text, View } from "react-native";

import type { Schema } from "@techshop/api-client";

import { Body, Button, Card, Field, Notice, Screen, Title } from "@techshop/mobile-ui";
import { useSession } from "@/features/auth/SessionProvider";
import { performAction, type ActionInput } from "@/features/jobs/actions";
import { LabelScanner } from "@/features/jobs/LabelScanner";
import { useJobs } from "@/features/jobs/useJobs";
import { tokens, useColors } from "@/theme";

type Reason = NonNullable<Schema<"Action">["reason"]>;
const REASONS: { value: Exclude<Reason, "">; label: string }[] = [
  { value: "customer_unreachable", label: "Customer unreachable" },
  { value: "customer_refused", label: "Customer refused" },
  { value: "wrong_address", label: "Wrong address" },
  { value: "rescheduled", label: "Customer rescheduled" },
  { value: "unsafe", label: "Unsafe to deliver" },
  { value: "other", label: "Other" },
];

async function position() {
  const pos = await Location.getLastKnownPositionAsync().catch(() => null);
  return pos ? { latitude: pos.coords.latitude, longitude: pos.coords.longitude, accuracyM: pos.coords.accuracy ?? undefined } : undefined;
}

/** One job, one primary action per state: pick up → start delivery → deliver (code) or couldn't deliver. */
export default function JobScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { session } = useSession();
  const { jobs, refresh, loading } = useJobs();
  const router = useRouter();
  const c = useColors();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "danger" | "warning"; text: string } | null>(null);
  const [otp, setOtp] = useState("");
  const [recipient, setRecipient] = useState("");
  const [failing, setFailing] = useState(false);
  const [reason, setReason] = useState<Exclude<Reason, ""> | null>(null);
  const [note, setNote] = useState("");

  const job = jobs.find((j) => j.id === id);
  // Still loading (from the API or the phone's cache): don't claim the job is gone yet.
  if (!job && loading) return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading job" />;
  if (!job) {
    return (
      <Screen underHeader>
        <Body muted>This job isn’t in your list any more. It may have been reassigned.</Body>
        <Button label="Back to today" onPress={() => (router.canGoBack() ? router.back() : router.replace("/"))} />
      </Screen>
    );
  }

  const act = async (input: ActionInput, success: string) => {
    setBusy(true);
    setMessage(null);
    const out = await performAction(session, job.id, input, await position());
    setBusy(false);
    if (out.kind === "done") {
      setMessage({ tone: "success", text: success });
      await refresh();
    } else if (out.kind === "queued") {
      setMessage({ tone: "warning", text: `${success} Saved on the phone; it syncs when you have signal.` });
    } else {
      setMessage({ tone: "danger", text: out.message });
    }
  };

  const navigate = () => {
    const q = job.latitude && job.longitude ? `${job.latitude},${job.longitude}` : encodeURIComponent(`${job.address}, ${job.city}`);
    void Linking.openURL(Platform.OS === "ios" ? `maps://?daddr=${q}` : `geo:0,0?q=${q}`);
  };

  return (
    <Screen underHeader>
      <Title>{job.recipientName}</Title>
      <Body>{[job.address, job.city, job.stateCode].filter(Boolean).join(", ")}</Body>
      {job.landmark ? <Body muted>Landmark: {job.landmark}</Body> : null}
      {job.orderNumber ? <Body small muted>Order {job.orderNumber} · attempt {job.attempts + 1}</Body> : null}
      {job.status === "assigned" && job.pickupName ? <Card><Body>Pick up from {job.pickupName}</Body>{job.pickupAddress ? <Body small muted>{job.pickupAddress}</Body> : null}</Card> : null}
      <Button label="Navigate" onPress={navigate} />
      {message ? <Notice tone={message.tone}>{message.text}</Notice> : null}

      {job.status === "assigned" ? <LabelScanner busy={busy} onLabel={(labelCode) => act({ kind: "picked_up", labelCode }, "Picked up. The customer has been sent their delivery code.")} /> : null}

      {job.status === "picked_up" ? <Button label="Start delivery" variant="primary" busy={busy} onPress={() => act({ kind: "en_route" }, "On your way. The customer can now follow you.")} /> : null}

      {(job.status === "picked_up" || job.status === "en_route") && !failing ? (
        <View style={styles.block}>
          <Field label="Customer’s delivery code" value={otp} onChangeText={(t) => setOtp(t.replace(/\D/g, "").slice(0, 6))} keyboardType="number-pad" maxLength={6} hint="Ask the customer for the 6-digit code from their SMS." />
          <Field label="Who received it (optional)" value={recipient} onChangeText={setRecipient} autoCapitalize="words" />
          <Button label="Delivered" variant={job.status === "en_route" ? "primary" : "secondary"} busy={busy} disabled={otp.length !== 6} onPress={() => act({ kind: "delivered", otp, recipientName: recipient.trim() || undefined }, "Delivered. Well done!")} />
          <Button label="Couldn’t deliver" onPress={() => setFailing(true)} />
        </View>
      ) : null}

      {failing ? (
        <View style={styles.block}>
          <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>
            What happened?
          </Text>
          {REASONS.map((r) => (
            <Pressable key={r.value} accessibilityRole="radio" accessibilityState={{ selected: reason === r.value }} onPress={() => setReason(r.value)} style={[styles.choice, { borderColor: reason === r.value ? c.accent : c.borderStrong, backgroundColor: reason === r.value ? c.accentSubtle : c.bg }]}>
              <Text style={{ color: c.text, fontSize: tokens.fontSize.lg }}>{r.label}</Text>
            </Pressable>
          ))}
          <Field label={reason === "other" ? "Note (required)" : "Note (optional)"} value={note} onChangeText={setNote} />
          <Button
            label="Report attempt"
            variant="danger"
            busy={busy}
            disabled={!reason || (reason === "other" && note.trim().length < 3)}
            onPress={() => act({ kind: "attempt_failed", reason: reason!, note: note.trim() || undefined }, "Attempt recorded. Dispatch will tell you what to do next.")}
          />
          <Button label="Cancel" onPress={() => setFailing(false)} />
        </View>
      ) : null}
    </Screen>
  );
}

const styles = StyleSheet.create({
  block: { gap: tokens.space[3] },
  choice: { minHeight: tokens.control.lg, borderWidth: 1, borderRadius: tokens.radius.md, justifyContent: "center", paddingHorizontal: tokens.space[4] },
});
