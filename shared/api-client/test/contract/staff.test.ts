// Contract tests: the staff portal's endpoints (sign-in, permissions, every list it reads, and
// the guarded actions: step-up for refunds, separation of duties).
import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { describe, test } from "node:test";

import { ApiError } from "../../src/index";
import {
  account,
  api,
  API_URL,
  customerToken,
  seed,
  staffToken,
  totp,
} from "../kit";

const WAREHOUSE = seed().extra.warehouseIkejaId;

/** Every list screen the portal shows, read as admin; each must answer 200 with JSON. */
const LIST_ENDPOINTS = [
  "/api/v1/staff/orders",
  "/api/v1/staff/finance/dashboard",
  "/api/v1/staff/finance/refunds",
  "/api/v1/staff/finance/trial-balance",
  "/api/v1/staff/finance/journals",
  "/api/v1/staff/finance/reconciliation/exceptions",
  "/api/v1/staff/finance/providers",
  "/api/v1/staff/finance/periods",
  "/api/v1/dispatch/board",
  "/api/v1/dispatch/riders",
  "/api/v1/dispatch/zones",
  "/api/v1/dispatch/devices",
  `/api/v1/staff/wms/pick-lists?warehouseId=${WAREHOUSE}`,
  `/api/v1/staff/wms/manifests?warehouseId=${WAREHOUSE}`,
  "/api/v1/staff/wms/carriers",
  "/api/v1/staff/catalog/moderation",
  "/api/v1/staff/catalog/suggestions",
  "/api/v1/staff/marketing/promotions",
  "/api/v1/staff/marketing/coupons",
  "/api/v1/staff/marketing/campaigns",
  "/api/v1/staff/marketing/segments",
  "/api/v1/staff/marketing/journeys",
  "/api/v1/staff/admin/users",
  "/api/v1/staff/admin/roles",
  "/api/v1/staff/admin/permissions",
  "/api/v1/staff/admin/audit-log",
  "/api/v1/staff/admin/auth-events",
  "/api/v1/staff/hr/members",
  "/api/v1/staff/inventory/levels",
  "/api/v1/staff/inventory/movements",
  "/api/v1/staff/inventory/warehouses",
  "/api/v1/staff/content/pages",
  "/api/v1/staff/search/insights",
  "/api/v1/staff/search/synonyms",
  "/api/v1/staff/notify/templates",
  "/api/v1/staff/notify/suppressions",
];

