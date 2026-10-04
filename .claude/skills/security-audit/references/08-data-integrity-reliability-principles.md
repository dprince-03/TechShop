# 08 — Data Integrity & Reliability Principles

The core principles behind correct, consistent, and resilient systems: ACID, BASE, CAP, idempotency, SLAs, RPO/RTO, and secure design principles. Each section has an explanation and an audit checklist.

---

## 1. ACID (Relational Transactions)

| Property | Meaning | Example |
|---|---|---|
| **Atomicity** | All operations in a transaction succeed or none do | Debit and credit in a transfer both happen or both roll back |
| **Consistency** | Transactions move DB from one valid state to another (constraints hold) | Balance can't go below zero if a CHECK constraint exists |
| **Isolation** | Concurrent transactions don't interfere in unsafe ways | Two withdrawals don't both read the same old balance |
| **Durability** | Committed data survives crashes | Write-ahead log flushed before commit returns |

### Isolation levels and anomalies

| Level | Dirty read | Non-repeatable read | Phantom read | Lost update / write skew |
|---|---|---|---|---|
| Read Uncommitted | Possible | Possible | Possible | Possible |
| Read Committed (PostgreSQL default) | Prevented | Possible | Possible | Possible |
| Repeatable Read (MySQL InnoDB default) | Prevented | Prevented | Possible* | Possible* |
| Serializable | Prevented | Prevented | Prevented | Prevented |

\* Behavior varies by engine: PostgreSQL's Repeatable Read (snapshot isolation) prevents phantoms but allows write skew; MySQL uses gap locks in some cases.

