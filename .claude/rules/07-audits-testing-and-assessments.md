# 07 — Audits, Testing & Assessments

How to verify that the controls in files 00–06 actually work: testing types, when to run them, tools, and how to track remediation.

---

## 1. Testing Types at a Glance

| Type | What it is | Finds | When |
|---|---|---|---|
| **SAST** | Static analysis of source code | Injection, unsafe APIs, hardcoded secrets | Every PR |
| **SCA** | Software composition analysis | Vulnerable/malicious dependencies, licence issues | Every PR + daily |
| **Secret scanning** | Search code/history for credentials | Leaked keys | Pre-commit + every push |
| **IaC scanning** | Static analysis of Terraform/K8s/Dockerfiles | Misconfigurations | Every PR |
| **Container scanning** | Scan images for vulnerable packages | OS/library CVEs | Every build + registry |
| **DAST** | Black-box testing of running app | XSS, injection, headers, misconfig | Staging, weekly/nightly |
| **IAST** | Instrumented runtime analysis during tests | Data-flow vulnerabilities with low false positives | During QA/integration tests |
| **RASP** | Runtime self-protection in production | Blocks attacks in real time | Production (high-risk apps) |
| **Fuzzing** | Random/malformed input generation | Crashes, parsing bugs, edge cases | Parsers, APIs, file handlers; continuous |
| **Manual code review** | Human security review | Logic flaws, authZ gaps | Sensitive changes; periodically |
| **Penetration testing** | Simulated attack by skilled testers | Chained, real-world exploitable issues | Annually + major releases |
| **Red teaming** | Goal-based adversary simulation (incl. social engineering) | Detection & response gaps | Mature orgs, annually |
| **Vulnerability scanning** | Automated network/host scanning | Open ports, outdated services | Monthly / continuous |
| **Configuration audit** | Compare configs to benchmarks | Drift, insecure settings | Quarterly / continuous |
| **Load / stress testing** | High-volume traffic simulation | DoS weaknesses, resource exhaustion | Before launches, major changes |
| **Chaos engineering** | Inject failures deliberately | Resilience gaps | Mature systems, scheduled |

---

## 2. Security Code Review

- [ ] `[HIGH]` Security checklist used for every PR touching auth, payments, file handling, crypto, or permissions
- [ ] `[HIGH]` Automated SAST findings triaged (true positive / false positive / accepted risk)
- [ ] `[MEDIUM]` Periodic deep review of high-risk modules by a security-focused reviewer
- [ ] `[MEDIUM]` Review covers business logic, not just syntax-level issues

**Tools:** Semgrep, CodeQL, SonarQube, gosec, SpotBugs + FindSecBugs, Bandit (Python), Brakeman (Rails)

---

## 3. Vulnerability Scanning

- [ ] `[HIGH]` External attack surface scanned (all public IPs, domains, subdomains)
- [ ] `[HIGH]` Internal network scanned (authenticated scans where possible)
- [ ] `[MEDIUM]` Asset discovery to find forgotten hosts/subdomains
- [ ] `[MEDIUM]` **ASV scans** quarterly if PCI DSS applies (by an Approved Scanning Vendor)

**Tools:** Nessus, OpenVAS/Greenbone, Qualys, Nuclei, Nmap, Amass, subfinder

---

## 4. Penetration Testing

### Approaches
| Approach | Tester knowledge | Best for |
|---|---|---|
| **Black box** | None | Simulating external attackers |
| **Grey box** | User accounts, some docs | Most cost-effective for apps |
| **White box** | Full source, architecture | Deepest coverage |

### Scope checklist
- [ ] Web application pentest (OWASP WSTG methodology)
- [ ] API pentest (OWASP API Top 10, authZ matrix testing)
- [ ] Mobile app pentest, static + dynamic (OWASP MASTG)
- [ ] Cloud configuration review
- [ ] Network/infrastructure pentest (external and internal)
- [ ] Multi-tenant isolation testing
- [ ] Business logic and payment flow testing
- [ ] Social engineering / phishing simulation (if in scope)

