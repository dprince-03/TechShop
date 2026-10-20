package httpx

import (
	"github.com/gin-gonic/gin"

	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// RequestID assigns (or forwards) X-Request-ID and stores request metadata for audit entries.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" || len(id) > 64 {
			id = crypto.RandomToken(12)
		}
		c.Header("X-Request-ID", id)
		ctx := WithRequestMeta(c.Request.Context(), RequestMeta{RequestID: id, IP: c.ClientIP(), UserAgent: c.Request.UserAgent()})
		c.Request = c.Request.WithContext(ctx)
		c.Set("request_id", id)
		c.Next()
	}
}

// SecurityHeaders sets conservative headers for an API that serves JSON only.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Cross-Origin-Resource-Policy", "same-site")
		c.Next()
	}
}
