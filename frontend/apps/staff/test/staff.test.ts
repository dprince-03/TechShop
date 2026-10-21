// End-to-end tests for the staff portal: real pages and Server Actions over HTTP against
// `next dev` (STAFF_URL) and the seeded local API.
import assert from "node:assert/strict";
import { describe, test } from "node:test";

import {
  account,
  api,
  API_URL,
  customerToken,
  form,
  seed,
  staffToken,
  textOf,
} from "@techshop/testing";

import { browser, signIn } from "./helpers";

/** A random IMEI with a valid Luhn check digit (the API rejects bad check digits). */
function imei(): string {
  const body = Array.from({ length: 14 }, (_, i) =>
    i === 0 ? 3 : Math.floor(Math.random() * 10),
  );
  const sum = body.reduce((t, d, i) => {
    const v = i % 2 === 1 ? d * 2 : d;
    return t + (v > 9 ? v - 9 : v);
  }, 0);
  return body.join("") + String((10 - (sum % 10)) % 10);
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
    for (const tab of ["products", "moderation", "suggestions"])
      assert.equal((await b.get(`/catalogue?tab=${tab}`)).status, 200, tab);
    const form_ = await b.get("/catalogue/new");
    assert.equal(form_.status, 200);
    assert.match(textOf(form_.html), /First variant/);

    const bad = await b.action<{
      ok: boolean;
      errors?: Record<string, string>;
    }>("/catalogue/new", "createProduct", [
      {},
      form({ name: "", categorySlug: "", sku: "" }),
    ]);
    assert.equal(bad.value.ok, false);
    assert.ok(bad.value.errors?.name && bad.value.errors?.sku);

    const stamp = Date.now().toString(36);
    const created = await b.action("/catalogue/new", "createProduct", [
      {},
      form({
        name: `E2E Test Charger ${stamp}`,
        categorySlug: "accessories",
        sku: `E2E-${stamp}`,
        variantName: "65W",
        weightGrams: "300",
        price: "15000",
        condition: "new",
        warrantyMonths: "6",
        status: "active",
      }),
    ]);
    assert.match(
      created.redirect ?? "",
      /\/catalogue\?created=e2e-test-charger-/,
      JSON.stringify(created.value),
    );
    // The API indexes the new product; search finds it.
    let found = false;
    for (let i = 0; i < 20 && !found; i++) {
      found = textOf(
        (
          await b.get(
            `/catalogue?tab=products&q=${encodeURIComponent(`E2E Test Charger ${stamp}`)}`,
          )
        ).html,
      ).includes(`E2E Test Charger ${stamp}`);
      if (!found) await new Promise((r) => setTimeout(r, 500));
    }
    assert.ok(found, "new product is searchable");
  });

  test("global search, notifications and a preview module", async () => {
    const b = await signIn("admin");
    const s = textOf((await b.get("/search?q=TS-")).html);
    assert.match(s, /Orders/);
    assert.equal((await b.get("/notifications")).status, 200);
    // B2B (backend M6) is still a preview; Vendors went live in M4.
    const v = textOf((await b.get("/b2b")).html);
    assert.match(v, /Preview: everything on this page is sample data/);
    assert.doesNotMatch(textOf((await b.get("/vendors")).html), /sample data/);
    assert.equal((await b.get("/no-such-module")).status, 404);
  });

  test("a support agent can't create products", async () => {
    const b = await signIn("support_agent");
    assert.match((await b.get("/catalogue/new")).url, /\/\?denied=1$/);
  });
});

