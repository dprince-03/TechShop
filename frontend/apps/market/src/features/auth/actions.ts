"use server";

import { redirect } from "next/navigation";

import { createAuthFlow, type AuthStep } from "@techshop/bff";

import { mergeGuestCart } from "@/features/cart/server";
import { bff } from "@/lib/bff";

const flow = createAuthFlow(bff);

/** One action for every sign-in step (the form's `intent` says which); works as a plain post. */
export async function signInStep(
  _prev: AuthStep,
  form: FormData,
): Promise<AuthStep> {
  const result = await flow.handle(form);
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
