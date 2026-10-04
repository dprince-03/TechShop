# Runbooks

## Table of contents
1. Sudden spike in 500 errors
2. Slow API / high latency
3. Database connection exhaustion
4. Memory leak / OOM kills
5. Failed deployment / app won't start
6. CORS errors in browser
7. Auth failures (401/403, token issues)
8. Webhooks not processing
9. Redis / queue problems
10. Nginx 502/504
11. Docker build failures
12. Mobile app API failures
13. Emails not delivered

---

## 1. Sudden spike in 500 errors
1. Check if it coincides with a deploy → **rollback first** if yes.
2. Error tracker (Sentry) → group by error type; check top stack trace.
3. Logs filtered by `level=error`, last 30 min; check dependency errors (DB, Redis, third-party).
4. Check DB/Redis health and connection counts.
5. Check config/env changes and expired credentials/certificates.
6. Fix, deploy, monitor error rate back to baseline.

## 2. Slow API / high latency
1. Identify which endpoints (APM/metrics p95/p99).
2. Check DB: `pg_stat_statements` top by total time; `EXPLAIN ANALYZE` slow queries; look for sequential scans and missing indexes.
3. Look for N+1 queries (many similar queries per request).
4. Check external API calls without timeouts.
5. Check CPU/memory saturation and event-loop lag (Node) / goroutine counts (Go).
6. Check cache hit rate.
7. Fix: index, batch queries, add caching, add timeouts, paginate.

## 3. Database connection exhaustion
Symptoms: `too many connections`, `remaining connection slots are reserved`, timeouts acquiring connections.
1. `SELECT state, count(*) FROM pg_stat_activity GROUP BY state;`
2. Look for many `idle in transaction` → code not committing/rolling back or not releasing connections.
3. Total pool size × instances must be < DB `max_connections` (minus reserve).
4. Long-running queries: `SELECT pid, now()-query_start, query FROM pg_stat_activity WHERE state='active' ORDER BY 2 DESC;`
5. Fix leaks (always release/defer close), reduce pool size, add PgBouncer, set `idle_in_transaction_session_timeout` and `statement_timeout`.

## 4. Memory leak / OOM kills
1. Confirm via container metrics: memory climbs steadily until restart.
2. Node: heap snapshots (`--inspect`, Chrome DevTools) comparing over time; common causes: unbounded in-memory caches, event listeners never removed, global arrays, large response buffering.
3. Go: `net/http/pprof` heap profile; goroutine leak (`/debug/pprof/goroutine`) from blocked channels or missing context cancellation.
4. Fix and add memory alerting.

## 5. Failed deployment / app won't start
1. Read build logs, then runtime logs from the first line after start.
2. Common: missing env var, wrong `PORT` binding (must listen on `0.0.0.0` and platform `PORT`), failed migration, health check path wrong/slow, out-of-memory during build.
3. Run the same image locally with prod-like env.
4. Roll back to last good release while fixing.

## 6. CORS errors in browser
1. Read the exact console message (missing header, origin not allowed, credentials issue, preflight failed).
2. Confirm the API is reachable at all (a 500/404 on preflight looks like CORS).
3. Server must return `Access-Control-Allow-Origin` exactly matching the frontend origin (scheme + host + port) — not `*` when using credentials.
4. Preflight `OPTIONS` must be handled before auth middleware.
5. With cookies: `credentials: 'include'` on fetch + `Access-Control-Allow-Credentials: true` + `SameSite=None; Secure` cookies for cross-site.
6. Check Nginx/CDN isn't stripping or duplicating headers.

## 7. Auth failures
1. 401 vs 403: unauthenticated vs unauthorized.
2. Decode JWT (locally, never paste prod tokens into websites): check `exp`, `iss`, `aud`, clock skew.
3. Signing secret/key mismatch between services or environments.
4. Cookie not sent: domain/path/SameSite/Secure mismatch, or HTTP vs HTTPS.
5. Refresh flow: rotation reuse detection may have revoked the session family.
6. Proxies stripping `Authorization` header.

## 8. Webhooks not processing
1. Provider dashboard: delivery attempts and response codes.
2. Endpoint reachable publicly? Correct URL per environment?
3. Signature verification failing: using raw body (not parsed JSON) for HMAC? Correct secret per environment?
4. Handler timing out: respond 2xx fast, process async.
5. Duplicate processing: dedupe by event ID.
6. Replay failed events from provider dashboard after fix.

## 9. Redis / queue problems
1. `redis-cli INFO` — memory, evictions, connected clients; `SLOWLOG GET`.
2. Evictions → maxmemory too low or missing TTLs.
3. Queue backlog growing → workers down, slow, or crashing; check dead-letter queue.
4. Jobs failing repeatedly → poison message; move to DLQ and inspect payload.

## 10. Nginx 502/504
- 502: upstream down or wrong port/socket → check app is running and listening on expected address.
- 504: upstream too slow → find slow endpoint; tune `proxy_read_timeout` only after fixing root cause.
- `tail -f /var/log/nginx/error.log`; `nginx -t` after config changes.

## 11. Docker build failures
- Cache issues: `--no-cache` to confirm.
- Native module builds failing on Alpine → install build deps or use `-slim` image.
- `COPY` failing → path not in build context or excluded by `.dockerignore`.
- Architecture mismatch (ARM Mac vs AMD64 server) → `--platform linux/amd64`.

## 12. Mobile app API failures
1. Same request works from Postman? Then client-side issue.
2. Android: cleartext blocked (HTTP), emulator localhost is `10.0.2.2`.
3. iOS: ATS blocking non-HTTPS.
4. Certificate pinning mismatch after cert rotation.
5. Old app versions calling removed endpoints → version-gate and force-update.

## 13. Emails not delivered
1. Provider logs: accepted, bounced, deferred, spam complaints.
2. SPF/DKIM/DMARC alignment (MXToolbox).
3. Sending from unverified domain/sender.
4. Sandbox mode (e.g., SES sandbox) restricting recipients.
