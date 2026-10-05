package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs each request using structured logging.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		level := slog.LevelInfo
		switch status := c.Writer.Status(); {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		if c.Request.URL.Path == "/livez" || c.Request.URL.Path == "/readyz" {
			return // health checks sampled away
		}
		logger.Log(c.Request.Context(), level, "request",
			slog.String("request_id", c.GetString("request_id")),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
