import "server-only";

import { cookies, headers } from "next/headers";

import {
  ApiError,
  createApiClient,
  normaliseNigerianPhone,
  type Schema,
} from "@techshop/api-client";

import { seal, unseal } from "./crypto";
import { safeReturnPath } from "./paths";
import { forwardedFor } from "./client-ip";
import type { Bff } from "./server";

/**
 * Multi-step sign-in shared by the web apps (phone OTP, password, new-device OTP, TOTP).
 * Challenge IDs and signup/MFA tokens stay server-side in a short-lived sealed cookie, so the
 * page only ever holds what the person typed.
 */
type FlowState = {
  kind: "otp" | "signup" | "device" | "totp";
  challengeId?: string;
  signupToken?: string;
  mfaToken?: string;
  /** Masked destination shown back to the person ("0803 ••• 4567"). */
  hint?: string;
  phone?: string;
  /** Epoch ms after which a new code may be requested. */
  resendAt?: number;
  /** Where to go after sign-in (already validated). */
  next?: string;
};

const FLOW_COOKIE = "ts_auth_flow";
const FLOW_TTL_SECONDS = 15 * 60;

export type AuthStep =
  | { step: "start"; error?: string }
  | { step: "code"; hint: string; resendAfter: number; error?: string }
  | { step: "profile"; error?: string }
  | { step: "device"; hint: string; error?: string }
  | { step: "totp"; error?: string }
  | { step: "done"; next: string };

export const maskPhone = (e164: string) =>
  `0${e164.slice(4, 7)} ••• ${e164.slice(-4)}`;

/** Messages for the API's problem codes; anything else gets a generic line (no enumeration). */
function problemMessage(e: unknown, fallback: string): string {
  if (e instanceof ApiError) {
    if (e.isUnavailable)
      return "We couldn’t reach TechShop. Check your connection and try again.";
    switch (e.code) {
      case "rate_limited":
      case "otp_rate_limited":
        return "Too many attempts. Please wait a few minutes and try again.";
      case "code_expired":
        return "That code has expired. Request a new one.";
      case "code_reused":
        return "That code was already used. Wait for the next one.";
      case "signup_token_invalid":
        return "Your sign-up expired. Start again.";
      case "mfa_enrollment_required":
        return "Set up your authenticator app first, using the link in your invitation.";
    }
    if (e.status === 401 || e.status === 422 || e.status === 400)
      return fallback;
  }
  return "Something went wrong. Please try again.";
}

