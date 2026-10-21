// End-to-end tests for Seller Centre: real pages and Server Actions over HTTP against `next dev`
// (SELLER_URL) and the seeded local API (API_URL, SEED_FILE). Staff decisions (KYC approval,
// listing moderation) are made through the API, as the staff app would.
import assert from "node:assert/strict";
import { randomInt } from "node:crypto";
import path from "node:path";
import { describe, test } from "node:test";
import { fileURLToPath } from "node:url";

import {
  api,
  form,
  latestCode,
  peekCode,
  placePaidOrder,
  seed,
  staffToken,
  textOf,
  WebClient,
} from "@techshop/testing";

const APP_DIR = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const BASE = process.env.SELLER_URL ?? "http://localhost:4104";
const browser = () => new WebClient(BASE, APP_DIR);
const freshPhone = () => `+23481${randomInt(10_000_000, 99_999_999)}`;

type Result = {
  ok?: boolean;
  message?: string;
  errors?: Record<string, string>;
};

/** A new seller signs up with phone OTP through the real sign-in form. */
async function signUp(
  b: WebClient,
  first: string,
  last: string,
  next = "/centre",
) {
  const phone = freshPhone();
  await b.get(`/sign-in?next=${encodeURIComponent(next)}`);
  const before = await peekCode(phone);
  const s1 = await b.action<{ step: string }>("/sign-in", "signInStep", [
    { step: "start" },
    form({ intent: "request-code", phone, next }),
  ]);
  assert.equal(s1.value.step, "code", s1.raw.slice(-300));
  const code = await latestCode(phone, before);
  const s2 = await b.action<{ step: string }>("/sign-in", "signInStep", [
    { step: "code" },
    form({ intent: "verify-code", code, next }),
  ]);
  assert.equal(s2.value.step, "profile", s2.raw.slice(-300));
  const s3 = await b.action("/sign-in", "signInStep", [
    { step: "profile" },
    form({ intent: "profile", firstName: first, lastName: last, next }),
  ]);
  assert.equal(
    s3.redirect,
    next,
    `signup should land on ${next}: ${s3.raw.slice(-300)}`,
  );
}

/** FormData with text fields and small image files (what the browser posts). */
function withFiles(
  fields: Record<string, string>,
  files: Record<string, string>,
) {
  const f = form(fields);
  for (const [name, filename] of Object.entries(files))
    f.set(
      name,
      new File([new Uint8Array([0xff, 0xd8, 0xff, 0xe0, 1, 2, 3])], filename, {
        type: "image/jpeg",
      }),
    );
  return f;
}

