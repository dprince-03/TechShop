import * as Crypto from "expo-crypto";
import * as SecureStore from "expo-secure-store";

import type { TokenStore } from "@techshop/mobile-auth";

const REFRESH_KEY = "ts.customer.refreshToken";
const DEVICE_KEY = "ts.customer.deviceId";

/** Refresh token in the Keychain / Keystore, readable only while the phone is unlocked. */
export const secureTokenStore: TokenStore = {
  get: () => SecureStore.getItemAsync(REFRESH_KEY),
  set: (v) => SecureStore.setItemAsync(REFRESH_KEY, v, { keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY }),
  clear: () => SecureStore.deleteItemAsync(REFRESH_KEY),
};

/** A random ID for this install (shown in the customer's signed-in devices list). */
export async function deviceId(): Promise<string> {
  const existing = await SecureStore.getItemAsync(DEVICE_KEY);
  if (existing) return existing;
  const id = `app-${Crypto.randomUUID()}`;
  await SecureStore.setItemAsync(DEVICE_KEY, id);
  return id;
}

const CART_KEY = "ts.customer.cartToken";

/** The guest cart's token (the API issues it on the first add; merged into the account on sign-in). */
export const cartTokenStore = {
  get: () => SecureStore.getItemAsync(CART_KEY),
  set: (v: string) => SecureStore.setItemAsync(CART_KEY, v),
  clear: () => SecureStore.deleteItemAsync(CART_KEY),
};