export function createAuthFlow(bff: Bff) {
  const cfg = () => bff.config();
  // Sign-in calls carry the shopper's IP so the API's per-IP limits apply per person.
  const api = async () =>
    createApiClient({
      baseUrl: cfg().apiUrl,
      headers: forwardedFor(await headers(), cfg()),
    });

  async function device() {
    const ua = (await headers()).get("user-agent") ?? "";
    const browser = /Edg\//.test(ua)
      ? "Edge"
      : /Chrome\//.test(ua)
        ? "Chrome"
        : /Firefox\//.test(ua)
          ? "Firefox"
          : /Safari\//.test(ua)
            ? "Safari"
            : "Browser";
    return {
      audience: cfg().audience,
      platform: "web" as const,
      deviceName: `${browser} (${cfg().app} web)`,
    };
  }

  async function saveFlow(state: FlowState) {
    (await cookies()).set(FLOW_COOKIE, await seal(state, cfg().secret), {
      httpOnly: true,
      secure: cfg().secure,
      sameSite: "lax",
      path: "/",
      maxAge: FLOW_TTL_SECONDS,
    });
  }

  async function readFlow(): Promise<FlowState | null> {
    return unseal<FlowState>(
      (await cookies()).get(FLOW_COOKIE)?.value,
      cfg().secret,
    );
  }

  async function clearFlow() {
    (await cookies()).delete(FLOW_COOKIE);
  }

  async function finish(
    tokens: Schema<"TokenPair">,
    next: string,
  ): Promise<AuthStep> {
    await bff.signIn(tokens);
    await clearFlow();
    return { step: "done", next };
  }

  /** Step 1 (phone): send a code. Unknown numbers get a sign-up code, so nothing is revealed. */
  async function requestCode(
    rawPhone: string,
    next: string,
  ): Promise<AuthStep> {
    const phone = normaliseNigerianPhone(rawPhone);
    if (!phone)
      return {
        step: "start",
        error: "Enter a Nigerian mobile number, e.g. 0803 123 4567.",
      };
    try {
      const r = await (
        await api()
      ).post("/api/v1/auth/otp/request", {
        body: { phone, purpose: "login" },
      });
      await saveFlow({
        kind: "otp",
        challengeId: r.challengeId,
        phone,
        hint: maskPhone(phone),
        resendAt: Date.now() + r.resendAfter * 1000,
        next,
      });
      return {
        step: "code",
        hint: maskPhone(phone),
        resendAfter: r.resendAfter,
      };
    } catch (e) {
      return {
        step: "start",
        error: problemMessage(e, "Enter a valid mobile number."),
      };
    }
  }

  async function resendCode(): Promise<AuthStep> {
    const flow = await readFlow();
    if (!flow?.phone)
      return { step: "start", error: "Start again: enter your phone number." };
    if (flow.resendAt && flow.resendAt > Date.now())
      return {
        step: "code",
        hint: flow.hint ?? "",
        resendAfter: Math.ceil((flow.resendAt - Date.now()) / 1000),
        error: "Please wait before requesting another code.",
      };
    return requestCode(flow.phone, flow.next ?? "/");
  }

  /** Step 2 (code): sign in, or continue to the profile step for a new customer. */
  async function verifyCode(code: string): Promise<AuthStep> {
    const flow = await readFlow();
    if (!flow?.challengeId || flow.kind !== "otp")
      return {
        step: "start",
        error: "Your code expired. Enter your phone number again.",
      };
    const resendAfter = Math.max(
      0,
      Math.ceil(((flow.resendAt ?? 0) - Date.now()) / 1000),
    );
    if (!/^\d{6}$/.test(code.trim()))
      return {
        step: "code",
        hint: flow.hint ?? "",
        resendAfter,
        error: "Enter the 6-digit code we sent you.",
      };
    try {
      const r = await (
        await api()
      ).post("/api/v1/auth/otp/verify", {
        body: {
          challengeId: flow.challengeId,
          code: code.trim(),
          ...(await device()),
        },
      });
      if (r.status === "signed_in" && r.tokens)
        return finish(r.tokens, flow.next ?? "/");
      if (r.status === "mfa_required" && r.mfaToken) {
        await saveFlow({ kind: "totp", mfaToken: r.mfaToken, next: flow.next });
        return { step: "totp" };
      }
      await saveFlow({ ...flow, kind: "signup", signupToken: r.signupToken });
      return { step: "profile" };
    } catch (e) {
      return {
        step: "code",
        hint: flow.hint ?? "",
        resendAfter,
        error: problemMessage(
          e,
          "That code isn’t right. Check it and try again.",
        ),
      };
    }
  }

  /** Step 3 (new customers): name, then signed in. */
  async function completeProfile(
    firstName: string,
    lastName: string,
  ): Promise<AuthStep> {
    const flow = await readFlow();
    if (!flow?.signupToken)
      return { step: "start", error: "Your sign-up expired. Start again." };
    const first = firstName.trim();
    const last = lastName.trim();
    if (!first || !last || first.length > 60 || last.length > 60)
      return { step: "profile", error: "Enter your first and last name." };
    try {
      const t = await (
        await api()
      ).post("/api/v1/auth/otp/signup", {
        body: {
          signupToken: flow.signupToken,
          firstName: first,
          lastName: last,
          ...(await device()),
        },
      });
      return finish(t, flow.next ?? "/");
    } catch (e) {
      return {
        step: "profile",
        error: problemMessage(e, "Check your name and try again."),
      };
    }
  }

  /** Password sign-in; may continue to a new-device code or an authenticator code. */
  async function password(
    identifier: string,
    pw: string,
    next: string,
  ): Promise<AuthStep> {
    const id = identifier.trim();
    if (!id || !pw)
      return {
        step: "start",
        error: "Enter your email or phone and your password.",
      };
    const phone = normaliseNigerianPhone(id);
    try {
      const r = await (
        await api()
      ).post("/api/v1/auth/password", {
        body: {
          identifier: phone ?? id.toLowerCase(),
          password: pw,
          ...(await device()),
        },
      });
      if (r.status === "signed_in" && r.tokens) return finish(r.tokens, next);
      if (r.status === "otp_required" && r.challengeId && r.mfaToken) {
        await saveFlow({
          kind: "device",
          challengeId: r.challengeId,
          mfaToken: r.mfaToken,
          next,
        });
        return { step: "device", hint: "your phone" };
      }
      if (r.status === "mfa_required" && r.mfaToken) {
        await saveFlow({ kind: "totp", mfaToken: r.mfaToken, next });
        return { step: "totp" };
      }
      return {
        step: "start",
        error: "We couldn’t sign you in. Please try again.",
      };
    } catch (e) {
      return {
        step: "start",
        error: problemMessage(e, "Those details don’t match an account."),
      };
    }
  }

  async function verifyDevice(code: string): Promise<AuthStep> {
    const flow = await readFlow();
    if (flow?.kind !== "device" || !flow.challengeId || !flow.mfaToken)
      return { step: "start", error: "Your sign-in expired. Start again." };
    try {
      const t = await (
        await api()
      ).post("/api/v1/auth/mfa/otp", {
        body: {
          challengeId: flow.challengeId,
          mfaToken: flow.mfaToken,
          code: code.trim(),
          ...(await device()),
        },
      });
      return finish(t, flow.next ?? "/");
    } catch (e) {
      return {
        step: "device",
        hint: "your phone",
        error: problemMessage(e, "That code isn’t right."),
      };
    }
  }

  async function verifyTotp(code: string): Promise<AuthStep> {
    const flow = await readFlow();
    if (flow?.kind !== "totp" || !flow.mfaToken)
      return { step: "start", error: "Your sign-in expired. Start again." };
    if (!/^\d{6}$/.test(code.trim()))
      return {
        step: "totp",
        error: "Enter the 6-digit code from your authenticator app.",
      };
    try {
      const t = await (
        await api()
      ).post("/api/v1/auth/mfa/totp", {
        body: {
          mfaToken: flow.mfaToken,
          code: code.trim(),
          ...(await device()),
        },
      });
      return finish(t, flow.next ?? "/");
    } catch (e) {
      return {
        step: "totp",
        error: problemMessage(
          e,
          "That code isn’t right. Codes change every 30 seconds.",
        ),
      };
    }
  }

  async function verifyRecovery(code: string): Promise<AuthStep> {
    const flow = await readFlow();
    if (flow?.kind !== "totp" || !flow.mfaToken)
      return { step: "start", error: "Your sign-in expired. Start again." };
    try {
      const t = await (
        await api()
      ).post("/api/v1/auth/mfa/recovery", {
        body: {
          mfaToken: flow.mfaToken,
          code: code.trim(),
          ...(await device()),
        },
      });
      return finish(t, flow.next ?? "/");
    } catch (e) {
      return {
        step: "totp",
        error: problemMessage(e, "That recovery code isn’t valid."),
      };
    }
  }

  /**
   * Runs the step a sign-in form posted (`intent` field), so each app's Server Action is a
   * one-liner. `next` is validated here, never trusted from the form.
   */
  async function handle(form: FormData): Promise<AuthStep> {
    const field = (name: string) => String(form.get(name) ?? "");
    const next = safeReturnPath(field("next"));
    switch (field("intent")) {
      case "request-code":
        return requestCode(field("phone"), next);
      case "resend-code":
        return resendCode();
      case "verify-code":
        return verifyCode(field("code"));
      case "profile":
        return completeProfile(field("firstName"), field("lastName"));
      case "password":
        return password(field("identifier"), field("password"), next);
      case "device-code":
        return verifyDevice(field("code"));
      case "totp":
        return verifyTotp(field("code"));
      case "restart":
        await clearFlow();
        return { step: "start" };
      default:
        return { step: "start" };
    }
  }

  return {
    requestCode,
    resendCode,
    verifyCode,
    completeProfile,
    password,
    verifyDevice,
    verifyTotp,
    verifyRecovery,
    clearFlow,
    handle,
  };
}
