# 01 — Backend Security

For APIs and server-side services (Node.js/Express, Go/Gin, Java/Spring, etc.), databases, caches, queues, and background jobs. Maps to **OWASP ASVS** and the **OWASP API Security Top 10 (2023)**.

---

## 1. Authentication

- [ ] `[CRITICAL]` Passwords hashed with **argon2id** (preferred) or **bcrypt** (cost ≥ 12); never MD5/SHA1/plain SHA256
- [ ] `[CRITICAL]` Login, signup, and reset endpoints rate-limited and protected against credential stuffing
- [ ] `[HIGH]` Password policy follows **NIST SP 800-63B**: min 8 (prefer 12+) chars, check against breached-password lists, no forced periodic rotation
- [ ] `[HIGH]` MFA available (TOTP, WebAuthn/passkeys); required for admins
- [ ] `[HIGH]` Generic error messages ("Invalid credentials"), no user enumeration via login, signup, or reset
- [ ] `[HIGH]` Account lockout or progressive delay after failed attempts
- [ ] `[HIGH]` Password reset tokens: single-use, random (≥128 bits), short-lived (≤ 1 hour), hashed in DB
- [ ] `[MEDIUM]` Re-authentication required for sensitive actions (change email/password, delete account, payouts)
- [ ] `[MEDIUM]` Notify users of new logins, password changes, and MFA changes

### JWT / Tokens
- [ ] `[CRITICAL]` Signature always verified; `alg: none` rejected; algorithm explicitly allow-listed
- [ ] `[CRITICAL]` Strong signing secret (≥256 bits) or asymmetric keys (RS256/ES256/EdDSA)
- [ ] `[HIGH]` Short-lived access tokens (5–15 min) + refresh tokens
- [ ] `[HIGH]` Refresh token rotation with reuse detection (revoke family on reuse)
- [ ] `[HIGH]` Validate `exp`, `nbf`, `iss`, `aud` claims
- [ ] `[HIGH]` Revocation mechanism exists (denylist or token versioning on logout/password change)
- [ ] `[MEDIUM]` No sensitive data in JWT payload (it is only base64, not encrypted)

### Sessions (if cookie-based)
- [ ] `[HIGH]` Session ID regenerated on login and privilege change
- [ ] `[HIGH]` Cookies: `HttpOnly`, `Secure`, `SameSite=Lax` or `Strict`
- [ ] `[MEDIUM]` Idle and absolute session timeouts
- [ ] `[MEDIUM]` Server-side session invalidation on logout

### OAuth2 / OIDC / SSO
- [ ] `[CRITICAL]` Authorization Code flow **with PKCE**; no Implicit flow
- [ ] `[HIGH]` `state` parameter validated (CSRF), `nonce` validated (OIDC)
- [ ] `[HIGH]` Exact redirect URI matching (no wildcards)
- [ ] `[HIGH]` ID token signature and claims verified
- [ ] `[MEDIUM]` **SAML** assertions: signature validation, XML signature wrapping protection, replay prevention
- [ ] `[MEDIUM]` **SCIM** provisioning secured and deprovisioning tested

---

## 2. Authorization

- [ ] `[CRITICAL]` Every endpoint enforces authorization server-side (never rely on frontend hiding)
- [ ] `[CRITICAL]` **BOLA/IDOR** check: user A cannot access user B's resource by changing an ID
- [ ] `[CRITICAL]` **BFLA** check: regular users cannot call admin functions
- [ ] `[CRITICAL]` Multi-tenant isolation: every query scoped by `tenant_id` (or DB-level row-level security)
- [ ] `[HIGH]` Centralized authorization logic (middleware/policy engine), not scattered `if` checks
- [ ] `[HIGH]` RBAC or ABAC model documented; roles reviewed periodically
- [ ] `[HIGH]` Use non-guessable IDs (UUIDv4/ULID) for external references (defense in depth, not a substitute for authZ)
- [ ] `[MEDIUM]` Property-level authorization: users can't read/write fields they shouldn't (OWASP API3)

**Tools:** Casbin, OPA (Open Policy Agent), Oso, Cerbos, PostgreSQL RLS

