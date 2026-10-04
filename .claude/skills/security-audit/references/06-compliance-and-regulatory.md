# 06 — Compliance & Regulatory

A map of the audit reports, certifications, frameworks, and laws that commonly apply to software. **You rarely need all of them.** Use the scoping table first, then the detailed sections.

> Not legal advice. Regulations change and applicability depends on your jurisdiction, industry, users, and contracts. Verify current versions and confirm obligations with a qualified compliance professional or lawyer.

---

## 0. Scoping: Which Ones Apply to You?

| If you... | Likely relevant |
|---|---|
| Sell B2B SaaS to companies | SOC 2, ISO 27001, vendor questionnaires (SIG/CAIQ) |
| Process data that affects clients' financial statements (payroll, billing, accounting) | SOC 1 |
| Handle personal data of people in Nigeria | NDPA 2023, GAID 2025 |
| Handle personal data of EU/UK residents | GDPR / UK GDPR, ePrivacy |
| Have California users (and meet thresholds) | CCPA/CPRA |
| Accept, store, or transmit card payments | PCI DSS |
| Operate fintech/payments in Nigeria | CBN frameworks and guidelines, NDPA |
| Handle US health data (as covered entity or business associate) | HIPAA, optionally HITRUST |
| Sell to US federal agencies | FedRAMP, NIST SP 800-53, FISMA |
| Defense contractors (US) | CMMC, NIST SP 800-171 |
| Publicly traded companies (US) | SOX |
| EU financial entities and their ICT providers | DORA |
| EU essential/important entities | NIS2 |
| Build or deploy AI systems in the EU | EU AI Act; ISO/IEC 42001 optional |
| Sell products with digital elements in the EU | Cyber Resilience Act |
| Serve children | COPPA (US), UK Age Appropriate Design Code, GDPR Art. 8 |
| Education records (US) | FERPA |
| Public-facing digital services | WCAG 2.2, ADA, Section 508, European Accessibility Act |

---

## 1. SOC Reports (AICPA)

SOC reports are **attestation reports** by an independent CPA firm, not certifications.

| Report | Purpose | Audience |
|---|---|---|
| **SOC 1** | Controls relevant to clients' **financial reporting** (ICFR); based on SSAE 18 | Clients' auditors, finance teams |
| **SOC 2** | Controls against the **Trust Services Criteria** | Customers, prospects (under NDA) |
| **SOC 3** | General-use public summary of SOC 2 | Public / marketing |
| SOC for Cybersecurity | Entity-wide cybersecurity risk management program | Boards, investors |
| SOC for Supply Chain | Controls in production/distribution of goods | Customers, partners |

**Type I vs Type II**
- **Type I**: controls are suitably *designed* at a point in time
- **Type II**: controls *operated effectively* over a period (typically 3–12 months); what most customers want

### SOC 2 Trust Services Criteria
- [ ] **Security** (Common Criteria, mandatory): access control, change management, risk assessment, monitoring, incident response
- [ ] **Availability**: uptime, capacity, DR, backups
- [ ] **Processing Integrity**: complete, accurate, timely, authorized processing
- [ ] **Confidentiality**: protection of confidential business information
- [ ] **Privacy**: personal information lifecycle (notice, consent, retention, disposal)

### SOC 2 readiness checklist
- [ ] Choose TSC in scope and define system boundary
- [ ] Written policies (see file 10)
- [ ] Risk assessment completed annually
- [ ] Access reviews quarterly; onboarding/offboarding evidence
- [ ] Change management evidence (PRs, approvals, deploy logs)
- [ ] Vendor risk management records
- [ ] Security awareness training records
- [ ] Incident response plan + test evidence
- [ ] Backup/restore test evidence
- [ ] Vulnerability scanning and pentest reports
- [ ] Engage auditor; consider readiness assessment first

**Automation platforms:** Vanta, Drata, Secureframe, Sprinto, Thoropass

---

## 2. ISO Standards

These are **certifications** issued by accredited certification bodies (except guidance-only standards noted).

