package middleware

import (
	"github.com/gin-gonic/gin"
)

// SetHeaders applies global HTTP headers for security and caching
func SetHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Equivalent to: res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

		// Equivalent to: res.set('Pragma', 'no-cache');
		c.Header("Pragma", "no-cache")

		// Equivalent to: res.set('X-Content-Type-Options', 'nosniff');
		c.Header("X-Content-Type-Options", "nosniff")

		c.Next()
	}
}