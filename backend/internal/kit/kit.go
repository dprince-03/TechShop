// Package kit holds the shared dependencies every domain module receives from the composition
// root (internal/app), and the small Module interface modules implement.
package kit

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/idempotency"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/ratelimit"
	"github.com/dprince-03/techshop/backend/internal/rbac"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/storage"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// Deps are shared by all modules.
type Deps struct {
	Cfg     *config.Config
	Logger  *slog.Logger
	Pool    *pgxpool.Pool
	Runner  *uow.Runner
	Q       *store.Queries // read-only queries outside a transaction
	Keys    *auth.Keyring
	RBAC    *rbac.Checker
	Cursors *httpx.Cursors
	Limiter *ratelimit.Limiter
	Idem    *idempotency.Store
	Storage storage.Port
	Sealer  crypto.Sealer
	Hub     *realtime.Hub
}

// Routes gives modules what they need to register endpoints.
type Routes struct {
	Public   *httpx.Guard // public API (/api/v1)
	Internal *httpx.Guard // internal listener (/internal), never routed by nginx
	Gin      *gin.Engine  // raw routes (SSE streams, dev helpers)
	API      huma.API
}

// Module is implemented by every domain package.
type Module interface {
	Name() string
	Routes(r *Routes)
	Jobs(reg *jobs.Registry)
	Events(relay *outbox.Relay)
}
