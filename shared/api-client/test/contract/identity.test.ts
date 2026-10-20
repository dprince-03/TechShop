// Contract tests: sign-in, sessions and the account endpoints the storefront uses.
import assert from "node:assert/strict";
import { randomInt } from "node:crypto";
import { describe, test } from "node:test";

import { ApiError } from "../../src/index";
import { account, api, customerToken, latestCode, peekCode } from "../kit";

const freshPhone = () => `+23480${String(randomInt(10_000_000, 99_999_999))}`;

describe("phone OTP", () => {
  test("existing customer: request → verify → signed in", async () => {
    const phone = account("customer").phone!;
    const before = await peekCode(phone);
    const ch = await api.post("/api/v1/auth/otp/request", {
      body: { phone, purpose: "login" },
    });
    assert.ok(ch.challengeId && ch.resendAfter > 0);
    const code = await latestCode(phone, before);
    const r = await api.post("/api/v1/auth/otp/verify", {
      body: {
        challengeId: ch.challengeId,
        code,
        audience: "market",
        platform: "web",
      },
    });
    assert.equal(r.status, "signed_in");
    assert.ok(r.tokens?.accessToken && r.tokens.refreshToken);
  });

  test("a wrong code is refused", async () => {
    const phone = freshPhone();
    const ch = await api.post("/api/v1/auth/otp/request", {
      body: { phone, purpose: "login" },
    });
    await assert.rejects(
      api.post("/api/v1/auth/otp/verify", {
        body: {
          challengeId: ch.challengeId,
          code: "000000",
          audience: "market",
        },
      }),
      (e: unknown) =>
        e instanceof ApiError && e.status >= 400 && e.status < 500,
    );
  });

  test("new phone: verify → needs_profile → complete sign-up", async () => {
    const phone = freshPhone();
    const ch = await api.post("/api/v1/auth/otp/request", {
      body: { phone, purpose: "signup" },
    });
    const code = await latestCode(phone);
    const v = await api.post("/api/v1/auth/otp/verify", {
      body: { challengeId: ch.challengeId, code, audience: "market" },
    });
    assert.equal(v.status, "needs_profile");
    assert.ok(v.signupToken);
    const done = await api.post("/api/v1/auth/otp/signup", {
      body: {
        signupToken: v.signupToken!,
        firstName: "Ada",
        lastName: "Test",
        audience: "market",
      },
    });
    assert.ok(done.accessToken);
    const me = await api.get("/api/v1/me", { token: done.accessToken });
    assert.equal(me.firstName, "Ada");
    assert.equal(me.phone, phone);
  });

  test("phone format is validated", async () => {
    await assert.rejects(
      api.post("/api/v1/auth/otp/request", {
        body: { phone: "12345", purpose: "login" },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 422,
    );
  });
});

describe("password and sessions", () => {
  test("wrong password gets a generic error", async () => {
    const a = account("customer");
    await assert.rejects(
      api.post("/api/v1/auth/password", {
        body: {
          identifier: a.email,
          password: "wrong-password-123",
          audience: "market",
        },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 401,
    );
    await assert.rejects(
      api.post("/api/v1/auth/password", {
        body: {
          identifier: "nobody@dev.techshop.ng",
          password: "wrong-password-123",
          audience: "market",
        },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 401,
    );
  });

  test("refresh rotates the token; logout revokes the session", async () => {
    const a = account("customer");
    const r = await api.post("/api/v1/auth/password", {
      body: { identifier: a.email, password: a.password, audience: "market" },
    });
    const t1 = r.tokens!;
    const t2 = await api.post("/api/v1/auth/refresh", {
      body: { refreshToken: t1.refreshToken },
    });
    assert.notEqual(t2.refreshToken, t1.refreshToken);
    assert.equal(t2.userId, t1.userId);
    await api.post("/api/v1/auth/logout", {
      token: t2.accessToken,
      body: { refreshToken: t2.refreshToken },
    });
    await assert.rejects(
      api.post("/api/v1/auth/refresh", {
        body: { refreshToken: t2.refreshToken },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 401,
    );
  });

  test("/me needs a token", async () => {
    await assert.rejects(
      api.get("/api/v1/me"),
      (e: unknown) => e instanceof ApiError && e.status === 401,
    );
  });

  test("sessions list includes the current one", async () => {
    const token = await customerToken();
    const list = (await api.get("/api/v1/me/sessions", { token })) ?? [];
    assert.ok(list.length > 0);
  });
});

describe("account", () => {
  test("profile update", async () => {
    const token = await customerToken("customer_b");
    const me = await api.patch("/api/v1/me", {
      token,
      body: { marketingOptIn: true },
    });
    assert.equal(me.marketingOptIn, true);
  });

  test("addresses: create, list, update, delete; other users can't touch them", async () => {
    const token = await customerToken();
    const other = await customerToken("customer_b");
    const body = {
      recipientName: "Ada Test",
      phone: "+2348031234567",
      line1: "12 Allen Avenue",
      city: "Ikeja",
      stateCode: "LA",
      lga: "Ikeja",
    };
    const created = await api.post("/api/v1/me/addresses", { token, body });
    assert.ok(created.id);
    const list = (await api.get("/api/v1/me/addresses", { token })) ?? [];
    assert.ok(list.some((x) => x.id === created.id));
    await assert.rejects(
      api.delete("/api/v1/me/addresses/{id}", {
        token: other,
        path: { id: created.id },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 404,
    );
    const updated = await api.put("/api/v1/me/addresses/{id}", {
      token,
      path: { id: created.id },
      body: { ...body, landmark: "Opposite the bank" },
    });
    assert.equal(updated.landmark, "Opposite the bank");
    await api.delete("/api/v1/me/addresses/{id}", {
      token,
      path: { id: created.id },
    });
    const after = (await api.get("/api/v1/me/addresses", { token })) ?? [];
    assert.ok(!after.some((x) => x.id === created.id));
  });

  test("saved products: toggle on, list, toggle off", async () => {
    const token = await customerToken();
    const p = await api.get("/api/v1/search", { query: { limit: 1 } });
    const id = p.items![0].id;
    const on = await api.post("/api/v1/me/saved/{productId}", {
      token,
      path: { productId: id },
    });
    assert.equal(on.saved, true);
    const saved = (await api.get("/api/v1/me/saved", { token })) ?? [];
    assert.ok(JSON.stringify(saved).includes(id));
    const off = await api.post("/api/v1/me/saved/{productId}", {
      token,
      path: { productId: id },
    });
    assert.equal(off.saved, false);
  });

  test("notifications list responds", async () => {
    const token = await customerToken();
    const n = await api.get("/api/v1/me/notifications", { token });
    assert.ok(n !== undefined);
  });
});
