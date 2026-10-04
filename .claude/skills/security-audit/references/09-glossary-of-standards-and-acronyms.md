# 09 — Glossary of Standards & Acronyms

Quick reference for every acronym and standard used across this checklist set, grouped by category.

---

## Data Integrity & Reliability

| Term | Meaning |
|---|---|
| **ACID** | Atomicity, Consistency, Isolation, Durability — guarantees for database transactions |
| **BASE** | Basically Available, Soft state, Eventual consistency — model for distributed/NoSQL systems |
| **CAP** | Consistency, Availability, Partition tolerance — pick C or A during a partition |
| **PACELC** | Extension of CAP: if Partition choose A/C, Else choose Latency/Consistency |
| **2PC** | Two-Phase Commit — distributed transaction protocol |
| **CDC** | Change Data Capture — streaming database changes |
| **SLA** | Service Level Agreement — contractual availability/performance promise |
| **SLO** | Service Level Objective — internal reliability target |
| **SLI** | Service Level Indicator — the measured metric |
| **RPO** | Recovery Point Objective — max acceptable data loss |
| **RTO** | Recovery Time Objective — max acceptable downtime |
| **MTTR** | Mean Time To Recovery/Repair |
| **MTBF** | Mean Time Between Failures |
| **DR** | Disaster Recovery |
| **BCP** | Business Continuity Plan |
| **PITR** | Point-In-Time Recovery |

## Security Principles & Models

| Term | Meaning |
|---|---|
| **CIA** | Confidentiality, Integrity, Availability |
| **AAA** | Authentication, Authorization, Accounting |
| **STRIDE** | Spoofing, Tampering, Repudiation, Information disclosure, Denial of service, Elevation of privilege |
| **PASTA** | Process for Attack Simulation and Threat Analysis |
| **DFD** | Data Flow Diagram |
| **RBAC** | Role-Based Access Control |
| **ABAC** | Attribute-Based Access Control |
| **ReBAC** | Relationship-Based Access Control |
| **PoLP** | Principle of Least Privilege |
| **ZTA** | Zero Trust Architecture |
| **SoD** | Separation of Duties |

## Authentication & Identity

| Term | Meaning |
|---|---|
| **MFA / 2FA** | Multi-Factor / Two-Factor Authentication |
| **TOTP** | Time-based One-Time Password |
| **OTP** | One-Time Password |
| **SSO** | Single Sign-On |
| **OAuth 2.0** | Authorization delegation framework |
| **OIDC** | OpenID Connect — identity layer on OAuth 2.0 |
| **PKCE** | Proof Key for Code Exchange — protects OAuth authorization code flow |
| **JWT** | JSON Web Token |
| **JWS / JWE** | JSON Web Signature / Encryption |
| **JWKS** | JSON Web Key Set |
| **SAML** | Security Assertion Markup Language — XML-based SSO |
| **SCIM** | System for Cross-domain Identity Management — user provisioning |
| **FIDO2** | Passwordless authentication standard (WebAuthn + CTAP) |
| **WebAuthn** | Browser API for public-key authentication |
| **Passkeys** | FIDO credentials synced across devices |
| **IdP** | Identity Provider |
| **IAM** | Identity and Access Management |
| **PAM** | Privileged Access Management |

## Vulnerabilities & Attacks

| Term | Meaning |
|---|---|
| **XSS** | Cross-Site Scripting |
| **CSRF / XSRF** | Cross-Site Request Forgery |
| **SSRF** | Server-Side Request Forgery |
| **SQLi** | SQL Injection |
| **XXE** | XML External Entity |
| **SSTI** | Server-Side Template Injection |
| **RCE** | Remote Code Execution |
| **LFI / RFI** | Local / Remote File Inclusion |
| **IDOR** | Insecure Direct Object Reference |
| **BOLA** | Broken Object Level Authorization |
| **BFLA** | Broken Function Level Authorization |
| **ReDoS** | Regular Expression Denial of Service |
| **DoS / DDoS** | (Distributed) Denial of Service |
| **MITM / AitM** | Man/Adversary-in-the-Middle |
| **ATO** | Account Takeover |
| **CSWSH** | Cross-Site WebSocket Hijacking |
| **CVE** | Common Vulnerabilities and Exposures — unique vulnerability IDs |
| **CWE** | Common Weakness Enumeration — weakness categories |
| **CVSS** | Common Vulnerability Scoring System (0–10 severity) |
| **EPSS** | Exploit Prediction Scoring System — likelihood of exploitation |
| **KEV** | CISA Known Exploited Vulnerabilities catalog |
| **NVD** | National Vulnerability Database |
| **OSV** | Open Source Vulnerabilities database |
| **0-day** | Vulnerability unknown to the vendor / no patch available |

