// Contract tests: the logistics app's endpoints, end to end through one delivery.
import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { describe, test } from "node:test";

import { ApiError } from "../../src/index";
import { account, api, customerToken, latestCode, peekCode, pickAndPack, placePaidOrder, riderToken, seed, staffToken } from "../kit";

const now = () => new Date().toISOString();

describe("rider app", () => {
  test("shift → assigned job → pickup by label → offline batch → location → delivery code → shift end", async () => {
    const number = await placePaidOrder("customer_b", seed().extra.listing_p2);
    const packed = await pickAndPack(number);
    assert.equal(packed.handoff, "rider");
    const jobId = packed.jobId!;
    const label = packed.parcel.labelCode;

    const rider = await riderToken();
    const riderId = account("rider").userId;
    const dispatcher = await staffToken("dispatcher");

    // End any shift left open by an earlier run, then start one (consent is required).
    await api.post("/api/v1/rider/shift/end", { token: rider }).catch(() => undefined);
    await assert.rejects(api.post("/api/v1/rider/shift/start", { token: rider, body: { locationConsent: false } }), (e: unknown) => e instanceof ApiError && e.code === "consent_required");
    const shift = await api.post("/api/v1/rider/shift/start", { token: rider, body: { locationConsent: true, latitude: 6.6018, longitude: 3.3515 } });
    assert.ok(shift.id && !shift.endedAt);

    await api.post("/api/v1/dispatch/jobs/{id}/assign", { token: dispatcher, path: { id: jobId }, body: { riderId } });
    const jobs = (await api.get("/api/v1/rider/jobs", { token: rider })) ?? [];
    const job = jobs.find((j) => j.id === jobId);
    assert.ok(job, "assigned job is in the rider's list");
    assert.doesNotMatch(JSON.stringify(jobs), /\+234\d{10}/, "rider never sees customer phone numbers");

    const act = (body: Record<string, unknown>) =>
      api.post("/api/v1/rider/jobs/{id}/actions", { token: rider, path: { id: jobId }, body: { clientEventId: randomUUID(), occurredAt: now(), latitude: 6.6, longitude: 3.35, ...body } as never });
    await assert.rejects(act({ kind: "picked_up", labelCode: "TSL-WRONG1" }), (e: unknown) => e instanceof ApiError && e.code === "label_mismatch");
    const phone = account("customer_b").phone!;
    const beforeCode = await peekCode(phone, "delivery code");
    await act({ kind: "picked_up", labelCode: label });
    const otp = await latestCode(phone, beforeCode, "delivery code");

    // The offline queue's batch endpoint applies each event once.
    const ev = { clientEventId: randomUUID(), jobId, kind: "en_route" as const, occurredAt: now(), offline: true };
    const first = await api.post("/api/v1/rider/events/batch", { token: rider, body: { events: [ev] } });
    const again = await api.post("/api/v1/rider/events/batch", { token: rider, body: { events: [ev] } });
    assert.equal(first.results?.[0].status, "accepted");
    assert.equal(again.results?.[0].status, "duplicate");
    const bad = await api.post("/api/v1/rider/events/batch", { token: rider, body: { events: [{ clientEventId: randomUUID(), jobId: randomUUID(), kind: "en_route", occurredAt: now() }] } });
    assert.equal(bad.results?.[0].status, "rejected", "an unknown job is rejected, not fatal");

    await api.post("/api/v1/rider/location", { token: rider, body: { points: [{ latitude: 6.59, longitude: 3.34, recordedAt: new Date(Date.now() - 30_000).toISOString() }, { latitude: 6.595, longitude: 3.345, recordedAt: now() }] } });

    await assert.rejects(act({ kind: "delivered" }), (e: unknown) => e instanceof ApiError && e.code === "proof_required");
    await assert.rejects(act({ kind: "delivered", otp: otp === "000000" ? "111111" : "000000" }), (e: unknown) => e instanceof ApiError && e.code === "otp_wrong");
    const done = await act({ kind: "delivered", otp, recipientName: "Ada" });
    assert.ok(done);

    const ctoken = await customerToken("customer_b");
    let status = "";
    for (let i = 0; i < 30 && status !== "delivered"; i++) {
      status = (await api.get("/api/v1/me/orders/{number}", { token: ctoken, path: { number } })).status;
      if (status !== "delivered") await new Promise((r) => setTimeout(r, 300));
    }
    assert.equal(status, "delivered");
    const ended = await api.post("/api/v1/rider/shift/end", { token: rider });
    assert.ok(ended.endedAt);
  });

  test("a customer token can't use rider endpoints", async () => {
    await assert.rejects(api.get("/api/v1/rider/jobs", { token: await customerToken() }), (e: unknown) => e instanceof ApiError && (e.status === 401 || e.status === 403));
  });
});
