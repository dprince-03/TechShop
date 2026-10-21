// End-to-end tests for the market app: real pages and Server Actions over HTTP against
// `next dev` (MARKET_URL) and the seeded local API (API_URL, SEED_FILE).
import assert from "node:assert/strict";
import { randomInt } from "node:crypto";
import path from "node:path";
import { describe, test } from "node:test";
import { fileURLToPath } from "node:url";

import {
  API_URL,
  api,
  account,
  deliveredOrder,
  form,
  latestCode,
  peekCode,
  seed,
  textOf,
  tokenForPhone,
  WebClient,
} from "@techshop/testing";

const APP_DIR = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const BASE = process.env.MARKET_URL ?? "http://localhost:4102";
const browser = () => new WebClient(BASE, APP_DIR);
const listing = (n: number) => seed().extra[`listing_p${n}`];

type CartResult = { ok: boolean; count?: number; message: string };

/** Signs a seeded customer in with email + password through the real sign-in form action. */
async function signInWithPassword(b: WebClient, role: string, next = "/") {
  await b.get(`/sign-in?next=${encodeURIComponent(next)}`);
  const a = account(role);
  return b.action<{ step: string }>("/sign-in", "signInStep", [
    { step: "start" },
    form({
      intent: "password",
      identifier: a.email,
      password: a.password,
      next,
    }),
  ]);
}

const freshPhone = () => `+23480${randomInt(10_000_000, 99_999_999)}`;

/** Signs a browser in with phone OTP through the real sign-in form action. */
async function signIn(b: WebClient, phone: string, next = "/") {
  await b.get(`/sign-in?next=${encodeURIComponent(next)}`);
  const before = await peekCode(phone);
  const step1 = await b.action<{ step: string }>("/sign-in", "signInStep", [
    { step: "start" },
    form({ intent: "request-code", phone, next }),
  ]);
  assert.equal(step1.value.step, "code");
  const code = await latestCode(phone, before);
  return b.action<{ step: string }>("/sign-in", "signInStep", [
    { step: "code" },
    form({ intent: "verify-code", code, next }),
  ]);
}

describe("catalogue pages", () => {
  test("home shows live categories and products", async () => {
    const r = await browser().get("/");
    assert.equal(r.status, 200);
    const text = textOf(r.html);
    assert.match(text, /Shop by category/);
    assert.match(text, /Phones/);
    assert.match(text, /₦[\d,]+/);
    assert.doesNotMatch(
      text,
      /Preview: products, sellers, prices and figures on this page are sample data/,
    );
  });

  test("category, subcategory, search with filters, deals, best sellers", async () => {
    const b = browser();
    for (const p of [
      "/c/phones",
      "/c/phones/android",
      "/search?q=phone&sort=price_asc",
      "/search?q=nova&condition=new&inStock=1",
      "/deals",
      "/best-sellers",
    ]) {
      const r = await b.get(p);
      assert.equal(r.status, 200, p);
      assert.match(textOf(r.html), /\d+ products?/, p);
    }
  });

  test("product page shows the API's offer, specs and buy box", async () => {
    const r = await browser().get("/p/orbit-one-15-max");
    assert.equal(r.status, 200);
    const text = textOf(r.html);
    assert.match(text, /Orbit One 15 Max/);
    assert.match(text, /₦2,150,000/);
    assert.match(text, /Sold by TechShop/);
    assert.match(text, /Add to cart/);
  });

  test("unknown product and category are 404", async () => {
    assert.equal((await browser().get("/p/no-such-product")).status, 404);
    assert.equal((await browser().get("/c/no-such-category")).status, 404);
  });

  test("help centre lists API articles and an article renders", async () => {
    const b = browser();
    assert.match(
      textOf((await b.get("/help")).html),
      /How do I track my order\?/,
    );
    const a = await b.get("/help/articles/track-order");
    assert.equal(a.status, 200);
    assert.match(textOf(a.html), /Use Track order with your order number\./);
  });
});

