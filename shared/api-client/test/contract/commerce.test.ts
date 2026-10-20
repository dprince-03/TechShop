// Contract tests: cart, checkout, order placement, fake payment, orders and tracking.
import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { describe, test } from "node:test";

import { ApiError } from "../../src/index";
import { account, api, API_URL, customerToken, seed } from "../kit";

const listing = (n: number) => seed().extra[`listing_p${n}`];

/** A listing with stock that has no flash price (flash prices need sign-in and have limits). */
const plainListing = () => listing(2);

async function addressFor(token: string) {
  const list = (await api.get("/api/v1/me/addresses", { token })) ?? [];
  const lagos = list.find((a) => a.stateCode === "LA");
  if (lagos) return lagos.id;
  const a = await api.post("/api/v1/me/addresses", {
    token,
    body: { recipientName: "Ada Test", phone: "+2348031234567", line1: "12 Allen Avenue", city: "Ikeja", stateCode: "LA", lga: "Ikeja" },
  });
  return a.id;
}

/** A well-formed hash that doesn't match the current price. */
const stale = (hash: string) => hash.slice(0, -1) + (hash.endsWith("0") ? "1" : "0");

async function emptyCart(token: string) {
  const cart = await api.get("/api/v1/cart", { token });
  for (const l of cart.lines ?? []) await api.delete("/api/v1/cart/items/{id}", { token, path: { id: l.cartItemId } });
}

describe("guest cart", () => {
  test("add issues a cart token; get, update, remove work with it", async () => {
    const r = await api.request.post("/api/v1/cart/items", { body: { listingId: plainListing(), quantity: 1 } });
    const token = r.headers.get("x-cart-token") ?? r.data.cartToken;
    assert.ok(token, "guest cart token");
    assert.equal(r.data.lines?.length, 1);
    const headers = { "X-Cart-Token": token };
    const got = await api.get("/api/v1/cart", { headers });
    assert.equal(got.lines?.[0].quantity, 1);
    const line = got.lines![0].cartItemId;
    const upd = await api.patch("/api/v1/cart/items/{id}", { headers, path: { id: line }, body: { quantity: 2 } });
    assert.equal(upd.lines?.[0].quantity, 2);
    assert.equal(upd.itemsKobo, upd.lines![0].unitPriceKobo * 2);
    const del = await api.delete("/api/v1/cart/items/{id}", { headers, path: { id: line } });
    assert.equal(del.lines?.length ?? 0, 0);
  });

  test("a different cart token can't see the cart", async () => {
    const r = await api.request.post("/api/v1/cart/items", { body: { listingId: plainListing(), quantity: 1 } });
    const other = await api.get("/api/v1/cart", { headers: { "X-Cart-Token": "not-a-real-token-000000000000" } }).catch((e) => e);
    if (other instanceof ApiError) assert.ok(other.status >= 400);
    else assert.equal(other.lines?.length ?? 0, 0);
    assert.ok(r.data.lines?.length);
  });

  test("bad quantity and unknown listing are refused", async () => {
    await assert.rejects(api.post("/api/v1/cart/items", { body: { listingId: plainListing(), quantity: 0 } }), (e: unknown) => e instanceof ApiError && e.status === 422);
    await assert.rejects(api.post("/api/v1/cart/items", { body: { listingId: randomUUID(), quantity: 1 } }), (e: unknown) => e instanceof ApiError && e.status >= 400 && e.status < 500);
  });

  test("guest cart merges into the customer's cart on sign-in", async () => {
    const token = await customerToken("customer_b");
    await emptyCart(token);
    const r = await api.request.post("/api/v1/cart/items", { body: { listingId: plainListing(), quantity: 1 } });
    const cartToken = r.headers.get("x-cart-token") ?? r.data.cartToken!;
    const merged = await api.post("/api/v1/cart/merge", { token, body: { cartToken } });
    assert.ok((merged.lines ?? []).some((l) => l.listingId === plainListing()));
    await emptyCart(token);
  });
});

