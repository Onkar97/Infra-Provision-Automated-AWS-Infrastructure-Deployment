package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OtherRoutes handles 404 Not Found for undefined routes
// Usage in main.go: r.NoRoute(middleware.OtherRoutes())
func OtherRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Equivalent to: return res.status(404).end();
		c.Status(http.StatusNotFound)
	}
}