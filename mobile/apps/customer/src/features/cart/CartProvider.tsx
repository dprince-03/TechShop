import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";

import { ApiError, type Schema } from "@techshop/api-client";

import { useSession } from "@/features/auth/SessionProvider";
import { cartTokenStore } from "@/lib/storage";

type Cart = Schema<"CartView">;
type Result = string | null;

type Ctx = {
  cart: Cart | null;
  count: number;
  loading: boolean;
  add: (listingId: string, quantity?: number) => Promise<Result>;
  setQuantity: (itemId: string, quantity: number) => Promise<Result>;
  remove: (itemId: string) => Promise<Result>;
  reload: () => Promise<void>;
};

const CartCtx = createContext<Ctx | null>(null);

const failure = (e: unknown, fallback: string): string => {
  if (e instanceof ApiError) {
    if (e.isUnavailable) return "No connection. Try again when you have signal.";
    if (e.code === "out_of_stock" || e.code === "quantity_too_large") return "Sorry, there isn’t enough stock for that quantity.";
    if (e.code === "unavailable") return "This item is no longer available.";
  }
  return fallback;
};

/**
 * The cart for the whole app. Signed in: the customer's cart (bearer token). Guest: the cart the
 * API keyed by X-Cart-Token. On sign-in the guest cart merges into the account.
 */
export function CartProvider({ children }: { children: ReactNode }) {
  const { session, status } = useSession();
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(false);

  const caller = useCallback(async () => {
    const token = status === "signed-in" ? null : await cartTokenStore.get();
    return { headers: token ? { "X-Cart-Token": token } : undefined, guestWithoutCart: status !== "signed-in" && !token };
  }, [status]);

  const reload = useCallback(async () => {
    const c = await caller();
    if (c.guestWithoutCart) return setCart(null);
    setLoading(true);
    try {
      setCart(await session.api.get("/api/v1/cart", { headers: c.headers }));
    } catch {
      setCart(null);
    } finally {
      setLoading(false);
    }
  }, [caller, session]);

  // Merge the guest cart once the customer signs in, then show the account's cart.
  useEffect(() => {
    const t = setTimeout(async () => {
      if (status === "signed-in") {
        const guest = await cartTokenStore.get();
        if (guest) {
          await session.api.post("/api/v1/cart/merge", { body: { cartToken: guest } }).catch(() => undefined);
          await cartTokenStore.clear();
        }
      }
      if (status !== "restoring") await reload();
    }, 0);
    return () => clearTimeout(t);
  }, [status, session, reload]);

  const value: Ctx = {
    cart,
    count: (cart?.lines ?? []).reduce((n, l) => n + l.quantity, 0),
    loading,
    reload,
    async add(listingId, quantity = 1) {
      try {
        const c = await caller();
        const r = await session.api.request.post("/api/v1/cart/items", { headers: c.headers, body: { listingId, quantity } });
        const token = r.headers.get("x-cart-token") ?? r.data.cartToken;
        if (status !== "signed-in" && token) await cartTokenStore.set(token);
        setCart(r.data);
        return null;
      } catch (e) {
        return failure(e, "We couldn’t add that to your cart.");
      }
    },
    async setQuantity(itemId, quantity) {
      if (!Number.isInteger(quantity) || quantity < 1 || quantity > 99) return "Choose a quantity between 1 and 99.";
      try {
        setCart(await session.api.patch("/api/v1/cart/items/{id}", { headers: (await caller()).headers, path: { id: itemId }, body: { quantity } }));
        return null;
      } catch (e) {
        return failure(e, "We couldn’t change the quantity.");
      }
    },
    async remove(itemId) {
      try {
        setCart(await session.api.delete("/api/v1/cart/items/{id}", { headers: (await caller()).headers, path: { id: itemId } }));
        return null;
      } catch (e) {
        return failure(e, "We couldn’t remove that item.");
      }
    },
  };

  return <CartCtx.Provider value={value}>{children}</CartCtx.Provider>;
}

export function useCart(): Ctx {
  const v = useContext(CartCtx);
  if (!v) throw new Error("useCart must be used inside CartProvider");
  return v;
}
