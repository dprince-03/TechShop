// Contract tests: the customer mobile app's sign-in audience and the reads it makes.
import assert from "node:assert/strict";
import { randomInt } from "node:crypto";
import { test } from "node:test";

import { api, latestCode } from "../kit";

test("customer_app: new number → code → profile → signed in → /me → refresh → logout", async () => {
  const phone = `+23481${randomInt(10_000_000, 99_999_999)}`;
  const ch = await api.post("/api/v1/auth/otp/request", { body: { phone, purpose: "login" } });
  const code = await latestCode(phone);
  const v = await api.post("/api/v1/auth/otp/verify", { body: { challengeId: ch.challengeId, code, audience: "customer_app", deviceId: "app-contract-test", deviceName: "android phone", platform: "android" } });
  assert.equal(v.status, "needs_profile");
  const t = await api.post("/api/v1/auth/otp/signup", { body: { signupToken: v.signupToken!, firstName: "Chidi", lastName: "App", audience: "customer_app", deviceId: "app-contract-test", platform: "android" } });
  const me = await api.get("/api/v1/me", { token: t.accessToken });
  assert.equal(me.firstName, "Chidi");
  assert.equal(me.isRider, false);
  const t2 = await api.post("/api/v1/auth/refresh", { body: { refreshToken: t.refreshToken } });
  assert.notEqual(t2.refreshToken, t.refreshToken);
  const sessions = (await api.get("/api/v1/me/sessions", { token: t2.accessToken })) ?? [];
  assert.ok(sessions.some((s) => s.audience === "customer_app"));
  await api.post("/api/v1/auth/logout", { token: t2.accessToken, body: { refreshToken: t2.refreshToken } });
});
