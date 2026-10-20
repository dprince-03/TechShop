import * as Crypto from "expo-crypto";
import * as SecureStore from "expo-secure-store";
import Storage from "expo-sqlite/kv-store";

import type { TokenStore } from "@/features/auth/session";
import type { QueuedEvent, QueueStore } from "@/features/sync/queue";

const REFRESH_KEY = "ts.refreshToken";
const DEVICE_KEY = "ts.deviceId";
const QUEUE_KEY = "ts.syncQueue";

/** Refresh token in the Keychain / Keystore; only readable while the phone is unlocked. */
export const secureTokenStore: TokenStore = {
  get: () => SecureStore.getItemAsync(REFRESH_KEY),
  set: (v) => SecureStore.setItemAsync(REFRESH_KEY, v, { keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY }),
  clear: () => SecureStore.deleteItemAsync(REFRESH_KEY),
};

/** A random ID for this install, sent at sign-in so dispatch can approve this phone once. */
export async function deviceId(): Promise<string> {
  const existing = await SecureStore.getItemAsync(DEVICE_KEY);
  if (existing) return existing;
  const id = `phone-${Crypto.randomUUID()}`;
  await SecureStore.setItemAsync(DEVICE_KEY, id);
  return id;
}

/** The offline queue, persisted in SQLite so actions survive an app restart. */
export const queueStore: QueueStore = {
  async load() {
    const raw = await Storage.getItem(QUEUE_KEY);
    return raw ? (JSON.parse(raw) as QueuedEvent[]) : [];
  },
  async save(items) {
    await Storage.setItem(QUEUE_KEY, JSON.stringify(items));
  },
};

/** Removes everything this app stored about jobs (on sign-out: customer data must not linger). */
export async function clearLocalData() {
  await Storage.removeItem(QUEUE_KEY);
}

export const newEventId = () => Crypto.randomUUID();
