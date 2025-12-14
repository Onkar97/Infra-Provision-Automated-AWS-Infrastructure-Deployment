package routes

import (
	"my-project/controllers" // Update with your actual module path
	"my-project/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterImageRoutes defines the routes for image handling.
// In Node, you exported the router. In Go, we pass the engine (or a group) to a function.
func RegisterImageRoutes(router *gin.RouterGroup) {
	// Note: I am assuming 'router' here is already grouped with any base path
	// (e.g., "/v1/product") if you had one in your app.js.
	// If not, use 'router.POST("/v1/product/:productId/image", ...)'

	// 1. POST Image (Auth + File Upload)
	// Node: router.post(..., authenticateUser, upload.single('file'), createImage)
	// Go: The "upload" logic is handled INSIDE controllers.CreateImage
	router.POST("/:productId/image", middleware.AuthenticateUser(), controllers.CreateImage)

	// 2. GET All Images (Public)
	// Node: router.get(..., getAllImage)
	router.GET("/:productId/image", controllers.GetAllImage)

	// 3. GET Single Image (Public)
	// Node: router.get(..., getImage)
	router.GET("/:productId/image/:imageId", controllers.GetImage)

	// 4. DELETE Image (Auth)
	// Node: router.delete(..., authenticateUser, deleteImage)
	router.DELETE("/:productId/image/:imageId", middleware.AuthenticateUser(), controllers.DeleteImage)

	// 5. OPTIONS (Auth)
	// Node: router.options(..., authenticateUser, otherMethods)
	//router.OPTIONS("/:productId", middleware.AuthenticateUser(), controllers.OtherMethods)
}
