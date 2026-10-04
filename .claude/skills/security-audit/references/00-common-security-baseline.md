# 00 — Common Security Baseline

Controls that apply to **every** application: backend, frontend, mobile, or fullstack. Complete this file first.

---

## 1. Threat Modeling

- [ ] `[HIGH]` A threat model exists for the system and is updated on major architecture changes
- [ ] `[HIGH]` Data flow diagrams (DFDs) show trust boundaries, data stores, external entities
- [ ] `[MEDIUM]` Threats categorized using **STRIDE** (Spoofing, Tampering, Repudiation, Information disclosure, Denial of service, Elevation of privilege)
- [ ] `[MEDIUM]` High-risk systems analyzed with **PASTA** or attack trees
- [ ] `[MEDIUM]` Threats mapped to **MITRE ATT&CK** techniques where useful
- [ ] `[MEDIUM]` Each identified threat has a mitigation, owner, and status

**How to verify:** Review the threat model doc; confirm it matches current architecture; pick 3 threats and trace them to implemented controls.

**Tools:** OWASP Threat Dragon, Microsoft Threat Modeling Tool, IriusRisk, draw.io

---

## 2. Secure SDLC

- [ ] `[HIGH]` Security requirements defined at design stage (mapped to **OWASP ASVS** level 1/2/3)
- [ ] `[HIGH]` Mandatory peer code review with a security checklist
- [ ] `[HIGH]` SAST, SCA, and secret scanning run on every pull request
- [ ] `[MEDIUM]` Security champions assigned per team
- [ ] `[MEDIUM]` Maturity assessed with **OWASP SAMM** or **BSIMM**
- [ ] `[MEDIUM]` Practices aligned with **NIST SSDF (SP 800-218)**
- [ ] `[LOW]` Developers follow **OWASP Proactive Controls** and **Cheat Sheet Series**

**ASVS levels:**
| Level | Use for |
|-------|---------|
| L1 | All apps; low-assurance, fully testable via pentest |
| L2 | Apps handling sensitive data (most business apps) |
| L3 | Critical apps: finance, health, high-value transactions |

---

## 3. OWASP Top 10 (Web) Coverage

- [ ] `[CRITICAL]` Broken Access Control
- [ ] `[CRITICAL]` Cryptographic Failures
- [ ] `[CRITICAL]` Injection (SQL, NoSQL, OS command, LDAP, XSS)
- [ ] `[HIGH]` Insecure Design
- [ ] `[HIGH]` Security Misconfiguration
- [ ] `[HIGH]` Vulnerable and Outdated Components / Software Supply Chain Failures
- [ ] `[HIGH]` Identification and Authentication Failures
- [ ] `[HIGH]` Software and Data Integrity Failures
- [ ] `[MEDIUM]` Security Logging and Monitoring Failures
- [ ] `[MEDIUM]` Server-Side Request Forgery (SSRF) / Mishandling of Exceptional Conditions

> Check the current OWASP Top 10 edition at owasp.org; categories are revised every few years.

---

## 4. Data Classification & Inventory

- [ ] `[HIGH]` All data types inventoried (what, where stored, who accesses, retention)
- [ ] `[HIGH]` Data classified: **Public / Internal / Confidential / Restricted**
- [ ] `[HIGH]` PII, financial, health, and credentials flagged as Restricted
- [ ] `[MEDIUM]` Data flow mapping covers third parties and cross-border transfers
- [ ] `[MEDIUM]` Data minimization: only collect what is needed

---

## 5. Secrets Management

- [ ] `[CRITICAL]` No secrets in source code, Git history, Docker images, or mobile bundles
- [ ] `[CRITICAL]` Secrets stored in a vault (AWS Secrets Manager, HashiCorp Vault, Doppler, SSM Parameter Store)
- [ ] `[HIGH]` Secrets rotated on schedule and immediately after staff exit or leak
- [ ] `[HIGH]` Different secrets per environment (dev/staging/prod)
- [ ] `[HIGH]` Pre-commit secret scanning enabled
- [ ] `[MEDIUM]` Git history scanned and leaked secrets revoked (not just deleted)

**Tools:** gitleaks, TruffleHog, GitHub Secret Scanning, detect-secrets

