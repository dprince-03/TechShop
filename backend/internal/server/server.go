// Package server wires up the HTTP server and its routes.
package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dprince-03/techshop/backend/internal/config"
	"github.com/dprince-03/techshop/backend/internal/middleware"
)

type Server struct {
	cfg    *config.Config
	db     *pgxpool.Pool
	logger *slog.Logger
	router *gin.Engine
}

func New(cfg *config.Config, db *pgxpool.Pool, logger *slog.Logger) *Server {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
		middleware.Logger(logger),
		middleware.CORS(cfg.CORSAllowedOrigins),
	)

	s := &Server{cfg: cfg, db: db, logger: logger, router: router}
	s.registerRoutes()
	return s
}

func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:         ":" + s.cfg.Port,
		Handler:      s.router,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}
}
