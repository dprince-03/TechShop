"use server";

import { redirect } from "next/navigation";

import { createAuthFlow, safeReturnPath, type AuthStep } from "@techshop/bff";

import { mergeGuestCart } from "@/features/cart/server";
import { bff } from "@/lib/bff";

const flow = createAuthFlow(bff);

const field = (form: FormData, name: string) => String(form.get(name) ?? "");

/**
 * One action for every sign-in step; `intent` says which. Works as a plain form post too.
 * On success the guest cart is merged and the customer goes back where they came from.
 */
export async function signInStep(
  _prev: AuthStep,
  form: FormData,
): Promise<AuthStep> {
  const next = safeReturnPath(field(form, "next"));
  let result: AuthStep;
  switch (field(form, "intent")) {
    case "request-code":
      result = await flow.requestCode(field(form, "phone"), next);
      break;
    case "resend-code":
      result = await flow.resendCode();
      break;
    case "verify-code":
      result = await flow.verifyCode(field(form, "code"));
      break;
    case "profile":
      result = await flow.completeProfile(
        field(form, "firstName"),
        field(form, "lastName"),
      );
      break;
    case "password":
      result = await flow.password(
        field(form, "identifier"),
        field(form, "password"),
        next,
      );
      break;
    case "device-code":
      result = await flow.verifyDevice(field(form, "code"));
      break;
    case "restart":
      await flow.clearFlow();
      result = { step: "start" };
      break;
    default:
      result = { step: "start" };
  }
  if (result.step === "done") {
    await mergeGuestCart();
    redirect(result.next);
  }
  return result;
}

export async function signOut() {
  await bff.signOut();
  redirect("/");
}
