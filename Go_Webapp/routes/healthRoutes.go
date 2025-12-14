package routes

import (
	"my-project/controllers" // Import your controllers package

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes sets up the health check endpoints.
// This is equivalent to exporting the router in Node.js.
func RegisterHealthRoutes(router *gin.Engine) {

	// Express: router.all("/healthz", getHealth)
	// Go (Gin): router.Any("/healthz", ...)
	// This matches GET, POST, PUT, HEAD, etc.
	router.Any("/healthz", controllers.GetHealth)
}
