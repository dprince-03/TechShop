// Contract tests: the public catalogue, search and content endpoints the storefronts use.
import assert from "node:assert/strict";
import { describe, test } from "node:test";

import { ApiError } from "../../src/index";
import { api } from "../kit";

describe("catalogue", () => {
  test("GET /categories lists active top-level categories", async () => {
    const cats = (await api.get("/api/v1/categories")) ?? [];
    assert.ok(cats.length >= 5);
    assert.ok(cats.some((c) => c.slug === "phones"));
  });

  test("GET /categories/{slug}/products returns cards for the category", async () => {
    const r = await api.get("/api/v1/categories/{slug}/products", {
      path: { slug: "phones" },
    });
    assert.ok((r.items ?? []).length > 0);
    for (const p of r.items ?? [])
      assert.ok(p.slug && p.name && (p.price ?? 0) > 0);
  });

  test("GET /search filters, sorts and pages with a cursor", async () => {
    const first = await api.get("/api/v1/search", {
      query: { q: "phone", limit: 2, sort: "price_asc" },
    });
    assert.equal(first.items?.length, 2);
    assert.ok((first.items![0].price ?? 0) <= (first.items![1].price ?? 0));
    assert.ok(first.page.hasMore && first.page.nextCursor);
    const next = await api.get("/api/v1/search", {
      query: {
        q: "phone",
        limit: 2,
        sort: "price_asc",
        cursor: first.page.nextCursor,
      },
    });
    assert.notEqual(next.items?.[0]?.id, first.items![0].id);
    assert.ok(first.facets);
  });

  test("GET /search rejects a forged cursor", async () => {
    await assert.rejects(
      api.get("/api/v1/search", {
        query: { q: "phone", cursor: "forged.cursor" },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 422,
    );
  });

  test("GET /search/suggest returns product suggestions", async () => {
    const r = await api.get("/api/v1/search/suggest", { query: { q: "or" } });
    assert.ok((r.products ?? []).length > 0);
  });

  test("GET /products/{slug} returns offers with a buy box", async () => {
    const list = await api.get("/api/v1/search", { query: { limit: 1 } });
    const p = await api.get("/api/v1/products/{slug}", {
      path: { slug: list.items![0].slug },
    });
    assert.equal(p.slug, list.items![0].slug);
    assert.ok((p.offers ?? []).some((o) => o.buyBox));
    assert.ok((p.variants ?? []).length > 0);
  });

  test("GET /products/{slug} 404s with a problem code", async () => {
    await assert.rejects(
      api.get("/api/v1/products/{slug}", { path: { slug: "no-such-product" } }),
      (e: unknown) =>
        e instanceof ApiError && e.status === 404 && e.code === "not_found",
    );
  });

  test("GET /deals only returns discounted items", async () => {
    const r = await api.get("/api/v1/deals", { query: { limit: 10 } });
    for (const p of r.items ?? [])
      assert.ok((p.discountPct ?? 0) > 0 || p.promoPrice !== undefined);
  });

  test("GET /best-sellers responds", async () => {
    const r = await api.get("/api/v1/best-sellers", { query: { limit: 5 } });
    assert.ok(Array.isArray(r.items ?? []));
  });

  test("GET /brands lists brands", async () => {
    const r = (await api.get("/api/v1/brands")) ?? [];
    assert.ok(r.length > 0 && r[0].slug);
  });
});

describe("reference data and delivery", () => {
  test("GET /states lists the 37 states (36 + FCT)", async () => {
    const r = (await api.get("/api/v1/states")) ?? [];
    assert.equal(r.length, 37);
  });

  test("GET /delivery/estimate covers Lagos and not Kano", async () => {
    const la = await api.get("/api/v1/delivery/estimate", {
      query: { state: "LA" },
    });
    assert.equal(la.covered, true);
    assert.ok((la.feeKobo ?? 0) > 0);
    const kn = await api.get("/api/v1/delivery/estimate", {
      query: { state: "KN" },
    });
    assert.equal(kn.covered, false);
  });
});

describe("content", () => {
  test("GET /content/help/{site} lists articles and each one loads", async () => {
    const list =
      (await api.get("/api/v1/content/help/{site}", {
        path: { site: "market" },
      })) ?? [];
    assert.ok(list.length > 0);
    const a = await api.get("/api/v1/content/help/{site}/{slug}", {
      path: { site: "market", slug: list[0].slug },
    });
    assert.equal(a.slug, list[0].slug);
  });

  test("GET /content/banners needs a site and returns a list", async () => {
    const r = await api.get("/api/v1/content/banners", {
      query: { site: "market" },
    });
    assert.ok(Array.isArray(r ?? []));
  });

  test("GET /content/pages/{site}/{slug} 404s for unpublished pages", async () => {
    await assert.rejects(
      api.get("/api/v1/content/pages/{site}/{slug}", {
        path: { site: "market", slug: "terms" },
      }),
      (e: unknown) => e instanceof ApiError && e.status === 404,
    );
  });
});
