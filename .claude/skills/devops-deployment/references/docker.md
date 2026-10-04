# Docker

## Principles
- Multi-stage builds; final stage minimal (distroless/alpine/scratch).
- Non-root `USER`; pin base images (tag + ideally digest).
- Copy dependency manifests first for layer caching.
- `.dockerignore`: `.git`, `node_modules`, `.env*`, `dist`, `coverage`, `*.log`.
- No secrets in `ARG`/`ENV`; use BuildKit `--secret` for private registries.
- Add `HEALTHCHECK` (or rely on orchestrator probes).

## Templates
See `assets/Dockerfile.go`, `assets/Dockerfile.node`, `assets/Dockerfile.nextjs`, `assets/docker-compose.dev.yml`.

## Next.js
- Set `output: 'standalone'` in `next.config.js` for small images.

## Compose (local dev)
- App + Postgres + Redis (+ Mongo) with named volumes and healthchecks; `depends_on: condition: service_healthy`.
- Bind DB ports to `127.0.0.1` only.
