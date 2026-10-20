import Storage from "expo-sqlite/kv-store";
import { useCallback, useEffect, useState } from "react";

import { ApiError, type Schema } from "@techshop/api-client";

import { useSession } from "@/features/auth/SessionProvider";

export type Job = Schema<"RiderJob">;

const JOBS_KEY = "ts.jobs";

/** Groups for the Today screen, in the order a rider works through them. */
export function groupJobs(jobs: Job[]) {
  return {
    pickup: jobs.filter((j) => j.status === "assigned"),
    deliver: jobs.filter((j) => j.status === "picked_up" || j.status === "en_route"),
    done: jobs.filter((j) => ["delivered", "failed", "returned", "cancelled"].includes(j.status)),
  };
}

/**
 * The rider's jobs. The last good list is kept on the phone so Today still works offline;
 * it is cleared on sign-out (customer details must not linger).
 */
export function useJobs() {
  const { session } = useSession();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [offline, setOffline] = useState(false);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const list = (await session.api.get("/api/v1/rider/jobs")) ?? [];
      setJobs(list);
      setOffline(false);
      await Storage.setItem(JOBS_KEY, JSON.stringify(list));
    } catch (e) {
      if (e instanceof ApiError && e.isUnavailable) {
        setOffline(true);
        const raw = await Storage.getItem(JOBS_KEY);
        if (raw) setJobs(JSON.parse(raw) as Job[]);
      }
    } finally {
      setLoading(false);
    }
  }, [session]);

  useEffect(() => {
    // Deferred so the first load doesn't set state during the effect itself.
    const t = setTimeout(() => void refresh(), 0);
    return () => clearTimeout(t);
  }, [refresh]);

  return { jobs, refresh, loading, offline };
}

export async function clearJobCache() {
  await Storage.removeItem(JOBS_KEY);
}