## Threat Intelligence Frameworks

| Term | Meaning |
|---|---|
| **MITRE ATT&CK** | Knowledge base of adversary tactics and techniques |
| **MITRE D3FEND** | Knowledge graph of defensive countermeasures |
| **Cyber Kill Chain** | Lockheed Martin model of attack stages |
| **TTPs** | Tactics, Techniques, and Procedures |
| **IOC** | Indicator of Compromise |

## Testing & Tooling

| Term | Meaning |
|---|---|
| **SAST** | Static Application Security Testing |
| **DAST** | Dynamic Application Security Testing |
| **IAST** | Interactive Application Security Testing |
| **RASP** | Runtime Application Self-Protection |
| **SCA** | Software Composition Analysis |
| **IaC** | Infrastructure as Code |
| **CSPM** | Cloud Security Posture Management |
| **CNAPP** | Cloud-Native Application Protection Platform |
| **CWPP** | Cloud Workload Protection Platform |
| **WAF** | Web Application Firewall |
| **SIEM** | Security Information and Event Management |
| **SOAR** | Security Orchestration, Automation and Response |
| **EDR / XDR** | Endpoint / Extended Detection and Response |
| **IDS / IPS** | Intrusion Detection / Prevention System |
| **HIDS** | Host-based Intrusion Detection System |
| **MDM** | Mobile Device Management |
| **VAPT** | Vulnerability Assessment and Penetration Testing |
| **ASV** | Approved Scanning Vendor (PCI DSS) |
| **PTES** | Penetration Testing Execution Standard |
| **OSSTMM** | Open Source Security Testing Methodology Manual |
| **VDP** | Vulnerability Disclosure Policy |

## Supply Chain

| Term | Meaning |
|---|---|
| **SBOM** | Software Bill of Materials |
| **SPDX** | Software Package Data Exchange — SBOM format (Linux Foundation) |
| **CycloneDX** | SBOM format (OWASP) |
| **SLSA** | Supply-chain Levels for Software Artifacts |
| **Sigstore / cosign** | Tools for signing and verifying software artifacts |
| **OpenSSF** | Open Source Security Foundation |
| **VEX** | Vulnerability Exploitability eXchange — states whether a CVE affects a product |

## Cryptography & Network

| Term | Meaning |
|---|---|
| **TLS** | Transport Layer Security |
| **mTLS** | Mutual TLS (both sides present certificates) |
| **HSTS** | HTTP Strict Transport Security |
| **CSP** | Content Security Policy |
| **CORS** | Cross-Origin Resource Sharing |
| **SRI** | Subresource Integrity |
| **COOP / COEP / CORP** | Cross-Origin Opener / Embedder / Resource Policy |
| **AES-GCM** | Authenticated symmetric encryption mode |
| **RSA / ECDSA / EdDSA** | Asymmetric algorithms |
| **HMAC** | Hash-based Message Authentication Code |
| **KMS** | Key Management Service |
| **HSM** | Hardware Security Module |
| **PQC** | Post-Quantum Cryptography |
| **ML-KEM / ML-DSA** | NIST post-quantum key encapsulation / signature standards |
| **FIPS 140-3** | US standard for validating cryptographic modules |
| **OCSP** | Online Certificate Status Protocol |
| **CA** | Certificate Authority |
| **CT** | Certificate Transparency |
| **VPC** | Virtual Private Cloud |
| **NTP** | Network Time Protocol |

## DNS & Email

| Term | Meaning |
|---|---|
| **SPF** | Sender Policy Framework — authorized mail senders |
| **DKIM** | DomainKeys Identified Mail — email signatures |
| **DMARC** | Domain-based Message Authentication, Reporting & Conformance |
| **BIMI** | Brand Indicators for Message Identification — logo in inbox |
| **MTA-STS** | Mail Transfer Agent Strict Transport Security |
| **TLS-RPT** | TLS reporting for mail delivery |
| **DNSSEC** | DNS Security Extensions |
| **CAA** | Certification Authority Authorization DNS record |
| **security.txt** | Standard file for security contact info (RFC 9116) |

## Mobile