### Checklist
- [ ] `[CRITICAL]` Multi-step financial or inventory operations wrapped in transactions
- [ ] `[HIGH]` Isolation level chosen deliberately for critical flows (often `SERIALIZABLE` or explicit row locks)
- [ ] `[HIGH]` Serialization failures / deadlocks handled with retries
- [ ] `[HIGH]` No external side effects (emails, API calls) inside DB transactions without outbox pattern
- [ ] `[MEDIUM]` MongoDB multi-document transactions used where atomicity across documents is required (replica set needed)
- [ ] `[MEDIUM]` `synchronous_commit` / durability settings understood (don't disable for critical data)

---

## 2. BASE (Distributed / NoSQL Systems)

| Property | Meaning |
|---|---|
| **Basically Available** | System responds even during partial failures |
| **Soft state** | State may change over time without new input (replication catching up) |
| **Eventual consistency** | All replicas converge if no new updates occur |

- [ ] `[HIGH]` Features relying on eventually consistent stores tolerate stale reads
- [ ] `[HIGH]` Critical decisions (balances, permissions) read from the strongly consistent source
- [ ] `[MEDIUM]` Read-after-write handled (route reads to primary after writes, or use session consistency)

---

## 3. CAP Theorem & PACELC

**CAP:** During a network **P**artition, a distributed system must choose between **C**onsistency (all nodes return the same latest data) and **A**vailability (every request gets a response).

**PACELC:** If **P**artition → choose **A** or **C**; **E**lse (normal operation) → choose **L**atency or **C**onsistency.

| System (typical config) | Partition choice | Normal choice |
|---|---|---|
| PostgreSQL (single primary) | C | C |
| MongoDB (majority write concern) | C | C (tunable) |
| Cassandra / DynamoDB (default) | A | L |
| Redis (replication) | A (async replication can lose writes) | L |

- [ ] `[MEDIUM]` Each data store's CAP/PACELC behavior documented
- [ ] `[MEDIUM]` Write/read concerns (MongoDB), consistency levels (Cassandra), or strongly consistent reads (DynamoDB) configured per use case

---

## 4. Consistency Models

| Model | Guarantee |
|---|---|
| **Strong / Linearizable** | Every read sees the latest write |
| **Sequential** | All nodes see operations in the same order |
| **Causal** | Causally related operations seen in order |
| **Read-your-writes** | A user always sees their own updates |
| **Monotonic reads** | A user never sees older data after newer data |
| **Eventual** | Replicas converge eventually |

- [ ] `[MEDIUM]` Required consistency model defined per feature (e.g., read-your-writes for profile edits, strong for payments)

---

## 5. Idempotency

An operation is idempotent if performing it multiple times has the same effect as once.

- [ ] `[CRITICAL]` Payment, transfer, and order-creation endpoints accept an **Idempotency-Key** header
- [ ] `[HIGH]` Idempotency keys stored with request hash and response; replays return the stored response
- [ ] `[HIGH]` Webhook and queue consumers deduplicate by event/message ID
- [ ] `[HIGH]` Background jobs safe to retry
- [ ] `[MEDIUM]` HTTP semantics respected: GET, PUT, DELETE idempotent; POST made idempotent via keys where needed

---

## 6. Concurrency Control

| Strategy | How | Use when |
|---|---|---|
| **Optimistic locking** | Version column; update fails if version changed | Low contention, user-edited records |
| **Pessimistic locking** | `SELECT ... FOR UPDATE` | High contention, money, inventory |
| **Atomic operations** | `UPDATE stock = stock - 1 WHERE stock > 0` | Counters, inventory |
| **Unique constraints** | DB rejects duplicates | Preventing double redemption/signup |
| **Distributed locks** | Redis (Redlock caveats), etcd, ZooKeeper, DB advisory locks | Cross-instance cron, singletons |

- [ ] `[CRITICAL]` Race conditions tested on balance, inventory, coupon, and voting endpoints
- [ ] `[HIGH]` Check-then-act patterns replaced by atomic DB operations or locks
- [ ] `[MEDIUM]` Distributed lock usage has timeouts and fencing tokens where correctness matters

---

## 7. Distributed Transactions

| Pattern | Description | Trade-off |
|---|---|---|
| **Two-Phase Commit (2PC)** | Coordinator asks all participants to prepare, then commit | Strong consistency; blocking, poor availability |
| **Saga** | Sequence of local transactions with compensating actions | Available; requires careful compensation design |
| **Transactional Outbox** | Write event to an outbox table in the same DB transaction; relay publishes it | Reliable event publishing without dual writes |
| **Inbox pattern** | Consumers record processed message IDs | Exactly-once *effect* on consumption |
| **Change Data Capture (CDC)** | Stream DB changes (Debezium) | Decoupled, reliable propagation |

- [ ] `[HIGH]` No "dual writes" (DB write + message publish) without outbox/CDC
- [ ] `[HIGH]` Sagas have defined compensating actions and failure handling
- [ ] `[MEDIUM]` Message delivery semantics understood (at-most-once, at-least-once, effectively-once)

---

## 8. Data Validation & Referential Integrity

- [ ] `[HIGH]` Foreign keys, NOT NULL, UNIQUE, and CHECK constraints enforced at DB level (not only in app code)
- [ ] `[HIGH]` Cascading deletes reviewed (avoid accidental mass deletion)
- [ ] `[MEDIUM]` MongoDB schema validation (`$jsonSchema`) for critical collections
- [ ] `[MEDIUM]` Checksums/hashes for file integrity and data transfers
- [ ] `[MEDIUM]` Periodic reconciliation jobs detect drift between systems (e.g., ledger vs payment provider)
- [ ] `[MEDIUM]` Money stored as integer minor units (kobo, cents) or DECIMAL, never float
- [ ] `[MEDIUM]` Timestamps stored in UTC with timezone awareness

---

## 9. CIA Triad

| Pillar | Goal | Example controls |
|---|---|---|
| **Confidentiality** | Only authorized access | Encryption, access control, MFA |
| **Integrity** | Data is accurate and unaltered | Hashing, signatures, constraints, audit logs |
| **Availability** | Systems accessible when needed | Redundancy, backups, DDoS protection, scaling |

- [ ] `[MEDIUM]` Each critical asset rated for C, I, and A impact; controls proportionate

---

## 10. AAA

| Component | Question | Examples |
|---|---|---|
| **Authentication** | Who are you? | Passwords, MFA, passkeys, certificates |
| **Authorization** | What can you do? | RBAC, ABAC, policies |
| **Accounting** (Auditing) | What did you do? | Audit logs, usage records |

- [ ] `[HIGH]` All three implemented for every user-facing and admin interface

---

## 11. Non-Repudiation

- [ ] `[MEDIUM]` Critical actions (approvals, contracts, financial instructions) linked to authenticated identity with tamper-evident logs
- [ ] `[MEDIUM]` Digital signatures for documents/transactions where legally relevant
- [ ] `[LOW]` Hash-chained or append-only audit logs (e.g., ledger databases, S3 Object Lock)

---

## 12. Availability Targets: SLA, SLO, SLI

| Term | Meaning | Example |
|---|---|---|
| **SLI** (Indicator) | A measured metric | % of requests succeeding under 300 ms |
| **SLO** (Objective) | Internal target for an SLI | 99.9% monthly |
| **SLA** (Agreement) | Contractual promise with penalties | 99.5% monthly or service credits |
| **Error budget** | Allowed unreliability = 100% − SLO | 0.1% ≈ 43 minutes/month |

### Uptime cheat sheet
| Availability | Downtime per year | Per month |
|---|---|---|
| 99% | ~3.65 days | ~7.3 hours |
| 99.9% | ~8.77 hours | ~43.8 minutes |
| 99.95% | ~4.38 hours | ~21.9 minutes |
| 99.99% | ~52.6 minutes | ~4.4 minutes |
| 99.999% | ~5.26 minutes | ~26 seconds |

- [ ] `[MEDIUM]` SLIs and SLOs defined for critical user journeys
- [ ] `[MEDIUM]` SLA promised to customers is looser than internal SLO
- [ ] `[MEDIUM]` Error budget policy guides release pace

---

## 13. Recovery Targets: RPO & RTO

| Term | Meaning | Example |
|---|---|---|
| **RPO** (Recovery Point Objective) | Maximum acceptable data loss (time) | 15 minutes → backups/replication at least every 15 min |
| **RTO** (Recovery Time Objective) | Maximum acceptable downtime | 1 hour → must restore service within 1 hour |
| MTTR | Mean Time To Recovery/Repair | Measured average |
| MTBF | Mean Time Between Failures | Measured average |

- [ ] `[HIGH]` RPO and RTO defined per system and approved by the business
- [ ] `[HIGH]` Backup frequency and replication meet RPO (e.g., PostgreSQL PITR with WAL archiving)
- [ ] `[HIGH]` Restore drills prove RTO is achievable
- [ ] `[MEDIUM]` DR strategy chosen: backup & restore, pilot light, warm standby, or multi-site active-active

---

## 14. Resilience Patterns

- [ ] `[HIGH]` **Timeouts** on every outbound call (HTTP, DB, cache)
- [ ] `[HIGH]` **Retries with exponential backoff and jitter**, only for idempotent operations
- [ ] `[HIGH]` **Circuit breakers** around unreliable dependencies
- [ ] `[MEDIUM]` **Bulkheads**: isolate resource pools so one failure doesn't sink everything
- [ ] `[MEDIUM]` **Rate limiting & load shedding** to protect under overload
- [ ] `[MEDIUM]` **Graceful degradation**: serve cached/partial results when dependencies fail
- [ ] `[MEDIUM]` **Health checks**: liveness vs readiness separated
- [ ] `[MEDIUM]` **Graceful shutdown**: drain connections and finish in-flight work

**Libraries:** opossum, cockatiel (Node); sony/gobreaker, failsafe-go (Go); Resilience4j (Java)

---

## 15. Security Design Principles

| Principle | Meaning |
|---|---|
| **Least privilege** | Grant only the minimum access needed |
| **Defense in depth** | Multiple layers so one failure isn't fatal |
| **Zero trust** | Never trust by network location; verify every request |
| **Secure by default** | Safe configuration out of the box |
| **Fail securely** | Errors result in denial, not access |
| **Separation of duties** | No single person controls a critical process end-to-end |
| **Complete mediation** | Check authorization on every access, not just the first |
| **Economy of mechanism** | Keep security designs simple |
| **Open design** | Security doesn't depend on secrecy of design (Kerckhoffs's principle) |
| **Minimize attack surface** | Remove unused features, endpoints, ports, dependencies |
| **Psychological acceptability** | Security that users can actually follow |

- [ ] `[MEDIUM]` Architecture review confirms each principle is applied (or documented exceptions)

---

## 16. Twelve-Factor App (security-relevant factors)

| Factor | Security/reliability relevance |
|---|---|
| I. Codebase | One repo per app, tracked in version control |
| II. Dependencies | Explicitly declared and isolated (lockfiles) |
| III. Config | Config and secrets in environment, not code |
| IV. Backing services | Treat DBs/queues as attachable resources (easy rotation/failover) |
| V. Build, release, run | Strictly separated stages (reproducible, auditable releases) |
| VI. Processes | Stateless processes (easy scaling, no local session leaks) |
| IX. Disposability | Fast startup, graceful shutdown |
| X. Dev/prod parity | Fewer environment-specific surprises |
| XI. Logs | Logs as event streams to central aggregation |
| XII. Admin processes | One-off admin tasks run in the same controlled environment |

- [ ] `[LOW]` App reviewed against the twelve factors

---

## References
- *Designing Data-Intensive Applications* (Martin Kleppmann)
- Google SRE Book (SLOs, error budgets)
- Saltzer & Schroeder, "The Protection of Information in Computer Systems" (design principles)
- 12factor.net
- AWS Well-Architected Framework (Reliability & Security pillars)
