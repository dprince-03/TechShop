# Go / Gin Conventions

> **TechShop:** the plan in `docs/backend.md` §1 takes precedence: modular monolith with `internal/<domain>` modules, a unit-of-work helper, outbox + River jobs, `cmd/{api,worker,migrate,seed,openapi}`, goose migrations in `db/migrations`, sqlc queries in `db/queries` generated into `internal/store` (never hand-edited), `log/slog` only, config via `internal/config`. No Redis.

## Layout
```
cmd/api/main.go            # wiring only: config, DB, router, server
internal/
  config/                  # env loading + validation
  server/                  # router setup, middleware registration
  middleware/              # auth, request ID, logging, recovery (+ tenant in multi-tenant apps)
  <feature>/
    handler.go             # Gin handlers (HTTP only)
    service.go             # business logic
    repository.go          # DB access
    model.go               # domain structs
    dto.go                 # request/response structs + validation tags
    errors.go              # domain errors
  platform/                # db, logger, mailer clients (redis only if the project uses it)
migrations/
```

## Style
- `gofmt`/`goimports` always; `golangci-lint` clean.
- Accept interfaces, return structs. Define interfaces where they are *consumed* (in the service package), not where implemented.
- `context.Context` is the first parameter of every function that does I/O.
- Constructor functions: `NewService(repo Repository, log *slog.Logger) *Service`.
- Dependency injection by constructor; no package-level globals for DB/clients.

## Errors
```go
var ErrNotFound = errors.New("not found")

if err != nil {
    return fmt.Errorf("find user %s: %w", id, err)
}
```
- Use `errors.Is/As` to check. Map to HTTP in one place (an error-to-status helper).

## Handlers
```go
func (h *Handler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "invalid_request", err)
        return
    }
    user, err := h.svc.Create(c.Request.Context(), actorFrom(c), req) // actor carries user, roles and scope
    if err != nil {
        respondDomainError(c, err)
        return
    }
    c.JSON(http.StatusCreated, toUserResponse(user))
}
```
- Bind to DTOs, never to DB models (prevents mass assignment).
- Validate with `binding:"required,email"` tags (go-playground/validator).

## Server
- `gin.SetMode(gin.ReleaseMode)` in production; `SetTrustedProxies` configured.
- `http.Server` with `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`.
- Graceful shutdown on SIGINT/SIGTERM.

## Logging
- `log/slog` with JSON handler; add request ID middleware.

## DB
- `pgx` (pgxpool) or `database/sql`; parameterized queries only; `sqlc` is a good fit for type-safe queries.
- Transactions via a helper `WithTx(ctx, func(tx) error)`.

## Tooling
- `go vet`, `golangci-lint`, `gosec`, `govulncheck` in CI.
