import "server-only";

import { bff } from "@/lib/bff";

/** IDs of the signed-in customer's saved products (empty when signed out or unavailable). */
export async function savedProductIds(): Promise<Set<string>> {
  if (!(await bff.getSession())) return new Set();
  try {
    const api = await bff.api();
    const rows =
      (await api.get("/api/v1/me/saved", { init: { cache: "no-store" } })) ??
      [];
    return new Set(rows.map((r) => r.id));
  } catch {
    return new Set();
  }
}

/** The signed-in customer's saved products, newest first. */
export async function listSaved() {
  const api = await bff.api();
  return (
    (await api.get("/api/v1/me/saved", { init: { cache: "no-store" } })) ?? []
  );
}
