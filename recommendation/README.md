# TechShop recommendation system

Python service that trains and serves product recommendations. The full design is in
[`../docs/recommendations.md`](../docs/recommendations.md).

## Rules
- **Owns its own database, never Go's.** Python stores its data in a private Postgres
  (`recs-postgres`, Docker network only). It reads pseudonymised exports from the Go API's internal
  endpoints, and a Go job pulls the finished results through this service's API into Go's tables.
  `tests/test_no_db_access.py` fails if it is ever pointed at Go's database.
- Every model is **gated**: it only contributes once its data threshold is met and it beats
  the popularity baseline offline (`configs/gates.yaml`).
- The live service (`serving/`) is optional and off by default (feature flag `recs.live_service`
  in the API, 80 ms timeout, Go falls back to its own rules).

## Layout
- `configs/`: shelves, gates and model hyper-parameters
- `src/techshop_recs/`: `io` (API client + contracts), `data`, `features`, `models` (G1–G9, R1, R2),
  `gating`, `evaluation`, `pipelines`, `serving`
- `tests/`: synthetic fixtures and contract tests
- The Dockerfile lives in `infra/docker/recommendation.Dockerfile`, per the repo rule that all
  Docker files go under `infra/docker/`

## Status
Scaffold only: structure and stubs. Dependencies are declared in `pyproject.toml` but **not installed yet**.