**Test:** Create two users in two tenants; replay every request of user A with user B's token. Automate with Burp Autorize extension.

---

## 3. Input Validation & Injection

- [ ] `[CRITICAL]` **SQL injection**: parameterized queries / prepared statements everywhere; no string concatenation
- [ ] `[CRITICAL]` **NoSQL injection**: reject operator objects (`$gt`, `$ne`, `$where`) from user input in MongoDB
- [ ] `[CRITICAL]` **OS command injection**: avoid `exec`/shell; use argument arrays if unavoidable
- [ ] `[CRITICAL]` **SSRF**: allow-list outbound hosts; block internal IPs (169.254.169.254, 10/8, 172.16/12, 192.168/16, localhost, IPv6 equivalents); re-check after DNS resolution and redirects
- [ ] `[HIGH]` **Path traversal**: normalize paths, reject `../`, use allow-listed directories
- [ ] `[HIGH]` **XXE**: disable external entities in XML parsers
- [ ] `[HIGH]` **Insecure deserialization**: never deserialize untrusted data with native serializers (Java `ObjectInputStream`, `node-serialize`)
- [ ] `[HIGH]` **Template injection (SSTI)**: never render user input as a template
- [ ] `[HIGH]` **Prototype pollution** (Node.js): validate keys, use `Object.create(null)` or Maps, safe merge libraries
- [ ] `[HIGH]` Schema validation on all request bodies, params, query strings, headers
- [ ] `[MEDIUM]` **ReDoS**: avoid catastrophic regex on user input
- [ ] `[MEDIUM]` **LDAP / XPath / header / log injection** handled
- [ ] `[MEDIUM]` Unicode normalization before comparison (usernames, emails)

**Tools:** Zod, Joi, Yup (Node); `go-playground/validator` (Go); Jakarta Bean Validation (Java); sqlmap, NoSQLMap for testing

---

## 4. API Security (OWASP API Top 10 2023)

- [ ] `[CRITICAL]` API1 — Broken Object Level Authorization
- [ ] `[CRITICAL]` API2 — Broken Authentication
- [ ] `[HIGH]` API3 — Broken Object Property Level Authorization (mass assignment + excessive data exposure)
- [ ] `[HIGH]` API4 — Unrestricted Resource Consumption (rate limits, pagination caps, payload size, timeouts)
- [ ] `[HIGH]` API5 — Broken Function Level Authorization
- [ ] `[HIGH]` API6 — Unrestricted Access to Sensitive Business Flows (bots buying stock, mass signups)
- [ ] `[HIGH]` API7 — Server Side Request Forgery
- [ ] `[HIGH]` API8 — Security Misconfiguration
- [ ] `[MEDIUM]` API9 — Improper Inventory Management (old versions, undocumented/shadow endpoints)
- [ ] `[MEDIUM]` API10 — Unsafe Consumption of APIs (trusting third-party responses)

### Additional API controls
- [ ] `[HIGH]` Mass assignment blocked: explicit DTOs/allow-listed fields
- [ ] `[HIGH]` Responses return only necessary fields (no `SELECT *` serialized to client)
- [ ] `[HIGH]` Rate limiting per IP, per user, per API key; stricter on auth endpoints
- [ ] `[HIGH]` Request body size limits; upload size limits
- [ ] `[MEDIUM]` Pagination with maximum page size enforced
- [ ] `[MEDIUM]` OpenAPI spec maintained; actual endpoints match spec
- [ ] `[MEDIUM]` Deprecated API versions retired on schedule
- [ ] `[MEDIUM]` API keys: hashed in DB, scoped, revocable, rotatable

### GraphQL
- [ ] `[HIGH]` Query depth and complexity limits
- [ ] `[HIGH]` Introspection disabled in production
- [ ] `[HIGH]` Field-level authorization in resolvers
- [ ] `[MEDIUM]` Batching/alias abuse limited (brute force via aliases)
- [ ] `[MEDIUM]` Persisted queries for public clients

### gRPC
- [ ] `[HIGH]` TLS/mTLS enabled
- [ ] `[HIGH]` Auth interceptors on every service
- [ ] `[MEDIUM]` Reflection disabled in production; message size limits set

