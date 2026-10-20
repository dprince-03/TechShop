import "server-only";

import { bff } from "@/lib/bff";

/** Payment methods the API accepts right now (empty when unavailable). */
export async function paymentMethods(): Promise<string[]> {
  try {
    const api = await bff.api();
    return (
      (await api.get("/api/v1/checkout/payment-methods", {
        init: { cache: "no-store" },
      })) ?? []
    );
  } catch {
    return [];
  }
}

/** First quote for the default (or first) address, so the page opens with a real total. */
export async function initialQuote(addressId: string | undefined) {
  if (!addressId) return null;
  try {
    const api = await bff.api();
    return await api.post("/api/v1/checkout/preview", {
      body: { addressId },
      init: { cache: "no-store" },
    });
  } catch {
    return null;
  }
}
