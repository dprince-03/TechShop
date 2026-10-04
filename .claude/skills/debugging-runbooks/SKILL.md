---
name: debugging-runbooks
description: Systematic debugging and incident troubleshooting for backend, frontend, mobile, database, and infrastructure problems — reproducing issues, reading logs and stack traces, isolating root causes, and following runbooks for common failures (500 errors, slow APIs, DB connection exhaustion, memory leaks, failed deploys, CORS errors, auth/token failures, webhook failures, Redis/queue issues, Docker/Nginx errors). Use this skill whenever the user reports a bug, error message, crash, outage, "it's not working", unexpected behaviour, or performance degradation, or pastes a stack trace or log.
---

# Debugging & Runbooks

## Method (always)

1. **Clarify the symptom**: exact error text, where it appears, since when, who is affected, what changed recently (deploy, config, dependency, data, traffic).
2. **Reproduce**: smallest reliable reproduction; note environment (local/staging/prod).
3. **Gather evidence before guessing**: logs (with request ID), stack trace, metrics, recent commits, config diff.
4. **Form hypotheses, ranked by likelihood × ease of checking.** Test one at a time.
5. **Isolate**: binary search (git bisect, disable components, minimal input).
6. **Fix the root cause**, not just the symptom; add a regression test.
7. **Verify** in the environment where it failed.
8. **Prevent**: monitoring/alert, validation, or doc update. For production incidents, write a short blameless postmortem.

If the user is mid-outage: prioritize **mitigation first** (rollback, feature flag off, scale up, failover), root cause second.

## Reading a stack trace

- Find the first frame in *your* code (not framework/library frames).
- Read the error type and message literally; check the line and the data flowing into it.
- For Go panics: look for `nil pointer dereference`, index out of range, and the goroutine that panicked.
- For Node: `UnhandledPromiseRejection` means a missing `await`/`catch`; `ECONNREFUSED` means the target isn't reachable.

## Response format

```
**Likely cause:** <one sentence>
**Why I think so:** <evidence>
**Check:** <commands/queries to confirm>
**Fix:** <code/config change>
**Prevent:** <test, alert, or guard>
```
If evidence is insufficient, list the top 2–3 hypotheses and exactly what information would distinguish them.

## Runbooks

Read `references/runbooks.md` for step-by-step procedures for common failures.