| Standard | Covers | Certifiable? |
|---|---|---|
| **ISO/IEC 27001:2022** | Information Security Management System (ISMS) | Yes |
| **ISO/IEC 27002:2022** | Control guidance (93 controls in 4 themes) | No (guidance) |
| ISO/IEC 27005 | Information security risk management | No |
| **ISO/IEC 27017** | Cloud-specific security controls | Via 27001 extension |
| **ISO/IEC 27018** | Protection of PII in public clouds | Via 27001 extension |
| **ISO/IEC 27701** | Privacy Information Management System (PIMS) | Yes |
| ISO/IEC 27034 | Application security | No (guidance) |
| ISO/IEC 29147 / 30111 | Vulnerability disclosure / handling | No (guidance) |
| **ISO 22301** | Business continuity management | Yes |
| **ISO/IEC 42001** | AI management systems | Yes |
| **ISO 9001** | Quality management | Yes |
| ISO/IEC 20000-1 | IT service management | Yes |
| ISO/IEC 25010 | Software product quality model | No (model) |

### ISO 27001 readiness checklist
- [ ] Define ISMS scope and context (clauses 4–10)
- [ ] Information security policy and objectives
- [ ] Risk assessment and risk treatment plan
- [ ] **Statement of Applicability (SoA)** covering Annex A controls
- [ ] Implement controls: Organizational (37), People (8), Physical (14), Technological (34)
- [ ] Internal audit
- [ ] Management review
- [ ] Stage 1 (documentation) and Stage 2 (implementation) certification audits
- [ ] Annual surveillance audits; recertification every 3 years

---

## 3. NIST Frameworks (US, widely used globally)

| Framework | Use |
|---|---|
| **NIST CSF 2.0** | High-level program: Govern, Identify, Protect, Detect, Respond, Recover |
| **SP 800-53 Rev. 5** | Comprehensive control catalog (basis for FedRAMP, FISMA) |
| **SP 800-171** | Protecting Controlled Unclassified Information (CUI) in non-federal systems |
| **SP 800-63** | Digital identity (authentication assurance levels) |
| SP 800-61 | Incident handling |
| SP 800-30 | Risk assessment |
| **SP 800-218 (SSDF)** | Secure Software Development Framework |
| SP 800-207 | Zero Trust Architecture |
| **AI RMF** | AI risk management |
| FIPS 140-3 | Cryptographic module validation |

- [ ] Current state mapped to CSF 2.0 functions; target profile defined
- [ ] Gaps prioritized in a roadmap

---

## 4. US Government & Sector Regulations

- [ ] **FedRAMP**: cloud services for US federal agencies (Low/Moderate/High baselines on SP 800-53)
- [ ] **StateRAMP**: state/local government equivalent
- [ ] **FISMA**: federal agency information security
- [ ] **CMMC 2.0**: Department of Defense contractors (Levels 1–3, based on SP 800-171/172)
- [ ] **HIPAA**: Privacy Rule, Security Rule (administrative, physical, technical safeguards), Breach Notification Rule; sign **BAAs** with vendors
- [ ] **HITRUST CSF**: certifiable framework often used to demonstrate HIPAA alignment
- [ ] **SOX**: IT general controls (ITGCs) for financial reporting systems at public companies
- [ ] **GLBA**: financial institutions' customer data (Safeguards Rule)
- [ ] **FERPA**: student education records
- [ ] **COPPA**: online services directed at children under 13
- [ ] **CCPA/CPRA**: California consumer privacy (right to know, delete, correct, opt-out of sale/sharing)
- [ ] Other US state privacy laws (Virginia, Colorado, Connecticut, Texas, etc.) if thresholds met

---

## 5. Payments & Finance

- [ ] **PCI DSS v4.0.1**: 12 requirements for any entity storing, processing, or transmitting cardholder data
  - Determine merchant/service provider level and SAQ type (A, A-EP, D, etc.)
  - Minimize scope with hosted payment pages / tokenization
  - Requirements effective since 31 March 2025 include script management on payment pages (6.4.3, 11.6.1), stronger MFA, and targeted risk analyses