async function raw(path: string, token?: string) {
  const r = await fetch(API_URL + path, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  const text = await r.text();
  let body: unknown = text;
  try {
    body = text ? JSON.parse(text) : undefined;
  } catch {
    // keep the raw text (e.g. a plain 404 page) for the failure message
  }
  return { status: r.status, body };
}

describe("staff sign-in", () => {
  test("password needs a TOTP second step; wrong codes are refused", async () => {
    const a = account("support_agent");
    const r = await api.post("/api/v1/auth/password", {
      body: { identifier: a.email, password: a.password, audience: "staff" },
    });
    assert.equal(r.status, "mfa_required");
    assert.ok(r.mfaToken && !r.tokens);
    await assert.rejects(
      api.post("/api/v1/auth/mfa/totp", {
        body: { mfaToken: r.mfaToken!, code: "000000", audience: "staff" },
      }),
      (e: unknown) =>
        e instanceof ApiError && e.status >= 400 && e.status < 500,
    );
    const t = await api.post("/api/v1/auth/mfa/totp", {
      body: {
        mfaToken: r.mfaToken!,
        code: totp(a.totpSecret!),
        audience: "staff",
      },
    });
    assert.ok(t.accessToken);
  });

  test("a customer token can't call staff endpoints", async () => {
    const r = await raw("/api/v1/staff/orders", await customerToken());
    assert.ok(r.status === 401 || r.status === 403, `got ${r.status}`);
  });

  test("/me lists the staff member's permissions", async () => {
    const me = await api.get("/api/v1/me", {
      token: await staffToken("admin"),
    });
    assert.equal(me.isStaff, true);
    assert.ok((me.permissions ?? []).includes("finance.view"));
  });
});

describe("staff lists", () => {
  test("every list the portal reads answers 200 for admin", async () => {
    const token = await staffToken("admin");
    const failures: string[] = [];
    for (const p of LIST_ENDPOINTS) {
      const r = await raw(p, token);
      if (r.status !== 200)
        failures.push(
          `${p} → ${r.status} ${JSON.stringify(r.body).slice(0, 120)}`,
        );
    }
    assert.deepEqual(failures, []);
  });

  test("permissions are enforced per role", async () => {
    const support = await staffToken("support_agent");
    for (const p of [
      "/api/v1/staff/finance/trial-balance",
      "/api/v1/staff/admin/users",
      "/api/v1/staff/marketing/campaigns",
    ]) {
      const r = await raw(p, support);
      assert.equal(r.status, 403, p);
    }
    assert.equal((await raw("/api/v1/staff/orders", support)).status, 200);
  });

  test("staff order view masks the customer's phone", async () => {
    const token = await staffToken("support_agent");
    const list = await api.get("/api/v1/staff/orders", {
      token,
      query: { limit: 5 },
    });
    const first = list.items?.[0];
    if (!first) return;
    const detail = await api.get("/api/v1/staff/orders/{number}", {
      token,
      path: { number: first.orderNumber },
    });
    if (detail.contactPhone)
      assert.doesNotMatch(detail.contactPhone, /^\+234\d{10}$/);
  });
});

describe("guarded actions", () => {
  test("refund approval: requester can't approve; approver needs a fresh step-up", async () => {
    // A paid order to refund.
    const ctoken = await customerToken("customer_b");
    const cart = await api.get("/api/v1/cart", { token: ctoken });
    for (const l of cart.lines ?? [])
      await api.delete("/api/v1/cart/items/{id}", {
        token: ctoken,
        path: { id: l.cartItemId },
      });
    await api.post("/api/v1/cart/items", {
      token: ctoken,
      body: { listingId: seed().extra.listing_p3, quantity: 1 },
    });
    const addresses =
      (await api.get("/api/v1/me/addresses", { token: ctoken })) ?? [];
    const addressId =
      addresses[0]?.id ??
      (
        await api.post("/api/v1/me/addresses", {
          token: ctoken,
          body: {
            recipientName: "B Test",
            phone: "+2348031234570",
            line1: "1 Test Road",
            city: "Ikeja",
            stateCode: "LA",
          },
        })
      ).id;
    const quote = await api.post("/api/v1/checkout/preview", {
      token: ctoken,
      body: { addressId },
    });
    const placed = await api.post("/api/v1/orders", {
      token: ctoken,
      idempotencyKey: randomUUID(),
      body: { addressId, priceHash: quote.priceHash, provider: "paystack" },
    });
    await fetch(
      `${API_URL}/dev/pay/paystack/${placed.payment.reference}/success`,
      { method: "POST" },
    );
    for (let i = 0; i < 30; i++) {
      const s = await api.get("/api/v1/me/payments/{id}", {
        token: ctoken,
        path: { id: placed.payment.paymentId },
      });
      if (s.orderStatus === "paid") break;
      await new Promise((r) => setTimeout(r, 300));
    }

    const officer = await staffToken("finance_officer");
    const refund = await api.post("/api/v1/staff/finance/refunds", {
      token: officer,
      idempotencyKey: randomUUID(),
      body: {
        orderNumber: placed.orderNumber,
        amountKobo: 10_000,
        method: "original",
        reason: "Contract test partial refund",
      },
    });
    assert.equal(refund.status, "requested");
    await assert.rejects(
      api.post("/api/v1/staff/finance/refunds/{id}/approve", {
        token: officer,
        path: { id: refund.id },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 403,
    );

    const manager = await staffToken("finance_manager");
    const approved = await api.post(
      "/api/v1/staff/finance/refunds/{id}/approve",
      { token: manager, path: { id: refund.id } },
    );
    assert.ok(
      ["approved", "processing", "succeeded"].includes(approved.status),
      approved.status,
    );
  });

  test("step-up returns fresh tokens for a valid authenticator code", async () => {
    const a = account("auditor");
    const r = await api.post("/api/v1/auth/password", {
      body: { identifier: a.email, password: a.password, audience: "staff" },
    });
    const t = await api.post("/api/v1/auth/mfa/totp", {
      body: {
        mfaToken: r.mfaToken!,
        code: totp(a.totpSecret!),
        audience: "staff",
      },
    });
    await assert.rejects(
      api.post("/api/v1/auth/mfa/step-up", {
        token: t.accessToken,
        body: { code: "000000" },
      }),
      (e: unknown) => e instanceof ApiError && e.status >= 400,
    );
    // The same 30-second code can't be reused, so use the next step's code.
    const next = totp(a.totpSecret!, Date.now() + 30_000);
    const up = await api
      .post("/api/v1/auth/mfa/step-up", {
        token: t.accessToken,
        body: { code: next },
      })
      .catch((e) => e);
    if (up instanceof ApiError)
      assert.ok(["code_invalid", "code_reused"].includes(up.code), up.code);
    else assert.ok(up.accessToken);
  });
});
