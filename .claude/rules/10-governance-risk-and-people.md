# 10 — Governance, Risk & People

The organizational controls that SOC 2, ISO 27001, PCI DSS, and most enterprise customers require in addition to technical security. Often the biggest gap for small teams and startups.

---

## 1. Security Governance

- [ ] `[HIGH]` Named owner for security (CISO, security lead, or designated engineer for small teams)
- [ ] `[HIGH]` Leadership approves security policies and reviews security status at least annually
- [ ] `[MEDIUM]` RACI matrix for security responsibilities (who owns access reviews, patching, incident response, vendor reviews)
- [ ] `[MEDIUM]` Security objectives and KPIs defined (e.g., patch SLA compliance, training completion, mean time to remediate)
- [ ] `[MEDIUM]` Security budget allocated for tools, audits, and pentests

---

## 2. Internal Policy Set

Typical policies expected by SOC 2 / ISO 27001 auditors. Each should have an owner, version, approval date, and annual review.

- [ ] Information Security Policy (top-level)
- [ ] Acceptable Use Policy
- [ ] Access Control Policy
- [ ] Password & Authentication Policy
- [ ] Data Classification & Handling Policy
- [ ] Data Retention & Disposal Policy
- [ ] Encryption & Key Management Policy
- [ ] Secure Development (SDLC) Policy
- [ ] Change Management Policy
- [ ] Vulnerability & Patch Management Policy
- [ ] Logging & Monitoring Policy
- [ ] Incident Response Policy & Plan
- [ ] Business Continuity & Disaster Recovery Plan
- [ ] Backup Policy
- [ ] Vendor / Third-Party Risk Management Policy
- [ ] Asset Management Policy
- [ ] Physical Security Policy (offices, data handling)
- [ ] Remote Work & BYOD Policy
- [ ] HR Security Policy (screening, onboarding, offboarding, discipline)
- [ ] Privacy Policy (internal) & Data Protection Policy
- [ ] Risk Management Policy
- [ ] Code of Conduct
- [ ] AI Acceptable Use Policy (use of AI tools by staff, data allowed in prompts)

- [ ] `[HIGH]` Employees acknowledge key policies at hire and annually (signed/recorded)

---

## 3. Risk Management

### Risk assessment process
- [ ] `[HIGH]` Formal risk assessment at least annually and after major changes
- [ ] `[HIGH]` Methodology documented (**ISO/IEC 27005**, **NIST SP 800-30**, or **FAIR** for quantitative)
- [ ] `[HIGH]` Risks scored by likelihood × impact
- [ ] `[HIGH]` Each risk has a treatment decision: **Mitigate, Transfer, Avoid, or Accept**
- [ ] `[MEDIUM]` Accepted risks signed off by an accountable leader and reviewed periodically

### Risk register (fields)
| Field | Example |
|---|---|
| Risk ID | R-012 |
| Description | Leaked cloud credentials allow data exfiltration |
| Asset(s) | AWS production account |
| Threat / vulnerability | Long-lived keys on developer laptops |
| Likelihood (1–5) | 3 |
| Impact (1–5) | 5 |
| Inherent risk score | 15 |
| Existing controls | MFA on console |
| Treatment | Mitigate — move to SSO + short-lived credentials |
| Owner | DevOps lead |
| Target date | Q2 |
| Residual risk score | 6 |
| Status | In progress |

- [ ] `[MEDIUM]` Key Risk Indicators (KRIs) monitored
- [ ] `[LOW]` Cyber insurance evaluated as risk transfer

---

## 4. Third-Party / Vendor Risk Management (TPRM)

- [ ] `[HIGH]` Inventory of all vendors and SaaS tools that access company or customer data
- [ ] `[HIGH]` Vendors tiered by risk (data access, criticality)
- [ ] `[HIGH]` High-risk vendors reviewed before onboarding: SOC 2 / ISO 27001 reports, security questionnaire (**SIG** or **CAIQ**), DPA signed
- [ ] `[HIGH]` Contracts include security, confidentiality, breach notification, and right-to-audit clauses where appropriate
- [ ] `[MEDIUM]` Annual re-review of critical vendors
- [ ] `[MEDIUM]` Sub-processor list maintained for customers (GDPR/NDPA)
- [ ] `[MEDIUM]` Vendor offboarding: access revoked, data returned/deleted
- [ ] `[MEDIUM]` Concentration risk considered (single cloud region/provider dependencies)

### Responding to customer security questionnaires
- [ ] Maintain a reusable answer library
- [ ] Publish a **trust center** / security page to reduce questionnaire load
- [ ] Have SOC 2 report / ISO certificate / pentest summary ready to share under NDA

---

## 5. People Security (HR Controls)

### Before hire
- [ ] `[MEDIUM]` Background checks proportionate to role and legal in the jurisdiction
- [ ] `[HIGH]` Confidentiality / NDA and IP assignment agreements signed
- [ ] `[MEDIUM]` Security responsibilities included in job descriptions for relevant roles

### Onboarding
- [ ] `[HIGH]` Access provisioned via documented request and approval (least privilege, role-based)
- [ ] `[HIGH]` Security awareness training completed within first 30 days
- [ ] `[HIGH]` Policy acknowledgements recorded
- [ ] `[MEDIUM]` Company devices configured to standard (encryption, MDM, screen lock)

### During employment
- [ ] `[HIGH]` Annual security awareness training (phishing, social engineering, data handling, incident reporting)
- [ ] `[HIGH]` Secure coding training for engineers (OWASP Top 10, framework-specific)
- [ ] `[MEDIUM]` Phishing simulations
- [ ] `[MEDIUM]` Role changes trigger access review (remove old access)
- [ ] `[MEDIUM]` Disciplinary process for policy violations documented