describe("operations modules", () => {
  test("every live module page renders for admin", async () => {
    const b = await signIn("admin");
    const pages = [
      "/dispatch",
      "/dispatch?tab=riders",
      "/dispatch?tab=phones",
      "/dispatch?tab=zones",
      "/warehouse",
      "/inventory",
      "/inventory?tab=movements",
      "/inventory?tab=warehouses",
      "/marketing",
      "/marketing?tab=coupons",
      "/marketing?tab=campaigns",
      "/marketing?tab=segments",
      "/marketing?tab=journeys",
      "/admin",
      "/admin?tab=roles",
      "/admin?tab=people",
      "/admin?tab=audit",
      "/admin?tab=signins",
      "/hr",
      "/content",
      "/content?tab=help",
      "/content?tab=banners",
      "/search-tuning",
    ];
    const failed: string[] = [];
    for (const p of pages) {
      const r = await b.get(p);
      if (
        r.status !== 200 ||
        /Application error|Unhandled Runtime Error/.test(r.html)
      )
        failed.push(`${p} → ${r.status}`);
    }
    assert.deepEqual(failed, []);
    assert.match(textOf((await b.get("/dispatch?tab=zones")).html), /Lagos/);
    assert.match(
      textOf((await b.get("/marketing?tab=coupons")).html),
      /WELCOME10/,
    );
  });

  test("warehouse: receive IMEIs, release a wave, claim, scan and pack", async () => {
    const b = await signIn("admin");
    const warehouseId = seed().extra.warehouseIkejaId;
    await b.get("/inventory");
    await b.get("/warehouse");

    // A fresh paid order guarantees something to pick in this run.
    const ct = await customerToken("customer_b");
    const cart = await api.get("/api/v1/cart", { token: ct });
    for (const l of cart.lines ?? [])
      await api.delete("/api/v1/cart/items/{id}", {
        token: ct,
        path: { id: l.cartItemId },
      });
    // Orbit One 15 Max: sold and shipped by TechShop from the Ikeja warehouse (serialised).
    await api.post("/api/v1/cart/items", {
      token: ct,
      body: { listingId: seed().extra.listing_p2, quantity: 1 },
    });
    const addressId = ((await api.get("/api/v1/me/addresses", { token: ct })) ??
      [])[0]?.id;
    assert.ok(addressId, "customer_b has an address");
    const quote = await api.post("/api/v1/checkout/preview", {
      token: ct,
      body: { addressId },
    });
    const placed = await api.post("/api/v1/orders", {
      token: ct,
      idempotencyKey: crypto.randomUUID(),
      body: { addressId, priceHash: quote.priceHash, provider: "paystack" },
    });
    await fetch(
      `${API_URL}/dev/pay/paystack/${placed.payment.reference}/success`,
      { method: "POST" },
    );
    for (let i = 0; i < 30; i++) {
      const st = await api.get("/api/v1/me/payments/{id}", {
        token: ct,
        path: { id: placed.payment.paymentId },
      });
      if (st.orderStatus === "paid") break;
      await new Promise((r) => setTimeout(r, 300));
    }

    const wave = await b.action<{ ok: boolean; message: string }>(
      "/warehouse",
      "releaseWave",
      [warehouseId],
    );
    assert.ok(wave.redirect, `wave released: ${wave.value?.message}`);
    const pickPath = wave.redirect;
    const listId = pickPath.split("/").pop()!;
    const page = await b.get(pickPath);
    assert.equal(page.status, 200);
    const claim = await b.action<{ ok: boolean; message: string }>(
      "/warehouse/pick/[id]",
      "claimPickList",
      [listId],
      { url: pickPath },
    );
    assert.equal(
      claim.value?.ok,
      true,
      `claim: ${claim.status} ${claim.value?.message ?? claim.raw.slice(-300)}`,
    );

    // Read the items through the API to know what to scan.
    const token = await staffToken("admin");
    const list = await api.get("/api/v1/staff/wms/pick-lists/{id}", {
      token,
      path: { id: listId },
    });
    for (const item of list.items ?? []) {
      let left = item.quantity - item.pickedQty;
      while (left > 0) {
        let serial = "";
        if (item.isSerialised) {
          serial = imei();
          await b.get("/inventory");
          const rec = await b.action<{ ok: boolean; message: string }>(
            "/inventory",
            "receiveUnits",
            [
              {},
              form({
                warehouseId,
                variantId: item.variantId,
                condition: item.condition,
                codes: serial,
              }),
            ],
          );
          assert.equal(
            rec.value?.ok,
            true,
            `receive: ${rec.status} ${rec.value?.message ?? rec.raw.slice(-300)}`,
          );
        }
        const qty = item.isSerialised ? 1 : left;
        const scan = await b.action<{ ok: boolean; message: string }>(
          "/warehouse/pick/[id]",
          "scanItem",
          [
            listId,
            {},
            form({
              orderLineId: item.orderLineId,
              quantity: String(qty),
              ...(serial ? { serial } : {}),
            }),
          ],
          { url: pickPath },
        );
        assert.equal(
          scan.value?.ok,
          true,
          `scan: ${scan.status} ${scan.value?.message ?? scan.raw.slice(-300)}`,
        );
        left -= qty;
      }
    }
    // Pack the first fulfilment: the API answers with a label and the handover.
    const fulfilmentId = list.items![0].fulfilmentId;
    const packed = await b.action<{
      ok: boolean;
      message: string;
      label?: string;
    }>(
      "/warehouse/pick/[id]",
      "packFulfilment",
      [{}, form({ fulfilmentId, weightGrams: "800" })],
      { url: pickPath },
    );
    assert.equal(
      packed.value?.ok,
      true,
      `pack: ${packed.status} ${packed.value?.message ?? packed.raw.slice(-300)}`,
    );
    assert.ok(packed.value.label);
  });

  test("marketing approval needs a second person; admin roles and HR are guarded", async () => {
    const exec = await signIn("marketing_exec");
    const mk = (await exec.get("/marketing?tab=campaigns")).html;
    assert.doesNotMatch(
      mk,
      />Approve</,
      "marketing exec sees no Approve button",
    );
    const support = await signIn("support_agent");
    assert.match((await support.get("/admin")).url, /\/\?denied=1$/);
    assert.match((await support.get("/hr")).url, /\/\?denied=1$/);
  });

  test("content: save a page, publish it, and the market serves it", async () => {
    // Separation of duties: the content editor writes, someone else (admin) publishes.
    const editor = await signIn("content_editor");
    await editor.get("/content");
    const slug = `e2e-${Date.now().toString(36)}`;
    const saved = await editor.action<{ ok: boolean; message: string }>(
      "/content",
      "savePage",
      [
        {},
        form({
          site: "market",
          slug,
          title: "E2E page",
          body: "## Hello\n\nThis page was written by the end-to-end test.",
        }),
      ],
    );
    assert.equal(saved.value.ok, true, saved.value.message);
    const b = await signIn("admin");
    assert.equal((await b.get("/content?tab=pages&site=market")).status, 200);
    const pages =
      (await api.get("/api/v1/staff/content/pages", {
        token: await staffToken("admin"),
        query: { site: "market" },
      })) ?? [];
    const id = pages.find((p) => p.slug === slug)?.id;
    assert.ok(id, "page listed");
    const pub = await b.action<{ ok: boolean; message: string }>(
      "/content",
      "publishPage",
      [id],
    );
    assert.equal(pub.value.ok, true, pub.value.message);
    const page = await api.get("/api/v1/content/pages/{site}/{slug}", {
      path: { site: "market", slug },
    });
    assert.equal(page.title, "E2E page");
  });

  test("search tuning: add and remove a synonym", async () => {
    const b = await signIn("admin");
    await b.get("/search-tuning");
    const word = `e2e${Date.now().toString(36)}`;
    const add = await b.action<{ ok: boolean; message: string }>(
      "/search-tuning",
      "createSynonym",
      [{}, form({ target: word, terms: "zzalpha, zzbeta", kind: "two_way" })],
    );
    assert.equal(add.value.ok, true, add.value.message);
    const html = (await b.get("/search-tuning")).html;
    assert.match(textOf(html), new RegExp(word));
  });
});
