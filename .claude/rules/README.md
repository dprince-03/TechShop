# Security, Audit & Compliance Checklist Set

A complete reference for auditing backend, frontend, mobile, and fullstack software against security best practices, compliance frameworks, and data-integrity principles.

## Files

| # | File | Covers |
|---|------|--------|
| 00 | [00-common-security-baseline.md](00-common-security-baseline.md) | Controls every app needs regardless of type |
| 01 | [01-backend-security.md](01-backend-security.md) | APIs, auth, databases, queues, jobs, business logic |
| 02 | [02-frontend-security.md](02-frontend-security.md) | Browser security, headers, XSS/CSRF, React/Next.js risks |
| 03 | [03-mobile-security.md](03-mobile-security.md) | Android/iOS, MASVS, storage, pinning, store requirements |
| 04 | [04-fullstack-integration-security.md](04-fullstack-integration-security.md) | End-to-end flows, trust boundaries, multi-tenancy |
| 05 | [05-devsecops-infrastructure.md](05-devsecops-infrastructure.md) | CI/CD, containers, cloud, network, DNS, email |
| 06 | [06-compliance-and-regulatory.md](06-compliance-and-regulatory.md) | SOC 1/2/3, ISO, NIST, PCI, GDPR, NDPA, and more |
| 07 | [07-audits-testing-and-assessments.md](07-audits-testing-and-assessments.md) | Pentests, SAST/DAST/IAST, fuzzing, chaos, bug bounty |
| 08 | [08-data-integrity-reliability-principles.md](08-data-integrity-reliability-principles.md) | ACID, BASE, CAP, idempotency, SLAs, RPO/RTO |
| 09 | [09-glossary-of-standards-and-acronyms.md](09-glossary-of-standards-and-acronyms.md) | Every acronym used across the set |
| 10 | [10-governance-risk-and-people.md](10-governance-risk-and-people.md) | Risk management, policies, HR, change, assets |

## How to use

1. **Start with 00** for every project, then add the files for the app type(s) you are building.
2. **Fullstack apps** use 00 + 01 + 02 + 04 + 05 (and 03 if there is a mobile client).
3. **Use 06 to scope compliance**: pick only the frameworks that apply to your industry, region, and customers.
4. **Use 07 to plan testing cadence** and 10 to produce the evidence auditors ask for.
5. Each checklist item uses `- [ ]` so you can copy it into GitHub issues, Notion, or a tracker.

## Severity tags used

- `[CRITICAL]` must be fixed before production
- `[HIGH]` fix before or immediately after launch
- `[MEDIUM]` fix in the next planned cycle
- `[LOW]` hardening / nice to have

## Disclaimer

Regulations change and their applicability depends on your jurisdiction, industry, and data. This set is a technical engineering reference, not legal advice. Confirm regulatory obligations with a qualified lawyer or compliance professional, and verify current versions of each standard before an audit.
