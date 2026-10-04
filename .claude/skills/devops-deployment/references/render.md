# Render

- Services: Web Service (Docker or native runtime), Background Worker, Cron Job, Static Site, Managed PostgreSQL, Key Value (Redis-compatible).
- Prefer **Blueprints** (`render.yaml`) so infrastructure lives in Git.
- Env vars: use Environment Groups shared across services; mark secrets; link DB connection strings via `fromDatabase`.
- Health check path configured on the web service for zero-downtime deploys.
- Pre-deploy command for migrations (e.g., `npm run migrate` / `./migrate up`) so they run before the new version takes traffic.
- Use private network (internal hostnames) between services; restrict DB external access by IP allow-list.
- Preview environments for PRs: never point at production data.
- Check current Render docs for plan limits, region availability, and feature names — they change.

Minimal blueprint:
```yaml
services:
  - type: web
    name: api
    runtime: docker
    healthCheckPath: /readyz
    preDeployCommand: ./migrate up
    envVars:
      - key: DATABASE_URL
        fromDatabase: { name: main-db, property: connectionString }
      - fromGroup: api-secrets
databases:
  - name: main-db
```
