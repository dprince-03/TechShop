// Command migrate applies pending goose migrations (embedded in the binary). It only migrates
// up; rollbacks are never run against shared databases.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dprince-03/techshop/backend/internal/platform/config"
	"github.com/dprince-03/techshop/backend/internal/platform/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", slog.Any("error", err))
		os.Exit(1)
	}
	if err := database.MigrateWithOwner(context.Background(), cfg); err != nil {
		logger.Error("migrate", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("migrations applied")
}
