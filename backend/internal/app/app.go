// Package app is the composition root: it builds shared dependencies, every domain module,
// the public and internal HTTP servers, River and the outbox relay. Provider adapters are
// chosen here (fake/sandbox by default; no live calls).
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/idempotency"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/middleware"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/ratelimit"
	"github.com/dprince-03/techshop/backend/internal/rbac"
	"github.com/dprince-03/techshop/backend/internal/realtime"
	"github.com/dprince-03/techshop/backend/internal/storage"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// App is the assembled application.
type App struct {
	Deps     *kit.Deps
	Modules  []kit.Module
	Public   *gin.Engine
	Internal *gin.Engine
	API      huma.API
	IntAPI   huma.API
	River    *river.Client[pgx.Tx]
	Relay    *outbox.Relay
	registry *jobs.Registry
}

// Options control which parts run.
type Options struct {
	Work bool // run River workers and the outbox relay in this process
}

// New builds the application. pool may be nil when only the OpenAPI spec is needed.
func New(cfg *config.Config, pool *pgxpool.Pool, logger *slog.Logger, opts Options) (*App, error) {
	httpx.Install(logger)
	keys, err := auth.NewKeyring(cfg.JWTSigningKeys, cfg.JWTIssuer, cfg.IsDevelopment() || cfg.Env == "test")
	if err != nil {
		return nil, err
	}
	sealer, err := crypto.NewLocalSealer(cfg.LocalKEK)
	if err != nil {
		return nil, err
	}
	var q *store.Queries
	if pool != nil {
		q = store.New(pool)
	}
	var st storage.Port = storage.Fake{BaseURL: cfg.AppBaseURL}
	if cfg.StorageProvider == "s3" {
		st = storage.NewS3(storage.S3Config{Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, PublicBucket: cfg.S3BucketPublic, PrivateBucket: cfg.S3BucketPrivate,
			AccessKeyID: cfg.S3AccessKeyID, SecretAccessKey: cfg.S3SecretAccessKey, ForcePathStyle: cfg.S3ForcePathStyle, PublicBaseURL: cfg.PublicAssetBaseURL})
	}
	checker := rbac.NewChecker(q)
	d := &kit.Deps{
		Cfg: cfg, Logger: logger, Pool: pool, Q: q, Keys: keys, RBAC: checker,
		Cursors: httpx.NewCursors(cfg.CursorHMACKey), Limiter: ratelimit.New(q), Idem: idempotency.New(q),
		Storage: st, Sealer: sealer, Hub: realtime.NewHub(pool, logger),
	}
	d.Runner = &uow.Runner{Pool: pool, Logger: logger}

	a := &App{Deps: d, registry: jobs.NewRegistry()}
	a.Modules = buildModules(d)
	for _, m := range a.Modules {
		m.Jobs(a.registry)
	}
	if pool != nil {
		a.River, err = jobs.NewClient(pool, a.registry, opts.Work, cfg.WorkerConcurrency, logger)
		if err != nil {
			return nil, fmt.Errorf("river: %w", err)
		}
		d.Runner.River = a.River
		a.Relay = outbox.NewRelay(d.Runner, pool, logger)
		for _, m := range a.Modules {
			m.Events(a.Relay)
		}
		d.Hub.OnChannel(realtime.ChannelAuthz, checker.InvalidateUser)
		d.Hub.OnChannel(realtime.ChannelSessions, checker.InvalidateSessions)
	}
	a.buildHTTP()
	return a, nil
}

func newEngine(cfg *config.Config, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode) // request logging goes through slog, not Gin's debug output
	r := gin.New()
	_ = r.SetTrustedProxies(cfg.TrustedProxies)
	r.Use(gin.Recovery(), httpx.RequestID(), middleware.Logger(logger), httpx.SecurityHeaders())
	return r
}

func humaConfig(title string) huma.Config {
	c := huma.DefaultConfig(title, "1.0.0")
	c.CreateHooks = nil // no $schema links in responses
	c.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}
	c.Info.Description = "TechShop API. Money is integer kobo; errors are application/problem+json with a stable `code`."
	return c
}

func (a *App) buildHTTP() {
	cfg := a.Deps.Cfg
	a.Public = newEngine(cfg, a.Deps.Logger)
	a.Public.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	pc := humaConfig("TechShop API")
	if cfg.IsProduction() {
		pc.DocsPath = ""
	}
	a.API = humagin.New(a.Public, pc)

	a.Internal = newEngine(cfg, a.Deps.Logger)
	ic := humaConfig("TechShop internal API")
	ic.OpenAPIPath, ic.DocsPath = "/internal/openapi", ""
	a.IntAPI = humagin.New(a.Internal, ic)

	pub := &httpx.Guard{API: a.API, Keys: a.Deps.Keys, Sessions: a.Deps.RBAC, Perms: a.Deps.RBAC, StepUpMaxAge: cfg.StepUpMaxAge}
	in := &httpx.Guard{API: a.IntAPI, Keys: a.Deps.Keys, Perms: a.Deps.RBAC, StepUpMaxAge: cfg.StepUpMaxAge}
	routes := &kit.Routes{Public: pub, Internal: in, Gin: a.Public, API: a.API}
	for _, m := range a.Modules {
		m.Routes(routes)
	}
	a.registerHealth()
}

func (a *App) registerHealth() {
	live := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
	ready := func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if a.Deps.Pool == nil || a.Deps.Pool.Ping(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "database": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "up"})
	}
	a.Public.GET("/health", ready) // kept for existing checks: includes the database
	a.Public.GET("/livez", live)
	a.Public.GET("/readyz", ready)
	a.Internal.GET("/internal/healthz", ready)
	a.Internal.GET("/internal/metrics", func(c *gin.Context) {
		lag, _ := a.Deps.Q.PlatformOutboxLag(c.Request.Context())
		jobsSummary, _ := a.Deps.Q.PlatformListJobsSummary(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"outbox": lag, "jobs": jobsSummary})
	})
}

// Start runs background parts: the realtime hub (always), and River + the outbox relay when
// this process works jobs.
func (a *App) Start(ctx context.Context, work bool) error {
	if a.Deps.Pool == nil {
		return nil
	}
	if err := rbac.Sync(ctx, a.Deps.Q); err != nil {
		return fmt.Errorf("sync permissions: %w", err)
	}
	go a.Deps.Hub.Run(ctx)
	if !work {
		return nil
	}
	if err := a.River.Start(ctx); err != nil {
		return fmt.Errorf("start river: %w", err)
	}
	go a.Relay.Run(ctx)
	return nil
}

// Stop drains River.
func (a *App) Stop(ctx context.Context) {
	if a.River != nil {
		_ = a.River.Stop(ctx)
	}
}
