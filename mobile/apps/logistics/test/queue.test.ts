import assert from "node:assert/strict";
import { test } from "node:test";

import { backoff, BATCH_SIZE, dismiss, enqueue, MAX_DELAY_MS, memoryStore, syncOnce, type RiderEvent } from "../src/features/sync/queue";

const ev = (id: string, kind: RiderEvent["kind"] = "en_route"): RiderEvent => ({ clientEventId: id, jobId: "j1", kind, occurredAt: "2026-10-20T08:00:00Z" });

test("enqueue ignores a duplicate clientEventId", async () => {
  const s = memoryStore();
  await enqueue(s, ev("a"));
  await enqueue(s, ev("a"));
  assert.equal(s.items.length, 1);
});

test("accepted and duplicate results leave the queue; rejected become conflicts", async () => {
  const s = memoryStore();
  for (const id of ["a", "b", "c"]) await enqueue(s, ev(id));
  const r = await syncOnce(s, async () => [
    { clientEventId: "a", status: "accepted" },
    { clientEventId: "b", status: "duplicate" },
    { clientEventId: "c", status: "rejected", code: "job_reassigned", detail: "Job moved" },
  ]);
  assert.deepEqual(r, { sent: 3, accepted: 2, conflicts: 1, failed: false });
  assert.equal(s.items.length, 1);
  assert.equal(s.items[0].state, "conflict");
  assert.equal(s.items[0].error?.code, "job_reassigned");
  // Conflicts are never resent.
  let calls = 0;
  await syncOnce(s, async () => {
    calls += 1;
    return [];
  });
  assert.equal(calls, 0);
  await dismiss(s, "c");
  assert.equal(s.items.length, 0);
});

test("a failed request keeps every item and backs off; it is retried when due", async () => {
  const s = memoryStore();
  await enqueue(s, ev("a"));
  const t0 = 1_000_000;
  const r = await syncOnce(s, async () => Promise.reject(new Error("offline")), t0);
  assert.equal(r.failed, true);
  assert.equal(s.items[0].attempts, 1);
  assert.ok(s.items[0].nextAttemptAt > t0);
  // Not due yet: nothing is sent.
  let sent = 0;
  await syncOnce(s, async (e) => ((sent += e.length), []), t0 + 1);
  assert.equal(sent, 0);
  // Due: sent and accepted.
  const later = await syncOnce(s, async (e) => e.map((x) => ({ clientEventId: x.clientEventId, status: "accepted" as const })), s.items[0].nextAttemptAt);
  assert.equal(later.accepted, 1);
  assert.equal(s.items.length, 0);
});

test("items missing from the response stay pending; batches are capped", async () => {
  const s = memoryStore();
  for (let i = 0; i < BATCH_SIZE + 5; i++) await enqueue(s, ev(`e${i}`));
  let size = 0;
  await syncOnce(s, async (e) => {
    size = e.length;
    return [{ clientEventId: "e0", status: "accepted" }];
  });
  assert.equal(size, BATCH_SIZE);
  assert.equal(s.items.length, BATCH_SIZE + 4);
  assert.ok(s.items.every((i) => i.state === "pending"));
});

test("backoff grows and is capped at five minutes", () => {
  assert.ok(backoff(1, () => 1) <= backoff(4, () => 1));
  assert.equal(backoff(30, () => 1), MAX_DELAY_MS);
  assert.ok(backoff(1, () => 0) >= 1_000);
});
