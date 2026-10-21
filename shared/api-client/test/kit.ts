// Shared helpers for contract tests that call a real, seeded local API (APP_ENV=development).
//   API_URL    API origin (default http://127.0.0.1:8080)
//   SEED_FILE  JSON written by `cmd/seed` (accounts and fixture IDs)
import { createHmac } from "node:crypto";
import { readFileSync } from "node:fs";

import { createApiClient } from "../src/index";

export const API_URL = process.env.API_URL ?? "http://127.0.0.1:8080";

type SeedAccount = {
  role: string;
  userId: string;
  email: string;
  phone?: string;
  password: string;
  totpSecret?: string;
};
type Seed = { accounts: SeedAccount[]; extra: Record<string, string> };

let seedCache: Seed | undefined;
export function seed(): Seed {
  if (!seedCache) {
    const file = process.env.SEED_FILE;
    if (!file) throw new Error("SEED_FILE is not set");
    seedCache = JSON.parse(readFileSync(file, "utf8")) as Seed;
  }
  return seedCache;
}

export function account(role: string): SeedAccount {
  const a = seed().accounts.find((x) => x.role === role);
  if (!a) throw new Error(`seed account ${role} missing`);
  return a;
}

// A per-run client address, so the API's per-IP limits (OTP, tracking) don't trip across runs.
const runIp = `198.18.${Math.floor(Math.random() * 256)}.${1 + Math.floor(Math.random() * 254)}`;
export const api = createApiClient({ baseUrl: API_URL, headers: { "X-Forwarded-For": runIp } });

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

async function newestCode(to: string, about = ""): Promise<string> {
  const res = await fetch(
    `${API_URL}/dev/messages?to=${encodeURIComponent(to)}`,
  );
  const rows = (await res.json()) as { payload?: { body?: string } }[] | null;
  for (const row of rows ?? []) {
    const body = row.payload?.body ?? "";
    if (!body || (about && !body.includes(about))) continue;
    return body.match(/\b(\d{6})\b/)?.[1] ?? "";
  }
  return "";
}

/** The newest code already sent to `to` ("" if none), so a test can wait for a fresh one. */
export const peekCode = (to: string, about = "") => newestCode(to, about);

/** Newest 6-digit code sent to `to` via the dev message log, waiting for one other than `notBefore`. */
export async function latestCode(
  to: string,
  notBefore = "",
  about = "",
): Promise<string> {
  const deadline = Date.now() + 20_000;
  while (Date.now() < deadline) {
    const code = await newestCode(to, about);
    if (code && code !== notBefore) return code;
    await sleep(300);
  }
  throw new Error(`no code delivered to ${to}`);
}

/** RFC 6238 TOTP (SHA-1, 6 digits, 30 s) from a base32 secret. */
export function totp(secret: string, at = Date.now()): string {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  let bits = "";
  for (const ch of secret.replace(/=+$/, "").toUpperCase())
    bits += alphabet.indexOf(ch).toString(2).padStart(5, "0");
  const key = Buffer.from(bits.match(/.{8}/g)!.map((b) => parseInt(b, 2)));
  const counter = Buffer.alloc(8);
  counter.writeBigUInt64BE(BigInt(Math.floor(at / 1000 / 30)));
  const h = createHmac("sha1", key).update(counter).digest();
  const offset = h[h.length - 1] & 0xf;
  const code = (h.readUInt32BE(offset) & 0x7fffffff) % 1_000_000;
  return code.toString().padStart(6, "0");
}

/** Signs a seeded customer in with password (market audience) and returns the access token. */
export async function customerToken(role = "customer"): Promise<string> {
  const a = account(role);
  const r = await api.post("/api/v1/auth/password", {
    body: { identifier: a.email, password: a.password, audience: "market" },
  });
  if (r.status !== "signed_in" || !r.tokens)
    throw new Error(`customer ${role} not signed in: ${r.status}`);
  return r.tokens.accessToken;
}

// Each TOTP code works once per 30-second step, so staff tokens are cached per role.
const staffCache = new Map<string, { token: string; at: number }>();