---

## 6. Dependency & Supply Chain Security

- [ ] `[CRITICAL]` No dependencies with known critical CVEs (check **CISA KEV** list first)
- [ ] `[HIGH]` Lockfiles committed (`package-lock.json`, `go.sum`, `pom.xml` pinned)
- [ ] `[HIGH]` Automated dependency updates (Dependabot / Renovate)
- [ ] `[HIGH]` **SBOM** generated per release (**SPDX** or **CycloneDX** format)
- [ ] `[MEDIUM]` Open-source licences reviewed (GPL/AGPL obligations understood)
- [ ] `[MEDIUM]` Typosquatting and abandoned-package checks
- [ ] `[MEDIUM]` Vulnerabilities prioritized using **CVSS** + **EPSS** + KEV, not CVSS alone
- [ ] `[LOW]` **OpenSSF Scorecard** reviewed for critical dependencies

**Tools:** `npm audit`, `govulncheck`, OWASP Dependency-Check, Snyk, Trivy, Grype, Syft, Socket.dev

---

## 7. Logging, Monitoring & Alerting

- [ ] `[HIGH]` Security events logged: logins (success/fail), privilege changes, access denials, admin actions, data exports
- [ ] `[HIGH]` Logs never contain passwords, tokens, full card numbers, or unmasked PII
- [ ] `[HIGH]` Centralized log aggregation with access control
- [ ] `[HIGH]` Alerts on anomalies (brute force, spikes in 4xx/5xx, privilege escalation)
- [ ] `[MEDIUM]` Log retention defined (commonly 90 days hot, 1 year archive; check regulations)
- [ ] `[MEDIUM]` Logs are tamper-evident (write-once storage or hashing)
- [ ] `[MEDIUM]` Time synchronized (NTP) across all systems

**Tools:** ELK/OpenSearch, Grafana Loki, Datadog, CloudWatch, Sentry, Wazuh

---

## 8. Incident Response

- [ ] `[HIGH]` Written incident response plan (roles, severity levels, escalation, contacts)
- [ ] `[HIGH]` Breach notification timelines documented (e.g., GDPR 72 hours; NDPA 72 hours to NDPC)
- [ ] `[MEDIUM]` Runbooks for common incidents (credential leak, DDoS, ransomware, data exposure)
- [ ] `[MEDIUM]` Tabletop exercise at least annually
- [ ] `[MEDIUM]` Post-incident reviews (blameless) with tracked action items

**Framework:** NIST SP 800-61 (Incident Handling Guide)

---

## 9. Backup, DR & Business Continuity

- [ ] `[CRITICAL]` Automated backups of all production data
- [ ] `[HIGH]` Backups encrypted and stored in a separate account/region
- [ ] `[HIGH]` Restore tested at least quarterly
- [ ] `[HIGH]` **RPO** and **RTO** defined per system (see file 08)
- [ ] `[MEDIUM]` Immutable/air-gapped backup copy for ransomware resilience
- [ ] `[MEDIUM]` Business continuity plan aligned with **ISO 22301** if required

---

## 10. Access Control Principles

- [ ] `[CRITICAL]` Least privilege for users, services, and admins
- [ ] `[HIGH]` Separation of duties (dev cannot deploy to prod alone; approver ≠ requester)
- [ ] `[HIGH]` MFA required for all admin, cloud console, Git, and CI/CD access
- [ ] `[HIGH]` Default deny on all access decisions
- [ ] `[MEDIUM]` Zero-trust approach: verify every request, don't trust network location

---

## 11. Security Documentation

- [ ] `[MEDIUM]` Architecture diagram (current)
- [ ] `[MEDIUM]` Security policy and secure coding standard
- [ ] `[MEDIUM]` Code review checklist
- [ ] `[MEDIUM]` Data retention and deletion policy
- [ ] `[LOW]` Security decisions recorded in ADRs (Architecture Decision Records)

---

## References

- OWASP ASVS, OWASP Top 10, OWASP SAMM, OWASP Cheat Sheet Series
- NIST SSDF (SP 800-218), NIST SP 800-61
- CISA Known Exploited Vulnerabilities Catalog
