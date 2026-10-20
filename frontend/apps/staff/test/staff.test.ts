// End-to-end tests for the staff portal: real pages and Server Actions over HTTP against
// `next dev` (STAFF_URL) and the seeded local API.
import assert from "node:assert/strict";
import path from "node:path";
import { describe, test } from "node:test";
import { fileURLToPath } from "node:url";

import { account, form, textOf, totp, WebClient } from "@techshop/testing";

const APP_DIR = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const BASE = process.env.STAFF_URL ?? "http://localhost:3004";
const browser = () => new WebClient(BASE, APP_DIR);

// Each TOTP code works once; remember the last step used per account across tests.
const usedStep = new Map<string, number>();
async function freshCode(secret: string) {
  for (;;) {
    const step = Math.floor(Date.now() / 30_000);
    if (usedStep.get(secret) !== step) {
      usedStep.set(secret, step);
      return totp(secret);
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
}

/** Signs a seeded staff member in through the real two-step sign-in action. */
async function signIn(role: string, next = "/") {
  const b = browser();
  const a = account(role);
  await b.get(`/sign-in?next=${encodeURIComponent(next)}`);
  const step1 = await b.action<{ step: string }>("/sign-in", "staffSignIn", [
    { step: "start" },
    form({ intent: "password", email: a.email, password: a.password, next }),
  ]);
  assert.equal(step1.value.step, "totp", JSON.stringify(step1.value));
  let step2 = await b.action<{ step: string; error?: string }>(
    "/sign-in",
    "staffSignIn",
    [
      { step: "totp" },
      form({ intent: "totp", code: await freshCode(a.totpSecret!), next }),
    ],
  );
  if (!step2.redirect && /already used/.test(step2.value?.error ?? "")) {
    // An earlier run used this 30-second code; take the next one.
    usedStep.set(a.totpSecret!, Math.floor(Date.now() / 30_000));
    step2 = await b.action("/sign-in", "staffSignIn", [
      { step: "totp" },
      form({ intent: "totp", code: await freshCode(a.totpSecret!), next }),
    ]);
  }
  assert.equal(step2.redirect, next, JSON.stringify(step2.value));
  return b;
}

describe("staff sign-in", () => {
  test("signed-out visitors are sent to sign-in with a return path", async () => {
    const r = await browser().get("/finance");
    assert.match(r.url, /\/sign-in\?next=%2Ffinance$/);
  });

  test("a wrong password and a wrong code are refused", async () => {
    const b = browser();
    const a = account("support_agent");
    await b.get("/sign-in");
    const bad = await b.action<{ step: string; error?: string }>(
      "/sign-in",
      "staffSignIn",
      [
        { step: "start" },
        form({
          intent: "password",
          email: a.email,
          password: "wrong-password-0",
        }),
      ],
    );
    assert.equal(bad.value.step, "start");
    const ok = await b.action<{ step: string }>("/sign-in", "staffSignIn", [
      { step: "start" },
      form({ intent: "password", email: a.email, password: a.password }),
    ]);
    assert.equal(ok.value.step, "totp");
    const wrong = await b.action<{ step: string; error?: string }>(
      "/sign-in",
      "staffSignIn",
      [{ step: "totp" }, form({ intent: "totp", code: "000000" })],
    );
    assert.equal(wrong.value.step, "totp");
    assert.ok(wrong.value.error);
    assert.ok(!b.cookie("ts_session"));
  });
});

describe("portal", () => {
  test("admin: dashboard, orders, finance tabs all render live data", async () => {
    const b = await signIn("admin");
    const home = textOf((await b.get("/")).html);
    assert.match(home, /Good day,/);
    assert.match(home, /Awaiting payment|Paid, to fulfil/);
    assert.doesNotMatch(home, /Sample role/);

    const orders = await b.get("/orders");
    assert.equal(orders.status, 200);
    const number = textOf(orders.html).match(/TS-\d+/)?.[0];
    assert.ok(number, "an order is listed");
    const detail = textOf((await b.get(`/orders/${number}`)).html);
    assert.match(detail, new RegExp(`Order ${number}`));
    assert.match(detail, /Payments/);

    for (const tab of [
      "overview",
      "refunds",
      "reconciliation",
      "ledger",
      "periods",
    ]) {
      const r = await b.get(`/finance?tab=${tab}`);
      assert.equal(r.status, 200, tab);
    }
    assert.match(
      textOf((await b.get("/finance?tab=ledger")).html),
      /Trial balance\s*Balanced/,
    );
  });

  test("support agent sees only permitted modules and can't open finance", async () => {
    const b = await signIn("support_agent");
    const home = (await b.get("/")).html;
    assert.match(home, /href="\/orders"/);
    assert.doesNotMatch(home, /href="\/finance"/);
    const fin = await b.get("/finance");
    assert.match(fin.url, /\/\?denied=1$/);
  });

  test("sign out ends the session", async () => {
    const b = await signIn("auditor");
    await b.get("/account");
    const out = await b.action("/account", "signOut", []);
    assert.equal(out.redirect, "/sign-in");
    assert.match((await b.get("/")).url, /\/sign-in/);
  });
});

describe("catalogue, search and previews", () => {
  test("catalogue tabs render; a product is created with a listing and is searchable", async () => {
    const b = await signIn("catalog_editor");
    for (const tab of ["products", "moderation", "suggestions"]) assert.equal((await b.get(`/catalogue?tab=${tab}`)).status, 200, tab);
    const form_ = await b.get("/catalogue/new");
    assert.equal(form_.status, 200);
    assert.match(textOf(form_.html), /First variant/);

    const bad = await b.action<{ ok: boolean; errors?: Record<string, string> }>("/catalogue/new", "createProduct", [{}, form({ name: "", categorySlug: "", sku: "" })]);
    assert.equal(bad.value.ok, false);
    assert.ok(bad.value.errors?.name && bad.value.errors?.sku);

    const stamp = Date.now().toString(36);
    const created = await b.action("/catalogue/new", "createProduct", [
      {},
      form({ name: `E2E Test Charger ${stamp}`, categorySlug: "accessories", sku: `E2E-${stamp}`, variantName: "65W", weightGrams: "300", price: "15000", condition: "new", warrantyMonths: "6", status: "active" }),
    ]);
    assert.match(created.redirect ?? "", /\/catalogue\?created=e2e-test-charger-/, JSON.stringify(created.value));
    // The API indexes the new product; search finds it.
    let found = false;
    for (let i = 0; i < 20 && !found; i++) {
      found = textOf((await b.get(`/catalogue?tab=products&q=${encodeURIComponent(`E2E Test Charger ${stamp}`)}`)).html).includes(`E2E Test Charger ${stamp}`);
      if (!found) await new Promise((r) => setTimeout(r, 500));
    }
    assert.ok(found, "new product is searchable");
  });

  test("global search, notifications and a preview module", async () => {
    const b = await signIn("admin");
    const s = textOf((await b.get("/search?q=TS-")).html);
    assert.match(s, /Orders/);
    assert.equal((await b.get("/notifications")).status, 200);
    const v = textOf((await b.get("/vendors")).html);
    assert.match(v, /Preview: everything on this page is sample data/);
    assert.equal((await b.get("/no-such-module")).status, 404);
  });

  test("a support agent can't create products", async () => {
    const b = await signIn("support_agent");
    assert.match((await b.get("/catalogue/new")).url, /\/\?denied=1$/);
  });
});