### Engagement checklist
- [ ] Rules of engagement signed (scope, timing, contacts, out-of-scope systems)
- [ ] Test accounts for every role and at least two tenants
- [ ] Cloud provider pentest policies reviewed (AWS allows most testing without pre-approval; check current policy)
- [ ] Production vs staging decision made (staging must mirror prod)
- [ ] Report includes: executive summary, CVSS-scored findings, reproduction steps, remediation guidance
- [ ] Retest after remediation; retest letter obtained (useful for customers and auditors)

**Frequency:** At least annually and after significant changes (required by PCI DSS; expected by SOC 2 and ISO auditors).

**Methodologies:** OWASP WSTG, OWASP MASTG, PTES, OSSTMM, NIST SP 800-115

---

## 5. Red Team Exercises

- [ ] Objectives defined (e.g., access production customer data, gain admin)
- [ ] Includes phishing, credential attacks, lateral movement where agreed
- [ ] Measures detection and response, not just prevention
- [ ] **Purple teaming**: red and blue teams collaborate to improve detections
- [ ] Mapped to **MITRE ATT&CK**; defenses mapped to **MITRE D3FEND**

---

## 6. Mobile App Security Assessment

- [ ] Static analysis of APK/IPA (secrets, insecure configs, permissions)
- [ ] Dynamic analysis on rooted/jailbroken device
- [ ] Traffic interception and pinning bypass attempts
- [ ] Local storage inspection
- [ ] Reverse engineering resistance evaluation
- [ ] Results mapped to MASVS controls

**Tools:** MobSF, Frida, objection, jadx, Burp Suite

---

## 7. Configuration & Cloud Audits

- [ ] `[HIGH]` CIS Benchmark scans for cloud, OS, containers, Kubernetes
- [ ] `[HIGH]` IAM permission review (unused permissions, overly broad policies)
- [ ] `[MEDIUM]` Continuous CSPM (cloud security posture management)

**Tools:** Prowler, ScoutSuite, AWS Security Hub, kube-bench, kubescape, Lynis, docker-bench-security

---

## 8. Access Reviews

- [ ] `[HIGH]` Quarterly review of: production access, cloud console, admin panels, Git org, CI/CD, databases, SaaS tools
- [ ] `[HIGH]` Terminated users removed within defined SLA (e.g., same day)
- [ ] `[MEDIUM]` Privileged access recertified by managers
- [ ] `[MEDIUM]` Evidence (export + sign-off) stored for auditors

---

## 9. Dependency Audits

| Ecosystem | Command / tool |
|---|---|
| Node.js | `npm audit`, `pnpm audit`, `yarn npm audit`, Snyk, Socket.dev |
| Go | `govulncheck ./...` |
| Java | OWASP Dependency-Check, `mvn dependency-check:check`, Snyk |
| Python | `pip-audit`, Safety |
| Containers | Trivy, Grype |
| Multi-language | OSV-Scanner, Dependabot, Renovate |

- [ ] `[HIGH]` Critical/high vulnerabilities with fixes patched within SLA
- [ ] `[MEDIUM]` Unfixable vulnerabilities assessed for reachability and documented as accepted risk
- [ ] `[MEDIUM]` Licence compliance report generated

---

## 10. Fuzzing

- [ ] `[MEDIUM]` Fuzz parsers, file processors, protocol handlers, and public APIs
- [ ] `[MEDIUM]` Go native fuzzing (`go test -fuzz`) for critical functions
- [ ] `[LOW]` Continuous fuzzing for open-source components (OSS-Fuzz)

**Tools:** Go fuzzing, Jazzer (Java), jsfuzz/Jazzer.js (JS), AFL++, libFuzzer, RESTler and Schemathesis (APIs)

---

## 11. Load, Stress & Abuse Testing

