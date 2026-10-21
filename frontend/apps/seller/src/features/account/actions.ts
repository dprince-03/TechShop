"use server";

import { failure } from "@techshop/bff";

import { bff } from "@/lib/bff";

export type SetupState = {
  ok?: boolean;
  message?: string;
  secret?: string;
  otpauthUrl?: string;
  recoveryCodes?: string[];
};

/** Starts authenticator setup: a new secret to add to the app. */
export async function startAuthenticator(): Promise<SetupState> {
  if (!(await bff.getSession({ refresh: true })))
    return { ok: false, message: "Your session ended. Sign in again." };
  try {
    const s = await (await bff.api()).post("/api/v1/me/mfa/totp/setup");
    return { ok: true, secret: s.secret, otpauthUrl: s.otpauthUrl };
  } catch (e) {
    return { ok: false, message: failure(e).message };
  }
}

/** Turns the authenticator on with the first code; the recovery codes are shown once. */
export async function confirmAuthenticator(
  prev: SetupState,
  form: FormData,
): Promise<SetupState> {
  const code = String(form.get("code") ?? "").replace(/\D/g, "");
  if (code.length !== 6)
    return {
      ...prev,
      ok: false,
      message: "Enter the 6-digit code from the app.",
    };
  try {
    const r = await (
      await bff.api()
    ).post("/api/v1/me/mfa/totp/confirm", {
      body: { code },
    });
    return { ok: true, recoveryCodes: r.recoveryCodes ?? [] };
  } catch (e) {
    return { ...prev, ok: false, message: failure(e).message };
  }
}