describe("guest cart", () => {
  test("add → cart page → change quantity → remove", async () => {
    const b = browser();
    await b.get("/p/orbit-one-15-max");
    const add = await b.action<CartResult>(
      "/p/[slug]",
      "addToCart",
      [listing(2), 1],
      { url: "/p/orbit-one-15-max" },
    );
    assert.equal(add.value.ok, true, add.value.message);
    assert.ok(b.cookie("ts_cart"), "guest cart cookie set");

    let cart = await b.get("/cart");
    assert.match(cart.html, /aria-label="Cart, 1 item"/);
    // The RSC payload escapes quotes inside the HTML, so allow an optional backslash.
    const itemId = cart.html.match(/cartItemId\\?":\\?"([0-9a-f-]{36})/)?.[1];
    assert.ok(itemId, "cart item id rendered");

    const upd = await b.action<CartResult>("/cart", "setQuantity", [itemId, 3]);
    assert.equal(upd.value.ok, true);
    assert.equal(upd.value.count, 3);
    const bad = await b.action<CartResult>("/cart", "setQuantity", [itemId, 0]);
    assert.equal(bad.value.ok, false);

    const del = await b.action<CartResult>("/cart", "removeItem", [itemId]);
    assert.equal(del.value.count, 0);
    cart = await b.get("/cart");
    assert.match(textOf(cart.html), /Your cart is empty/);
  });

  test("checkout asks a guest to sign in first", async () => {
    const r = await browser().get("/checkout");
    assert.match(r.url, /\/sign-in\?next=%2Fcheckout/);
  });
});

