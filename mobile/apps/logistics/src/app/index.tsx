import { Redirect, useFocusEffect, useRouter } from "expo-router";
import { useCallback, useState } from "react";
import { ActivityIndicator, RefreshControl, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { Body, Button, Card, Notice, Title } from "@techshop/mobile-ui";
import { useSession } from "@/features/auth/SessionProvider";
import { groupJobs, useJobs, type Job } from "@/features/jobs/useJobs";
import { useShift } from "@/features/shift/useShift";
import { useSync } from "@/features/sync/useSync";
import { tokens, useColors } from "@/theme";

const time = (iso?: string) => (iso ? new Date(iso).toLocaleTimeString("en-NG", { hour: "2-digit", minute: "2-digit", timeZone: "Africa/Lagos" }) : "");

/** Today: the shift toggle, then jobs grouped Pickup → To deliver → Done. Works offline from cache. */
export default function Today() {
  const { status, me } = useSession();
  const shift = useShift();
  const { jobs, refresh, loading, offline } = useJobs();
  const sync = useSync();
  const router = useRouter();
  const c = useColors();
  const [disclose, setDisclose] = useState(false);

  // Each screen keeps its own copy of jobs and the queue; coming back to Today (e.g. after a
  // pickup on the job screen) reloads both so the list and the "waiting to sync" count are current.
  const reloadQueue = sync.reload;
  useFocusEffect(
    useCallback(() => {
      void refresh();
      void reloadQueue();
    }, [refresh, reloadQueue]),
  );

  if (status === "restoring") return <ActivityIndicator style={{ flex: 1 }} accessibilityLabel="Loading" />;
  if (status === "signed-out") return <Redirect href="/sign-in" />;

  const g = groupJobs(jobs);
  const section = (title: string, list: Job[]) =>
    list.length ? (
      <View style={styles.section}>
        <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.xl, fontWeight: "600" }}>
          {title} ({list.length})
        </Text>
        {list.map((j) => (
          <Card key={j.id} label={`${j.kind === "return" ? "Return" : "Delivery"} for ${j.recipientName}, ${j.address}`} onPress={() => router.push({ pathname: "/job/[id]", params: { id: j.id } })}>
            <Text style={{ color: c.text, fontSize: tokens.fontSize.lg, fontWeight: "600" }}>{j.recipientName}</Text>
            <Body muted>{[j.address, j.city].filter(Boolean).join(", ")}</Body>
            {j.windowStart ? <Body small>Window {time(j.windowStart)}–{time(j.windowEnd)}</Body> : null}
            {j.orderNumber ? <Body small muted>{j.orderNumber}</Body> : null}
          </Card>
        ))}
      </View>
    ) : null;

  return (
    <SafeAreaView style={[styles.screen, { backgroundColor: c.bg }]} edges={["top", "left", "right"]}>
      <ScrollView contentContainerStyle={styles.content} refreshControl={<RefreshControl refreshing={loading} onRefresh={refresh} />}>
        <Title>Hello{me ? `, ${me.firstName}` : ""}</Title>
        {offline ? <Notice tone="warning">No signal. Showing your last job list; actions will sync later.</Notice> : null}
        {shift.error ? <Notice tone="danger">{shift.error}</Notice> : null}
        {shift.onShift ? (
          <Button label="End shift" busy={shift.busy} onPress={shift.end} accessibilityHint="Stops sharing your location" />
        ) : (
          <>
            {disclose ? (
              <Card>
                <Text accessibilityRole="header" style={{ color: c.text, fontSize: tokens.fontSize.lg, fontWeight: "600" }}>
                  Location while on shift
                </Text>
                <Body>TechShop Logistics uses your location while you are on shift, so dispatch can give you nearby jobs and customers can see their parcel is on the way.</Body>
                <Body muted small>It stops when you end your shift. Raw location points are deleted after 90 days.</Body>
                <Button label="Continue" variant="primary" busy={shift.busy} onPress={() => { setDisclose(false); void shift.start(); }} />
                <Button label="Not now" onPress={() => setDisclose(false)} />
              </Card>
            ) : (
              <Button
                label="Start shift"
                variant="primary"
                busy={shift.busy}
                onPress={async () => ((await shift.needsDisclosure()) ? setDisclose(true) : shift.start())}
                accessibilityHint="Shares your location with dispatch while on shift"
              />
            )}
          </>
        )}
        {sync.pending || sync.conflicts.length ? (
          <Button label={`${sync.pending} waiting to sync${sync.conflicts.length ? ` · ${sync.conflicts.length} need attention` : ""}`} onPress={() => router.push("/sync")} />
        ) : null}
        {section("To pick up", g.pickup)}
        {section("To deliver", g.deliver)}
        {section("Done today", g.done)}
        {!jobs.length && !loading ? <Body muted>{shift.onShift ? "No jobs yet. Dispatch will assign them here." : "Start your shift to receive jobs."}</Body> : null}
        <Button label="Profile" onPress={() => router.push("/profile")} />
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1 },
  content: { padding: tokens.space[4], gap: tokens.space[4] },
  section: { gap: tokens.space[3] },
});
