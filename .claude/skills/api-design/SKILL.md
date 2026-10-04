---
name: api-design
description: Design and document HTTP APIs — REST resource modeling, URL and naming conventions, status codes, standard response and error envelopes, pagination, filtering, sorting, versioning, idempotency, rate-limit headers, webhooks, and OpenAPI specs (plus GraphQL/gRPC basics). Use this skill whenever the user is designing endpoints, writing or updating an OpenAPI/Swagger spec, reviewing an API for consistency, defining request/response formats, or integrating frontend/mobile clients with a backend.
---

# API Design

Consistency beats cleverness: every endpoint in a project should look like it was designed by one person. If the project already has conventions, follow them and only use these defaults to fill gaps.

## Resource modeling

- Nouns, plural, lowercase, kebab-case: `/api/v1/invoice-items`.
- Nest only one level for ownership: `/projects/{id}/members`. Deeper → top-level resource with filter.
- Actions that don't map to CRUD: sub-resource verbs as POST: `POST /invoices/{id}/send`, `POST /orders/{id}/cancel`.
- JSON fields: `snake_case` or `camelCase` — pick one per project and never mix.
- IDs as strings in JSON (UUIDs/ULIDs), timestamps ISO 8601 UTC (`2026-10-04T12:00:00Z`), money as integer minor units + currency.

## Methods & status codes

| Method | Use | Success |
|---|---|---|
| GET | Read | 200 |
| POST | Create / action | 201 (+ `Location`) / 200 / 202 for async |
| PUT | Full replace | 200 |
| PATCH | Partial update | 200 |
| DELETE | Remove | 204 |

| Code | When |
|---|---|
| 400 | Malformed request / validation failed |
| 401 | Not authenticated |
| 403 | Authenticated but not allowed |
| 404 | Not found **or** not visible to this user/tenant (avoid leaking existence) |
| 409 | Conflict (duplicate, version mismatch, invalid state transition) |
| 422 | Semantically invalid (if project distinguishes from 400) |
| 429 | Rate limited (+ `Retry-After`) |
| 500 | Unexpected server error (generic message only) |
| 503 | Temporarily unavailable |

## Envelopes

Success:
```json
{ "data": { ... } }
{ "data": [ ... ], "meta": { "page": 1, "limit": 20, "total": 134 } }
```

Error (one shape everywhere):
```json
{
  "error": {
    "code": "validation_failed",
    "message": "One or more fields are invalid.",
    "details": [ { "field": "email", "issue": "must be a valid email" } ],
    "request_id": "req_8f2c..."
  }
}
```
Use stable machine-readable `code` values clients can switch on. RFC 9457 Problem Details is an acceptable alternative if the project uses it.

## Pagination, filtering, sorting

- Offset: `?page=2&limit=20` (simple admin lists). Max `limit` enforced (e.g., 100).
- Cursor: `?cursor=<opaque>&limit=20` → `meta.next_cursor` (feeds, large/real-time datasets).
- Filters: `?status=paid&created_after=2026-01-01`; search: `?q=`.
- Sorting: `?sort=-created_at,name` (minus = descending); allow-list sortable fields.
- Sparse fields optional: `?fields=id,name`.

## Versioning

- URL prefix `/api/v1`. Additive changes don't need a new version; breaking changes do.
- Breaking = removing/renaming fields, changing types/semantics, new required inputs.
- Deprecate with `Deprecation` and `Sunset` headers and a changelog entry.

## Reliability

- **Idempotency**: `Idempotency-Key` header on POSTs that create money movements/orders; return the stored response for repeats.
- **Concurrency**: optimistic locking via `ETag` + `If-Match` or a `version` field → 409/412 on mismatch.
- **Async jobs**: 202 + `Location: /jobs/{id}` for long operations.
- **Rate limits**: return `RateLimit-*` (or `X-RateLimit-*`) headers and 429 with `Retry-After`.

## Security basics

- Auth via `Authorization: Bearer <token>` or secure cookies; never tokens in query strings.
- Never return more fields than the client needs; never echo secrets.
- Validate everything; reject unknown fields on writes.

## Webhooks (outgoing)

- Event envelope: `{ "id", "type": "invoice.paid", "created_at", "data": {...} }`.
- Sign with HMAC-SHA256 over timestamp + body; header e.g. `X-Signature: t=...,v1=...`.
- Retries with exponential backoff; receivers dedupe by `id`.

## OpenAPI

- Keep `openapi.yaml` in the repo as the contract; update it in the same PR as the code.
- Start from `assets/openapi-template.yaml`.
- Define reusable `components/schemas` (Error, PaginationMeta) and `securitySchemes`.
- Lint with Spectral/Redocly; generate client types (openapi-typescript, oapi-codegen) where useful.

## GraphQL / gRPC quick rules

- GraphQL: depth/complexity limits, no introspection in prod, field-level authZ, Relay-style cursor pagination, errors in `extensions.code`.
- gRPC: proto package versioning (`v1`), canonical status codes, deadlines on every call, never reuse field numbers.

## Review checklist

- [ ] Naming consistent with rest of API
- [ ] Correct methods & status codes
- [ ] Standard envelopes and error codes
- [ ] Pagination with max limit on all lists
- [ ] AuthN/AuthZ defined per endpoint
- [ ] Idempotency on money-moving POSTs
- [ ] OpenAPI updated
- [ ] No breaking change without version bump
