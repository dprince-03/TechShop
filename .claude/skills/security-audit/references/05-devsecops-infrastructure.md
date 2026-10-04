# 05 — DevSecOps & Infrastructure

CI/CD, source control, containers, cloud (AWS focus, applicable to others), networking, reverse proxies, DNS, email, and observability.

---

## 1. Source Control (Git/GitHub)

- [ ] `[CRITICAL]` MFA enforced for all org members
- [ ] `[HIGH]` Branch protection on `main`/`production`: required reviews, status checks, no force-push, no direct push
- [ ] `[HIGH]` CODEOWNERS for sensitive paths (auth, infra, CI config)
- [ ] `[HIGH]` Secret scanning + push protection enabled
- [ ] `[MEDIUM]` Signed commits / tags (GPG, SSH, or Sigstore gitsign)
- [ ] `[MEDIUM]` Least-privilege repo access; outside collaborators reviewed quarterly
- [ ] `[MEDIUM]` Deploy keys and personal access tokens scoped and expiring (prefer fine-grained tokens / GitHub Apps)

---

## 2. CI/CD Pipeline Security

- [ ] `[CRITICAL]` CI secrets scoped per environment; production secrets only on protected branches/environments
- [ ] `[CRITICAL]` Pull requests from forks cannot access secrets (`pull_request_target` used carefully)
- [ ] `[HIGH]` Third-party GitHub Actions pinned to full commit SHA, not tags
- [ ] `[HIGH]` OIDC federation to cloud (no long-lived AWS keys in CI)
- [ ] `[HIGH]` `GITHUB_TOKEN` permissions set to minimum (`permissions:` block)
- [ ] `[HIGH]` Production deploys require approval (environment protection rules)
- [ ] `[HIGH]` Script injection: untrusted inputs (`github.event.*.title`, branch names) never interpolated into `run:` directly
- [ ] `[MEDIUM]` Self-hosted runners isolated, ephemeral, not used for public repos
- [ ] `[MEDIUM]` Build provenance generated (**SLSA** level 2+; GitHub artifact attestations)
- [ ] `[MEDIUM]` Artifacts and images signed (**Sigstore / cosign**) and verified at deploy

### Security gates in pipeline
| Stage | Check | Tools |
|---|---|---|
| Pre-commit | Secrets, lint | gitleaks, pre-commit, Husky |
| PR | SAST | Semgrep, CodeQL, SonarQube, gosec |
| PR | SCA | Dependabot, Snyk, Trivy, govulncheck, npm audit |
| PR | IaC scan | Checkov, tfsec/Trivy, KICS |
| Build | Container scan | Trivy, Grype, Docker Scout |
| Build | SBOM | Syft, CycloneDX tools |
| Staging | DAST | OWASP ZAP, Nuclei, Burp Enterprise |
| Deploy | Signature verify | cosign, Kyverno, admission controllers |

- [ ] `[HIGH]` Defined policy for which findings block merges (e.g., critical/high with fix available)

---

## 3. GitOps

- [ ] `[HIGH]` Git is the single source of truth; manual changes in clusters/cloud detected (drift detection)
- [ ] `[HIGH]` GitOps controller (Argo CD / Flux) has least-privilege access
- [ ] `[HIGH]` Secrets in Git encrypted (SOPS, Sealed Secrets) or referenced from external vault (External Secrets Operator)
- [ ] `[MEDIUM]` Argo CD UI/API not public; SSO + RBAC enabled

---

## 4. Container Security (Docker)

- [ ] `[CRITICAL]` Containers run as non-root user (`USER` directive)
- [ ] `[HIGH]` Minimal base images (distroless, Alpine, `scratch` for Go binaries, Chainguard)
- [ ] `[HIGH]` Multi-stage builds; no build tools, source, or `.env` in final image
- [ ] `[HIGH]` `.dockerignore` excludes `.git`, `.env`, `node_modules`, secrets
- [ ] `[HIGH]` Base images pinned by digest and rebuilt regularly
- [ ] `[HIGH]` No secrets in `ENV`, `ARG`, or image layers (use BuildKit secrets)
- [ ] `[MEDIUM]` Read-only root filesystem; drop all Linux capabilities, add only needed ones
- [ ] `[MEDIUM]` No `--privileged`; Docker socket never mounted into containers
- [ ] `[MEDIUM]` Resource limits (CPU/memory) set
- [ ] `[MEDIUM]` `HEALTHCHECK` defined
- [ ] `[LOW]` Images scanned in registry continuously

