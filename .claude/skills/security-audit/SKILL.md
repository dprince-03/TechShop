---
name: security-audit
description: Perform security reviews and audits of code, pull requests, modules, configs, or whole applications (backend, frontend, mobile, fullstack, infrastructure) against OWASP ASVS/Top 10, API Top 10, MASVS, CIS benchmarks, and compliance needs (SOC 2, ISO 27001, PCI DSS, GDPR, NDPA). Use this skill whenever the user asks to audit, security-review, harden, pentest-prep, check for vulnerabilities, assess compliance readiness, or asks "is this secure?" — and also proactively when reviewing auth, payments, file uploads, multi-tenant queries, or infrastructure config.
---

# Security Audit

Audit defensively and produce actionable, prioritized findings. This skill is for protecting the user's own systems.

## Step 1 — Scope

Determine what is being audited and pick the reference checklists:

| Target | Read |
|---|---|
| Any system (always) | `references/00-common-security-baseline.md` |
| APIs, servers, DBs, queues | `references/01-backend-security.md` |
| Web UI, React/Next.js | `references/02-frontend-security.md` |
| Android/iOS/RN/Flutter | `references/03-mobile-security.md` |
| Cross-component flows, multi-tenancy, payments | `references/04-fullstack-integration-security.md` |
| CI/CD, Docker, cloud, Nginx, DNS, email | `references/05-devsecops-infrastructure.md` |
| Compliance readiness | `references/06-compliance-and-regulatory.md` |
| Testing plan / pentest prep | `references/07-audits-testing-and-assessments.md` |
| Transactions, races, idempotency, DR | `references/08-data-integrity-reliability-principles.md` |
| Policies, risk, HR, vendors | `references/10-governance-risk-and-people.md` |
| Unknown acronym | `references/09-glossary-of-standards-and-acronyms.md` |

Read only the sections relevant to the code in front of you — the files are long.

## Step 2 — Quick automated pass (when a repo is available)

Run `scripts/quick_scan.sh <path>` for a fast grep-based sweep (hardcoded secrets, dangerous functions, weak crypto, debug flags, permissive CORS). Treat hits as leads to verify, not findings. Suggest real tools for deeper coverage (Semgrep, gitleaks, Trivy, npm audit, govulncheck).

## Step 3 — Manual review priorities

Review in this order — these cause the most real-world breaches:
1. **Authorization**: IDOR/BOLA, missing role checks, cross-tenant access
2. **Authentication**: token validation, session handling, reset flows
3. **Injection**: SQL/NoSQL/command/SSRF/template
4. **Secrets & config**: hardcoded keys, debug mode, permissive CORS, exposed admin
5. **Business logic**: client-trusted prices, race conditions, missing idempotency
6. **Data exposure**: over-fetching responses, PII in logs, verbose errors
7. **Dependencies & infra**: known CVEs, container/cloud misconfig

## Step 4 — Report

Use this exact structure:

```
# Security Audit — <target>
**Scope:** ...  **Date:** ...  **Method:** manual review + <tools>

## Summary
<2–4 sentences: overall posture, count by severity, top risk>

| Severity | Count |
|---|---|
| Critical | n |
| High | n |
| Medium | n |
| Low | n |

## Findings
### [CRITICAL] F-01: <short title>
- **Location:** `path/file.ext:line`
- **Category:** OWASP A01 Broken Access Control / CWE-639
- **Issue:** what is wrong
- **Impact:** what an attacker could do
- **Evidence:** minimal code excerpt
- **Fix:** concrete corrected code or config
- **Verify:** how to confirm the fix (test case)

## Passed checks
<brief list of notable things done well>

## Recommendations
<prioritized next steps, tooling, and which checklist sections remain unreviewed>
```

## Severity guide

- **Critical**: remotely exploitable, leads to data breach, account takeover, RCE, or money loss with no special access
- **High**: exploitable with low-privilege access or meaningful preconditions; serious data exposure
- **Medium**: defense-in-depth gaps, limited impact, or hard to exploit
- **Low**: hardening, best-practice deviations

## Rules

- Never claim something is secure without having checked it; list what was *not* reviewed.
- Prefer fixes in the project's existing style and libraries.
- Don't produce working exploit payloads beyond what's needed to demonstrate the issue to the owner (a simple test request is fine).
- Compliance statements are engineering guidance, not legal advice.
