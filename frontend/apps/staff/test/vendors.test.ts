// End-to-end tests for the marketplace staff screens: vendor KYC review (with documents and the
// audited ID reveal) and review moderation. A fresh seller and customer are set up through the API.
import assert from "node:assert/strict";
import { randomInt } from "node:crypto";
import { describe, test } from "node:test";

import { api, latestCode, peekCode, textOf } from "@techshop/testing";

import { signIn } from "./helpers";

const phone = () => `+23481${randomInt(10_000_000, 99_999_999)}`;

/** Signs a brand-new person up by phone OTP on an audience and returns their access token. */
async function signUp(audience: "seller" | "market", first: string) {
  const p = phone();
  const before = await peekCode(p);
  const req = await api.post("/api/v1/auth/otp/request", {
    body: { phone: p, purpose: "login" },
  });
  const code = await latestCode(p, before);
  const v = await api.post("/api/v1/auth/otp/verify", {
    body: { challengeId: req.challengeId, code, audience },
  });
  const s = await api.post("/api/v1/auth/otp/signup", {
    body: {
      signupToken: v.signupToken!,
      firstName: first,
      lastName: "Test",
      audience,
      deviceName: "e2e",
    },
  });
  return s.accessToken;
}

/** A seller with a submitted verification and a payout account. */
async function sellerAwaitingReview() {
  const token = await signUp("seller", "Amaka");
  const name = `Amaka Phones ${randomInt(1000, 9999)}`;
  const shop = await api.post("/api/v1/seller/shops", {
    body: {
      type: "individual",
      displayName: name,
      stateCode: "AN",
      city: "Onitsha",
    },
    token,
  });
  const file = async (n: string) => {
    const up = await api.post("/api/v1/files/uploads", {
      body: {
        purpose: "kyc_document",
        contentType: "image/jpeg",
        sizeBytes: 2048,
        name: n,
      },
      token,
    });
    await api.post("/api/v1/files/{id}/complete", {
      path: { id: up.fileId },
      body: {},
      token,
    });
    return up.fileId;
  };
  const documents = [
    { kind: "id_front" as const, fileId: await file("id.jpg") },
    { kind: "selfie" as const, fileId: await file("me.jpg") },
  ];
  const idNumber = `${randomInt(10_000_000_000, 99_999_999_999)}`;
  const kyc = await api.post("/api/v1/seller/shops/{shopId}/kyc", {
    path: { shopId: shop.id },
    body: { idType: "nin", idNumber, documents },
    token,
  });
  await api.post("/api/v1/seller/shops/{shopId}/bank-accounts", {
    path: { shopId: shop.id },
    body: {
      bankCode: "058",
      accountNumber: `02${randomInt(10_000_000, 99_999_999)}`,
      accountName: "Amaka Test",
    },
    token,
  });
  return { shopId: shop.id, name, kycId: kyc.id, idNumber, documents };
}

type Result = { ok: boolean; message: string };

describe("vendors", () => {
  test("a KYC reviewer opens documents, reveals the ID with a reason and approves", async () => {
    const s = await sellerAwaitingReview();
    const b = await signIn("kyc_reviewer", "/vendors");
    const list = textOf((await b.get("/vendors?tab=to_review")).html);
    assert.match(list, new RegExp(s.name));

    const page = `/vendors/${s.shopId}`;
    const detail = textOf((await b.get(page)).html);
    assert.match(detail, new RegExp(`ending ${s.idNumber.slice(-4)}`));
    assert.doesNotMatch(detail, new RegExp(s.idNumber));
    assert.match(detail, /Matches the holder/);

    // A document opens through a short-lived link requested on click.
    const doc = await b.get(`${page}/documents/${s.documents[0].fileId}`);
    assert.equal(doc.status, 200);
    assert.match(doc.url, /\/dev\/files\//);

    const short = await b.action<Result>(
      "/vendors/[id]",
      "revealId",
      [s.kycId, "why"],
      { url: page },
    );
    assert.equal(short.value.ok, false);
    const reveal = await b.action<Result>(
      "/vendors/[id]",
      "revealId",
      [s.kycId, "Comparing the NIN with the bank record"],
      { url: page },
    );
    assert.equal(reveal.value.ok, true, reveal.value.message);
    assert.match(reveal.value.message, new RegExp(s.idNumber));

    const noNote = await b.action<Result>(
      "/vendors/[id]",
      "reviewKyc",
      [s.kycId, "needs_more_info", ""],
      { url: page },
    );
    assert.equal(noNote.value.ok, false);
    const ok = await b.action<Result>(
      "/vendors/[id]",
      "reviewKyc",
      [s.kycId, "approved", ""],
      { url: page },
    );
    assert.equal(ok.value.ok, true, ok.value.message);
    assert.match(textOf((await b.get(page)).html), /Approved/);
  });

  test("staff without vendor access are turned away", async () => {
    const b = await signIn("dispatcher");
    const r = await b.get("/vendors");
    assert.match(r.url, /\/\?denied=1$/);
  });
});

describe("review moderation", () => {
  test("a held review is published from the catalogue's Reviews tab", async () => {
    const customer = await signUp("market", "Tobi");
    const held = await api.post("/api/v1/products/{slug}/reviews", {
      path: { slug: "nova-x5-pro-5g" },
      body: {
        rating: 2,
        body: `Call me for a cheaper one: 0803 555 ${randomInt(1000, 9999)}`,
      },
      token: customer,
    });
    assert.equal(held.status, "pending");

    const b = await signIn("catalog_editor", "/catalogue?tab=reviews");
    const queue = textOf((await b.get("/catalogue?tab=reviews")).html);
    assert.match(queue, /Call me for a cheaper one/);
    assert.match(queue, /contact details or links/);
    const r = await b.action<Result>(
      "/catalogue",
      "decideReview",
      [held.id, "hide"],
      { url: "/catalogue?tab=reviews" },
    );
    assert.equal(r.value.ok, true, r.value.message);
    assert.doesNotMatch(
      textOf((await b.get("/catalogue?tab=reviews")).html),
      new RegExp(held.id),
    );
  });
});