**Benchmark:** CIS Docker Benchmark (test with `docker-bench-security`)

---

## 5. Kubernetes (if used)

- [ ] `[CRITICAL]` API server not publicly exposed (or restricted by IP/private endpoint)
- [ ] `[HIGH]` RBAC least privilege; no `cluster-admin` for apps or CI
- [ ] `[HIGH]` Pod Security Standards (`restricted` profile) enforced
- [ ] `[HIGH]` Network Policies: default deny between namespaces
- [ ] `[HIGH]` Secrets encrypted at rest (KMS); consider external secret stores
- [ ] `[MEDIUM]` Admission control (Kyverno / OPA Gatekeeper) for image signatures, registries, privileged pods
- [ ] `[MEDIUM]` Runtime threat detection (Falco)
- [ ] `[MEDIUM]` Service account tokens not auto-mounted unless needed

**Benchmark:** CIS Kubernetes Benchmark (kube-bench); scan with kubescape

---

## 6. Cloud Security Posture (AWS)

### Identity (IAM)
- [ ] `[CRITICAL]` Root account: MFA on, no access keys, used only for break-glass
- [ ] `[CRITICAL]` No long-lived IAM user keys where roles/SSO can be used (IAM Identity Center)
- [ ] `[HIGH]` Least-privilege policies; no `*:*`; use IAM Access Analyzer
- [ ] `[HIGH]` Separate AWS accounts for prod/staging/dev (AWS Organizations + SCPs)
- [ ] `[MEDIUM]` Unused roles, users, and keys removed (credential report reviewed)

### Storage
- [ ] `[CRITICAL]` S3 Block Public Access enabled at account level (exceptions documented)
- [ ] `[HIGH]` S3 encryption, versioning, and access logging enabled for important buckets
- [ ] `[HIGH]` RDS/EBS/snapshots encrypted and not public

### Network
- [ ] `[CRITICAL]` No security groups open to `0.0.0.0/0` on SSH (22), RDP (3389), DB ports (5432, 3306, 27017, 6379)
- [ ] `[HIGH]` Databases and caches in private subnets
- [ ] `[HIGH]` SSH replaced by SSM Session Manager or bastion with MFA
- [ ] `[MEDIUM]` VPC Flow Logs enabled
- [ ] `[MEDIUM]` VPC endpoints for AWS services where sensible

### Detection
- [ ] `[HIGH]` CloudTrail enabled in all regions, logs to protected bucket
- [ ] `[HIGH]` GuardDuty enabled
- [ ] `[MEDIUM]` AWS Config + Security Hub with CIS AWS Foundations Benchmark
- [ ] `[MEDIUM]` Billing alarms (detect cryptomining / abuse)
- [ ] `[MEDIUM]` IMDSv2 required on EC2

**Tools:** Prowler, ScoutSuite, Steampipe, AWS Security Hub, Checkov

### PaaS (Render, Vercel, Railway, Heroku, etc.)
- [ ] `[HIGH]` Team accounts with MFA; least-privilege roles
- [ ] `[HIGH]` Environment variables scoped per environment
- [ ] `[MEDIUM]` Preview deployments protected and not connected to prod data
- [ ] `[MEDIUM]` Managed DBs restricted by IP / private networking

---

## 7. Infrastructure as Code

- [ ] `[HIGH]` All infra defined in code (Terraform, Pulumi, CDK, CloudFormation)
- [ ] `[HIGH]` IaC scanned in CI (Checkov, Trivy, KICS)
- [ ] `[HIGH]` Terraform state stored remotely, encrypted, with locking and restricted access (state contains secrets)
- [ ] `[MEDIUM]` Plans reviewed in PRs before apply

---

## 8. Nginx / Reverse Proxy Hardening

- [ ] `[HIGH]` TLS 1.2/1.3 only; strong cipher suites (Mozilla SSL Config Generator "intermediate" profile)
- [ ] `[HIGH]` HTTP → HTTPS redirect; HSTS header
- [ ] `[HIGH]` `server_tokens off;`
- [ ] `[HIGH]` Request size limits (`client_max_body_size`) and timeouts
- [ ] `[HIGH]` Rate limiting (`limit_req_zone`, `limit_conn_zone`) on auth and expensive routes
- [ ] `[HIGH]` Correct forwarding headers (`X-Forwarded-For`, `X-Real-IP`) and client-supplied versions overwritten
- [ ] `[MEDIUM]` Security headers added (see file 02)
- [ ] `[MEDIUM]` Directory listing off (`autoindex off`); hidden files denied (`location ~ /\. { deny all; }`)
- [ ] `[MEDIUM]` OCSP stapling enabled
- [ ] `[MEDIUM]` Default server block returns 444 for unknown hosts

