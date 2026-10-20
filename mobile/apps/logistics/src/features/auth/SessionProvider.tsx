import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { Platform } from "react-native";

import { ApiError, createApiClient, normaliseNigerianPhone, type Schema } from "@techshop/api-client";

import { env } from "@/lib/env";
import { clearLocalData, deviceId, secureTokenStore } from "@/lib/storage";

import { RiderSession } from "./session";

type Me = Schema<"Me">;
type Status = "restoring" | "signed-out" | "signed-in";

type SessionContext = {
  status: Status;
  me: Me | null;
  session: RiderSession;
  /** Step 1: send a code. Returns an error message, or null. */
  requestCode: (phone: string) => Promise<string | null>;
  /** Step 2: check the code and sign in. Returns an error message, or null. */
  verifyCode: (code: string) => Promise<string | null>;
  signOut: () => Promise<void>;
};

const Ctx = createContext<SessionContext | null>(null);

const message = (e: unknown, fallback: string) => {
  if (e instanceof ApiError) {
    if (e.isUnavailable) return "No connection to TechShop. Check your data and try again.";
    if (e.code === "otp_rate_limited" || e.code === "rate_limited") return "Too many tries. Wait a few minutes.";
    if (e.code === "code_expired") return "That code expired. Ask for a new one.";
  }
  return fallback;
};

/** Holds the rider's session for the whole app: restore on launch, sign in with phone OTP, sign out. */
export function SessionProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<Status>("restoring");
  const [me, setMe] = useState<Me | null>(null);
  const [challengeId, setChallengeId] = useState<string | null>(null);

  const session = useMemo(
    () =>
      new RiderSession(env.apiUrl, secureTokenStore, undefined, {
        onSignedOut: () => {
          setMe(null);
          setStatus("signed-out");
        },
      }),
    [],
  );

  const loadMe = useCallback(async () => {
    const profile = await session.api.get("/api/v1/me");
    if (!profile.isRider) throw new ApiError(403, { code: "not_a_rider" });
    setMe(profile);
    setStatus("signed-in");
  }, [session]);

  useEffect(() => {
    // Restore: a stored refresh token signs the rider straight back in (offline keeps them in).
    void (async () => {
      if (!(await session.hasSession())) return setStatus("signed-out");
      try {
        await loadMe();
      } catch (e) {
        setStatus(e instanceof ApiError && e.isUnavailable ? "signed-in" : "signed-out");
      }
    })();
  }, [session, loadMe]);

  const value: SessionContext = {
    status,
    me,
    session,
    async requestCode(raw) {
      const phone = normaliseNigerianPhone(raw);
      if (!phone) return "Enter your mobile number, e.g. 0803 123 4567.";
      try {
        const r = await createApiClient({ baseUrl: env.apiUrl }).post("/api/v1/auth/otp/request", { body: { phone, purpose: "login" } });
        setChallengeId(r.challengeId);
        return null;
      } catch (e) {
        return message(e, "We couldn’t send a code to that number.");
      }
    },
    async verifyCode(code) {
      if (!challengeId) return "Ask for a code first.";
      if (!/^\d{6}$/.test(code)) return "Enter the 6-digit code.";
      try {
        const r = await createApiClient({ baseUrl: env.apiUrl }).post("/api/v1/auth/otp/verify", {
          body: { challengeId, code, audience: "logistics", deviceId: await deviceId(), deviceName: `${Platform.OS} phone`, platform: Platform.OS === "ios" ? "ios" : "android" },
        });
        if (r.status !== "signed_in" || !r.tokens) return "This number isn’t registered as a TechShop rider.";
        await session.accept(r.tokens);
        await loadMe();
        setChallengeId(null);
        return null;
      } catch (e) {
        if (e instanceof ApiError && e.code === "not_a_rider") {
          await session.signOut();
          return "This number isn’t registered as a TechShop rider.";
        }
        return message(e, "That code isn’t right. Check it and try again.");
      }
    },
    async signOut() {
      await session.signOut();
      await clearLocalData();
    },
  };

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useSession(): SessionContext {
  const v = useContext(Ctx);
  if (!v) throw new Error("useSession must be used inside SessionProvider");
  return v;
}
