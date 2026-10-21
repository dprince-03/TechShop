"use server";

import { redirect } from "next/navigation";

import { failure, type ActionState } from "@techshop/bff";

import { act, nairaToKobo, text } from "@/lib/actions";
import { bff } from "@/lib/bff";

type Condition = "new" | "uk_used" | "refurbished" | "open_box";
type Grade = "" | "A" | "B" | "C";
type Warranty = "" | "manufacturer" | "techshop" | "seller" | "none";

const int = (form: FormData, name: string) => {
  const v = text(form, name, 10);
  return v === "" ? undefined : Number.parseInt(v, 10);
};

/** Creates a listing; the automatic checks decide whether it goes live or to review. */
export async function createListing(
  shopId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const price = nairaToKobo(text(form, "price", 20));
  const stock = int(form, "sellerStock");
  const condition = text(form, "condition") as Condition;
  const errors: Record<string, string> = {};
  if (!text(form, "variantId")) errors.variantId = "Choose the exact model.";
  if (price === null)
    errors.price = "Enter a price in naira, for example 250000.";
  if (stock === undefined || Number.isNaN(stock) || stock < 0)
    errors.sellerStock = "Enter how many you have (0 or more).";
  if (condition !== "new" && !text(form, "conditionNotes"))
    errors.conditionNotes = "Describe the condition.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };
  if (!(await bff.getSession({ refresh: true })))
    return { ok: false, message: "Your session ended. Sign in again." };
  let id: string;
  try {
    const api = await bff.api();
    const r = await api.post("/api/v1/seller/shops/{shopId}/listings", {
      path: { shopId },
      body: {
        variantId: text(form, "variantId"),
        condition,
        grade: (text(form, "grade") as Grade) || undefined,
        price: price!,
        sellerStock: stock!,
        handlingDays: int(form, "handlingDays"),
        warrantyProvider:
          (text(form, "warrantyProvider") as Warranty) || undefined,
        warrantyMonths: int(form, "warrantyMonths"),
        conditionNotes: text(form, "conditionNotes", 500) || undefined,
        batteryHealthPct: int(form, "batteryHealthPct"),
      },
    });
    id = r.listing.id;
  } catch (e) {
    return failure(e);
  }
  redirect(`/centre/${shopId}/listings/${id}`);
}

export async function setPrice(
  shopId: string,
  listingId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const price = nairaToKobo(text(form, "price", 20));
  const was = text(form, "compareAt", 20);
  const compareAt = was ? nairaToKobo(was) : undefined;
  if (price === null)
    return {
      ok: false,
      message: "Check the highlighted fields.",
      errors: { price: "Enter a price in naira." },
    };
  if (compareAt === null)
    return {
      ok: false,
      message: "Check the highlighted fields.",
      errors: {
        compareAt: "Enter the previous price in naira, or leave it empty.",
      },
    };
  return act(
    (api) =>
      api.patch("/api/v1/seller/shops/{shopId}/listings/{listingId}/price", {
        path: { shopId, listingId },
        body: { price, compareAt: compareAt ?? undefined },
      }),
    "Price saved.",
  );
}

export async function setStock(
  shopId: string,
  listingId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const stock = int(form, "stock");
  if (stock === undefined || Number.isNaN(stock) || stock < 0)
    return {
      ok: false,
      message: "Check the highlighted fields.",
      errors: { stock: "Enter 0 or more." },
    };
  return act(
    (api) =>
      api.put("/api/v1/seller/shops/{shopId}/listings/{listingId}/stock", {
        path: { shopId, listingId },
        body: { stock },
      }),
    "Stock saved.",
  );
}

export async function pauseListing(
  shopId: string,
  listingId: string,
  paused: boolean,
): Promise<ActionState> {
  return act(
    (api) =>
      paused
        ? api.post("/api/v1/seller/shops/{shopId}/listings/{listingId}/pause", {
            path: { shopId, listingId },
          })
        : api.post(
            "/api/v1/seller/shops/{shopId}/listings/{listingId}/resume",
            { path: { shopId, listingId } },
          ),
    paused ? "Taken off sale." : "Back on sale.",
  );
}

export async function resubmitListing(
  shopId: string,
  listingId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  return act(
    (api) =>
      api.put("/api/v1/seller/shops/{shopId}/listings/{listingId}", {
        path: { shopId, listingId },
        body: {
          grade: (text(form, "grade") as Grade) || undefined,
          conditionNotes: text(form, "conditionNotes", 500) || undefined,
          batteryHealthPct: int(form, "batteryHealthPct"),
          warrantyProvider:
            (text(form, "warrantyProvider") as Exclude<Warranty, "">) ||
            "manufacturer",
          warrantyMonths: int(form, "warrantyMonths") ?? 0,
          handlingDays: int(form, "handlingDays") ?? 1,
        },
      }),
    "Resubmitted.",
  );
}
