// Package middleware contains Gin middleware.
package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS allows the configured web origins (no origins = no CORS headers at all).
func CORS(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Idempotency-Key", "X-Request-ID", "X-Cart-Token", "X-Anonymous-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "Idempotent-Replayed"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
