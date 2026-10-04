# Redis Key Design

- Namespaced keys: `<app>:<env>:<tenant>:<entity>:<id>` e.g. `trp:prod:t_123:session:abc`.
- Always set TTLs on cache keys; document TTL per key type.
- Don't use Redis as the system of record unless persistence (AOF) is configured and understood.
- Avoid `KEYS *` in code; use `SCAN`.
- Rate limiting: sliding window with sorted sets or fixed window with `INCR` + `EXPIRE`.
- Distributed locks: `SET key value NX PX <ttl>` with unique value; release via Lua compare-and-delete.
- Invalidate caches on write; prefer cache-aside pattern.
