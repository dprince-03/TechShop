// Command migrate applies pending goose migrations (embedded in the binary). It only migrates
// up; rollbacks are never run against shared databases.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/internal/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", slog.Any("error", err))
		os.Exit(1)
	}
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		logger.Error("migrate", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("migrations applied")
}