### Offboarding
- [ ] `[CRITICAL]` All access revoked within defined SLA (same day for involuntary exits): SSO, cloud, Git, CI/CD, DBs, SaaS, VPN, shared accounts
- [ ] `[HIGH]` Shared secrets the person knew are rotated
- [ ] `[HIGH]` Company devices returned and wiped
- [ ] `[MEDIUM]` Offboarding checklist completed and stored as evidence
- [ ] `[MEDIUM]` Contractors and freelancers follow the same process

---

## 6. Change Management

- [ ] `[HIGH]` All production changes go through a tracked process (PR, ticket, or change request)
- [ ] `[HIGH]` Changes peer-reviewed and approved by someone other than the author
- [ ] `[HIGH]` Changes tested before production (CI, staging)
- [ ] `[HIGH]` Deployment logs show who deployed what and when
- [ ] `[MEDIUM]` Emergency change process defined, with retroactive review
- [ ] `[MEDIUM]` Rollback plan for significant changes
- [ ] `[MEDIUM]` Infrastructure changes managed via IaC with the same review process
- [ ] `[MEDIUM]` Database migrations reviewed and backed up beforehand

**Evidence auditors ask for:** sample of PRs with approvals, CI results, deployment records, and linked tickets.

---

## 7. Asset Management

- [ ] `[HIGH]` Inventory of hardware (laptops, phones, servers) with owner and status
- [ ] `[HIGH]` Inventory of software, SaaS, cloud accounts, domains, and certificates
- [ ] `[HIGH]` Inventory of data stores (see file 00, data inventory)
- [ ] `[MEDIUM]` Assets classified by criticality
- [ ] `[MEDIUM]` Secure disposal of devices and media (certified wiping/destruction)
- [ ] `[MEDIUM]` Shadow IT discovery (unapproved SaaS tools)

---

## 8. Endpoint & Device Management

- [ ] `[HIGH]` Full-disk encryption on all company laptops (FileVault, BitLocker, LUKS)
- [ ] `[HIGH]` Automatic screen lock and strong device passwords
- [ ] `[HIGH]` OS and browser auto-updates enabled
- [ ] `[MEDIUM]` **MDM** enrollment (Jamf, Kandji, Intune, Google Endpoint Management, Fleet)
- [ ] `[MEDIUM]` Endpoint protection / EDR installed
- [ ] `[MEDIUM]` Remote wipe capability for lost/stolen devices
- [ ] `[MEDIUM]` BYOD rules: separate work profile, minimum security requirements
- [ ] `[LOW]` USB/removable media restrictions where data sensitivity demands

---

## 9. Physical Security

- [ ] `[MEDIUM]` Office access control (badges, visitor logs) if applicable
- [ ] `[MEDIUM]` Clean desk / clear screen practices
- [ ] `[MEDIUM]` Cloud providers' physical controls covered by their SOC 2/ISO reports (documented under shared responsibility)
- [ ] `[LOW]` Secure storage for printed sensitive documents; shredding

---

## 10. Business Continuity Governance

- [ ] `[HIGH]` Business Impact Analysis (BIA) identifies critical processes and dependencies
- [ ] `[HIGH]` RPO/RTO approved by business owners (see file 08)
- [ ] `[MEDIUM]` Key-person risk addressed (documentation, shared access to critical systems via break-glass accounts)
- [ ] `[MEDIUM]` Communication plan for outages (status page, customer comms templates)
- [ ] `[MEDIUM]` BCP/DR tested annually and updated

---

## 11. Compliance Operations

- [ ] `[HIGH]` **Controls register**: every control mapped to frameworks, owner, evidence, and test frequency
- [ ] `[HIGH]` Evidence collected continuously, not just before audits
- [ ] `[MEDIUM]` Internal audit or self-assessment before external audits
- [ ] `[MEDIUM]` Management review meetings documented (ISO 27001 requirement)
- [ ] `[MEDIUM]` Regulatory change monitoring (new laws, framework versions)
- [ ] `[MEDIUM]` Compliance calendar (renewals, audits, filings such as NDPA annual audit returns where applicable)

**Tools:** Vanta, Drata, Secureframe, Sprinto, Eramba (open source), CISO Assistant (open source), spreadsheets for early stage

---

## 12. Security Culture

- [ ] `[MEDIUM]` Easy, blame-free channel to report security concerns and mistakes
- [ ] `[MEDIUM]` Security champions program across teams
- [ ] `[LOW]` Recognition for good security behavior (reported phishing, found bugs)
- [ ] `[LOW]` Regular security updates/newsletters to staff

---

## 13. Minimum Viable Governance (for startups and small teams)

If you're just starting, prioritize these first:

1. [ ] MFA everywhere + password manager for the team
2. [ ] Information Security Policy + Acceptable Use Policy + Incident Response Plan
3. [ ] Onboarding/offboarding checklist with access revocation
4. [ ] Vendor list with DPAs for vendors handling personal data
5. [ ] Simple risk register (top 10 risks)
6. [ ] Annual security awareness training
7. [ ] PR-based change management with required reviews
8. [ ] Quarterly access reviews (even a spreadsheet)
9. [ ] Device encryption + auto-updates on all laptops
10. [ ] Privacy policy, terms, and a security contact (`security.txt`)

---

## References
- ISO/IEC 27001:2022 clauses 4–10 and Annex A (Organizational, People, Physical controls)
- AICPA Trust Services Criteria (CC1–CC9)
- NIST CSF 2.0 — Govern function
- ISO/IEC 27005, NIST SP 800-30, FAIR Institute
- Shared Assessments SIG, CSA CAIQ
