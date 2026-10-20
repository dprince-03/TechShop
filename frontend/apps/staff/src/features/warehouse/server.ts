import "server-only";

import { ApiError } from "@techshop/api-client";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

export async function warehouses() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/inventory/warehouses", { init: NO_STORE })) ?? []
  );
}

export async function pickLists(warehouseId: string) {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/wms/pick-lists", {
      query: { warehouseId },
      init: NO_STORE,
    })) ?? []
  );
}

export async function pickList(id: string) {
  try {
    return await (
      await bff.api()
    ).get("/api/v1/staff/wms/pick-lists/{id}", {
      path: { id },
      init: NO_STORE,
    });
  } catch (e) {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      return null;
    throw e;
  }
}

export async function manifests(warehouseId: string) {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/wms/manifests", {
      query: { warehouseId },
      init: NO_STORE,
    })) ?? []
  );
}

export async function manifest(id: string) {
  try {
    return await (
      await bff.api()
    ).get("/api/v1/staff/wms/manifests/{id}", { path: { id }, init: NO_STORE });
  } catch (e) {
    if (e instanceof ApiError && (e.status === 404 || e.status === 422))
      return null;
    throw e;
  }
}

export async function carriers() {
  return (
    (await (
      await bff.api()
    ).get("/api/v1/staff/wms/carriers", { init: NO_STORE })) ?? []
  );
}
