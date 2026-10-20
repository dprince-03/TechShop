// Offline queue for rider actions (docs/mobile.md §4.3). Pure TypeScript: storage and the network
// are passed in, so the logic is unit-tested in Node and runs the same on the phone.
import type { Schema } from "@techshop/api-client";

export type RiderEvent = Schema<"Action">;
export type SyncResult = Schema<"Result">;

export type QueuedEvent = {
  event: RiderEvent;
  /** Failed send attempts so far (network or server errors only). */
  attempts: number;
  /** Epoch ms before which this item is not retried. */
  nextAttemptAt: number;
  /** "conflict" = the API refused it (e.g. job reassigned); kept for the rider to see, never retried. */
  state: "pending" | "conflict";
  error?: { code?: string; detail?: string };
};

export interface QueueStore {
  load(): Promise<QueuedEvent[]>;
  save(items: QueuedEvent[]): Promise<void>;
}

/** Sends a batch; throws on network/5xx so the engine backs off. */
export type SendBatch = (events: RiderEvent[]) => Promise<SyncResult[]>;

export const BATCH_SIZE = 50;
const BASE_DELAY_MS = 2_000;
export const MAX_DELAY_MS = 5 * 60_000;

/** Exponential backoff with full jitter, capped at 5 minutes. */
export function backoff(attempts: number, random = Math.random): number {
  const ceiling = Math.min(MAX_DELAY_MS, BASE_DELAY_MS * 2 ** Math.max(0, attempts - 1));
  return Math.round(ceiling * (0.5 + random() / 2));
}

/** Adds an action to the queue (the UI shows it as "not yet synced"). */
export async function enqueue(store: QueueStore, event: RiderEvent): Promise<void> {
  const items = await store.load();
  if (items.some((i) => i.event.clientEventId === event.clientEventId)) return;
  items.push({ event, attempts: 0, nextAttemptAt: 0, state: "pending" });
  await store.save(items);
}

export type SyncSummary = { sent: number; accepted: number; conflicts: number; failed: boolean };

/**
 * Sends due pending items in order (oldest first) and applies the per-item results:
 * accepted/duplicate → removed; rejected → conflict (kept, not retried); a failed request →
 * every sent item backs off. Items are never dropped silently.
 */
export async function syncOnce(store: QueueStore, send: SendBatch, now = Date.now()): Promise<SyncSummary> {
  const items = await store.load();
  const due = items.filter((i) => i.state === "pending" && i.nextAttemptAt <= now).slice(0, BATCH_SIZE);
  if (!due.length) return { sent: 0, accepted: 0, conflicts: 0, failed: false };

  let results: SyncResult[];
  try {
    results = await send(due.map((i) => i.event));
  } catch {
    for (const i of due) {
      i.attempts += 1;
      i.nextAttemptAt = now + backoff(i.attempts);
    }
    await store.save(items);
    return { sent: due.length, accepted: 0, conflicts: 0, failed: true };
  }

  const byId = new Map(results.map((r) => [r.clientEventId, r]));
  let accepted = 0;
  let conflicts = 0;
  const remaining: QueuedEvent[] = [];
  for (const item of items) {
    const r = byId.get(item.event.clientEventId);
    if (!r) {
      remaining.push(item);
    } else if (r.status === "accepted" || r.status === "duplicate") {
      accepted += 1;
    } else {
      conflicts += 1;
      remaining.push({ ...item, state: "conflict", error: { code: r.code, detail: r.detail } });
    }
  }
  await store.save(remaining);
  return { sent: due.length, accepted, conflicts, failed: false };
}

/** Clears a conflict the rider has acknowledged (e.g. after handing the parcel back). */
export async function dismiss(store: QueueStore, clientEventId: string): Promise<void> {
  await store.save((await store.load()).filter((i) => i.event.clientEventId !== clientEventId));
}

export function memoryStore(initial: QueuedEvent[] = []): QueueStore & { items: QueuedEvent[] } {
  const s = {
    items: initial,
    async load() {
      return s.items.map((i) => ({ ...i }));
    },
    async save(items: QueuedEvent[]) {
      s.items = items.map((i) => ({ ...i }));
    },
  };
  return s;
}
