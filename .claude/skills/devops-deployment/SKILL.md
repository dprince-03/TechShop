---
name: devops-deployment
description: Containerize, configure, and deploy applications — Dockerfiles, docker-compose, Nginx reverse proxy configs, GitHub Actions CI/CD pipelines, Render and AWS deployments (EC2, ECS, RDS, S3, CloudFront), environment variables and secrets, health checks, zero-downtime releases, and rollbacks. Use this skill whenever the user asks to dockerize, deploy, set up CI/CD, write a pipeline or workflow, configure Nginx/SSL/domains, move between hosting providers, or troubleshoot a deployment or build failure.
---

# DevOps & Deployment

## Workflow

1. Identify the target platform and what exists already (Dockerfile, compose, workflows, `render.yaml`, Terraform).
   - **TechShop:** every Docker file (Dockerfiles, compose files) lives in `infra/docker/`; there's no root `.env`; CI (`.github/workflows/ci.yml`) and deployment workflows are separate files; `make check` runs what CI runs; no Redis; hosting isn't decided yet, so ask before choosing a platform.
2. Read the relevant reference:
   - Docker & compose → `references/docker.md`
   - Nginx & TLS → `references/nginx.md`
   - GitHub Actions CI/CD → `references/github-actions.md`
   - Render → `references/render.md`
   - AWS → `references/aws.md`
3. Start from templates in `assets/` and adapt to the project.
4. Always deliver with: required env vars list, deploy steps, health check, and rollback steps.

## Principles

- **Build once, deploy many**: the same image/artifact moves from staging to production; config differs only via env vars.
- **Immutable releases**: no SSH-and-edit in production; changes go through Git and CI.
- **Secrets never in images, repos, or logs**: use platform secret stores (Render env groups, AWS Secrets Manager/SSM, GitHub environment secrets) and OIDC for cloud auth from CI.
- **Health checks**: `/healthz` (liveness: process up) and `/readyz` (readiness: database and other required dependencies reachable). Deploys wait on readiness.
- **Zero downtime**: rolling or blue/green; run backward-compatible DB migrations *before* the new code (see `database-migrations` skill).
- **Graceful shutdown**: handle SIGTERM, stop accepting new requests, finish in-flight work within the platform's grace period.
- **Observability on day one**: structured logs, error tracking (Sentry), uptime monitoring, basic metrics.
- **Least privilege**: CI tokens, cloud roles, and DB users scoped to what they need.

## Standard pipeline stages

1. Install & cache dependencies
2. Lint + type-check
3. Unit tests
4. Security: secret scan, SAST, dependency audit
5. Build image (multi-stage), scan image (Trivy)
6. Integration tests (service containers)
7. Push image (tagged with commit SHA)
8. Deploy to staging → smoke tests
9. Manual approval → deploy to production
10. Post-deploy smoke test + automatic rollback or alert on failure

## Deliverable checklist

- [ ] Dockerfile multi-stage, non-root, pinned base image, `.dockerignore`
- [ ] Env vars documented in `.env.example` (no real values)
- [ ] Health endpoints and platform health check configured
- [ ] Migrations step defined and ordered before app rollout
- [ ] CI runs tests and security scans before deploy
- [ ] HTTPS enforced, security headers set
- [ ] Rollback procedure written
- [ ] Logs/monitoring/alerts connected