describe("sign-in, account and checkout", () => {
  test("sign-in merges the guest cart, then account, checkout, payment and order", async () => {
    const b = browser();
    const phone = account("customer").phone!;
    await b.get("/p/orbit-one-15-max");
    const add = await b.action<CartResult>(
      "/p/[slug]",
      "addToCart",
      [listing(2), 1],
      { url: "/p/orbit-one-15-max" },
    );
    assert.equal(add.value.ok, true);

    // Sign in: the session cookie is set and the guest cart merges.
    const done = await signInWithPassword(b, "customer", "/cart");
    assert.equal(done.redirect, "/cart");
    assert.ok(b.cookie("ts_session"), "session cookie");
    assert.ok(!b.cookie("ts_auth_flow"), "sign-in flow cookie cleared");
    const cart = await b.get("/cart");
    const cartText = textOf(cart.html);
    assert.match(cartText, /Account/);
    assert.match(cartText, /Checkout/);

    // Account: profile and a new address.
    const acct = await b.get("/account");
    assert.equal(acct.status, 200);
    assert.match(textOf(acct.html), /Signed-in devices/);
    const prof = await b.action<{ ok: boolean }>("/account", "updateProfile", [
      {},
      form({ firstName: "Ada", lastName: "Customer" }),
    ]);
    assert.equal(prof.value.ok, true);
    const badAddr = await b.action<{
      ok: boolean;
      errors?: Record<string, string>;
    }>("/account", "saveAddress", [
      {},
      form({
        recipientName: "",
        phone: "123",
        line1: "",
        city: "",
        stateCode: "",
      }),
    ]);
    assert.equal(badAddr.value.ok, false);
    assert.ok(badAddr.value.errors?.phone && badAddr.value.errors?.stateCode);
    const addr = await b.action<{ ok: boolean; message: string }>(
      "/account",
      "saveAddress",
      [
        {},
        form({
          recipientName: "Ada Customer",
          phone: "0803 123 4567",
          line1: "5 Awolowo Road",
          city: "Ikoyi",
          stateCode: "LA",
          lga: "Eti-Osa",
          isDefault: "on",
        }),
      ],
    );
    assert.equal(addr.value.ok, true, addr.value.message);

    // Checkout: the page opens with a real quote for the default address.
    const co = await b.get("/checkout");
    assert.equal(co.status, 200);
    const coText = textOf(co.html);
    assert.match(coText, /Pay ₦[\d,]+/);
    const addressId =
      co.html.match(/name="address" value="([0-9a-f-]{36})"/)?.[1] ??
      co.html
        .match(
          /"isDefault":true[^}]*"id":"([0-9a-f-]{36})"|"id":"([0-9a-f-]{36})"[^}]*"isDefault":true/,
        )
        ?.slice(1)
        .find(Boolean);
    assert.ok(addressId, "address id on page");

    const quote = await b.action<{
      ok: boolean;
      quote: { priceHash: string; totalKobo: number };
    }>("/checkout", "previewCheckout", [addressId, ""]);
    assert.equal(quote.value.ok, true);
    const badCoupon = await b.action<{ ok: boolean; field?: string }>(
      "/checkout",
      "previewCheckout",
      [addressId, "NOSUCHCODE"],
    );
    assert.equal(badCoupon.value.ok, false);

    const placed = await b.action("/checkout", "placeOrder", [
      {
        addressId,
        coupon: "",
        priceHash: quote.value.quote.priceHash,
        provider: "paystack",
        idempotencyKey: crypto.randomUUID(),
      },
    ]);
    assert.ok(
      placed.redirect?.includes("/dev/pay/paystack/"),
      `redirected to provider: ${placed.redirect}`,
    );

    // Pay on the fake provider page, then the order page shows it as paid.
    const ref = placed.redirect!.split("/").pop()!;
    assert.equal(
      (
        await fetch(`${API_URL}/dev/pay/paystack/${ref}/success`, {
          method: "POST",
        })
      ).status,
      200,
    );
    const orders = await b.get("/orders");
    const number = textOf(orders.html).match(/TS-\d+/)?.[0];
    assert.ok(number, "order listed");
    let detail = "";
    for (let i = 0; i < 30 && !/Payment confirmed/.test(detail); i++) {
      detail = textOf((await b.get(`/orders/${number}?placed=1`)).html);
      if (!/Payment confirmed/.test(detail))
        await new Promise((r) => setTimeout(r, 300));
    }
    assert.match(detail, /Payment confirmed/);
    assert.match(detail, /Card \(Paystack\)\s*:\s*Paid/);

    // Guest tracking with the checkout phone finds it; a wrong phone doesn't.
    const guest = browser();
    await guest.get("/orders");
    const found = await guest.action<{ state: string }>(
      "/orders",
      "trackOrder",
      [{ state: "idle" }, form({ number, phone })],
    );
    assert.equal(found.value.state, "found");
    const wrong = await guest.action<{ state: string }>(
      "/orders",
      "trackOrder",
      [{ state: "idle" }, form({ number, phone: "0809 999 9999" })],
    );
    assert.equal(wrong.value.state, "error");

    // Another customer can't open the order.
    const other = browser();
    await signInWithPassword(other, "customer_b");
    assert.equal((await other.get(`/orders/${number}`)).status, 404);

    // Sign out revokes the session and clears the cookie.
    await b.get("/account");
    const out = await b.action("/account", "signOut", []);
    assert.equal(out.redirect, "/");
    assert.ok(!b.cookie("ts_session"));
    assert.match((await b.get("/account")).url, /\/sign-in/);
  });

  test("new phone number: OTP → profile → signed in", async () => {
    const b = browser();
    const r = await signIn(b, freshPhone());
    assert.equal(r.value.step, "profile");
    const done = await b.action("/sign-in", "signInStep", [
      { step: "profile" },
      form({ intent: "profile", firstName: "Tunde", lastName: "New" }),
    ]);
    assert.equal(done.redirect, "/");
    assert.match(textOf((await b.get("/account")).html), /Hello, Tunde/);
  });

  test("password sign-in and a wrong password", async () => {
    const b = browser();
    await b.get("/sign-in");
    const a = account("customer_b");
    const bad = await b.action<{ step: string; error?: string }>(
      "/sign-in",
      "signInStep",
      [
        { step: "start" },
        form({
          intent: "password",
          identifier: a.email,
          password: "not-the-password",
        }),
      ],
    );
    assert.equal(bad.value.step, "start");
    assert.match(bad.value.error ?? "", /don’t match/);
    const ok = await b.action("/sign-in", "signInStep", [
      { step: "start" },
      form({ intent: "password", identifier: a.email, password: a.password }),
    ]);
    assert.equal(ok.redirect, "/");
  });
});

