import "server-only";

import { refresh } from "next/cache";

import { failure, type ActionState } from "@techshop/bff";

import { bff } from "@/lib/bff";

type Api = Awaited<ReturnType<typeof bff.api>>;

/**
 * Runs a Seller Centre change for a Server Action: needs a session, turns API failures into
 * form messages, and refreshes the page data on success.
 */
export async function act(
  run: (api: Api) => Promise<unknown>,
  success: string,
): Promise<ActionState> {
  if (!(await bff.getSession({ refresh: true })))
    return { ok: false, message: "Your session ended. Sign in again." };
  try {
    await run(await bff.api());
  } catch (e) {
    return failure(e);
  }
  refresh();
  return { ok: true, message: success };
}

/** A trimmed text field, capped. */
export const text = (form: FormData, name: string, max = 200) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);

/** Naira typed by a person ("1,250,000" or "1250000.50") → kobo, or null if it isn't a positive amount. */
export function nairaToKobo(input: string): number | null {
  const clean = input.replace(/[₦,\s]/g, "");
  if (!/^\d+(\.\d{1,2})?$/.test(clean)) return null;
  const kobo = Math.round(Number(clean) * 100);
  return kobo > 0 ? kobo : null;
}
