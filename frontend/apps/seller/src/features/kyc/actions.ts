"use server";

import { refresh } from "next/cache";

import { failure, type ActionState } from "@techshop/bff";

import { bff } from "@/lib/bff";
import { text } from "@/lib/actions";

const ALLOWED = ["image/jpeg", "image/png", "application/pdf"];
const MAX_BYTES = 10 * 1024 * 1024; // the API's limit for kyc_document files

type Kind = "id_front" | "id_back" | "selfie" | "cac_certificate";
const LABELS: Record<Kind, string> = {
  id_front: "the front of your ID",
  id_back: "the back of your ID",
  selfie: "your selfie",
  cac_certificate: "your CAC certificate",
};

/**
 * Uploads the documents (server to storage, through presigned URLs) and submits the
 * verification. The browser never talks to the API or storage directly.
 */
export async function submitKyc(
  shopId: string,
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  if (!(await bff.getSession({ refresh: true })))
    return { ok: false, message: "Your session ended. Sign in again." };
  const business = text(form, "sellerType") === "business";
  const required: Kind[] = business
    ? ["id_front", "selfie", "cac_certificate"]
    : ["id_front", "selfie"];
  const errors: Record<string, string> = {};
  const files: { kind: Kind; file: File }[] = [];
  for (const kind of [
    "id_front",
    "id_back",
    "selfie",
    "cac_certificate",
  ] as Kind[]) {
    const f = form.get(kind);
    const file = f instanceof File && f.size > 0 ? f : null;
    if (!file) {
      if (required.includes(kind)) errors[kind] = `Add ${LABELS[kind]}.`;
      continue;
    }
    if (!ALLOWED.includes(file.type)) errors[kind] = "Use a JPG, PNG or PDF.";
    else if (file.size > MAX_BYTES)
      errors[kind] = "The file must be 10 MB or smaller.";
    else files.push({ kind, file });
  }
  const idNumber = text(form, "idNumber", 30);
  if (!idNumber) errors.idNumber = "Enter your ID number.";
  if (business && !text(form, "cacRcNumber", 20))
    errors.cacRcNumber = "Enter your CAC registration number.";
  if (Object.keys(errors).length)
    return { ok: false, message: "Check the highlighted fields.", errors };

  try {
    const api = await bff.api();
    const documents: { kind: Kind; fileId: string }[] = [];
    for (const { kind, file } of files) {
      const up = await api.post("/api/v1/files/uploads", {
        body: {
          purpose: "kyc_document",
          contentType: file.type,
          sizeBytes: file.size,
          name: file.name.slice(0, 200),
        },
      });
      const put = await fetch(up.uploadUrl, {
        method: "PUT",
        headers: { "Content-Type": file.type },
        body: file,
      });
      if (!put.ok)
        return {
          ok: false,
          message: `We couldn’t upload ${LABELS[kind]}. Try again.`,
        };
      await api.post("/api/v1/files/{id}/complete", {
        path: { id: up.fileId },
        body: {},
      });
      documents.push({ kind, fileId: up.fileId });
    }
    await api.post("/api/v1/seller/shops/{shopId}/kyc", {
      path: { shopId },
      body: {
        idType: text(form, "idType") as "nin",
        idNumber,
        cacRcNumber: text(form, "cacRcNumber", 20) || undefined,
        tin: text(form, "tin", 20) || undefined,
        documents,
      },
    });
  } catch (e) {
    return failure(e);
  }
  refresh();
  return {
    ok: true,
    message: "Thanks. We’ll review your verification and email you.",
  };
}