describe("security", () => {
  test("a sign-in return path can't point off-site", async () => {
    const b = browser();
    await b.get("/sign-in");
    const phone = freshPhone();
    const r = await signIn(b, phone, "//evil.example/steal");
    assert.equal(r.value.step, "profile");
    const done = await b.action("/sign-in", "signInStep", [
      { step: "profile" },
      form({ intent: "profile", firstName: "Safe", lastName: "Redirect" }),
    ]);
    assert.equal(done.redirect, "/");
  });

  test("Server Actions refuse a cross-site Origin", async () => {
    const b = browser();
    await b.get("/p/orbit-one-15-max");
    const r = await b.action<CartResult>(
      "/p/[slug]",
      "addToCart",
      [listing(2), 1],
      { url: "/p/orbit-one-15-max", origin: "https://evil.example" },
    );
    assert.notEqual(r.value?.ok, true);
    assert.ok(!b.cookie("ts_cart"), "no cart created");
  });

  test("a tampered session cookie is ignored and cleared", async () => {
    const r = await fetch(`${BASE}/account`, {
      headers: { cookie: "ts_session=v1.tampered" },
      redirect: "manual",
    });
    assert.ok([302, 303, 307].includes(r.status) || r.status === 200);
    const loc = r.headers.get("location") ?? "";
    assert.match(loc + (await r.text()), /sign-in/);
  });

  test("session cookie is httpOnly and SameSite=Lax", async () => {
    const b = browser();
    await b.get("/sign-in");
    const phone = freshPhone();
    const before = await peekCode(phone);
    await b.action("/sign-in", "signInStep", [
      { step: "start" },
      form({ intent: "request-code", phone }),
    ]);
    const code = await latestCode(phone, before);
    const id = b.actionId("/sign-in", "signInStep");
    const body = new FormData();
    body.set("_1_intent", "verify-code");
    body.set("_1_code", code);
    body.set("0", JSON.stringify([{ step: "code" }, "$K1"]));
    const res = await fetch(`${BASE}/sign-in`, {
      method: "POST",
      headers: {
        "Next-Action": id,
        Accept: "text/x-component",
        Origin: BASE,
        "x-real-ip": b.ip,
        cookie: `ts_auth_flow=${b.cookie("ts_auth_flow")}`,
      },
      body,
      redirect: "manual",
    });
    // A new number continues to the profile step, so complete it and read the session cookie there.
    const flow = res.headers
      .getSetCookie()
      .find((c) => c.startsWith("ts_auth_flow="))
      ?.split(";")[0];
    const body2 = new FormData();
    body2.set("_1_intent", "profile");
    body2.set("_1_firstName", "Cookie");
    body2.set("_1_lastName", "Check");
    body2.set("0", JSON.stringify([{ step: "profile" }, "$K1"]));
    const res2 = await fetch(`${BASE}/sign-in`, {
      method: "POST",
      headers: {
        "Next-Action": id,
        Accept: "text/x-component",
        Origin: BASE,
        "x-real-ip": b.ip,
        cookie: flow ?? "",
      },
      body: body2,
      redirect: "manual",
    });
    const set = res2.headers
      .getSetCookie()
      .find((c) => c.startsWith("ts_session="));
    assert.ok(set, "session cookie set");
    assert.match(set!, /HttpOnly/i);
    assert.match(set!, /SameSite=lax/i);
    assert.match(set!, /Path=\//);
    assert.doesNotMatch(set!, /Domain=/i);
  });
});

describe("content from the CMS", () => {
  test("legal page shows the outline, then the published text once staff publish it", async () => {
    const { staffToken } = await import("@techshop/testing");
    const before = textOf((await browser().get("/legal/cookies")).html);
    const marker = `E2E cookie policy ${Date.now().toString(36)}`;
    const editor = await staffToken("content_editor");
    const saved = await api.put("/api/v1/staff/content/pages/{site}/{slug}", {
      token: editor,
      path: { site: "market", slug: "cookies" },
      body: {
        title: "Cookie policy",
        blocks: [{ type: "paragraph", text: marker }],
      },
    });
    // Before publishing, the public page still shows the outline (drafts never leak).
    assert.doesNotMatch(
      textOf((await browser().get("/legal/cookies")).html),
      new RegExp(marker),
    );
    await api.post("/api/v1/staff/content/pages/{id}/publish", {
      token: await staffToken("admin"),
      path: { id: saved.id },
    });
    let after = "";
    for (let i = 0; i < 20 && !after.includes(marker); i++) {
      after = textOf((await browser().get("/legal/cookies")).html);
      if (!after.includes(marker))
        await new Promise((r) => setTimeout(r, 3500));
    }
    assert.match(after, new RegExp(marker));
    assert.notEqual(before, after);
  });
});

describe("reviews", () => {
  test("review a delivered item, rate the seller, see it on the product page and report one", async () => {
    // A fresh customer with a delivered TechShop order (fake carrier), signed in on the web too.
    const phone = freshPhone();
    const token = await tokenForPhone(phone, "market", "Chidi");
    const { orderNumber, fulfilmentId, lineId } = await deliveredOrder(
      token,
      listing(21),
    );
    // The API allows one code per number per minute, so the web session signs in with a
    // password set through the API instead of a second code.
    const password = `Pw-${randomInt(1e9, 9e9)}-${Date.now().toString(36)}`;
    await api.post("/api/v1/me/password", {
      token,
      body: { newPassword: password },
    });
    const b = browser();
    const next = `/orders/${orderNumber}`;
    await b.get(`/sign-in?next=${encodeURIComponent(next)}`);
    const signed = await b.action("/sign-in", "signInStep", [
      { step: "start" },
      form({ intent: "password", identifier: phone, password, next }),
    ]);
    assert.equal(signed.redirect, next, signed.raw.slice(-300));

    const orderPage = textOf((await b.get(`/orders/${orderNumber}`)).html);
    assert.match(orderPage, /Review/);
    assert.match(orderPage, /Rate TechShop/);

    const reviewPath = `/orders/${orderNumber}/review/${lineId}`;
    assert.match(textOf((await b.get(reviewPath)).html), /Write a review/);
    const noStars = await b.action<{
      ok: boolean;
      errors?: Record<string, string>;
    }>(
      "/orders/[number]/review/[lineId]",
      "reviewItem",
      [lineId, {}, form({ rating: "" })],
      { url: reviewPath },
    );
    assert.equal(noStars.value.ok, false);
    assert.ok(noStars.value.errors?.rating);
    const marker = `Bright bulbs ${Date.now().toString(36)}`;
    const posted = await b.action<{ ok: boolean; message: string }>(
      "/orders/[number]/review/[lineId]",
      "reviewItem",
      [
        lineId,
        {},
        form({
          rating: "5",
          title: "Great",
          body: `${marker}, easy to pair with the app.`,
        }),
      ],
      { url: reviewPath },
    );
    assert.equal(posted.value.ok, true, posted.value.message);
    assert.match(posted.value.message, /live/);

    const rated = await b.action<{ ok: boolean; message: string }>(
      "/orders/[number]",
      "rateSeller",
      [fulfilmentId, {}, form({ rating: "4", comment: "Quick" })],
      { url: `/orders/${orderNumber}` },
    );
    assert.equal(rated.value.ok, true, rated.value.message);
    const again = await b.action<{ ok: boolean; message: string }>(
      "/orders/[number]",
      "rateSeller",
      [fulfilmentId, {}, form({ rating: "4" })],
      { url: `/orders/${orderNumber}` },
    );
    assert.equal(again.value.ok, false);
    assert.match(again.value.message, /already rated/);

    // The review shows on the product page as a verified purchase.
    const product = textOf(
      (await browser().get("/p/lumen-smart-bulb-4-pack")).html,
    );
    assert.match(product, new RegExp(marker));
    assert.match(product, /Verified purchase/);

    // Someone else reports it; once is enough.
    const other = browser();
    const step = await signIn(
      other,
      freshPhone(),
      "/p/lumen-smart-bulb-4-pack",
    );
    assert.equal(step.value.step, "profile"); // a new number: finish sign-up
    await other.action("/sign-in", "signInStep", [
      { step: "profile" },
      form({
        intent: "profile",
        firstName: "Eze",
        lastName: "Reporter",
        next: "/p/lumen-smart-bulb-4-pack",
      }),
    ]);
    const reviews = await api.get("/api/v1/products/{slug}/reviews", {
      path: { slug: "lumen-smart-bulb-4-pack" },
    });
    const mine = (reviews.reviews.items ?? []).find((r) =>
      (r.body ?? "").includes(marker),
    )!;
    await other.get("/p/lumen-smart-bulb-4-pack");
    const report = await other.action<{ ok: boolean; message: string }>(
      "/p/[slug]",
      "reportReview",
      [mine.id, {}, form({ reason: "fake" })],
      { url: "/p/lumen-smart-bulb-4-pack" },
    );
    assert.equal(report.value.ok, true, report.value.message);
    const twice = await other.action<{ ok: boolean; message: string }>(
      "/p/[slug]",
      "reportReview",
      [mine.id, {}, form({ reason: "fake" })],
      { url: "/p/lumen-smart-bulb-4-pack" },
    );
    assert.equal(twice.value.ok, false);
  });
});
