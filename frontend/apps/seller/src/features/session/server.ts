import "server-only";

import { notFound, redirect } from "next/navigation";

import { signInPath } from "@techshop/bff";

import { bff } from "@/lib/bff";

/** The signed-in seller's API client, or a redirect to sign in (then back to `here`). */
export async function sellerApi(here: string) {
  const session = await bff.getSession();
  if (!session) redirect(signInPath(here));
  return bff.api();
}

export type Shop = Awaited<ReturnType<typeof myShops>>[number];

/** Shops the seller belongs to, with their role. */
export async function myShops(here: string) {
  const api = await sellerApi(here);
  return (
    (await api.get("/api/v1/seller/shops", { init: { cache: "no-store" } })) ??
    []
  );
}

/** One of the seller's shops by ID, or 404 (other shops' existence isn't revealed). */
export async function requireShop(shopId: string, here: string) {
  const shops = await myShops(here);
  const shop = shops.find((s) => s.id === shopId);
  if (!shop) notFound();
  return shop;
}
