# Node.js / TypeScript Testing

- Vitest (preferred for new projects) or Jest — match the project.
- API tests with Supertest against the Express `app` (not a running server).
```ts
it('returns 404 for another tenant\'s invoice', async () => {
  const other = await makeInvoice({ tenantId: tenantB.id });
  const res = await request(app).get(`/api/v1/invoices/${other.id}`).set(authAs(userA));
  expect(res.status).toBe(404);
});
```
- DB: Testcontainers or a dedicated test DB via docker-compose; reset with transactions or truncation in `beforeEach`.
- MongoDB: `mongodb-memory-server` acceptable for unit-ish tests; real Mongo for integration.
- Mock external HTTP with `msw` or `nock`; fake timers via `vi.useFakeTimers()`.
- Factories with `@faker-js/faker` seeded for determinism.