---

## 5. Data Protection & Cryptography

- [ ] `[CRITICAL]` TLS 1.2+ everywhere (prefer TLS 1.3); no plain HTTP in prod
- [ ] `[CRITICAL]` Encryption at rest for databases, backups, object storage
- [ ] `[HIGH]` Field-level encryption for highly sensitive fields (national IDs, BVN/NIN, health data)
- [ ] `[HIGH]` Keys managed in KMS/HSM; key rotation policy
- [ ] `[HIGH]` Approved algorithms only: AES-256-GCM, ChaCha20-Poly1305, RSA ≥ 2048 (prefer 3072), ECDSA P-256, Ed25519, SHA-256+
- [ ] `[HIGH]` Cryptographically secure randomness (`crypto.randomBytes`, `crypto/rand`, `SecureRandom`) for tokens, IDs, OTPs
- [ ] `[HIGH]` No custom cryptography
- [ ] `[MEDIUM]` **FIPS 140-3** validated modules if required (US government, some finance)
- [ ] `[MEDIUM]` Data masking/tokenization in non-prod environments
- [ ] `[LOW]` **Post-quantum readiness**: crypto inventory and agility plan (NIST PQC standards: ML-KEM, ML-DSA)

---

## 6. Database Security

- [ ] `[CRITICAL]` Database not publicly exposed to the internet
- [ ] `[CRITICAL]` App connects with least-privilege user (no superuser/root/`postgres`)
- [ ] `[HIGH]` Separate DB users for app, migrations, read-only analytics
- [ ] `[HIGH]` TLS between app and database
- [ ] `[HIGH]` **Row-Level Security** (PostgreSQL) for multi-tenant data where suitable
- [ ] `[HIGH]` MongoDB: auth enabled, bound to private interface, no default ports exposed
- [ ] `[MEDIUM]` Connection pool limits and statement timeouts
- [ ] `[MEDIUM]` Audit logging for sensitive tables (pgAudit, MySQL audit plugin)
- [ ] `[MEDIUM]` Migrations reviewed; no destructive migrations without backup
- [ ] `[MEDIUM]` Constraints (FK, unique, check) enforce integrity (see file 08)

---

## 7. Cache & Message Queue Security

- [ ] `[CRITICAL]` Redis: `requirepass`/ACLs enabled, not internet-exposed, `protected-mode yes`
- [ ] `[HIGH]` Dangerous Redis commands renamed/disabled (`FLUSHALL`, `CONFIG`, `KEYS`)
- [ ] `[HIGH]` TLS for Redis and queues in transit
- [ ] `[HIGH]` Queue (RabbitMQ, Kafka, SQS) access via least-privilege credentials
- [ ] `[MEDIUM]` Cache keys namespaced by tenant/user; no cross-tenant cache leaks
- [ ] `[MEDIUM]` Cache poisoning: don't cache responses varying on unkeyed headers
- [ ] `[MEDIUM]` Messages validated on consumption (treat as untrusted input)
- [ ] `[MEDIUM]` Dead-letter queues monitored

---

## 8. File Uploads

- [ ] `[CRITICAL]` Validate file type by content (magic bytes), not just extension or `Content-Type`
- [ ] `[CRITICAL]` Uploaded files never executed; stored outside web root
- [ ] `[HIGH]` Size limits enforced server-side
- [ ] `[HIGH]` Filenames regenerated server-side (UUID)
- [ ] `[HIGH]` Object storage (S3) private by default; access via short-lived signed URLs
- [ ] `[HIGH]` Malware scanning (ClamAV, cloud AV) for user uploads
- [ ] `[MEDIUM]` Images re-encoded to strip metadata (EXIF GPS) and payloads
- [ ] `[MEDIUM]` SVG sanitized or served with `Content-Disposition: attachment`
- [ ] `[MEDIUM]` Zip bomb / decompression limits

---

## 9. Error Handling

- [ ] `[HIGH]` No stack traces, SQL errors, file paths, or framework versions in responses
- [ ] `[HIGH]` Global error handler returns generic messages + correlation ID
- [ ] `[MEDIUM]` `X-Powered-By` and server version headers removed
- [ ] `[MEDIUM]` Fail securely: errors deny access, never grant it
- [ ] `[MEDIUM]` Unhandled promise rejections / panics caught and logged

