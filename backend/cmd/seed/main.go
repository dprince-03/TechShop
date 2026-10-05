// Command seed creates development data: one test account per role (customer, seller owner,
// business buyer, rider, dispatcher and every PROPOSED staff role), plus the sample catalogue
// and reference setup. It refuses to run unless APP_ENV=development. Credentials are printed as
// JSON to stdout (never committed); set SEED_PASSWORD to choose the shared dev password.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/dprince-03/techshop/backend/internal/app"
	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/internal/database"
	"github.com/dprince-03/techshop/backend/internal/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !cfg.IsDevelopment() {
		fmt.Fprintln(os.Stderr, "seed: refusing to run because APP_ENV is not development")
		os.Exit(2)
	}
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer pool.Close()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	a, err := app.New(cfg, pool, logger, app.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	seed.Catalog, seed.Inventory = app.Mods.Catalog, app.Mods.Inventory
	out, err := seed.Run(ctx, a.Deps, os.Getenv("SEED_PASSWORD"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed failed:", err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
