---
name: performance-optimization
description: Diagnose and improve application performance — slow API endpoints, database queries and indexes, N+1 problems, caching strategy (Redis, HTTP, CDN), concurrency, memory and CPU usage, frontend Core Web Vitals, bundle size, rendering performance, mobile app performance, and load testing. Use this skill whenever the user says something is slow, wants to speed up, optimize, scale, reduce costs or latency, improve Lighthouse/Web Vitals scores, or prepare for higher traffic.
---

# Performance Optimization

## Rule zero: measure first

Never optimize from intuition. Get a baseline number (p50/p95/p99 latency, query time, LCP, bundle size, memory), change one thing, measure again. Report before/after.

## Workflow

1. **Define the target**: e.g., "p95 < 300 ms for `GET /invoices`", "LCP < 2.5 s on 4G mobile".
2. **Profile to find the bottleneck** (usually one dominant cause).
3. **Fix the biggest bottleneck first**; re-measure.
4. **Guard against regression**: perf budget in CI, alert on latency.

## Backend checklist

**Database (most common culprit)**
- [ ] Enable `pg_stat_statements`; sort by total time
- [ ] `EXPLAIN (ANALYZE, BUFFERS)`; eliminate sequential scans on large tables with proper composite indexes (`tenant_id` first)
- [ ] Remove N+1: eager load / joins / batch `WHERE id IN (...)` / DataLoader for GraphQL
- [ ] Select only needed columns; paginate everything
- [ ] Keyset (cursor) pagination for deep pages instead of large `OFFSET`
- [ ] Connection pool sized correctly; PgBouncer for many instances
- [ ] Move heavy reporting to read replicas or materialized views

**Application**
- [ ] Timeouts on every outbound call; parallelize independent calls (`Promise.all`, goroutines + errgroup)
- [ ] Move slow non-critical work to background jobs (emails, PDFs, webhooks)
- [ ] Avoid blocking the Node event loop (CPU-heavy work → worker threads/queue)
- [ ] Go: profile with pprof (CPU, heap, goroutines); reduce allocations in hot paths
- [ ] Compress responses (gzip/brotli); avoid oversized JSON payloads

**Caching**
- [ ] Cache-aside in Redis for expensive, read-heavy, tolerably-stale data; tenant-namespaced keys; explicit TTLs; invalidate on write
- [ ] HTTP caching headers (`Cache-Control`, `ETag`) for public/static content
- [ ] CDN for static assets and public pages

## Frontend checklist (Core Web Vitals)

Targets: **LCP < 2.5 s**, **INP < 200 ms**, **CLS < 0.1**.
- [ ] Measure with Lighthouse + real-user data (Vercel Analytics, web-vitals library)
- [ ] LCP: optimize hero image (`next/image`, `priority`, modern formats, correct sizes), server-render above-the-fold, preconnect to critical origins
- [ ] INP: break up long tasks, reduce JS, avoid heavy synchronous handlers, debounce inputs
- [ ] CLS: set width/height on images/embeds, reserve space for dynamic content, `next/font`
- [ ] Bundle: analyze (`@next/bundle-analyzer`), code-split routes, dynamic import heavy components, drop unused libraries (moment → date-fns/dayjs, lodash → per-method)
- [ ] React: avoid unnecessary re-renders (memoization where measured, stable props, state colocated), virtualize long lists
- [ ] Server Components by default in Next.js to ship less JS

## Mobile

- [ ] Profile with Flipper/React Native DevTools or Flutter DevTools
- [ ] FlatList/ListView.builder with keys; avoid inline functions in hot lists; cache images
- [ ] Reduce startup time: lazy-load screens, defer non-critical init
- [ ] Minimize network calls; cache with React Query / local DB

## Load testing

- k6 / Artillery scripts for critical endpoints; ramp to 2–3× expected peak.
- Watch latency percentiles, error rate, DB CPU/connections, memory.
- Run against staging with production-like data volume.

## Report format

```
**Bottleneck:** ...
**Evidence:** profile/query plan/metric
**Change:** ...
**Result:** before → after (p95, LCP, etc.)
**Next opportunities:** ...
```
