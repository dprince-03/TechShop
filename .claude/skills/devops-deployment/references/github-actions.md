# GitHub Actions

Template: `assets/ci.yml`.

- Pin third-party actions to commit SHAs; set top-level `permissions: contents: read` and elevate per job.
- Use `concurrency` to cancel superseded runs.
- Cache: `actions/setup-node` with `cache: npm`, `actions/setup-go` with built-in cache.
- Service containers for Postgres/Redis in integration jobs.
- Environments (`staging`, `production`) with protection rules and required reviewers; secrets scoped per environment.
- AWS auth via OIDC: `aws-actions/configure-aws-credentials` with `role-to-assume` (no access keys).
- Never interpolate `${{ github.event.* }}` user-controlled values directly into `run:`; pass via `env:`.
- Render deploy: trigger deploy hook URL (stored as secret) after tests pass, or use Render's auto-deploy on branch with "wait for CI" enabled.
