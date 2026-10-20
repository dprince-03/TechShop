import { useCallback, useEffect, useRef, useState } from "react";

import { ApiError } from "@techshop/api-client";

/**
 * Loads one API read with loading / error / retry states (no screen is ever left blank).
 * `key` names the inputs (e.g. the search text); a new key loads again.
 */
export function useLoad<T>(load: () => Promise<T>, key: string) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const latest = useRef(load);
  useEffect(() => {
    latest.current = load;
  });

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      setData(await latest.current());
      setError(null);
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 404
          ? "Not found."
          : e instanceof ApiError && e.isUnavailable
            ? "No connection. Pull down to try again."
            : "Something went wrong. Try again.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const t = setTimeout(() => void reload(), 0);
    return () => clearTimeout(t);
  }, [key, reload]);

  return { data, error, loading, reload };
}
