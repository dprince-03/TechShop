"use server";

import { submitContact, type ActionState } from "@techshop/bff";

/** The public contact form: becomes a support ticket. */
export async function contact(
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  return submitContact("wholesale", form);
}
