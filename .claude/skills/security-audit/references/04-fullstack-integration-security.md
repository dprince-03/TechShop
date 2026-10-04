# 04 — Fullstack Integration Security

Checks that only show up when frontend, mobile, backend, and infrastructure work **together**. Run this after files 01–03, focusing on seams between components.

---

## 1. End-to-End Authentication Flow

- [ ] `[CRITICAL]` Trace login → token issue → token storage → API call → token validation → refresh → logout across every client
- [ ] `[HIGH]` Same auth rules applied to web, mobile, and any third-party API clients
- [ ] `[HIGH]` Token audience (`aud`) differs per client where appropriate
- [ ] `[HIGH]` Password change / account disable revokes sessions on **all** devices
- [ ] `[HIGH]` SSR (Next.js server side) forwards user identity securely; never uses a privileged service token on behalf of unauthenticated users
- [ ] `[MEDIUM]` BFF (Backend-for-Frontend) pattern considered to keep tokens out of the browser

**Test:** Log in on web and mobile; change password on web; confirm mobile session dies on next request.

---

## 2. Session Consistency Across Clients

- [ ] `[HIGH]` Device/session list available to users with remote revoke
- [ ] `[MEDIUM]` Concurrent session limits if required by policy
- [ ] `[MEDIUM]` Session timeouts consistent with risk (shorter for admin dashboards)

---

## 3. API Contract & Data Leakage

- [ ] `[HIGH]` API responses designed per client need; no "fat" responses filtered by the UI
- [ ] `[HIGH]` Shared types/schemas (OpenAPI, tRPC, GraphQL schema) reviewed for sensitive fields
- [ ] `[HIGH]` Error responses consistent and non-revealing across all services
- [ ] `[MEDIUM]` Contract tests ensure frontend and backend agree on validation rules

---

## 4. Trust Boundaries

- [ ] `[CRITICAL]` Every boundary documented: browser ↔ CDN ↔ load balancer ↔ API ↔ internal services ↔ DB ↔ third parties
- [ ] `[HIGH]` Internal services still authenticate each other (mTLS, signed service tokens); "internal network" is not trust
- [ ] `[HIGH]` Headers like `X-User-Id`, `X-Forwarded-For`, `X-Tenant-Id` stripped at the edge and only set by trusted proxies
- [ ] `[HIGH]` Admin panels on separate subdomain/app with stricter auth (MFA, IP allow-list or VPN/zero-trust access)
- [ ] `[MEDIUM]` Microservices follow least privilege for service-to-service calls

---

## 5. Multi-Tenant Isolation (End-to-End)

- [ ] `[CRITICAL]` Tenant resolved server-side from authenticated identity (not from a client-supplied header or subdomain alone)
- [ ] `[CRITICAL]` Tenant scoping applied in: API queries, caches, search indexes, file storage paths, queues, analytics, logs, exports, emails
- [ ] `[HIGH]` Background jobs carry and enforce tenant context
- [ ] `[HIGH]` Tenant-specific branding/config cannot inject scripts into other tenants
- [ ] `[HIGH]` Cross-tenant automated test suite in CI
- [ ] `[MEDIUM]` Per-tenant encryption keys for high-sensitivity customers
- [ ] `[MEDIUM]` Noisy-neighbor protection: per-tenant rate limits and quotas

**Test:** Two tenants, two users each. Attempt every read/write across tenants via API, search, file URLs, exports, and WebSocket channels.

---

## 6. Environment Separation

- [ ] `[CRITICAL]` Production data never copied to dev/staging unmasked
- [ ] `[CRITICAL]` Separate credentials, API keys, and cloud accounts/projects per environment
- [ ] `[HIGH]` Staging not publicly indexable; protected by auth
- [ ] `[HIGH]` Test payment keys never in production and vice versa
- [ ] `[MEDIUM]` Clear visual indicator in non-prod UIs to avoid mistakes

---

## 7. Feature Flags & Hidden Endpoints

- [ ] `[HIGH]` Feature flags evaluated server-side for access control; client flags only toggle UI
- [ ] `[HIGH]` Unreleased features' APIs protected even if UI is hidden
- [ ] `[HIGH]` Debug, test, seed, and health-detail endpoints disabled or protected in prod (`/debug`, `/swagger`, `/actuator`, `/graphql` playground)
- [ ] `[MEDIUM]` API inventory reconciled with discovered endpoints (crawl + JS bundle analysis)

