import { Body, Button, Card, Notice, Screen, Title } from "@/components/ui";
import { useSync } from "@/features/sync/useSync";

const KIND: Record<string, string> = { picked_up: "Picked up", en_route: "On the way", attempt_failed: "Couldn’t deliver", delivered: "Delivered", note: "Note" };

/** Actions saved on the phone: waiting to send, or refused by the server and needing the rider. */
export default function SyncScreen() {
  const sync = useSync();
  return (
    <Screen>
      <Title>Sync</Title>
      <Body muted>{sync.pending ? `${sync.pending} action${sync.pending === 1 ? "" : "s"} waiting for signal.` : "Everything is sent."}</Body>
      <Button label="Sync now" variant="primary" busy={sync.syncing} onPress={sync.syncNow} />
      {sync.conflicts.map((i) => (
        <Card key={i.event.clientEventId}>
          <Notice tone="danger">{i.error?.code === "job_reassigned" ? "This job was given to another rider. Hand the parcel back at the hub." : (i.error?.detail ?? "The server refused this action.")}</Notice>
          <Body>{KIND[i.event.kind] ?? i.event.kind} · {new Date(i.event.occurredAt).toLocaleString("en-NG", { timeZone: "Africa/Lagos" })}</Body>
          <Button label="I’ve dealt with it" onPress={() => sync.dismiss(i.event.clientEventId)} />
        </Card>
      ))}
      {sync.items
        .filter((i) => i.state === "pending")
        .map((i) => (
          <Card key={i.event.clientEventId}>
            <Body>{KIND[i.event.kind] ?? i.event.kind}</Body>
            <Body small muted>{i.attempts ? `Tried ${i.attempts} time${i.attempts === 1 ? "" : "s"}; will retry automatically.` : "Waiting to send."}</Body>
          </Card>
        ))}
    </Screen>
  );
}