export async function staffToken(role = "admin"): Promise<string> {
  const cached = staffCache.get(role);
  if (cached && Date.now() - cached.at < 4 * 60_000) return cached.token;
  const a = account(role);
  const r = await api.post("/api/v1/auth/password", {
    body: { identifier: a.email, password: a.password, audience: "staff" },
  });
  if (r.status !== "mfa_required" || !r.mfaToken)
    throw new Error(`staff ${role}: expected mfa_required, got ${r.status}`);
  try {
    const t = await api.post("/api/v1/auth/mfa/totp", {
      body: {
        mfaToken: r.mfaToken,
        code: totp(a.totpSecret!),
        audience: "staff",
      },
    });
    staffCache.set(role, { token: t.accessToken, at: Date.now() });
    return t.accessToken;
  } catch (e) {
    if ((e as { code?: string }).code !== "code_reused") throw e;
    await sleep(31_000);
    return staffToken(role);
  }
}

const IMEI = () => {
  const body = Array.from({ length: 14 }, (_, i) => (i === 0 ? 3 : Math.floor(Math.random() * 10)));
  const sum = body.reduce((t, d, i) => {
    const v = i % 2 === 1 ? d * 2 : d;
    return t + (v > 9 ? v - 9 : v);
  }, 0);
  return body.join("") + String((10 - (sum % 10)) % 10);
};

/** A customer's paid order for one listing (Lagos address), via cart → preview → order → fake pay. */
export async function placePaidOrder(role: string, listingId: string) {
  const token = await customerToken(role);
  const cart = await api.get("/api/v1/cart", { token });
  for (const l of cart.lines ?? []) await api.delete("/api/v1/cart/items/{id}", { token, path: { id: l.cartItemId } });
  await api.post("/api/v1/cart/items", { token, body: { listingId, quantity: 1 } });
  const a = account(role);
  const addresses = (await api.get("/api/v1/me/addresses", { token })) ?? [];
  const addressId =
    addresses.find((x) => x.stateCode === "LA")?.id ??
    (await api.post("/api/v1/me/addresses", { token, body: { recipientName: "Test Customer", phone: a.phone ?? "+2348031234567", line1: "12 Allen Avenue", city: "Ikeja", stateCode: "LA", lga: "Ikeja" } })).id;
  const quote = await api.post("/api/v1/checkout/preview", { token, body: { addressId } });
  const placed = await api.post("/api/v1/orders", { token, idempotencyKey: crypto.randomUUID(), body: { addressId, priceHash: quote.priceHash, provider: "paystack" } });
  await fetch(`${API_URL}/dev/pay/paystack/${placed.payment.reference}/success`, { method: "POST" });
  for (let i = 0; i < 40; i++) {
    const st = await api.get("/api/v1/me/payments/{id}", { token, path: { id: placed.payment.paymentId } });
    if (st.orderStatus === "paid") return placed.orderNumber;
    await sleep(300);
  }
  throw new Error(`order ${placed.orderNumber} not paid`);
}

/** Picks and packs an order's TechShop parcel in the Ikeja warehouse; returns the pack result. */
export async function pickAndPack(orderNumber: string) {
  const token = await staffToken("admin");
  const warehouseId = seed().extra.warehouseIkejaId;
  const order = await api.get("/api/v1/staff/orders/{number}", { token, path: { number: orderNumber } });
  const fulfilmentId = order.fulfilments![0].id;
  const list = await api.post("/api/v1/staff/wms/waves", { token, body: { warehouseId } });
  await api.post("/api/v1/staff/wms/pick-lists/{id}/claim", { token, path: { id: list.id } });
  for (const item of (list.items ?? []).filter((i) => i.fulfilmentId === fulfilmentId)) {
    let serial: string | undefined;
    if (item.isSerialised) {
      serial = IMEI();
      await api.post("/api/v1/staff/inventory/device-units", { token, body: { warehouseId, variantId: item.variantId, condition: item.condition as "new", units: [{ imei: serial }] } });
    }
    await api.post("/api/v1/staff/wms/pick-lists/{id}/scan", { token, path: { id: list.id }, body: { orderLineId: item.orderLineId, quantity: item.quantity, serial } });
  }
  return api.post("/api/v1/staff/wms/pack", { token, body: { fulfilmentId, weightGrams: 800 } });
}

