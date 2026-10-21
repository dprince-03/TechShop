"use server";

import { redirect } from "next/navigation";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";
import { runStaffAction, UUID, type StaffResult } from "@/lib/actions";

const api = () => bff.api({ refresh: true });

/** Releases a wave: the API groups paid TechShop lines in this warehouse into a pick list. */
export async function releaseWave(warehouseId: string): Promise<StaffResult> {
  if (!UUID.test(warehouseId))
    return { ok: false, message: "Choose a warehouse." };
  let id: string | undefined;
  const r = await runStaffAction(async () => {
    id = (
      await (
        await api()
      ).post("/api/v1/staff/wms/waves", { body: { warehouseId } })
    ).id;
  }, "Wave released.");
  if (r.ok && id) redirect(`/warehouse/pick/${id}`);
  if (!r.ok && /nothing|no lines|empty/i.test(r.message))
    return {
      ok: false,
      message: "Nothing to pick: no paid orders are waiting in this warehouse.",
    };
  return r;
}

export async function claimPickList(id: string): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown pick list." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/wms/pick-lists/{id}/claim", {
        path: { id },
      }),
    "You’re picking this list.",
  );
}

export type ScanState = { ok?: boolean; message?: string };

/** Records a picked line; serialised devices need the scanned IMEI/serial of a unit in stock. */
export async function scanItem(
  pickListId: string,
  _prev: ScanState,
  form: FormData,
): Promise<ScanState> {
  const orderLineId = String(form.get("orderLineId") ?? "");
  const quantity = Number(form.get("quantity") ?? 1);
  const serial = String(form.get("serial") ?? "").trim() || undefined;
  if (!UUID.test(pickListId) || !UUID.test(orderLineId))
    return { ok: false, message: "Unknown item." };
  if (!Number.isInteger(quantity) || quantity < 1)
    return { ok: false, message: "Quantity must be at least 1." };
  const r = await runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/wms/pick-lists/{id}/scan", {
        path: { id: pickListId },
        body: { orderLineId, quantity, serial },
      }),
    "Picked.",
  );
  return { ok: r.ok, message: r.message };
}

export type PackState = {
  ok?: boolean;
  message?: string;
  label?: string;
  handoff?: string;
};

/** Packs a fulfilment: weight is checked against the catalogue; a rider job or carrier booking follows. */
export async function packFulfilment(
  _prev: PackState,
  form: FormData,
): Promise<PackState> {
  const fulfilmentId = String(form.get("fulfilmentId") ?? "");
  const weightGrams = Number(form.get("weightGrams") ?? 0);
  if (!UUID.test(fulfilmentId))
    return { ok: false, message: "Unknown parcel." };
  if (!Number.isInteger(weightGrams) || weightGrams <= 0)
    return { ok: false, message: "Weigh the parcel (grams)." };
  try {
    const p = await (
      await api()
    ).post("/api/v1/staff/wms/pack", {
      body: {
        fulfilmentId,
        weightGrams,
        boxCode: String(form.get("boxCode") ?? "").trim() || undefined,
      },
    });
    return {
      ok: true,
      label: p.parcel.labelCode,
      handoff: p.handoff,
      message:
        p.handoff === "rider"
          ? `Packed. Label ${p.parcel.labelCode}: a rider job was created.`
          : `Packed. Label ${p.parcel.labelCode}: booked with the carrier.`,
    };
  } catch (e) {
    return {
      ok: false,
      message:
        e instanceof ApiError && e.detail
          ? e.detail
          : "We couldn’t pack it. Check the items are all picked.",
    };
  }
}

export type ManifestState = { ok?: boolean; message?: string };

export async function createManifest(
  warehouseId: string,
  _prev: ManifestState,
  form: FormData,
): Promise<ManifestState> {
  const labels = String(form.get("labels") ?? "")
    .split(/[\s,]+/)
    .map((l) => l.trim().toUpperCase())
    .filter(Boolean);
  const target = String(form.get("target") ?? "");
  const [kind, id] = target.split(":");
  if (!UUID.test(warehouseId))
    return { ok: false, message: "Choose a warehouse." };
  if (!labels.length)
    return { ok: false, message: "Scan or type at least one parcel label." };
  if (!UUID.test(id ?? "") || (kind !== "rider" && kind !== "carrier"))
    return { ok: false, message: "Choose who collects the parcels." };
  let manifestId: string | undefined;
  const r = await runStaffAction(async () => {
    manifestId = (
      await (
        await api()
      ).post("/api/v1/staff/wms/manifests", {
        body: {
          warehouseId,
          labels,
          riderId: kind === "rider" ? id : undefined,
          carrierId: kind === "carrier" ? id : undefined,
        },
      })
    ).id;
  }, "Manifest created.");
  if (r.ok && manifestId) redirect(`/warehouse/manifests/${manifestId}`);
  return { ok: r.ok, message: r.message };
}

export async function signManifest(
  id: string,
  name: string,
): Promise<StaffResult> {
  if (!UUID.test(id)) return { ok: false, message: "Unknown manifest." };
  if (name.trim().length < 2)
    return { ok: false, message: "Type the name of the person collecting." };
  return runStaffAction(
    async () =>
      (await api()).post("/api/v1/staff/wms/manifests/{id}/sign", {
        path: { id },
        body: { signedByName: name.trim().slice(0, 100) },
      }),
    "Signed. The parcels have left the warehouse.",
  );
}

export type DropoffState = { ok?: boolean; message?: string };

/** Receives a seller's packed parcel at this hub: the label is scanned, then a rider job or carrier booking follows. */
export async function receiveDropoff(
  warehouseId: string,
  _prev: DropoffState,
  form: FormData,
): Promise<DropoffState> {
  const label = String(form.get("label") ?? "")
    .trim()
    .toUpperCase();
  if (!UUID.test(warehouseId))
    return { ok: false, message: "Choose a warehouse." };
  if (!/^TSS-[0-9A-Z]{4,}$/.test(label))
    return {
      ok: false,
      message: "Scan or type the seller label (it starts with TSS-).",
    };
  try {
    const d = await (
      await api()
    ).post("/api/v1/staff/wms/dropoffs", { body: { warehouseId, label } });
    return {
      ok: true,
      message:
        d.handoff === "rider"
          ? `Received ${label} for order ${d.orderNumber}: a rider job was created.`
          : `Received ${label} for order ${d.orderNumber}: booked with the carrier.`,
    };
  } catch (e) {
    return {
      ok: false,
      message:
        e instanceof ApiError && e.status === 404
          ? `No parcel has the label ${label}. Check the code with the seller.`
          : e instanceof ApiError && e.detail
            ? e.detail
            : "We couldn’t receive it. Try again.",
    };
  }
}
