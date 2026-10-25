"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });
const text = (form: FormData, name: string, max = 200) =>
  String(form.get(name) ?? "")
    .trim()
    .slice(0, max);
const bad = (message: string): StaffResult => ({ ok: false, message });
const kobo = (form: FormData, name: string) => {
  const n = Number.parseFloat(text(form, name, 20).replace(/,/g, ""));
  return Number.isFinite(n) && n >= 0 ? Math.round(n * 100) : null; // typed in naira
};

// ---- Returns

export async function decideReturn(
  rma: string,
  decision: "approve" | "reject",
  reason: string,
): Promise<StaffResult> {
  if (decision === "reject" && reason.trim().length < 3)
    return bad("Tell the customer why (they'll see it).");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/returns/{rma}/decision", {
        path: { rma },
        body: { decision, reason: reason.trim().slice(0, 500) || undefined },
      }),
    decision === "approve"
      ? "Approved. The customer has been told how to send it back."
      : "Rejected. The customer sees your reason.",
  );
}

export async function receiveReturn(rma: string): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/returns/{rma}/receive", {
        path: { rma },
      }),
    "Received. Inspect each item next.",
  );
}

const GRADES = ["A", "B", "C", "faulty"] as const;
const DISPOSITIONS = [
  "restock_new",
  "restock_open_box",
  "repair",
  "return_to_supplier",
  "return_to_seller",
  "write_off",
] as const;

export async function inspectItem(
  rma: string,
  itemId: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const grade = GRADES.find((g) => g === text(form, "grade"));
  const disposition = DISPOSITIONS.find((d) => d === text(form, "disposition"));
  const warehouseId = text(form, "warehouseId", 40);
  if (!UUID.test(itemId) || !grade || !disposition || !UUID.test(warehouseId))
    return bad("Choose the grade, where it goes and the location.");
  return runStaffAction(
    async () =>
      (await api()).post(
        "/api/v1/staff/returns/{rma}/items/{itemId}/inspection",
        {
          path: { rma, itemId },
          body: {
            grade,
            disposition,
            warehouseId,
            scannedSerial:
              text(form, "scannedSerial", 40).replace(/\s+/g, "") || undefined,
            notes: text(form, "notes", 500) || undefined,
          },
        },
      ),
    "Inspection recorded.",
  );
}

export async function refundReturn(rma: string): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/returns/{rma}/refund", {
        path: { rma },
      }),
    "Refund raised. Finance approves it next.",
  );
}

export async function decideDispute(
  rma: string,
  decision: "upheld" | "rejected",
  note: string,
): Promise<StaffResult> {
  if (note.trim().length < 3)
    return bad("Explain the decision; the seller sees it.");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/returns/{rma}/dispute-decision", {
        path: { rma },
        body: { decision, note: note.trim().slice(0, 1000) },
      }),
    decision === "upheld"
      ? "Upheld. The seller is credited and TechShop bears the cost."
      : "Dispute rejected.",
  );
}

// ---- Warranty and repairs

export async function walkInClaim(
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const imeiOrSerial = text(form, "imeiOrSerial", 40).replace(/\s+/g, "");
  const storeId = text(form, "storeId", 40);
  const fault = text(form, "fault", 1000);
  if (imeiOrSerial.length < 3 || !UUID.test(storeId) || fault.length < 5)
    return bad(
      "Scan the IMEI or serial, choose the store and describe the fault.",
    );
  const lineId = text(form, "orderLineId", 40);
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/warranty/claims", {
        body: {
          imeiOrSerial,
          storeId,
          fault,
          orderLineId: UUID.test(lineId) ? lineId : undefined,
        },
      }),
    "Claim taken and checked in.",
  );
}

export async function receiveClaim(
  number: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const storeId = text(form, "storeId", 40);
  const scannedSerial = text(form, "scannedSerial", 40).replace(/\s+/g, "");
  if (!UUID.test(storeId) || scannedSerial.length < 3)
    return bad("Choose the store and scan the device's IMEI or serial.");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/warranty/claims/{number}/receive", {
        path: { number },
        body: { storeId, scannedSerial },
      }),
    "Checked in. The customer has been told.",
  );
}

