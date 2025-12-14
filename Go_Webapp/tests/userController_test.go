package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"my-project/controllers"
	"my-project/db"
	"my-project/middleware"
	"my-project/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupUserTestEnv() (*gin.Engine, *gorm.DB) {
	testDB := db.DB
	testDB.Exec("DELETE FROM users")

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	v1 := r.Group("/v1/user")

	// Public
	v1.POST("/", controllers.CreateUser)

	// Protected
	protected := v1.Group("/")
	protected.Use(middleware.AuthenticateUser())
	{
		protected.GET("/:userId", controllers.GetUser)
		protected.PUT("/:userId", controllers.UpdateUser)
	}

	return r, testDB
}

func TestUserController(t *testing.T) {

	// --- CreateUser Tests ---
	t.Run("POST /v1/user", func(t *testing.T) {
		router, db := setupUserTestEnv()

		t.Run("should create user 201", func(t *testing.T) {
			data := map[string]string{
				"username": fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
				"password": "Password123", "first_name": "F", "last_name": "L",
			}
			body, _ := json.Marshal(data)
			req, _ := http.NewRequest("POST", "/v1/user/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 201 {
				t.Logf("Failed Response Body: %s", w.Body.String())
			}
			assert.Equal(t, 201, w.Code)

			var u models.User
			db.Where("username = ?", data["username"]).First(&u)
			assert.NotEqual(t, "Password123", u.Password)
		})
	})

	// --- Protected Route Tests ---
	t.Run("GET /v1/user/:userId", func(t *testing.T) {
		router, db := setupUserTestEnv()

		password := "mySecretPass"
		hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := models.User{
			Username: "get@example.com", Password: string(hashed), FirstName: "Get", LastName: "User",
		}
		db.Create(&user)

		t.Run("should return 200 with Valid Basic Auth", func(t *testing.T) {
			// Generates /v1/user/1
			req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/user/%d", user.ID), nil)
			req.SetBasicAuth(user.Username, password)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 200 {
				// This log will tell us EXACTLY why it failed if it happens again
				t.Logf("GET Failed. Response Body: %s", w.Body.String())
			}
			assert.Equal(t, 200, w.Code)
		})

		t.Run("should return 401 with Invalid Password", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/user/%d", user.ID), nil)
			req.SetBasicAuth(user.Username, "wrongpassword")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 401, w.Code)
		})
	})

	t.Run("PUT /v1/user/:userId", func(t *testing.T) {
		router, db := setupUserTestEnv()

		password := "oldPass"
		hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := models.User{
			Username: "put@example.com", Password: string(hashed), FirstName: "Old", LastName: "Name",
		}
		db.Create(&user)

		t.Run("should update and return 204", func(t *testing.T) {
			updateData := map[string]string{
				"first_name": "Updated",
				"last_name":  "Name",
				"password":   "NewPassword123",
			}
			body, _ := json.Marshal(updateData)

			req, _ := http.NewRequest("PUT", fmt.Sprintf("/v1/user/%d", user.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth(user.Username, password)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 204 {
				t.Logf("PUT Failed. Code: %d, Body: %s", w.Code, w.Body.String())
			}
			assert.Equal(t, 204, w.Code)

			var updated models.User
			db.First(&updated, user.ID)
			assert.Equal(t, "Updated", updated.FirstName)
		})
	})
}
