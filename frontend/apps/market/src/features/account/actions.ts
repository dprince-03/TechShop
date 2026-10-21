"use server";

import { refresh } from "next/cache";

import { normaliseNigerianPhone } from "@techshop/api-client";
import { failure } from "@techshop/bff";

import { bff } from "@/lib/bff";

export type FormState = {
  ok?: boolean;
  message?: string;
  /** Field name → error, shown next to the field. */
  errors?: Record<string, string>;
};

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const text = (form: FormData, name: string, max = 120) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);

async function signedInApi() {
  const session = await bff.getSession({ refresh: true });
  if (!session) return null;
  return bff.api();
}

export async function updateProfile(
  _prev: FormState,
  form: FormData,
): Promise<FormState> {
  const firstName = text(form, "firstName", 60);
  const lastName = text(form, "lastName", 60);
  const errors: Record<string, string> = {};
  if (!firstName) errors.firstName = "Enter your first name.";
  if (!lastName) errors.lastName = "Enter your last name.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };
  const api = await signedInApi();
  if (!api) return { ok: false, message: "Your session ended. Sign in again." };
  try {
    await api.patch("/api/v1/me", {
      body: {
        firstName,
        lastName,
        marketingOptIn: form.get("marketingOptIn") === "on",
      },
    });
    refresh();
    return { ok: true, message: "Profile saved." };
  } catch (e) {
    return failure(e, "We couldn’t save your profile.");
  }
}

/** Creates an address, or updates one when the form carries an `id`. */
export async function saveAddress(
  _prev: FormState,
  form: FormData,
): Promise<FormState> {
  const id = text(form, "id", 36);
  const phone = normaliseNigerianPhone(text(form, "phone", 20));
  const body = {
    recipientName: text(form, "recipientName", 100),
    phone: phone ?? "",
    line1: text(form, "line1", 200),
    line2: text(form, "line2", 200) || undefined,
    city: text(form, "city", 100),
    stateCode: text(form, "stateCode", 2).toUpperCase(),
    lga: text(form, "lga", 100) || undefined,
    landmark: text(form, "landmark", 200) || undefined,
    label: text(form, "label", 40) || undefined,
    isDefault: form.get("isDefault") === "on",
  };
  const errors: Record<string, string> = {};
  if (!body.recipientName) errors.recipientName = "Enter the recipient’s name.";
  if (!phone)
    errors.phone = "Enter a Nigerian mobile number, e.g. 0803 123 4567.";
  if (!body.line1) errors.line1 = "Enter the street address.";
  if (!body.city) errors.city = "Enter the city or town.";
  if (!/^[A-Z]{2}$/.test(body.stateCode)) errors.stateCode = "Choose a state.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };
  const api = await signedInApi();
  if (!api) return { ok: false, message: "Your session ended. Sign in again." };
  try {
    if (id && UUID.test(id))
      await api.put("/api/v1/me/addresses/{id}", { path: { id }, body });
    else await api.post("/api/v1/me/addresses", { body });
    refresh();
    return { ok: true, message: id ? "Address updated." : "Address added." };
  } catch (e) {
    return failure(e, "We couldn’t save the address.");
  }
}

export async function deleteAddress(id: string): Promise<FormState> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown address." };
  const api = await signedInApi();
  if (!api) return { ok: false, message: "Your session ended. Sign in again." };
  try {
    await api.delete("/api/v1/me/addresses/{id}", { path: { id } });
    refresh();
    return { ok: true, message: "Address removed." };
  } catch (e) {
    return failure(e, "We couldn’t remove the address.");
  }
}

export async function revokeSession(id: string): Promise<FormState> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown device." };
  const api = await signedInApi();
  if (!api) return { ok: false, message: "Your session ended. Sign in again." };
  try {
    await api.delete("/api/v1/me/sessions/{id}", { path: { id } });
    refresh();
    return { ok: true, message: "Signed out of that device." };
  } catch (e) {
    return failure(e, "We couldn’t sign that device out.");
  }
}

export async function revokeOtherSessions(): Promise<FormState> {
  const api = await signedInApi();
  if (!api) return { ok: false, message: "Your session ended. Sign in again." };
  try {
    const r = await api.post("/api/v1/me/sessions/revoke-others");
    refresh();
    return {
      ok: true,
      message: `Signed out of ${r.revoked} other device${r.revoked === 1 ? "" : "s"}.`,
    };
  } catch (e) {
    return failure(e, "We couldn’t sign the other devices out.");
  }
}