| Term | Meaning |
|---|---|
| **MASVS** | OWASP Mobile Application Security Verification Standard |
| **MASTG** | OWASP Mobile Application Security Testing Guide |
| **ATS** | App Transport Security (iOS) |
| **ATT** | App Tracking Transparency (iOS) |
| **APK / AAB** | Android Package / Android App Bundle |
| **IPA** | iOS App Store Package |

## OWASP Projects

| Term | Meaning |
|---|---|
| **OWASP** | Open Worldwide Application Security Project |
| **ASVS** | Application Security Verification Standard |
| **WSTG** | Web Security Testing Guide |
| **SAMM** | Software Assurance Maturity Model |
| **CRS** | Core Rule Set (for ModSecurity/WAFs) |
| **OWASP Top 10** | Most critical web app risks |
| **API Top 10** | Most critical API risks |
| **Mobile Top 10** | Most critical mobile risks |
| **LLM Top 10** | Most critical risks for LLM applications |
| **Proactive Controls** | Developer-focused security techniques |

## Development & Operations

| Term | Meaning |
|---|---|
| **SDLC / SSDLC** | (Secure) Software Development Life Cycle |
| **SSDF** | NIST Secure Software Development Framework (SP 800-218) |
| **BSIMM** | Building Security In Maturity Model |
| **DevSecOps** | Integrating security into DevOps |
| **GitOps** | Git as the source of truth for deployments |
| **CI/CD** | Continuous Integration / Continuous Delivery |
| **ADR** | Architecture Decision Record |
| **BFF** | Backend-for-Frontend |
| **SRE** | Site Reliability Engineering |
| **ITIL** | IT Infrastructure Library — IT service management practices |
| **COBIT** | Control Objectives for Information and Related Technologies — IT governance |
| **ISO/IEC 25010** | Software product quality model |

## Audit Reports & Certifications

| Term | Meaning |
|---|---|
| **AICPA** | American Institute of Certified Public Accountants |
| **SSAE 18** | Attestation standard underlying SOC reports |
| **SOC 1** | Report on controls relevant to clients' financial reporting |
| **SOC 2** | Report on controls against Trust Services Criteria |
| **SOC 3** | Public summary of SOC 2 |
| **TSC** | Trust Services Criteria (Security, Availability, Processing Integrity, Confidentiality, Privacy) |
| **Type I / Type II** | Design at a point in time / operating effectiveness over a period |
| **ICFR** | Internal Control over Financial Reporting |
| **ITGC** | IT General Controls |
| **ISMS** | Information Security Management System (ISO 27001) |
| **PIMS** | Privacy Information Management System (ISO 27701) |
| **SoA** | Statement of Applicability (ISO 27001) |
| **ISO/IEC 27001** | ISMS certification standard |
| **ISO/IEC 27002** | Security controls guidance |
| **ISO/IEC 27005** | Security risk management |
| **ISO/IEC 27017** | Cloud security controls |
| **ISO/IEC 27018** | PII protection in public cloud |
| **ISO/IEC 27701** | Privacy management extension |
| **ISO/IEC 27034** | Application security |
| **ISO/IEC 29147** | Vulnerability disclosure |
| **ISO/IEC 30111** | Vulnerability handling |
| **ISO 22301** | Business continuity management |
| **ISO/IEC 42001** | AI management system |
| **ISO 9001** | Quality management |
| **CSA STAR** | Cloud Security Alliance Security, Trust, Assurance and Risk registry |
| **CCM** | CSA Cloud Controls Matrix |
| **CAIQ** | Consensus Assessments Initiative Questionnaire |
| **SIG** | Standardized Information Gathering questionnaire |
| **VPAT / ACR** | Voluntary Product Accessibility Template / Accessibility Conformance Report |
| **CIS** | Center for Internet Security (Benchmarks and Controls) |
| **HITRUST CSF** | Certifiable healthcare-focused security framework |
| **Cyber Essentials** | UK baseline security certification |

## US Frameworks & Laws