---

## 8. Real-Time Channels (WebSockets, SSE, Socket.IO)

- [ ] `[CRITICAL]` Connections authenticated at handshake
- [ ] `[CRITICAL]` Authorization checked per channel/room subscription and per message
- [ ] `[HIGH]` `Origin` header validated (Cross-Site WebSocket Hijacking)
- [ ] `[HIGH]` Message size and rate limits
- [ ] `[HIGH]` Incoming messages validated like API input
- [ ] `[MEDIUM]` Connections closed when token expires or user is revoked
- [ ] `[MEDIUM]` `wss://` only

---

## 9. Payment Flow Integrity

- [ ] `[CRITICAL]` Amount, currency, and items determined server-side; client sends only product IDs and quantities
- [ ] `[CRITICAL]` Payment confirmation verified server-to-server with provider (Paystack, Flutterwave, Stripe) before fulfilling
- [ ] `[CRITICAL]` Idempotency keys prevent duplicate charges and duplicate fulfillment
- [ ] `[HIGH]` Use provider-hosted fields/checkout to reduce PCI DSS scope (SAQ A vs SAQ D)
- [ ] `[HIGH]` Webhook and redirect-callback both handled; state machine prevents double processing
- [ ] `[HIGH]` Refunds and payouts require elevated authorization and audit logging
- [ ] `[MEDIUM]` Reconciliation job compares internal records with provider records

---

## 10. Account Lifecycle

- [ ] `[HIGH]` **Signup:** email/phone verification; bot protection; no enumeration
- [ ] `[HIGH]` **Recovery:** secure reset flow; recovery codes for MFA; support-desk verification process can't be socially engineered
- [ ] `[HIGH]` **Email change:** confirm via both old and new addresses
- [ ] `[HIGH]` **Deactivation:** sessions revoked, API keys disabled
- [ ] `[HIGH]` **Deletion:** data deleted or anonymized across DB, backups (per retention), search, caches, analytics, third parties; documented timeline
- [ ] `[MEDIUM]` **Data export** (portability) available and secured
- [ ] `[MEDIUM]` Orphaned resources cleaned up (files, tokens, webhooks)

---

## 11. File Flow End-to-End

- [ ] `[HIGH]` Direct-to-S3 uploads use pre-signed URLs with content-type and size conditions
- [ ] `[HIGH]` Upload → scan → approve pipeline before files are served
- [ ] `[HIGH]` Download URLs authorized per user/tenant and short-lived
- [ ] `[MEDIUM]` User-uploaded content served from a separate domain (sandbox origin)

---

## 12. Notifications (Email, SMS, Push)

- [ ] `[HIGH]` Templates escape user content (HTML email injection)
- [ ] `[HIGH]` No sensitive data or long-lived tokens in emails/SMS
- [ ] `[MEDIUM]` SMS/OTP endpoints rate-limited (SMS pumping / toll fraud)
- [ ] `[MEDIUM]` Magic links single-use and short-lived

---

## 13. Search & Analytics

- [ ] `[HIGH]` Search indexes (Elasticsearch, Meilisearch, Algolia) filtered by tenant and permissions
- [ ] `[HIGH]` Search engines not publicly exposed; API keys scoped (search-only keys for clients)
- [ ] `[MEDIUM]` Analytics events don't contain PII unless disclosed and consented

---

## 14. AI / LLM Features (if present)

- [ ] `[HIGH]` **Prompt injection**: user and retrieved content treated as untrusted; LLM output never executed directly
- [ ] `[HIGH]` LLM tools/function calls authorized as the end user, not with elevated service permissions
- [ ] `[HIGH]` RAG retrieval respects user/tenant permissions
- [ ] `[HIGH]` No secrets or other users' data in prompts/system prompts
- [ ] `[MEDIUM]` Output sanitized before rendering (XSS via LLM output)
- [ ] `[MEDIUM]` Rate limits and cost controls per user
- [ ] `[MEDIUM]` Data sent to LLM providers disclosed in privacy policy; provider data-retention settings reviewed
- [ ] `[MEDIUM]` Mapped to **OWASP Top 10 for LLM Applications**

---

## References
- OWASP ASVS (V1 Architecture, V4 Access Control)
- OWASP Multi-Tenant Security Cheat Sheet, WebSocket Security Cheat Sheet
- OWASP Top 10 for LLM Applications
- PCI DSS SAQ guidance
