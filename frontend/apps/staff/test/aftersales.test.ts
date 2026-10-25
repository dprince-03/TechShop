// End-to-end tests for the M5 staff screens: support replies and internal notes, return
// decisions, and the stolen-device blocklist. Customers are set up through the API.
import assert from "node:assert/strict";
import { randomInt } from "node:crypto";
import { describe, test } from "node:test";

import {
  api,
  deliveredOrder,
  form,
  seed,
  textOf,
  tokenForPhone,
} from "@techshop/testing";

import { signIn } from "./helpers";

const phone = () => `+23480${randomInt(10_000_000, 99_999_999)}`;
type Result = { ok: boolean; message: string };

describe("after-sales", () => {
  test("support: the queue, a reply the customer sees and a note they don't", async () => {
    const token = await tokenForPhone(phone(), "market", "Ifeoma");
    const t = await api.post("/api/v1/me/tickets", {
      token,
      body: {
        topic: "payment",
        subject: "Charged twice",
        message: "My card was charged twice for one order.",
      },
    });
    const b = await signIn("support_agent", "/support");
    assert.match(
      textOf((await b.get("/support?tab=urgent")).html),
      new RegExp(t.number),
    );
    const page = `/support/${t.number}`;
    assert.match(textOf((await b.get(page)).html), /Charged twice/);
    const note = await b.action<Result>(
      "/support/[number]",
      "replyToTicket",
      [
        t.number,
        {},
        form({
          body: "Checked the provider: one charge was reversed.",
          internal: "on",
        }),
      ],
      { url: page },
    );
    assert.equal(note.value.ok, true, note.value.message);
    const reply = await b.action<Result>(
      "/support/[number]",
      "replyToTicket",
      [
        t.number,
        {},
        form({
          body: "One charge was reversed; it reaches you in 3-5 days.",
          status: "pending_customer",
        }),
      ],
      { url: page },
    );
    assert.equal(reply.value.ok, true, reply.value.message);
    const seen = await api.get("/api/v1/me/tickets/{number}", {
      token,
      path: { number: t.number },
    });
    const bodies = (seen.messages ?? []).map((m) => m.body).join(" | ");
    assert.match(bodies, /reaches you in 3-5 days/);
    assert.doesNotMatch(bodies, /Checked the provider/);
    assert.equal(seen.status, "pending_customer");
  });

  test("returns: a change-of-mind return is approved from its page", async () => {
    const token = await tokenForPhone(phone(), "market", "Segun");
    const { orderNumber, lineId } = await deliveredOrder(
      token,
      seed().extra.listing_p21,
    );
    const r = await api.post("/api/v1/me/orders/{number}/returns", {
      token,
      path: { number: orderNumber },
      body: {
        items: [{ orderLineId: lineId, quantity: 1 }],
        reason: "change_of_mind",
        method: "store_dropoff",
        unopened: true,
      },
    });
    const b = await signIn("support_lead", "/returns");
    assert.match(
      textOf((await b.get("/returns?tab=requested")).html),
      new RegExp(r.rma),
    );
    const page = `/returns/${r.rma}`;
    assert.match(textOf((await b.get(page)).html), /Changed my mind/);
    const ok = await b.action<Result>(
      "/returns/[rma]",
      "decideReturn",
      [r.rma, "approve", ""],
      { url: page },
    );
    assert.equal(ok.value.ok, true, ok.value.message);
    const after = await api.get("/api/v1/me/returns/{rma}", {
      token,
      path: { rma: r.rma },
    });
    assert.equal(after.status, "approved");
  });

  test("risk: block a device and see it on the blocklist and the IMEI check", async () => {
    const b = await signIn("risk_analyst", "/risk/devices");
    const imei = luhn(`35${String(Date.now()).slice(-12)}`);
    const blocked = await b.action<Result>(
      "/risk/devices",
      "blockDevice",
      [
        {},
        form({
          imei,
          reason: "reported_stolen",
          source: "Customer with police extract",
        }),
      ],
      { url: "/risk/devices" },
    );
    assert.equal(blocked.value.ok, true, blocked.value.message);
    const page = textOf((await b.get(`/risk/devices?imei=${imei}`)).html);
    assert.match(page, new RegExp(imei));
    assert.match(page, /Blocked: Reported stolen/);
    assert.match(textOf((await b.get("/risk")).html), /Blocked devices/);
  });
});

/** Completes a 14-digit body with the Luhn check digit. */
function luhn(first14: string) {
  let sum = 0;
  for (let i = 0; i < 14; i++) {
    let d = Number(first14[13 - i]);
    if (i % 2 === 0) {
      d *= 2;
      if (d > 9) d -= 9;
    }
    sum += d;
  }
  return first14 + String((10 - (sum % 10)) % 10);
}
