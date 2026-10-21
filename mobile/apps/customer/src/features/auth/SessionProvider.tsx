import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { Platform } from "react-native";

import { ApiError, createApiClient, normaliseNigerianPhone, type Schema } from "@techshop/api-client";
import { MobileSession } from "@techshop/mobile-auth";

import { deviceId, secureTokenStore } from "@/lib/storage";

const baseUrl = (process.env.EXPO_PUBLIC_API_URL ?? "").replace(/\/$/, "");

type Me = Schema<"Me">;
type Status = "restoring" | "signed-out" | "signed-in";
/** What the sign-in screen shows next. */
export type Step = "phone" | "code" | "profile";

type Ctx = {
  status: Status;
  me: Me | null;
  session: MobileSession;
  step: Step;
  requestCode: (phone: string) => Promise<string | null>;
  verifyCode: (code: string) => Promise<string | null>;
  completeProfile: (first: string, last: string) => Promise<string | null>;
  restart: () => void;
  signOut: () => Promise<void>;
};

const SessionCtx = createContext<Ctx | null>(null);

const message = (e: unknown, fallback: string) => {
  if (e instanceof ApiError) {
    if (e.isUnavailable) return "No connection. Check your data and try again.";
    if (e.code === "otp_rate_limited" || e.code === "rate_limited") return "Too many tries. Wait a few minutes.";
    if (e.code === "code_expired") return "That code expired. Ask for a new one.";
    if (e.code === "signup_token_invalid") return "Your sign-up expired. Start again.";
  }
  return fallback;
};

/** Customer session: phone code (a new number creates an account), restore on launch, sign out. */
export function SessionProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<Status>("restoring");
  const [me, setMe] = useState<Me | null>(null);
  const [step, setStep] = useState<Step>("phone");
  const [challengeId, setChallengeId] = useState<string | null>(null);
  const [signupToken, setSignupToken] = useState<string | null>(null);

  const session = useMemo(
    () => new MobileSession(baseUrl, secureTokenStore, undefined, { onSignedOut: () => (setMe(null), setStatus("signed-out")) }),
    [],
  );
  const plain = useMemo(() => createApiClient({ baseUrl }), []);

  const loadMe = useCallback(async () => {
    setMe(await session.api.get("/api/v1/me"));
    setStatus("signed-in");
  }, [session]);

  useEffect(() => {
    void (async () => {
      if (!(await session.hasSession())) return setStatus("signed-out");
      try {
        await loadMe();
      } catch (e) {
        setStatus(e instanceof ApiError && e.isUnavailable ? "signed-in" : "signed-out");
      }
    })();
  }, [session, loadMe]);

  const device = async () => ({ audience: "customer_app" as const, deviceId: await deviceId(), deviceName: `${Platform.OS} phone`, platform: Platform.OS === "ios" ? ("ios" as const) : ("android" as const) });

  const value: Ctx = {
    status,
    me,
    session,
    step,
    async requestCode(raw) {
      const phone = normaliseNigerianPhone(raw);
      if (!phone) return "Enter your mobile number, e.g. 0803 123 4567.";
      try {
        setChallengeId((await plain.post("/api/v1/auth/otp/request", { body: { phone, purpose: "login" } })).challengeId);
        setStep("code");
        return null;
      } catch (e) {
        return message(e, "We couldn’t send a code to that number.");
      }
    },
    async verifyCode(code) {
      if (!challengeId) return "Ask for a code first.";
      if (!/^\d{6}$/.test(code)) return "Enter the 6-digit code.";
      try {
        const r = await plain.post("/api/v1/auth/otp/verify", { body: { challengeId, code, ...(await device()) } });
        if (r.status === "needs_profile" && r.signupToken) {
          setSignupToken(r.signupToken);
          setStep("profile");
          return null;
        }
        if (r.status === "mfa_required")
          return "This account uses an authenticator app, which this app doesn’t support yet. Sign in on the TechShop website.";
        if (!r.tokens) return "We couldn’t sign you in. Try again.";
        await session.accept(r.tokens);
        await loadMe();
        setStep("phone");
        return null;
      } catch (e) {
        return message(e, "That code isn’t right. Check it and try again.");
      }
    },
    async completeProfile(first, last) {
      if (!signupToken) return "Your sign-up expired. Start again.";
      if (!first.trim() || !last.trim()) return "Enter your first and last name.";
      try {
        await session.accept(await plain.post("/api/v1/auth/otp/signup", { body: { signupToken, firstName: first.trim().slice(0, 60), lastName: last.trim().slice(0, 60), ...(await device()) } }));
        await loadMe();
        setSignupToken(null);
        setStep("phone");
        return null;
      } catch (e) {
        return message(e, "We couldn’t create your account. Try again.");
      }
    },
    restart() {
      setStep("phone");
      setChallengeId(null);
      setSignupToken(null);
    },
    signOut: () => session.signOut(),
  };

  return <SessionCtx.Provider value={value}>{children}</SessionCtx.Provider>;
}

export function useSession(): Ctx {
  const v = useContext(SessionCtx);
  if (!v) throw new Error("useSession must be used inside SessionProvider");
  return v;
}
