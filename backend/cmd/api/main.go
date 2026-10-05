// Command api runs the public HTTP API (and, when WORKER_IN_PROCESS=true, the job worker) plus
// the internal listener (INTERNAL_ADDR) that nginx never routes.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dprince-03/techshop/backend/internal/app"
	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/internal/database"
	"github.com/dprince-03/techshop/backend/pkg/logging"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(); err != nil {
		logger.Error("api exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := logging.New(cfg.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	if cfg.MigrateOnStartup {
		if err := database.Migrate(ctx, pool); err != nil {
			return err
		}
		logger.Info("database migrations applied")
	}

	a, err := app.New(cfg, pool, logger, app.Options{Work: cfg.WorkerInProcess})
	if err != nil {
		return err
	}
	if err := a.Start(ctx, cfg.WorkerInProcess); err != nil {
		return err
	}
	defer a.Stop(context.Background())

	public := &http.Server{Addr: ":" + cfg.Port, Handler: a.Public, ReadTimeout: cfg.ReadTimeout, WriteTimeout: 0, ReadHeaderTimeout: cfg.ReadTimeout}
	internal := &http.Server{Addr: cfg.InternalAddr, Handler: a.Internal, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, ReadHeaderTimeout: cfg.ReadTimeout}
	errCh := make(chan error, 2)
	for _, srv := range []*http.Server{public, internal} {
		go func(s *http.Server) {
			logger.Info("listening", slog.String("addr", s.Addr), slog.String("env", cfg.Env))
			if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}
		}(srv)
	}
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	_ = internal.Shutdown(shutdownCtx)
	return public.Shutdown(shutdownCtx)
}