/** Signs the seeded rider in on their pre-approved phone (logistics audience). */
export async function riderToken(): Promise<string> {
  const a = account("rider");
  const before = await peekCode(a.phone!);
  const ch = await api.post("/api/v1/auth/otp/request", { body: { phone: a.phone!, purpose: "login" } });
  const code = await latestCode(a.phone!, before);
  const v = await api.post("/api/v1/auth/otp/verify", { body: { challengeId: ch.challengeId, code, audience: "logistics", deviceId: seed().extra.riderDeviceId, deviceName: "Contract test phone", platform: "android" } });
  if (!v.tokens) throw new Error("rider not signed in");
  return v.tokens.accessToken;
}

/** An access token for a phone number on an audience, signing the person up if they're new. */
export async function tokenForPhone(phone: string, audience: "market" | "seller" = "market", firstName = "Test"): Promise<string> {
  const before = await peekCode(phone);
  const ch = await api.post("/api/v1/auth/otp/request", { body: { phone, purpose: "login" } });
  const code = await latestCode(phone, before);
  const v = await api.post("/api/v1/auth/otp/verify", { body: { challengeId: ch.challengeId, code, audience } });
  if (v.tokens) return v.tokens.accessToken;
  const s = await api.post("/api/v1/auth/otp/signup", { body: { signupToken: v.signupToken!, firstName, lastName: "Customer", audience, deviceName: "e2e" } });
  return s.accessToken;
}

/**
 * A delivered order for one listing: bought for a Kano address (outside rider zones, so it ships
 * by the fake carrier), picked, packed, manifested and tracked to delivered. TechShop listings only.
 */
export async function deliveredOrder(token: string, listingId: string) {
  const cart = await api.get("/api/v1/cart", { token });
  for (const l of cart.lines ?? []) await api.delete("/api/v1/cart/items/{id}", { token, path: { id: l.cartItemId } });
  await api.post("/api/v1/cart/items", { token, body: { listingId, quantity: 1 } });
  const address = await api.post("/api/v1/me/addresses", { token, body: { recipientName: "Halima Bello", phone: "+2348031234570", line1: "5 Murtala Mohammed Way", city: "Kano", stateCode: "KN" } });
  const quote = await api.post("/api/v1/checkout/preview", { token, body: { addressId: address.id } });
  const placed = await api.post("/api/v1/orders", { token, idempotencyKey: crypto.randomUUID(), body: { addressId: address.id, priceHash: quote.priceHash, provider: "paystack" } });
  await fetch(`${API_URL}/dev/pay/paystack/${placed.payment.reference}/success`, { method: "POST" });
  const until = async (what: string, ok: () => Promise<boolean>) => {
    for (let i = 0; i < 60; i++) {
      if (await ok()) return;
      await sleep(300);
    }
    throw new Error(`timed out waiting for ${what}`);
  };
  const number = placed.orderNumber;
  const order = () => api.get("/api/v1/me/orders/{number}", { token, path: { number } });
  await until("payment", async () => (await order()).status === "paid");
  const packed = await pickAndPack(number);
  const staff = await staffToken("admin");
  const carriers = (await api.get("/api/v1/staff/wms/carriers", { token: staff })) ?? [];
  const carrierId = carriers.find((c) => c.code === "fake")!.id;
  const manifest = await api.post("/api/v1/staff/wms/manifests", { token: staff, body: { warehouseId: seed().extra.warehouseIkejaId, carrierId, labels: [packed.parcel!.labelCode] } });
  await api.post("/api/v1/staff/wms/manifests/{id}/sign", { token: staff, path: { id: manifest.id }, body: { signedByName: "Courier" } });
  for (const status of ["in_transit", "delivered"]) {
    const ev = (await (await fetch(`${API_URL}/dev/carriers/fake/${packed.shipment!.trackingNumber}/${status}`, { method: "POST" })).json()) as { raw: string; signature: string };
    await fetch(`${API_URL}/api/v1/webhooks/carriers/fake`, { method: "POST", headers: { "x-carrier-signature": ev.signature, "Content-Type": "application/json" }, body: ev.raw });
  }
  await until("delivery", async () => (await order()).status === "delivered");
  const f = (await order()).fulfilments![0];
  return { orderNumber: number, fulfilmentId: f.id, lineId: f.lines![0].id };
}
