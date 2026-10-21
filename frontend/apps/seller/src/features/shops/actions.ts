"use server";

import { redirect } from "next/navigation";

import { failure, type ActionState } from "@techshop/bff";

import { bff } from "@/lib/bff";
import { text } from "@/lib/actions";

/** Opens a shop for the signed-in seller, then goes to its checklist. */
export async function openShop(
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  if (!(await bff.getSession({ refresh: true })))
    return { ok: false, message: "Your session ended. Sign in again." };
  const type = text(form, "type") === "business" ? "business" : "individual";
  let id: string;
  try {
    const api = await bff.api();
    const shop = await api.post("/api/v1/seller/shops", {
      body: {
        type,
        displayName: text(form, "displayName", 60),
        stateCode: text(form, "stateCode", 2),
        city: text(form, "city", 60),
      },
    });
    id = shop.id;
  } catch (e) {
    return failure(e);
  }
  redirect(`/centre/${id}`);
}
