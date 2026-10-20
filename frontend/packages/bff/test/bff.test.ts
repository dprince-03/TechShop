import assert from "node:assert/strict";
import { test } from "node:test";

import { loadConfig } from "../src/config";
import { seal, unseal } from "../src/crypto";
import { safeReturnPath, signInPath } from "../src/paths";
import {
  decodeSession,
  encodeSession,
  needsRefresh,
  refreshSession,
  type Session,
} from "../src/session";

const cfg = loadConfig("test", "market", {
  NODE_ENV: "development",
  API_URL: "http://api.test",
});

const session = (over: Partial<Session> = {}): Session => ({
  sessionId: "s1",
  userId: "u1",
  accessToken: "a1",
  accessExpiresAt: Date.now() + 10 * 60_000,
  refreshToken: "r1",
  refreshExpiresAt: Date.now() + 30 * 86_400_000,
  ...over,
});

test("seal/unseal round-trips and rejects tampering or a different key", async () => {
  const sealed = await seal({ a: 1 }, "k".repeat(40));
  assert.deepEqual(await unseal(sealed, "k".repeat(40)), { a: 1 });
  assert.equal(await unseal(sealed, "j".repeat(40)), null);
  const flipped =
    sealed.slice(0, -2) + (sealed.endsWith("A") ? "B" : "A") + sealed.slice(-1);
  assert.equal(await unseal(flipped, "k".repeat(40)), null);
  assert.equal(await unseal("garbage", "k".repeat(40)), null);
  assert.equal(await unseal(undefined, "k".repeat(40)), null);
});

test("sealing is randomised (same value, different cookie)", async () => {
  assert.notEqual(
    await seal({ a: 1 }, "k".repeat(40)),
    await seal({ a: 1 }, "k".repeat(40)),
  );
});

test("config refuses missing or weak secrets in production only", () => {
  assert.throws(
    () =>
      loadConfig("m", "market", {
        NODE_ENV: "production",
        API_URL: "https://api",
      }),
    /SESSION_SECRET is required/,
  );
  assert.throws(
    () =>
      loadConfig("m", "market", {
        NODE_ENV: "production",
        API_URL: "https://api",
        SESSION_SECRET: "short",
      }),
    /at least 32/,
  );
  assert.throws(
    () =>
      loadConfig("m", "market", {
        NODE_ENV: "production",
        API_URL: "https://api",
        SESSION_SECRET: "change-me-change-me-change-me-change-me",
      }),
    /at least 32/,
  );
  const prod = loadConfig("m", "market", {
    NODE_ENV: "production",
    API_URL: "https://api",
    SESSION_SECRET: "x9".repeat(24),
  });
  assert.equal(prod.cookieName, "__Host-ts_session");
  assert.equal(prod.secure, true);
  assert.equal(cfg.cookieName, "ts_session");
  assert.throws(
    () => loadConfig("m", "market", { NODE_ENV: "development" }),
    /API_URL/,
  );
});

test("expired sessions decode to null; near-expiry access tokens need a refresh", async () => {
  assert.ok(await decodeSession(await encodeSession(session(), cfg), cfg));
  assert.equal(
    await decodeSession(
      await encodeSession(session({ refreshExpiresAt: Date.now() - 1 }), cfg),
      cfg,
    ),
    null,
  );
  assert.equal(needsRefresh(session()), false);
  assert.equal(
    needsRefresh(session({ accessExpiresAt: Date.now() + 30_000 })),
    true,
  );
});

test("refresh: 401 means revoked, 5xx or network means unavailable", async () => {
  const reply = (status: number) => async () =>
    new Response(JSON.stringify({ code: "x" }), { status });
  assert.equal(
    (await refreshSession(cfg, session(), reply(401))).kind,
    "revoked",
  );
  assert.equal(
    (await refreshSession(cfg, session(), reply(503))).kind,
    "unavailable",
  );
  assert.equal(
    (
      await refreshSession(cfg, session(), async () =>
        Promise.reject(new Error("down")),
      )
    ).kind,
    "unavailable",
  );
});

test("return paths only allow same-site relative paths", () => {
  for (const bad of [
    "//evil.com",
    "/\\evil.com",
    "https://evil.com",
    "https:evil.com",
    "evil",
    "",
    undefined,
    "/a\nb",
  ]) {
    assert.equal(safeReturnPath(bad), "/", String(bad));
  }
  assert.equal(safeReturnPath("/orders?x=1"), "/orders?x=1");
  assert.equal(signInPath("/orders"), "/sign-in?next=%2Forders");
  assert.equal(signInPath("//evil.com"), "/sign-in");
});
