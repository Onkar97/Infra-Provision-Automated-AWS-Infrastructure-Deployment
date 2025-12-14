package routes

import (
	"my-project/controllers" // Update with your actual module path
	"my-project/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterProductRoutes registers the CRUD endpoints for products.
func RegisterProductRoutes(router *gin.RouterGroup) {

	// 1. Create Product (Auth required)
	// Node: router.post("/", authenticateUser, createProduct)
	router.POST("/", middleware.AuthenticateUser(), controllers.CreateProduct)

	// 2. Get Single Product (Public)
	// Node: router.get("/:productId", getProduct)
	router.GET("/:productId", controllers.GetProduct)

	// 3. Get All Products (Public)
	// Node: router.get("/", getAllProduct)
	router.GET("/", controllers.GetAllProduct)

	// 4. Update Product - PUT (Auth required)
	// Node: router.put("/:productId", authenticateUser, updatePutProduct)
	router.PUT("/:productId", middleware.AuthenticateUser(), controllers.UpdatePutProduct)

	// 5. Update Product - PATCH (Auth required)
	// Node: router.patch("/:productId", authenticateUser, updatePatchProduct)
	router.PATCH("/:productId", middleware.AuthenticateUser(), controllers.UpdatePatchProduct)

	// 6. Delete Product (Auth required)
	// Node: router.delete("/:productId", authenticateUser, deleteProduct)
	router.DELETE("/:productId", middleware.AuthenticateUser(), controllers.DeleteProduct)

	// 7. Options (Auth required)
	// Node: router.options("/:productId", authenticateUser, otherMethods)
	router.OPTIONS("/:productId", middleware.AuthenticateUser(), controllers.OtherMethods)
}
