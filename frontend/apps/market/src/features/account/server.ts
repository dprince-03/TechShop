import "server-only";

import { bff } from "@/lib/bff";

const NO_STORE = { cache: "no-store" } as const;

/** Everything the account page shows, loaded in parallel for the signed-in customer. */
export async function loadAccount() {
  const api = await bff.api();
  const [me, addresses, sessions] = await Promise.all([
    api.get("/api/v1/me", { init: NO_STORE }),
    api.get("/api/v1/me/addresses", { init: NO_STORE }),
    api.get("/api/v1/me/sessions", { init: NO_STORE }),
  ]);
  return { me, addresses: addresses ?? [], sessions: sessions ?? [] };
}

export async function listAddresses() {
  const api = await bff.api();
  return (await api.get("/api/v1/me/addresses", { init: NO_STORE })) ?? [];
}
