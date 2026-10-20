import * as Location from "expo-location";
import Storage from "expo-sqlite/kv-store";
import { useCallback, useEffect, useState } from "react";
import { AppState } from "react-native";

import { ApiError, type Schema } from "@techshop/api-client";

import { registerDevice, useSession } from "@/features/auth/SessionProvider";

type Shift = Schema<"LogisticsRiderShift">;

const SHIFT_KEY = "ts.shift";
const FIX_TIMEOUT_MS = 8_000;

/** A fresh fix, or the last known one if none arrives within 8 s (weak signal, indoors). */
async function currentPosition() {
  const fresh = Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.Balanced }).catch(() => null);
  const timeout = new Promise<null>((r) => setTimeout(() => r(null), FIX_TIMEOUT_MS));
  return (await Promise.race([fresh, timeout])) ?? (await Location.getLastKnownPositionAsync().catch(() => null));
}
const LOCATION_EVERY_MS = 60_000;

/**
 * The rider's shift. The API has no "current shift" read, so the open shift is remembered on the
 * phone from the start/end responses. Location is sent only while on shift and the app is open
 * (background tracking needs the Play/App Store declarations first: docs/mobile.md §4.4).
 */
export function useShift() {
  const { session } = useSession();
  const [shift, setShift] = useState<Shift | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    void Storage.getItem(SHIFT_KEY).then((raw) => raw && setShift(JSON.parse(raw) as Shift));
  }, []);

  const remember = async (s: Shift | null) => {
    setShift(s);
    if (s) await Storage.setItem(SHIFT_KEY, JSON.stringify(s));
    else await Storage.removeItem(SHIFT_KEY);
  };

  const start = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const perm = await Location.requestForegroundPermissionsAsync();
      if (perm.status !== "granted") {
        setError("Allow location to start a shift: dispatch needs it to route deliveries.");
        return;
      }
      const pos = await currentPosition();
      const s = await session.api.post("/api/v1/rider/shift/start", {
        body: { locationConsent: true, latitude: pos?.coords.latitude, longitude: pos?.coords.longitude },
      });
      await remember(s);
    } catch (e) {
      if (e instanceof ApiError && e.code === "device_not_approved") {
        await registerDevice(session);
        setError("This phone isn’t approved yet. Ask dispatch to approve it (Profile shows its ID).");
      }
      else if (e instanceof ApiError && e.isUnavailable) setError("No connection. A shift can only start online.");
      else setError(e instanceof ApiError && e.detail ? e.detail : "We couldn’t start your shift.");
    } finally {
      setBusy(false);
    }
  }, [session]);

  const end = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      await session.api.post("/api/v1/rider/shift/end");
      await remember(null);
    } catch (e) {
      // Already closed on the server (e.g. the 12-hour limit): forget it here too.
      if (e instanceof ApiError && e.status >= 400 && e.status < 500) await remember(null);
      else setError("No connection. Try ending the shift again when you’re online.");
    } finally {
      setBusy(false);
    }
  }, [session]);

  // While on shift and in the foreground, send the latest position every minute.
  useEffect(() => {
    if (!shift) return;
    let active = AppState.currentState === "active";
    const send = async () => {
      if (!active) return;
      const pos = await Location.getLastKnownPositionAsync().catch(() => null);
      if (!pos) return;
      await session.api
        .post("/api/v1/rider/location", {
          body: { points: [{ latitude: pos.coords.latitude, longitude: pos.coords.longitude, accuracyM: pos.coords.accuracy ?? undefined, recordedAt: new Date(pos.timestamp).toISOString() }] },
        })
        .catch(() => undefined);
    };
    void send();
    const timer = setInterval(send, LOCATION_EVERY_MS);
    const sub = AppState.addEventListener("change", (s) => (active = s === "active"));
    return () => {
      clearInterval(timer);
      sub.remove();
    };
  }, [shift, session]);

  /** True until the rider has granted location: show the disclosure before the OS prompt. */
  const needsDisclosure = useCallback(async () => (await Location.getForegroundPermissionsAsync()).status !== "granted", []);

  return { shift, onShift: Boolean(shift), start, end, busy, error, needsDisclosure };
}
