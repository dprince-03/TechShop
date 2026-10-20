import { useCallback, useEffect, useRef, useState } from "react";
import { AppState } from "react-native";

import { ApiError } from "@techshop/api-client";

import { useSession } from "@/features/auth/SessionProvider";
import { queueStore } from "@/lib/storage";

import { dismiss, syncOnce, type QueuedEvent } from "./queue";

const EVERY_MS = 60_000;

/** Runs the offline queue: on launch, every minute, when the app returns to the foreground, and on demand. */
export function useSync() {
  const { session, status } = useSession();
  const [items, setItems] = useState<QueuedEvent[]>([]);
  const [syncing, setSyncing] = useState(false);
  const running = useRef(false);

  const reload = useCallback(async () => setItems(await queueStore.load()), []);

  const syncNow = useCallback(async () => {
    if (running.current || status !== "signed-in") return;
    running.current = true;
    setSyncing(true);
    try {
      await syncOnce(queueStore, async (events) => {
        try {
          return (await session.api.post("/api/v1/rider/events/batch", { body: { events } })).results ?? [];
        } catch (e) {
          // Network or server trouble: the engine backs off. Other refusals reject the whole batch.
          if (e instanceof ApiError && !e.isUnavailable && e.status < 500) {
            return events.map((ev) => ({ clientEventId: ev.clientEventId, status: "rejected" as const, code: e.code, detail: e.detail }));
          }
          throw e;
        }
      });
    } finally {
      running.current = false;
      setSyncing(false);
      await reload();
    }
  }, [session, status, reload]);

  useEffect(() => {
    const first = setTimeout(() => void reload().then(syncNow), 0);
    const timer = setInterval(syncNow, EVERY_MS);
    const sub = AppState.addEventListener("change", (s) => s === "active" && void syncNow());
    return () => {
      clearTimeout(first);
      clearInterval(timer);
      sub.remove();
    };
  }, [reload, syncNow]);

  return {
    items,
    pending: items.filter((i) => i.state === "pending").length,
    conflicts: items.filter((i) => i.state === "conflict"),
    syncing,
    syncNow,
    reload,
    dismiss: async (id: string) => {
      await dismiss(queueStore, id);
      await reload();
    },
  };
}
