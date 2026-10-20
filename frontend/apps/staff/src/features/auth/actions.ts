"use server";

import { redirect } from "next/navigation";

import { createAuthFlow, safeReturnPath, type AuthStep } from "@techshop/bff";

import { bff } from "@/lib/bff";

const flow = createAuthFlow(bff);
const field = (form: FormData, name: string) => String(form.get(name) ?? "");

/** Staff sign-in: email + password, then the authenticator (or a recovery) code. */
export async function staffSignIn(
  _prev: AuthStep,
  form: FormData,
): Promise<AuthStep> {
  const next = safeReturnPath(field(form, "next"));
  let result: AuthStep;
  switch (field(form, "intent")) {
    case "password":
      result = await flow.password(
        field(form, "email"),
        field(form, "password"),
        next,
      );
      break;
    case "totp":
      result = await flow.verifyTotp(field(form, "code"));
      break;
    case "recovery":
      result = await flow.verifyRecovery(field(form, "code"));
      break;
    case "restart":
      await flow.clearFlow();
      result = { step: "start" };
      break;
    default:
      result = { step: "start" };
  }
  if (result.step === "done") redirect(result.next);
  // Staff accounts never use phone codes here; anything else means start again.
  if (
    result.step === "code" ||
    result.step === "profile" ||
    result.step === "device"
  )
    return { step: "start", error: "Use your staff email and password." };
  return result;
}

/** Confirms an authenticator code so the next sensitive action is allowed (step-up). */
export async function confirmStepUp(
  code: string,
): Promise<{ ok: boolean; message: string }> {
  const ok = await bff.stepUp(code);
  return ok
    ? { ok: true, message: "Confirmed." }
    : {
        ok: false,
        message: "That code isn’t right. Codes change every 30 seconds.",
      };
}

export async function signOut() {
  await bff.signOut();
  redirect("/sign-in");
}
