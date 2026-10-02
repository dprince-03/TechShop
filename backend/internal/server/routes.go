package server

import "github.com/dprince-03/techshop/backend/internal/handler"

func (s *Server) registerRoutes() {
	health := handler.NewHealthHandler(s.db)
	s.router.GET("/health", health.Check)

	v1 := s.router.Group("/api/v1")
	_ = v1 // Register versioned API routes here.
}