| Term | Meaning |
|---|---|
| **NIST** | National Institute of Standards and Technology |
| **CSF** | NIST Cybersecurity Framework |
| **SP 800-53** | Security and privacy controls catalog |
| **SP 800-63** | Digital identity guidelines |
| **SP 800-171** | Protecting CUI in non-federal systems |
| **CUI** | Controlled Unclassified Information |
| **AI RMF** | NIST AI Risk Management Framework |
| **FedRAMP** | Federal Risk and Authorization Management Program |
| **StateRAMP** | State-level equivalent of FedRAMP |
| **FISMA** | Federal Information Security Modernization Act |
| **CMMC** | Cybersecurity Maturity Model Certification (DoD) |
| **HIPAA** | Health Insurance Portability and Accountability Act |
| **BAA** | Business Associate Agreement (HIPAA) |
| **PHI** | Protected Health Information |
| **SOX** | Sarbanes-Oxley Act |
| **GLBA** | Gramm-Leach-Bliley Act |
| **FERPA** | Family Educational Rights and Privacy Act |
| **COPPA** | Children's Online Privacy Protection Act |
| **CCPA / CPRA** | California Consumer Privacy Act / California Privacy Rights Act |
| **ADA** | Americans with Disabilities Act |
| **Section 508** | US federal accessibility requirement |
| **CISA** | Cybersecurity and Infrastructure Security Agency |

## Payments & Finance

| Term | Meaning |
|---|---|
| **PCI DSS** | Payment Card Industry Data Security Standard |
| **PCI SSF** | PCI Secure Software Framework (replaced PA-DSS) |
| **PA-DSS** | Payment Application Data Security Standard (retired) |
| **P2PE** | Point-to-Point Encryption |
| **SAQ** | Self-Assessment Questionnaire (PCI DSS) |
| **QSA** | Qualified Security Assessor (PCI DSS) |
| **CHD** | Cardholder Data |
| **PAN** | Primary Account Number (card number) |
| **PSD2** | EU Payment Services Directive 2 |
| **SCA** | Strong Customer Authentication (PSD2) — not to be confused with Software Composition Analysis |
| **SWIFT CSP** | SWIFT Customer Security Programme |
| **DORA** | EU Digital Operational Resilience Act |
| **KYC / AML** | Know Your Customer / Anti-Money Laundering |

## EU, UK & International

| Term | Meaning |
|---|---|
| **GDPR** | EU General Data Protection Regulation |
| **UK GDPR** | UK version of GDPR (with Data Protection Act 2018) |
| **ePrivacy / PECR** | EU cookie & e-marketing rules / UK equivalent |
| **NIS2** | EU Network and Information Security Directive 2 |
| **CRA** | EU Cyber Resilience Act |
| **EU AI Act** | EU regulation on artificial intelligence |
| **eIDAS** | EU regulation on electronic identification and trust services |
| **EAA** | European Accessibility Act |
| **EN 301 549** | EU ICT accessibility standard |
| **AADC** | UK Age Appropriate Design Code |
| **DPO** | Data Protection Officer |
| **DPA** | Data Processing Agreement (also: Data Protection Authority/Act) |
| **DPIA** | Data Protection Impact Assessment |
| **ROPA** | Record of Processing Activities |
| **TIA** | Transfer Impact Assessment |
| **LIA** | Legitimate Interests Assessment |
| **SCCs** | Standard Contractual Clauses (EU) |
| **IDTA** | International Data Transfer Agreement (UK) |
| **PII** | Personally Identifiable Information |
| **DSAR** | Data Subject Access Request |
| **WCAG** | Web Content Accessibility Guidelines |

## Nigeria & Africa

| Term | Meaning |
|---|---|
| **NDPA** | Nigeria Data Protection Act 2023 |
| **NDPC** | Nigeria Data Protection Commission |
| **GAID** | NDPC General Application and Implementation Directive (2025) |
| **NDPR** | Nigeria Data Protection Regulation 2019 (predecessor framework) |
| **CBN** | Central Bank of Nigeria |
| **NITDA** | National Information Technology Development Agency |
| **NCC** | Nigerian Communications Commission |
| **SEC** | Securities and Exchange Commission (Nigeria) |
| **FCCPC / FCCPA** | Federal Competition and Consumer Protection Commission / Act |
| **BVN** | Bank Verification Number |
| **NIN** | National Identification Number |
| **POPIA** | South Africa Protection of Personal Information Act |
| **Malabo Convention** | African Union Convention on Cyber Security and Personal Data Protection |

## Governance & Risk

| Term | Meaning |
|---|---|
| **GRC** | Governance, Risk, and Compliance |
| **FAIR** | Factor Analysis of Information Risk — quantitative risk model |
| **TPRM** | Third-Party Risk Management |
| **AUP** | Acceptable Use Policy |
| **BYOD** | Bring Your Own Device |
| **RACI** | Responsible, Accountable, Consulted, Informed |
| **KRI / KPI** | Key Risk Indicator / Key Performance Indicator |
| **CISO** | Chief Information Security Officer |
| **IR** | Incident Response |
