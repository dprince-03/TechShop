import assert from "node:assert/strict";
import { test } from "node:test";

import { MobileSession, type TokenStore } from "../src/session";

const memory = (initial: string | null = null): TokenStore & { value: string | null } => {
  const s = {
    value: initial,
    async get() {
      return s.value;
    },
    async set(v: string) {
      s.value = v;
    },
    async clear() {
      s.value = null;
    },
  };
  return s;
};

const pair = (n: number, ttlMs = 600_000) => ({
  accessToken: `a${n}`,
  refreshToken: `r${n}`,
  accessExpiresAt: new Date(Date.now() + ttlMs).toISOString(),
  refreshExpiresAt: new Date(Date.now() + 7 * 86_400_000).toISOString(),
  sessionId: "s",
  userId: "u",
});

const json = (status: number, body: unknown) => new Response(JSON.stringify(body), { status, headers: { "content-type": "application/json" } });

test("restores from the stored refresh token and sends the new access token", async () => {
  const store = memory("r0");
  const seen: string[] = [];
  const fetchImpl: typeof fetch = async (input, init) => {
    const url = String(input);
    if (url.endsWith("/auth/refresh")) return json(200, pair(1));
    seen.push(new Headers(init?.headers).get("authorization") ?? "");
    return json(200, []);
  };
  const s = new MobileSession("http://api", store, fetchImpl);
  await s.api.get("/api/v1/rider/jobs");
  assert.deepEqual(seen, ["Bearer a1"]);
  assert.equal(store.value, "r1", "rotated refresh token stored");
});

test("concurrent calls share one refresh (rotation must not race)", async () => {
  let refreshes = 0;
  const fetchImpl: typeof fetch = async (input) => {
    if (String(input).endsWith("/auth/refresh")) {
      refreshes += 1;
      await new Promise((r) => setTimeout(r, 20));
      return json(200, pair(refreshes));
    }
    return json(200, []);
  };
  const s = new MobileSession("http://api", memory("r0"), fetchImpl);
  await Promise.all([s.api.get("/api/v1/rider/jobs"), s.api.get("/api/v1/rider/jobs"), s.api.get("/api/v1/rider/jobs")]);
  assert.equal(refreshes, 1);
});

test("a 401 refreshes once and retries", async () => {
  let calls = 0;
  const fetchImpl: typeof fetch = async (input) => {
    if (String(input).endsWith("/auth/refresh")) return json(200, pair(2));
    calls += 1;
    return calls === 1 ? json(401, { code: "unauthenticated" }) : json(200, []);
  };
  const s = new MobileSession("http://api", memory("r1"), fetchImpl);
  await s.accept(pair(1));
  await s.api.get("/api/v1/rider/jobs");
  assert.equal(calls, 2);
});

test("a refused refresh signs out; a network error keeps the token", async () => {
  let signedOut = false;
  const refused = new MobileSession("http://api", memory("r0"), async () => json(401, { code: "refresh_reuse_detected" }), { onSignedOut: () => (signedOut = true) });
  assert.equal(await refused.refresh(), false);
  assert.equal(signedOut, true);
  assert.equal(await refused.hasSession(), false);

  const store = memory("r0");
  const offline = new MobileSession("http://api", store, async () => Promise.reject(new TypeError("network")));
  assert.equal(await offline.refresh(), false);
  assert.equal(store.value, "r0");
});
