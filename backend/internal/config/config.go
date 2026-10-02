// Package config loads application configuration from the environment.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Env  string `env:"APP_ENV" envDefault:"development"`
	Port string `env:"PORT" envDefault:"8080"`

	DatabaseURL      string `env:"DATABASE_URL,required"`
	MigrateOnStartup bool   `env:"DB_MIGRATE_ON_STARTUP" envDefault:"false"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://localhost:3001,http://localhost:3002,http://localhost:3003,http://localhost:3004,http://techshop.localhost,http://market.techshop.localhost,http://wholesale.techshop.localhost,http://seller.techshop.localhost,http://staff.techshop.localhost"`

	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

// Load reads a local .env file if present, then parses environment variables.
func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	return &cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
