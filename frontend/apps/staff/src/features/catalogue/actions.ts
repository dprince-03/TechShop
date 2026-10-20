"use server";

import { redirect } from "next/navigation";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

type Decision = "approved" | "changes_requested" | "rejected" | "suspended";
const DECISIONS: Decision[] = [
  "approved",
  "changes_requested",
  "rejected",
  "suspended",
];

export async function moderateListing(
  id: string,
  decision: string,
  note: string,
): Promise<StaffResult> {
  const d = DECISIONS.find((x) => x === decision);
  if (!UUID.test(id) || !d)
    return { ok: false, message: "Unknown listing or decision." };
  if (d !== "approved" && note.trim().length < 3)
    return { ok: false, message: "Tell the seller why." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/catalog/listings/{id}/decision",
        {
          path: { id },
          body: { decision: d, note: note.trim().slice(0, 500) || undefined },
        },
      ),
    d === "approved"
      ? "Listing approved and live."
      : "Decision sent to the seller.",
  );
}

export async function decideSuggestion(
  id: string,
  decision: "accepted" | "rejected",
): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown suggestion." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/catalog/suggestions/{id}/decision",
        { path: { id }, body: { decision } },
      ),
    decision === "accepted" ? "Suggestion applied." : "Suggestion rejected.",
  );
}

export async function reindex(): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post("/api/v1/staff/catalog/reindex"),
    "Search index rebuilt.",
  );
}

export type CreateState = {
  ok?: boolean;
  message?: string;
  errors?: Record<string, string>;
};

const slugify = (s: string) =>
  s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "")
    .slice(0, 80);

const naira = (v: string) => {
  const n = Number(v.replace(/[,\s₦]/g, ""));
  return Number.isFinite(n) && n > 0 ? Math.round(n * 100) : undefined;
};

/**
 * Creates a product with one variant and, when a price is given, TechShop's own listing for it.
 * Two calls: if the listing fails, the product still exists as a draft and the message says so.
 */
export async function createProduct(
  _prev: CreateState,
  form: FormData,
): Promise<CreateState> {
  const get = (k: string, max = 200) =>
    String(form.get(k) ?? "")
      .trim()
      .slice(0, max);
  const name = get("name", 150);
  const slug = slugify(get("slug") || name);
  const categorySlug = get("categorySlug", 80);
  const sku = get("sku", 60).toUpperCase();
  const variantName = get("variantName", 100) || "Standard";
  const weight = Number(get("weightGrams", 6));
  const price = naira(get("price", 20));
  const condition = get("condition", 20) || "new";
  const warrantyMonths = Number(get("warrantyMonths", 3) || 0);
  const status = get("status", 10) === "active" ? "active" : "draft";

  const errors: Record<string, string> = {};
  if (name.length < 2) errors.name = "Enter the product name.";
  if (!categorySlug) errors.categorySlug = "Choose a category.";
  if (sku.length < 2) errors.sku = "Enter a SKU (at least 2 characters).";
  if (get("weightGrams") && (!Number.isInteger(weight) || weight <= 0))
    errors.weightGrams = "Weight in whole grams.";
  if (get("price") && !price) errors.price = "Enter a price in naira.";
  if (
    !Number.isInteger(warrantyMonths) ||
    warrantyMonths < 0 ||
    warrantyMonths > 60
  )
    errors.warrantyMonths = "0 to 60 months.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };

  const api = await bff.api({ refresh: true });
  let productSlug: string;
  try {
    const created = await api.post("/api/v1/staff/catalog/products", {
      body: {
        name,
        slug,
        categorySlug,
        brandName: get("brandName", 80) || undefined,
        model: get("model", 80) || undefined,
        description: get("description", 4000) || undefined,
        status,
        variants: [
          {
            sku,
            name: variantName,
            weightGrams: weight || undefined,
            isSerialised: form.get("isSerialised") === "on",
          },
        ],
      },
    });
    productSlug = created.product.slug;
    if (price) {
      await api.post("/api/v1/staff/catalog/listings", {
        body: {
          variantId: created.variants![0].id,
          price,
          condition: condition as
            "new" | "uk_used" | "refurbished" | "open_box",
          fulfilledBy: "techshop",
          warrantyMonths,
          warrantyProvider: warrantyMonths > 0 ? "techshop" : "none",
        },
      });
    }
  } catch (e) {
    if (e instanceof ApiError) {
      const fieldErrors: Record<string, string> = {};
      for (const f of e.fieldErrors) {
        const k = (f.location ?? "")
          .replace(/^body\./, "")
          .replace(/^variants\[0\]\./, "");
        if (k) fieldErrors[k] = f.message ?? "Check this field.";
      }
      if (e.status === 409)
        return {
          ok: false,
          message: "A product with that web address or SKU already exists.",
          errors: { slug: "Already used." },
        };
      return {
        ok: false,
        message: e.detail ?? "We couldn’t create the product.",
        errors: fieldErrors,
      };
    }
    return { ok: false, message: "We couldn’t create the product." };
  }
  redirect(`/catalogue?created=${encodeURIComponent(productSlug)}`);
}