describe("checkout and orders", () => {
  test("payment methods are listed for signed-in customers", async () => {
    await assert.rejects(api.get("/api/v1/checkout/payment-methods"), (e: unknown) => e instanceof ApiError && e.status === 401);
    const m = (await api.get("/api/v1/checkout/payment-methods", { token: await customerToken() })) ?? [];
    assert.ok(m.includes("paystack"));
  });

  test("preview → place (idempotent) → fake payment → paid → in my orders", async () => {
    const token = await customerToken();
    await emptyCart(token);
    await api.post("/api/v1/cart/items", { token, body: { listingId: plainListing(), quantity: 1 } });
    const addressId = await addressFor(token);
    const quote = await api.post("/api/v1/checkout/preview", { token, body: { addressId } });
    assert.equal(quote.purchasable, true);
    assert.ok(quote.totalKobo > 0 && quote.deliveryKobo >= 0);

    await assert.rejects(
      api.post("/api/v1/orders", { token, idempotencyKey: randomUUID(), body: { addressId, priceHash: stale(quote.priceHash), provider: "paystack" } }),
      (e: unknown) => e instanceof ApiError && e.status === 409 && e.code === "price_changed",
    );

    const key = randomUUID();
    const body = { addressId, priceHash: quote.priceHash, provider: "paystack" as const };
    const placed = await api.post("/api/v1/orders", { token, idempotencyKey: key, body });
    const again = await api.request.post("/api/v1/orders", { token, idempotencyKey: key, body });
    assert.equal(again.data.orderNumber, placed.orderNumber);
    assert.equal(placed.totalKobo, quote.totalKobo);
    assert.ok(placed.payment.checkoutUrl?.includes("/dev/pay/"));

    const ref = placed.payment.reference!;
    const pay = await fetch(`${API_URL}/dev/pay/paystack/${ref}/success`, { method: "POST" });
    assert.equal(pay.status, 200);

    let status = "";
    for (let i = 0; i < 40 && status !== "paid"; i++) {
      const s = await api.get("/api/v1/me/payments/{id}", { token, path: { id: placed.payment.paymentId } });
      status = s.orderStatus;
      if (status !== "paid") await new Promise((r) => setTimeout(r, 300));
    }
    assert.equal(status, "paid");

    const detail = await api.get("/api/v1/me/orders/{number}", { token, path: { number: placed.orderNumber } });
    assert.equal(detail.status, "paid");
    assert.ok((detail.fulfilments ?? []).length > 0);
    const mine = await api.get("/api/v1/me/orders", { token, query: { limit: 10 } });
    assert.ok((mine.items ?? []).some((o) => o.orderNumber === placed.orderNumber));

    // Another customer can't read it.
    const other = await customerToken("customer_b");
    await assert.rejects(api.get("/api/v1/me/orders/{number}", { token: other, path: { number: placed.orderNumber } }), (e: unknown) => e instanceof ApiError && e.status === 404);

    // Guest tracking needs the right phone.
    const phone = account("customer").phone!;
    const t = await api.get("/api/v1/track/{number}", { path: { number: placed.orderNumber }, query: { phone } });
    assert.equal(t.orderNumber, placed.orderNumber);
    await assert.rejects(api.get("/api/v1/track/{number}", { path: { number: placed.orderNumber }, query: { phone: "+2348099999999" } }), (e: unknown) => e instanceof ApiError && e.status === 404);
  });

  test("an unpaid order can be cancelled by its owner", async () => {
    const token = await customerToken();
    await emptyCart(token);
    await api.post("/api/v1/cart/items", { token, body: { listingId: plainListing(), quantity: 1 } });
    const addressId = await addressFor(token);
    const quote = await api.post("/api/v1/checkout/preview", { token, body: { addressId } });
    const placed = await api.post("/api/v1/orders", { token, idempotencyKey: randomUUID(), body: { addressId, priceHash: quote.priceHash, provider: "paystack" } });
    const cancelled = await api.post("/api/v1/me/orders/{number}/cancel", { token, path: { number: placed.orderNumber }, body: { reason: "Changed my mind" } });
    assert.equal(cancelled.status, "cancelled");
  });

  test("placing an order needs a signed-in customer and an idempotency key", async () => {
    await assert.rejects(api.post("/api/v1/orders", { idempotencyKey: randomUUID(), body: { addressId: randomUUID(), priceHash: "x", provider: "paystack" } }), (e: unknown) => e instanceof ApiError && e.status === 401);
    const token = await customerToken();
    const r = await fetch(`${API_URL}/api/v1/orders`, { method: "POST", headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }, body: JSON.stringify({ addressId: randomUUID(), priceHash: "x", provider: "paystack" }) });
    assert.ok(r.status === 400 || r.status === 422, `missing key → ${r.status}`);
  });
});
