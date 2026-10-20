"use server";

import { refresh } from "next/cache";

import { bff } from "@/lib/bff";

export type SaveResult =
  | { ok: true; saved: boolean }
  | { ok: false; signIn?: boolean; message: string };

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/** Saves or un-saves a product for the signed-in customer. */
export async function toggleSaved(productId: string): Promise<SaveResult> {
  if (!UUID.test(productId)) return { ok: false, message: "Unknown product." };
  const session = await bff.getSession({ refresh: true });
  if (!session)
    return { ok: false, signIn: true, message: "Sign in to save items." };
  try {
    const api = await bff.api();
    const r = await api.post("/api/v1/me/saved/{productId}", {
      path: { productId },
    });
    refresh();
    return { ok: true, saved: r.saved };
  } catch {
    return { ok: false, message: "We couldn’t update your saved items." };
  }
}
