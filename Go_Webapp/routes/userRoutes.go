package routes

import (
	"my-project/controllers" // Update with your actual module path
	"my-project/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers the user management endpoints.
func RegisterUserRoutes(router *gin.RouterGroup) {

	// 1. Create User (Public)
	// Node: router.post("/", createUser)
	router.POST("/", controllers.CreateUser)

	// 2. Verify Email (Public)
	// Node: router.get('/verifyEmail', verifyEmail)
	// Query params (like ?token=xyz) are handled inside the controller in Gin.
	router.GET("/verifyEmail", controllers.VerifyEmail)

	// 3. Get User Details (Auth required)
	// Node: router.get("/:userId", authenticateUser, getUser)
	router.GET("/:userId", middleware.AuthenticateUser(), controllers.GetUser)

	// 4. Update User (Auth required)
	// Node: router.put("/:userId", authenticateUser, updateUser)
	router.PUT("/:userId", middleware.AuthenticateUser(), controllers.UpdateUser)

	// 5. Other Methods (HEAD, OPTIONS, PATCH) - (Auth required)
	// Node: router.head/options/patch("/:userId", authenticateUser, otherMethods)
	// In your Node code, you explicitly routed these to a handler (likely to return 405 or specific headers).
	router.HEAD("/:userId", middleware.AuthenticateUser(), controllers.OtherMethods)
	router.OPTIONS("/:userId", middleware.AuthenticateUser(), controllers.OtherMethods)
	router.PATCH("/:userId", middleware.AuthenticateUser(), controllers.OtherMethods)
}
