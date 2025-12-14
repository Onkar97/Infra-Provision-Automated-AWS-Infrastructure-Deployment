package routes

import (
	"my-project/controllers" // Import your specific controllers
	"my-project/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes defines all application routes in one place.
func RegisterRoutes(router *gin.Engine) {

	// --- Health Check ---
	// Node: router.all("/healthz", getHealth);
	router.Any("/healthz", controllers.GetHealth)

	// --- User Routes ---

	// Node: router.post("/user", createUser);
	router.POST("/user", controllers.CreateUser)

	// Node: router.get("/user/:userId", authenticateUser, getUser);
	router.GET("/user/:userId", middleware.AuthenticateUser(), controllers.GetUser)

	// Node: router.put("/user/:userId", authenticateUser, updateUser);
	router.PUT("/user/:userId", middleware.AuthenticateUser(), controllers.UpdateUser)

	// Node: router.head("/user/:userId", otherMethods); (No Auth)
	router.HEAD("/user/:userId", controllers.OtherMethods)

	// Node: router.options("/user/:userId", otherMethods); (No Auth)
	router.OPTIONS("/user/:userId", controllers.OtherMethods)

	// Node: router.patch("/user/:userId", authenticateUser, otherMethods); (With Auth)
	router.PATCH("/user/:userId", middleware.AuthenticateUser(), controllers.OtherMethods)

	// --- Product Routes ---

	// Node: router.post("/product", authenticateUser, createProduct);
	router.POST("/product", middleware.AuthenticateUser(), controllers.CreateProduct)

	// Node: router.get("/product/:productId", authenticateUser, getProduct);
	// Note: In this specific file, you applied Auth to GET product (unlike previous snippets).
	router.GET("/product/:productId", middleware.AuthenticateUser(), controllers.GetProduct)

	// Node: router.put("/product/:productId", authenticateUser, updateProduct);
	router.PUT("/product/:productId", middleware.AuthenticateUser(), controllers.UpdatePutProduct)

	// Node: router.patch("/product/:productId", authenticateUser, updateProduct);
	// Note: You reused 'updateProduct' for both PUT and PATCH here.
	router.PATCH("/product/:productId", middleware.AuthenticateUser(), controllers.UpdatePatchProduct)

	// Node: router.delete("/product/:productId", authenticateUser, deleteProduct);
	router.DELETE("/product/:productId", middleware.AuthenticateUser(), controllers.DeleteProduct)

	// Node: router.options("/product/:productId", otherMethods); (No Auth)
	router.OPTIONS("/product/:productId", controllers.OtherMethods)
}