describe("Seller Centre", () => {
  test("sign up, open a shop, verify, add a payout account, list a product", async () => {
    const b = browser();
    await signUp(b, "Ngozi", "Eke");
    const empty = await b.get("/centre");
    assert.match(textOf(empty.html), /You don’t have a shop yet/);

    // Open a shop.
    const name = `Ngozi Gadgets ${randomInt(1000, 9999)}`;
    await b.get("/centre/open-shop");
    const opened = await b.action("/centre/open-shop", "openShop", [
      {},
      form({
        type: "individual",
        displayName: name,
        stateCode: "LA",
        city: "Surulere",
      }),
    ]);
    assert.match(
      opened.redirect ?? "",
      /^\/centre\/[0-9a-f-]{36}$/,
      opened.raw.slice(-300),
    );
    const shopId = opened.redirect!.split("/")[2];
    const home = textOf((await b.get(`/centre/${shopId}`)).html);
    assert.match(home, new RegExp(name));
    assert.match(home, /Verify your identity/);

    // Verification: required documents are checked before anything is uploaded.
    const kycPage = `/centre/${shopId}/verification`;
    await b.get(kycPage);
    const missing = await b.action<Result>(
      "/centre/[shopId]/verification",
      "submitKyc",
      [
        shopId,
        {},
        withFiles(
          {
            sellerType: "individual",
            idType: "passport",
            idNumber: "A12345678",
          },
          { id_front: "id.jpg" },
        ),
      ],
      { url: kycPage },
    );
    assert.equal(missing.value.ok, false);
    assert.ok(missing.value.errors?.selfie, JSON.stringify(missing.value));
    const kyc = await b.action<Result>(
      "/centre/[shopId]/verification",
      "submitKyc",
      [
        shopId,
        {},
        withFiles(
          {
            sellerType: "individual",
            idType: "passport",
            idNumber: "A12345678",
          },
          { id_front: "id.jpg", selfie: "me.jpg" },
        ),
      ],
      { url: kycPage },
    );
    assert.equal(kyc.value.ok, true, JSON.stringify(kyc.value));
    const kycText = textOf((await b.get(kycPage)).html);
    assert.match(kycText, /Submitted/);
    assert.match(kycText, /ending 5678/);
    assert.doesNotMatch(kycText, /A12345678/);

    // Payout account (dev name enquiry echoes the typed name).
    const payPage = `/centre/${shopId}/payouts`;
    await b.get(payPage);
    const badAcct = await b.action<Result>(
      "/centre/[shopId]/payouts",
      "addBankAccount",
      [
        shopId,
        {},
        form({
          bankCode: "058",
          accountNumber: "12345",
          accountName: "Ngozi Eke",
        }),
      ],
      { url: payPage },
    );
    assert.ok(badAcct.value.errors?.accountNumber);
    const acct = await b.action<Result>(
      "/centre/[shopId]/payouts",
      "addBankAccount",
      [
        shopId,
        {},
        form({
          bankCode: "058",
          accountNumber: `01${randomInt(10_000_000, 99_999_999)}`,
          accountName: "Ngozi Eke",
        }),
      ],
      { url: payPage },
    );
    assert.equal(acct.value.ok, true, JSON.stringify(acct.value));
    assert.match(textOf((await b.get(payPage)).html), /Name matches/);

    // A reviewer approves the shop (staff API).
    const reviewer = await staffToken("kyc_reviewer");
    const vendor = await api.get("/api/v1/staff/vendors/{id}", {
      path: { id: shopId },
      token: reviewer,
    });
    await api.post("/api/v1/staff/vendors/kyc/{id}/review", {
      path: { id: vendor.kyc!.id },
      body: { decision: "approved" },
      token: reviewer,
    });
    assert.match(
      textOf((await b.get(`/centre/${shopId}`)).html),
      /You’re ready to sell/,
    );
    const payouts = textOf((await b.get(payPage)).html);
    assert.match(payouts, /Your balance/);
    assert.match(payouts, /No payouts yet/);
    assert.match(payouts, /Payouts are on hold/); // the account was added minutes ago

    // List a product: search, choose, offer. New shops' listings go to review.
    const newPage = `/centre/${shopId}/listings/new`;
    assert.match(
      textOf((await b.get(`${newPage}?q=nova`)).html),
      /Nova X5 Pro/,
    );
    const offer = await b.get(`${newPage}?product=nova-x5-pro-5g`);
    assert.match(textOf(offer.html), /Create listing/);
    const bad = await b.action<Result>(
      "/centre/[shopId]/listings/new",
      "createListing",
      [
        shopId,
        {},
        form({
          variantId: seed().extra.variant_p1,
          condition: "uk_used",
          price: "abc",
          sellerStock: "2",
        }),
      ],
      { url: newPage },
    );
    assert.ok(
      bad.value.errors?.price && bad.value.errors?.conditionNotes,
      JSON.stringify(bad.value),
    );
    const created = await b.action(
      "/centre/[shopId]/listings/new",
      "createListing",
      [
        shopId,
        {},
        form({
          variantId: seed().extra.variant_p1,
          condition: "new",
          price: "885,000",
          sellerStock: "4",
          handlingDays: "1",
          warrantyProvider: "manufacturer",
          warrantyMonths: "12",
        }),
      ],
      { url: newPage },
    );
    assert.match(
      created.redirect ?? "",
      new RegExp(`^/centre/${shopId}/listings/[0-9a-f-]{36}$`),
      created.raw.slice(-300),
    );
    const listingPath = created.redirect!;
    const listingId = listingPath.split("/").pop()!;
    assert.match(textOf((await b.get(listingPath)).html), /In review/);

    const editor = await staffToken("catalog_editor");
    await api.post("/api/v1/staff/catalog/listings/{id}/decision", {
      path: { id: listingId },
      body: { decision: "approved" },
      token: editor,
    });
    const live = textOf((await b.get(listingPath)).html);
    assert.match(live, /On sale/);
    assert.match(live, /₦885,000/);

    // Price and stock changes.
    const price = await b.action<Result>(
      "/centre/[shopId]/listings/[listingId]",
      "setPrice",
      [shopId, listingId, {}, form({ price: "880000", compareAt: "" })],
      { url: listingPath },
    );
    assert.equal(price.value.ok, true, JSON.stringify(price.value));
    const stock = await b.action<Result>(
      "/centre/[shopId]/listings/[listingId]",
      "setStock",
      [shopId, listingId, {}, form({ stock: "7" })],
      { url: listingPath },
    );
    assert.equal(stock.value.ok, true, JSON.stringify(stock.value));
    const after = textOf((await b.get(listingPath)).html);
    assert.match(after, /₦880,000/);
    const list = textOf(
      (await b.get(`/centre/${shopId}/listings?status=active`)).html,
    );
    assert.match(list, /Nova X5 Pro/);

    // Orders and reviews pages render for a new shop.
    assert.match(
      textOf((await b.get(`/centre/${shopId}/orders`)).html),
      /No orders here/,
    );
    assert.match(
      textOf((await b.get(`/centre/${shopId}/reviews`)).html),
      /No reviews yet/,
    );

    // An order: accept, pack, then the drop-off panel shows the parcel code, the deadline and
    // the hubs; once hub staff scan it in, the seller sees it at the hub.
    const number = await placePaidOrder("customer", listingId);
    const ops = await staffToken("ops_manager");
    const order = await api.get("/api/v1/staff/orders/{number}", {
      path: { number },
      token: ops,
    });
    const orderId = order.fulfilments![0].id;
    const orderPage = `/centre/${shopId}/orders/${orderId}`;
    assert.match(textOf((await b.get(orderPage)).html), /New — accept it/);
    for (const name of ["acceptOrder", "packOrder"]) {
      const r = await b.action<Result>(
        "/centre/[shopId]/orders/[orderId]",
        name,
        [shopId, orderId, {}, form({})],
        { url: orderPage },
      );
      assert.equal(r.value.ok, true, `${name}: ${JSON.stringify(r.value)}`);
    }
    const packed = textOf((await b.get(orderPage)).html);
    const label = packed.match(/TSS-[0-9A-Z]{8}/)?.[0];
    assert.ok(label, packed.slice(0, 600));
    assert.match(packed, /Drop it off by/);
    assert.match(packed, /Ikeja warehouse/);
    assert.match(packed, /Wuse hub/);
    assert.doesNotMatch(packed, /Victoria Island store/);
    await api.post("/api/v1/staff/wms/dropoffs", {
      body: { warehouseId: seed().extra.warehouseIkejaId, label },
      token: await staffToken("warehouse_lead"),
    });
    assert.match(textOf((await b.get(orderPage)).html), /At the TechShop hub/);

    // Another seller can't open this shop's pages.
    const other = browser();
    await signUp(other, "Kunle", "Ade");
    assert.equal((await other.get(`/centre/${shopId}`)).status, 404);
    assert.equal((await other.get(listingPath)).status, 404);
  });

  test("Seller Centre pages need a session", async () => {
    const r = await browser().get("/centre");
    assert.match(r.url, /\/sign-in\?next=%2Fcentre$/);
    assert.match(textOf(r.html), /Sign in to Seller Centre/);
  });
});
