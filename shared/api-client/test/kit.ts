// Shared helpers for contract tests that call a real, seeded local API (APP_ENV=development).
//   API_URL    API origin (default http://127.0.0.1:8080)
//   SEED_FILE  JSON written by `cmd/seed` (accounts and fixture IDs)
import { createHmac } from "node:crypto";
import { readFileSync } from "node:fs";

import { createApiClient } from "../src/index";

export const API_URL = process.env.API_URL ?? "http://127.0.0.1:8080";

type SeedAccount = { role: string; userId: string; email: string; phone?: string; password: string; totpSecret?: string };
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

export const api = createApiClient({ baseUrl: API_URL });

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

/** Newest 6-digit code sent to `to` via the dev message log, waiting for one other than `notBefore`. */
export async function latestCode(to: string, notBefore = "", about = ""): Promise<string> {
  const deadline = Date.now() + 20_000;
  while (Date.now() < deadline) {
    const res = await fetch(`${API_URL}/dev/messages?to=${encodeURIComponent(to)}`);
    const rows = (await res.json()) as { payload?: { body?: string } }[] | null;
    for (const row of rows ?? []) {
      const body = row.payload?.body ?? "";
      if (!body || (about && !body.includes(about))) continue;
      const m = body.match(/\b(\d{6})\b/);
      if (m && m[1] !== notBefore) return m[1];
      break;
    }
    await sleep(300);
  }
  throw new Error(`no code delivered to ${to}`);
}

/** RFC 6238 TOTP (SHA-1, 6 digits, 30 s) from a base32 secret. */
export function totp(secret: string, at = Date.now()): string {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  let bits = "";
  for (const ch of secret.replace(/=+$/, "").toUpperCase()) bits += alphabet.indexOf(ch).toString(2).padStart(5, "0");
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
  if (r.status !== "signed_in" || !r.tokens) throw new Error(`customer ${role} not signed in: ${r.status}`);
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
  if (r.status !== "mfa_required" || !r.mfaToken) throw new Error(`staff ${role}: expected mfa_required, got ${r.status}`);
  try {
    const t = await api.post("/api/v1/auth/mfa/totp", {
      body: { mfaToken: r.mfaToken, code: totp(a.totpSecret!), audience: "staff" },
    });
    staffCache.set(role, { token: t.accessToken, at: Date.now() });
    return t.accessToken;
  } catch (e) {
    if ((e as { code?: string }).code !== "code_reused") throw e;
    await sleep(31_000);
    return staffToken(role);
  }
}
