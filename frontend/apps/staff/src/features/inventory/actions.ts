"use server";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

type Condition = "new" | "uk_used" | "refurbished" | "open_box";
const CONDITIONS: Condition[] = ["new", "uk_used", "refurbished", "open_box"];

export type FormResult = StaffResult;

/** Adjusts stock (+ receive, − write off) with a reason; the API records a movement. */
export async function adjustStock(
  _prev: FormResult,
  form: FormData,
): Promise<FormResult> {
  const warehouseId = String(form.get("warehouseId") ?? "");
  const variantId = String(form.get("variantId") ?? "").trim();
  const delta = Number(form.get("delta") ?? 0);
  const reason = String(form.get("reason") ?? "adjustment") as
    "adjustment" | "write_off" | "purchase_receipt";
  const condition =
    CONDITIONS.find((c) => c === form.get("condition")) ?? "new";
  const note = String(form.get("note") ?? "")
    .trim()
    .slice(0, 300);
  if (!UUID.test(warehouseId) || !UUID.test(variantId))
    return { ok: false, message: "Choose a warehouse and paste a variant ID." };
  if (!Number.isInteger(delta) || delta === 0)
    return {
      ok: false,
      message: "Enter a whole number other than 0 (negative to remove).",
    };
  if (note.length < 3)
    return { ok: false, message: "Say why (it’s kept on the movement)." };
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/inventory/adjustments",
        { body: { warehouseId, variantId, delta, reason, condition, note } },
      ),
    `Stock ${delta > 0 ? "added" : "removed"}.`,
  );
}

/** Receives serialised devices: one IMEI or serial per line; each becomes a sellable unit. */
export async function receiveUnits(
  _prev: FormResult,
  form: FormData,
): Promise<FormResult> {
  const warehouseId = String(form.get("warehouseId") ?? "");
  const variantId = String(form.get("variantId") ?? "").trim();
  const condition =
    CONDITIONS.find((c) => c === form.get("condition")) ?? "new";
  const lines = String(form.get("codes") ?? "")
    .split(/\n+/)
    .map((l) => l.trim())
    .filter(Boolean);
  if (!UUID.test(warehouseId) || !UUID.test(variantId))
    return { ok: false, message: "Choose a warehouse and paste a variant ID." };
  if (!lines.length || lines.length > 500)
    return {
      ok: false,
      message: "Scan between 1 and 500 IMEIs or serial numbers.",
    };
  // 15-digit numbers are IMEIs; anything else is a serial number.
  const units = lines.map((c) =>
    /^\d{15}$/.test(c) ? { imei: c } : { serial: c },
  );
  return runStaffAction(
    async () =>
      (await bff.api({ refresh: true })).post(
        "/api/v1/staff/inventory/device-units",
        { body: { warehouseId, variantId, condition, units } },
      ),
    `${units.length} unit${units.length === 1 ? "" : "s"} received.`,
  );
}
