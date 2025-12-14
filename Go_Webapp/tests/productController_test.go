package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"my-project/controllers"
	"my-project/db"
	"my-project/middleware" // <--- Import Real Middleware
	"my-project/models"
)

func setupProductTestEnv() (*gin.Engine, *models.User, *gorm.DB) {
	testDB := db.DB

	// 1. Clean DB
	testDB.Exec("DELETE FROM image")
	testDB.Exec("DELETE FROM products")
	testDB.Exec("DELETE FROM users")

	// 2. Create User with KNOWN Password
	password := "password123"
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Username:  "product.test@example.com",
		Password:  string(hashedPwd), // Save HASHED password in DB
		FirstName: "Test",
		LastName:  "User",
	}
	testDB.Create(&user)

	// 3. Setup Router with REAL Middleware
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	v1 := r.Group("/v1/product")

	// Public
	v1.GET("/:productId", controllers.GetProduct)
	v1.GET("/", controllers.GetAllProduct)

	// Protected - Use REAL Basic Auth Middleware
	protected := v1.Group("/")
	protected.Use(middleware.AuthenticateUser())
	{
		protected.POST("/", controllers.CreateProduct)
		protected.PUT("/:productId", controllers.UpdatePutProduct)
		protected.PATCH("/:productId", controllers.UpdatePatchProduct)
		protected.DELETE("/:productId", controllers.DeleteProduct)
	}

	// Hack: We return the "plain text" password in the user struct
	// so the test knows what to send in Basic Auth headers
	user.Password = password
	return r, &user, testDB
}

func TestProductController(t *testing.T) {

	t.Run("POST /v1/product (CreateProduct)", func(t *testing.T) {
		router, user, db := setupProductTestEnv()

		validProduct := map[string]interface{}{
			"name": "New Gadget", "description": "Desc", "sku": "GAD-001",
			"manufacturer": "Gadgets Inc.", "quantity": 25,
		}

		t.Run("should create a product and return 201", func(t *testing.T) {
			body, _ := json.Marshal(validProduct)
			req, _ := http.NewRequest("POST", "/v1/product/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// USE BASIC AUTH
			req.SetBasicAuth(user.Username, user.Password)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 201, w.Code)

			var count int64
			db.Model(&models.Product{}).Where("owner_user_id = ?", user.ID).Count(&count)
			assert.Equal(t, int64(1), count)
		})

		t.Run("should return 401 if auth is missing", func(t *testing.T) {
			body, _ := json.Marshal(validProduct)
			req, _ := http.NewRequest("POST", "/v1/product/", bytes.NewBuffer(body))
			// No Basic Auth Set
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 401, w.Code)
		})
	})

	t.Run("DELETE /v1/product/:productId", func(t *testing.T) {
		router, user, db := setupProductTestEnv()

		// Create product owned by user
		product := models.Product{Name: "To Delete", OwnerUserID: user.ID}
		db.Create(&product)

		t.Run("should delete and return 204", func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/v1/product/%d", product.ID), nil)

			// USE BASIC AUTH
			req.SetBasicAuth(user.Username, user.Password)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 204, w.Code)
		})
	})
}