export async function rejectClaim(
  number: string,
  reason: string,
): Promise<StaffResult> {
  if (reason.trim().length < 3)
    return bad("Tell the customer why it isn't covered.");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/warranty/claims/{number}/reject", {
        path: { number },
        body: { reason: reason.trim().slice(0, 500) },
      }),
    "Marked not covered.",
  );
}

export async function startRepair(
  number: string,
  diagnosis: string,
): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/warranty/claims/{number}/repairs", {
        path: { number },
        body: { diagnosis: diagnosis.trim().slice(0, 1000) || undefined },
      }),
    "Sent to the workshop.",
  );
}

const REPAIR = [
  "diagnosing",
  "awaiting_parts",
  "repairing",
  "done",
  "unrepairable",
] as const;

export async function updateRepair(
  jobId: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const status = REPAIR.find((s) => s === text(form, "status"));
  if (!UUID.test(jobId) || !status) return bad("Choose the job's new status.");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/warranty/repairs/{id}", {
        path: { id: jobId },
        body: {
          status,
          diagnosis: text(form, "diagnosis", 1000) || undefined,
          notes: text(form, "notes", 1000) || undefined,
        },
      }),
    "Repair updated.",
  );
}

export async function resolveClaim(
  number: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const resolution =
    text(form, "resolution") === "refunded" ? "refunded" : "replaced";
  const warehouseId = text(form, "warehouseId", 40);
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/warranty/claims/{number}/resolution", {
        path: { number },
        body: {
          resolution,
          replacementSerial:
            text(form, "replacementSerial", 40).replace(/\s+/g, "") ||
            undefined,
          warehouseId: UUID.test(warehouseId) ? warehouseId : undefined,
          note: text(form, "note", 500) || undefined,
        },
      }),
    resolution === "refunded"
      ? "Refund raised; finance approves it."
      : "Replacement ready to collect.",
  );
}

export async function collectClaim(number: string): Promise<StaffResult> {
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/warranty/claims/{number}/collect", {
        path: { number },
      }),
    "Collected.",
  );
}

// ---- Trade-ins

export async function estimateTradeIn(
  ref: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const estimate = kobo(form, "estimate");
  const variantId = text(form, "variantId", 40);
  if (estimate === null) return bad("Enter the estimate in naira.");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/trade-ins/{ref}/estimate", {
        path: { ref },
        body: {
          estimate,
          variantId: UUID.test(variantId) ? variantId : undefined,
        },
      }),
    "Estimate sent to the customer.",
  );
}

export async function receiveTradeIn(
  ref: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const storeId = text(form, "storeId", 40);
  if (!UUID.test(storeId)) return bad("Choose the store.");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/trade-ins/{ref}/receive", {
        path: { ref },
        body: {
          storeId,
          imei: text(form, "imei", 20).replace(/\s+/g, "") || undefined,
          serial: text(form, "serial", 40) || undefined,
        },
      }),
    "Device taken in. Inspect it next.",
  );
}

const CONDITIONS = ["like_new", "good", "fair", "faulty"] as const;

export async function inspectTradeIn(
  ref: string,
  _prev: unknown,
  form: FormData,
): Promise<StaffResult> {
  const condition = CONDITIONS.find((c) => c === text(form, "condition"));
  if (!condition) return bad("Choose the grade.");
  const variantId = text(form, "variantId", 40);
  const offer = text(form, "offer") ? kobo(form, "offer") : undefined;
  if (offer === null)
    return bad(
      "Enter the offer in naira, or leave it empty to price from the catalogue.",
    );
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/trade-ins/{ref}/inspection", {
        path: { ref },
        body: {
          condition,
          variantId: UUID.test(variantId) ? variantId : undefined,
          offer,
          notes: text(form, "notes", 1000) || undefined,
        },
      }),
    "Offer sent. The customer has 7 days to answer.",
  );
}

/** Finance records the bank transfer (fresh authenticator code; not whoever priced it). */
export async function payTradeIn(
  ref: string,
  reference: string,
): Promise<StaffResult> {
  if (reference.trim().length < 3)
    return bad("Enter the bank transfer reference.");
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/trade-ins/{ref}/paid", {
        path: { ref },
        body: { reference: reference.trim().slice(0, 80) },
      }),
    "Recorded as paid. The customer has been told.",
  );
}