**Test:** SSL Labs (target A or A+), testssl.sh

---

## 9. Network Security, WAF & DDoS

- [ ] `[HIGH]` WAF in front of public apps (Cloudflare, AWS WAF, ModSecurity + OWASP CRS)
- [ ] `[HIGH]` DDoS protection (Cloudflare, AWS Shield)
- [ ] `[HIGH]` Origin servers accept traffic only from CDN/WAF IPs
- [ ] `[MEDIUM]` Network segmentation: public, application, and data tiers
- [ ] `[MEDIUM]` Egress filtering to limit data exfiltration and SSRF impact
- [ ] `[MEDIUM]` Bot management on high-value endpoints

---

## 10. TLS Certificate Management

- [ ] `[HIGH]` Automated issuance and renewal (Let's Encrypt / ACM / Cloudflare)
- [ ] `[HIGH]` Expiry monitoring and alerts
- [ ] `[MEDIUM]` Private keys protected; not committed or shared
- [ ] `[MEDIUM]` Certificate Transparency monitoring for your domains (crt.sh, Cert Spotter)

---

## 11. DNS & Domain Security

- [ ] `[HIGH]` Registrar account: MFA, registrar lock, transfer lock
- [ ] `[HIGH]` **Subdomain takeover** checks: no dangling CNAMEs to deleted S3/Heroku/Vercel/GitHub Pages resources
- [ ] `[MEDIUM]` **CAA records** restrict which CAs can issue certificates
- [ ] `[MEDIUM]` **DNSSEC** enabled where supported
- [ ] `[MEDIUM]` DNS provider access restricted with MFA
- [ ] `[LOW]` Lookalike/typosquat domain monitoring

**Tools:** subjack, nuclei takeover templates, dnsrecon, Amass

---

## 12. Email Security

- [ ] `[HIGH]` **SPF** record lists only authorized senders, ends with `-all` or `~all`
- [ ] `[HIGH]` **DKIM** signing for all sending services (Google Workspace, SES, SendGrid, Mailgun, etc.)
- [ ] `[HIGH]` **DMARC** published, moving from `p=none` → `quarantine` → `reject`; reports monitored
- [ ] `[MEDIUM]` Non-sending domains have `v=spf1 -all` and `p=reject` DMARC
- [ ] `[LOW]` **BIMI** for brand logo (requires DMARC enforcement)
- [ ] `[LOW]` **MTA-STS** and **TLS-RPT** for inbound mail TLS

**Tools:** MXToolbox, dmarcian, Google Postmaster Tools

---

## 13. Server Hardening & Patching

- [ ] `[CRITICAL]` OS and packages patched; critical patches within defined SLA (e.g., 7 days for critical)
- [ ] `[HIGH]` SSH: key-only auth, root login disabled, non-default user, fail2ban
- [ ] `[HIGH]` Host firewall (ufw/iptables/nftables) default deny inbound
- [ ] `[HIGH]` Unused services and ports disabled
- [ ] `[MEDIUM]` Automatic security updates (unattended-upgrades) where safe
- [ ] `[MEDIUM]` CIS Benchmark for the OS (Ubuntu, Amazon Linux)
- [ ] `[MEDIUM]` Endpoint detection / host intrusion detection (Wazuh, OSSEC, CrowdStrike)

**Tools:** Lynis, OpenSCAP, CIS-CAT

---

## 14. Observability Stack Security

- [ ] `[HIGH]` Dashboards (Grafana, Kibana, Prometheus) not publicly accessible; SSO enforced
- [ ] `[HIGH]` Log access restricted (logs can contain sensitive data)
- [ ] `[MEDIUM]` Retention periods match policy and regulation
- [ ] `[MEDIUM]` Error tracking (Sentry) scrubs PII and secrets
- [ ] `[MEDIUM]` Uptime and synthetic monitoring for critical paths

---

## References
- CIS Benchmarks (AWS, Docker, Kubernetes, Linux)
- SLSA framework, NIST SSDF (SP 800-218)
- OWASP Top 10 CI/CD Security Risks
- Mozilla Server Side TLS guidelines
- M3AAWG / DMARC.org email authentication guidance
