// Command worker runs background jobs (River), scheduled jobs and the outbox relay.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dprince-03/techshop/backend/internal/app"
	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/internal/database"
	"github.com/dprince-03/techshop/backend/pkg/logging"
)

func main() {
	cfg, err := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err != nil {
		logger.Error("config", slog.Any("error", err))
		os.Exit(1)
	}
	logger = logging.New(cfg.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()
	a, err := app.New(cfg, pool, logger, app.Options{Work: true})
	if err != nil {
		logger.Error("app", slog.Any("error", err))
		os.Exit(1)
	}
	if err := a.Start(ctx, true); err != nil {
		logger.Error("start", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("worker running")
	<-ctx.Done()
	a.Stop(context.Background())
}
