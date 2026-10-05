// Command openapi writes the public OpenAPI 3.1 spec (backend/api/openapi.yaml) without a
// database. The frontend's shared/api-client is generated from this file.
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/dprince-03/techshop/backend/internal/app"
	"github.com/dprince-03/techshop/backend/internal/config"
)

func main() {
	out := "api/openapi.yaml"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	cfg := &config.Config{Env: "development", JWTIssuer: "https://api.techshop.ng", OTPHMACKey: "x", CursorHMACKey: "x", TokenHMACKey: "x",
		BlindIndexKey: "x", LocalKEK: "openapi-generation-only", AppBaseURL: "http://localhost:8080"}
	a, err := app.New(cfg, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), app.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	b, err := a.API.OpenAPI().YAML()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(out, b, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d paths)\n", out, len(a.API.OpenAPI().Paths))
}