- [ ] **PCI SSF** (Secure Software Standard; replaced PA-DSS) for payment software vendors
- [ ] **PCI PIN** / **PCI P2PE** if handling PINs or point-to-point encryption
- [ ] **PSD2 / Strong Customer Authentication** (EU/UK payments); PSD3/PSR developments to monitor
- [ ] **SWIFT Customer Security Programme (CSP)** for SWIFT users
- [ ] **DORA**: EU financial entities' digital operational resilience, ICT risk, incident reporting, testing, third-party risk (applies from 17 January 2025)

### PCI DSS 12 Requirements
1. Install and maintain network security controls
2. Apply secure configurations
3. Protect stored account data
4. Protect cardholder data with strong cryptography during transmission
5. Protect systems against malware
6. Develop and maintain secure systems and software
7. Restrict access by business need to know
8. Identify users and authenticate access
9. Restrict physical access
10. Log and monitor all access
11. Test security regularly
12. Support security with organizational policies and programs

---

## 6. EU & UK

- [ ] **GDPR / UK GDPR**: lawful basis, transparency, data subject rights, DPO (if required), DPIAs, ROPA, breach notification within 72 hours, international transfer safeguards, privacy by design
- [ ] **ePrivacy Directive / PECR (UK)**: cookies and electronic marketing consent
- [ ] **NIS2 Directive**: cybersecurity risk management and incident reporting for essential/important entities
- [ ] **DORA**: see Payments & Finance
- [ ] **EU AI Act**: risk-based obligations for AI systems (prohibited practices, high-risk requirements, transparency, general-purpose AI); phased application 2025–2027
- [ ] **Cyber Resilience Act (CRA)**: security requirements for products with digital elements; vulnerability reporting obligations begin before full application in late 2027
- [ ] **eIDAS / eIDAS 2.0**: electronic identification and trust services (e-signatures, EUDI Wallet)
- [ ] **European Accessibility Act**: accessibility of certain products and services (applies from 28 June 2025)
- [ ] **UK Cyber Essentials / Cyber Essentials Plus**: baseline certification (firewalls, secure config, access control, malware protection, patching); often required for UK public-sector contracts
- [ ] **UK Age Appropriate Design Code (Children's Code)**

---

## 7. Nigeria & Africa

### Nigeria
- [ ] **Nigeria Data Protection Act (NDPA) 2023**
  - Lawful basis for processing; consent requirements
  - Data subject rights (access, rectification, erasure, portability, objection)
  - Data Protection Officer for data controllers/processors of major importance
  - Registration with the **Nigeria Data Protection Commission (NDPC)** if classified as of major importance
  - Data Protection Impact Assessments for high-risk processing
  - Breach notification to NDPC within 72 hours where required; notify affected data subjects where high risk
  - Cross-border transfer safeguards
  - Annual compliance audit filing where required
- [ ] **NDPC General Application and Implementation Directive (GAID) 2025**: operational detail for implementing the NDPA
- [ ] **Cybercrimes (Prohibition, Prevention, etc.) Act** (2015, amended 2024): incident reporting, critical infrastructure duties
- [ ] **CBN** frameworks for regulated financial institutions and fintechs, including the Risk-Based Cybersecurity Framework and guidelines on operational resilience, open banking, and payment service providers
- [ ] **NITDA** guidelines (e.g., IT standards, cloud/hosting, local content where applicable)
- [ ] **NCC** regulations for telecom-related services (e.g., consumer code, data/SIM rules)
- [ ] **SEC Nigeria** rules if operating in digital assets/investment platforms
- [ ] **FCCPA** (Federal Competition and Consumer Protection Act): consumer protection in digital transactions

### Other African data protection laws (if serving those markets)
- [ ] South Africa — **POPIA**
- [ ] Kenya — **Data Protection Act 2019**
- [ ] Ghana — **Data Protection Act 2012**
- [ ] Rwanda, Egypt, Uganda, Morocco and others as applicable
- [ ] **AU Malabo Convention** (cyber security and personal data protection) awareness

---

## 8. Cloud & Vendor Assurance

- [ ] **CSA STAR** (Cloud Security Alliance Security, Trust, Assurance and Risk) registry (Level 1 self-assessment, Level 2 third-party)
- [ ] **CSA Cloud Controls Matrix (CCM)** mapping
- [ ] **CAIQ** (Consensus Assessments Initiative Questionnaire) completed for customers
- [ ] **SIG** (Shared Assessments Standardized Information Gathering) questionnaire readiness
- [ ] **CIS Controls v8** implementation group (IG1/IG2/IG3) chosen
- [ ] **CIS Benchmarks** applied (see file 05)
- [ ] **Shared responsibility model** documented for each cloud/PaaS provider
- [ ] Sub-processor list maintained and published (for GDPR/NDPA customers)
- [ ] Provider compliance reports (AWS Artifact, etc.) collected

---

## 9. Accessibility

- [ ] **WCAG 2.2** Level AA (the common legal/contractual target)
- [ ] **ADA** (US) — courts often reference WCAG for websites
- [ ] **Section 508** (US federal procurement)
- [ ] **EN 301 549** (EU ICT accessibility standard)
- [ ] **European Accessibility Act**
- [ ] Accessibility statement published; **VPAT/ACR** prepared for enterprise/government sales

---

## 10. Required Policy Documents (public-facing)

- [ ] Privacy Policy / Privacy Notice
- [ ] Terms of Service / Terms of Use
- [ ] Cookie Policy (+ consent mechanism)
- [ ] Acceptable Use Policy
- [ ] Data Processing Agreement (DPA) template for B2B customers
- [ ] Sub-processor list
- [ ] Service Level Agreement (SLA)
- [ ] Security page / Trust center
- [ ] Vulnerability Disclosure Policy and `security.txt`
- [ ] Accessibility Statement
- [ ] Refund / payment terms (for commerce)

Internal policies are covered in file 10.

---

## 11. Data Subject Rights Workflows

- [ ] **Right of access** (export of personal data)
- [ ] **Right to rectification**
- [ ] **Right to erasure** ("right to be forgotten")
- [ ] **Right to restrict processing**
- [ ] **Right to data portability** (machine-readable format)
- [ ] **Right to object** (including direct marketing)
- [ ] Rights related to **automated decision-making / profiling**
- [ ] **Opt-out of sale/sharing** (CCPA/CPRA)
- [ ] Identity verification before fulfilling requests
- [ ] Response deadlines tracked (e.g., GDPR one month, extendable; check NDPA/GAID and CCPA timelines)
- [ ] Request log maintained as evidence

---

## 12. Privacy Records & Assessments

- [ ] **ROPA** — Record of Processing Activities
- [ ] **DPIA** — Data Protection Impact Assessment for high-risk processing
- [ ] **TIA** — Transfer Impact Assessment for international transfers
- [ ] **LIA** — Legitimate Interest Assessment where relying on legitimate interests
- [ ] International transfer mechanisms: **SCCs** (EU Standard Contractual Clauses), UK IDTA/Addendum, adequacy decisions, NDPA transfer bases
- [ ] Data retention schedule per data category
- [ ] Breach register (including breaches not reported to regulators)
- [ ] Consent records

---

## 13. Privacy Engineering Controls

- [ ] **Privacy by design and by default**
- [ ] **Data minimization** and purpose limitation
- [ ] **Pseudonymization** (reversible with separately held key) where possible
- [ ] **Anonymization** (irreversible) for analytics and research datasets
- [ ] **Consent management platform** (OneTrust, Cookiebot, Osano, open-source alternatives)
- [ ] **Data residency / localization** requirements identified per customer and jurisdiction
- [ ] Automated retention enforcement (TTL jobs, lifecycle rules)

---

## Compliance Evidence Tips

- Automate evidence collection (CI logs, access reviews, config snapshots) instead of screenshots before audits
- Map controls once to many frameworks (SOC 2 ↔ ISO 27001 ↔ NIST CSF overlap heavily) using a common control framework
- Keep a single **controls register**: control ID, description, owner, frameworks mapped, evidence location, test frequency