---

## 10. Business Logic

- [ ] `[CRITICAL]` Prices, totals, discounts calculated server-side only
- [ ] `[CRITICAL]` **Race conditions**: double-spend, double-redeem, coupon reuse prevented (locks, unique constraints, idempotency keys)
- [ ] `[HIGH]` Workflow steps enforced in order (can't skip payment/verification)
- [ ] `[HIGH]` Negative/zero/huge quantities and amounts rejected
- [ ] `[HIGH]` Limits on referrals, promos, free trials (abuse prevention)
- [ ] `[MEDIUM]` Currency precision handled with integers/decimals, never floats

**Test:** Send 20 parallel identical requests (Burp Turbo Intruder, `xargs -P`) to redeem/transfer endpoints.

---

## 11. Webhooks & Third-Party Integrations

- [ ] `[CRITICAL]` Incoming webhooks: HMAC signature verified with constant-time comparison
- [ ] `[HIGH]` Timestamp checked to prevent replay; event IDs deduplicated
- [ ] `[HIGH]` Payment status confirmed by calling provider API, not trusting webhook body alone (e.g., Paystack/Flutterwave/Stripe verify endpoint)
- [ ] `[HIGH]` Outgoing webhooks: signed, sent to validated URLs (SSRF protection)
- [ ] `[MEDIUM]` Third-party API responses validated (OWASP API10)
- [ ] `[MEDIUM]` Third-party credentials scoped and rotated

---

## 12. Background Jobs & Cron

- [ ] `[HIGH]` Jobs run with least privilege
- [ ] `[HIGH]` Job payloads validated; no user-controlled code execution
- [ ] `[MEDIUM]` Jobs are idempotent (safe to retry)
- [ ] `[MEDIUM]` Failures alerted; retries bounded with backoff
- [ ] `[MEDIUM]` Distributed locks prevent duplicate cron runs across instances

---

## 13. Audit Logging

- [ ] `[HIGH]` Record actor, action, target, timestamp, IP, user agent, result, tenant
- [ ] `[HIGH]` Covers auth events, permission changes, data exports, deletions, admin actions, config changes
- [ ] `[MEDIUM]` Append-only storage; app cannot delete its own audit logs
- [ ] `[MEDIUM]` Audit logs searchable and retained per policy

---

## 14. Framework-Specific Quick Checks

### Node.js / Express
- [ ] `helmet` configured; `express.json({ limit })` set
- [ ] `trust proxy` set correctly behind Nginx/load balancer (affects rate limiting and IPs)
- [ ] `express-rate-limit` with Redis store in multi-instance deployments
- [ ] No `eval`, `new Function`, `child_process.exec` with user input
- [ ] Run as non-root; `NODE_ENV=production`

### Go / Gin
- [ ] `gin.SetMode(gin.ReleaseMode)` in prod
- [ ] `SetTrustedProxies` configured
- [ ] HTTP server timeouts set (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ReadHeaderTimeout`)
- [ ] `govulncheck` and `gosec` in CI
- [ ] `html/template` (not `text/template`) for HTML output

### Java / Spring
- [ ] Spring Security configured; actuator endpoints restricted
- [ ] Jackson polymorphic typing disabled for untrusted input
- [ ] SpotBugs + FindSecBugs in CI

---

## Tools Summary

| Purpose | Tools |
|---|---|
| SAST | Semgrep, CodeQL, SonarQube, gosec, SpotBugs/FindSecBugs, eslint-plugin-security |
| API testing | Burp Suite, OWASP ZAP, Postman, 42Crunch, Schemathesis |
| Injection testing | sqlmap, NoSQLMap, commix |
| AuthZ testing | Burp Autorize, AuthMatrix |
| Load / abuse | k6, Artillery, Turbo Intruder |

## References
- OWASP API Security Top 10 (2023)
- OWASP ASVS
- NIST SP 800-63B (Digital Identity — Authentication)
- OWASP Cheat Sheets: Authentication, Authorization, Password Storage, JWT, SSRF, Input Validation