- [ ] `[HIGH]` Rate limits verified under load
- [ ] `[HIGH]` Expensive endpoints (search, reports, exports, uploads) tested for resource exhaustion
- [ ] `[MEDIUM]` Auto-scaling and cost limits behave correctly
- [ ] `[MEDIUM]` Graceful degradation under overload

**Tools:** k6, Artillery, Locust, Gatling, JMeter, Vegeta

---

## 12. Chaos Engineering

- [ ] `[MEDIUM]` Steady-state metrics defined before experiments
- [ ] `[MEDIUM]` Experiments: kill instances, inject latency, fail DB/cache, expire certificates, revoke credentials
- [ ] `[MEDIUM]` Start in staging; limit blast radius in production
- [ ] `[LOW]` Game days scheduled with on-call team

**Tools:** Chaos Mesh, LitmusChaos, AWS Fault Injection Service, Gremlin, Toxiproxy

---

## 13. Bug Bounty & Vulnerability Disclosure

- [ ] `[HIGH]` **Vulnerability Disclosure Policy (VDP)** published (scope, safe harbor, how to report)
- [ ] `[HIGH]` `/.well-known/security.txt` with contact, policy URL, expiry (RFC 9116)
- [ ] `[MEDIUM]` Internal handling process aligned with **ISO/IEC 29147** (disclosure) and **ISO/IEC 30111** (handling)
- [ ] `[MEDIUM]` Triage SLA defined (e.g., acknowledge within 3 business days)
- [ ] `[LOW]` Paid bug bounty program once basics are mature

**Platforms:** HackerOne, Bugcrowd, Intigriti, YesWeHack

---

## 14. Tabletop Incident Response Drills

- [ ] `[MEDIUM]` At least annual tabletop with engineering, leadership, legal, and comms
- [ ] `[MEDIUM]` Scenarios: leaked AWS keys, ransomware, database dump posted online, insider misuse, third-party breach
- [ ] `[MEDIUM]` Test regulator notification timelines (GDPR, NDPA 72 hours)
- [ ] `[MEDIUM]` Lessons learned documented and actioned

---

## 15. Vulnerability Management & Remediation Tracking

### Prioritization inputs
- **CVSS**: severity score (0–10)
- **EPSS**: probability of exploitation in the next 30 days
- **CISA KEV**: known actively exploited vulnerabilities (fix first)
- **Reachability**: is the vulnerable code actually used?
- **Asset criticality**: internet-facing? holds sensitive data?

### Suggested remediation SLAs
| Severity | Internet-facing | Internal |
|---|---|---|
| Critical / KEV | 7 days | 14 days |
| High | 30 days | 60 days |
| Medium | 90 days | 120 days |
| Low | Best effort / next cycle | Best effort |

### Tracking checklist
- [ ] All findings in one tracker (Jira, GitHub Issues, DefectDojo)
- [ ] Each finding: ID, source, severity, owner, due date, status
- [ ] Exceptions/risk acceptances approved and time-limited
- [ ] Monthly metrics: open by severity, mean time to remediate, SLA breaches
- [ ] Retest evidence attached before closing

**Tools:** DefectDojo, Faraday, Jira, GitHub Security Overview

---

## 16. Recommended Annual Calendar (small–mid team)

| Cadence | Activity |
|---|---|
| Every PR | SAST, SCA, secrets, IaC scan, code review |
| Every build | Container scan, SBOM |
| Weekly | DAST on staging, dependency update review |
| Monthly | Vulnerability scan, metrics review, patching cycle |
| Quarterly | Access reviews, backup restore test, config audit, (PCI ASV scan) |
| Twice yearly | Threat model refresh, policy review |
| Annually | Pentest, tabletop exercise, risk assessment, security training, DR test, SOC 2/ISO audit |

---

## References
- OWASP WSTG, OWASP MASTG, PTES, NIST SP 800-115
- FIRST CVSS, FIRST EPSS, CISA KEV
- RFC 9116 (security.txt), ISO/IEC 29147, ISO/IEC 30111
