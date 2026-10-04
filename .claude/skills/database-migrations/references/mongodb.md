# MongoDB Notes

## Design
- Model around access patterns: embed data read together and bounded in size; reference data that grows unbounded or is shared.
- Avoid unbounded arrays (16 MB document limit, performance).
- Include `tenantId` on every multi-tenant document and in compound indexes: `{ tenantId: 1, createdAt: -1 }`.

## Validation
```js
db.runCommand({ collMod: "invoices", validator: { $jsonSchema: {
  bsonType: "object", required: ["tenantId", "amountMinor", "currency"],
  properties: { amountMinor: { bsonType: "long", minimum: 1 }, currency: { bsonType: "string", minLength: 3, maxLength: 3 } }
}}, validationLevel: "strict" });
```

## Indexes
- Build on large collections during low traffic; unique indexes include `tenantId`.
- TTL indexes for expiring data (sessions, OTPs): `{ expiresAt: 1 }, { expireAfterSeconds: 0 }`.

## Migrations
- Use `migrate-mongo` (or project tool); write idempotent scripts; batch updates with `bulkWrite`.

## Transactions
- Require replica set; use for multi-document atomic changes; keep them short.

## Security
- Auth enabled, bind to private network, sanitize `$` operators from user input, least-privilege users per app.
