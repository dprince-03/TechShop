// Unit tests for the client plumbing (no network).
import assert from "node:assert/strict";
import { test } from "node:test";

import { ApiError, buildPath, buildQuery, createApiClient } from "../src/index";

test("buildPath encodes values and rejects missing ones", () => {
  assert.equal(buildPath("/api/v1/products/{slug}", { slug: "a b/c" }), "/api/v1/products/a%20b%2Fc");
  assert.throws(() => buildPath("/api/v1/products/{slug}", {}), /slug/);
});

test("buildQuery drops empty values and repeats arrays", () => {
  assert.equal(buildQuery({ q: "phone", empty: "", none: undefined, n: 0, tags: ["a", "b"], on: true }), "?q=phone&n=0&tags=a&tags=b&on=true");
  assert.equal(buildQuery({}), "");
});

test("errors become ApiError with the problem code and request ID", async () => {
  const api = createApiClient({
    baseUrl: "http://api.test/",
    fetch: async () =>
      new Response(JSON.stringify({ code: "out_of_stock", detail: "Sold out", status: 409 }), {
        status: 409,
        headers: { "content-type": "application/problem+json", "x-request-id": "req-1" },
      }),
  });
  await assert.rejects(api.get("/api/v1/categories"), (e: unknown) => {
    assert.ok(e instanceof ApiError);
    assert.equal(e.status, 409);
    assert.equal(e.code, "out_of_stock");
    assert.equal(e.requestId, "req-1");
    return true;
  });
});

test("sends bearer token, idempotency key and JSON body", async () => {
  let seen: Request | undefined;
  const api = createApiClient({
    baseUrl: "http://api.test",
    getToken: () => "default-token",
    fetch: async (input, init) => {
      seen = new Request(input, init);
      return new Response(null, { status: 204 });
    },
  });
  await api.post("/api/v1/auth/logout", { body: { refreshToken: "r" }, idempotencyKey: "k1", token: "t1" });
  assert.equal(seen?.method, "POST");
  assert.equal(seen?.headers.get("authorization"), "Bearer t1");
  assert.equal(seen?.headers.get("idempotency-key"), "k1");
  assert.deepEqual(await seen?.json(), { refreshToken: "r" });
});

test("network failures are status 0 and unavailable", async () => {
  const api = createApiClient({ baseUrl: "http://api.test", fetch: async () => Promise.reject(new TypeError("refused")) });
  await assert.rejects(api.get("/api/v1/categories"), (e: unknown) => e instanceof ApiError && e.status === 0 && e.isUnavailable);
});
