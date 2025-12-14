package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"my-project/controllers"
	"my-project/db"     // <--- Added to access the global DB connection
	"my-project/models" // Assuming you have a HealthCheck model defined here
)

// setupHealthTestEnv prepares the router and database for health checks
func setupHealthTestEnv() (*gin.Engine, *gorm.DB) {
	// 1. Get the global DB connection initialized by TestMain
	testDB := db.DB

	// 2. Setup Router
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Register the Health Route
	// Note: We use r.Any to capture all methods so we can test 405s manually if needed
	r.Any("/healthz", controllers.GetHealth)

	return r, testDB
}

func TestHealthController(t *testing.T) {

	t.Run("GET /healthz", func(t *testing.T) {
		router, database := setupHealthTestEnv()

		t.Run("should return 200 OK and insert a record", func(t *testing.T) {
			// Clear table before test
			// We check if database is not nil to avoid panic if setup failed
			if database != nil {
				database.Exec("DELETE FROM health_checks")
			}

			req, _ := http.NewRequest("GET", "/healthz", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)

			// Verify DB insertion
			if database != nil {
				var count int64
				database.Model(&models.HealthCheck{}).Count(&count)
				assert.Equal(t, int64(1), count)
			}
		})

		// --- Method Not Allowed Tests ---
		methods := []string{"POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
		for _, method := range methods {
			t.Run(fmt.Sprintf("should return 405 for %s", method), func(t *testing.T) {
				req, _ := http.NewRequest(method, "/healthz", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, 405, w.Code)
			})
		}

		// --- Bad Request Tests ---

		t.Run("should return 400 Bad Request for a query string", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/healthz?bad=param", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 400, w.Code)
		})

		t.Run("should return 400 Bad Request for a body", func(t *testing.T) {
			// Gin doesn't auto-reject bodies on GET, so your controller must check Content-Length
			req, _ := http.NewRequest("GET", "/healthz", strings.NewReader("some body"))
			req.Header.Set("Content-Length", "9")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 400, w.Code)
		})

		t.Run("should return 400 Bad Request for an authorization header", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/healthz", nil)
			req.Header.Set("Authorization", "Bearer token")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 400, w.Code)
		})

		t.Run("should return 503 Service Unavailable if database query fails", func(t *testing.T) {
			if database == nil {
				t.Skip("Database not initialized, skipping 503 test")
			}

			// Simulate DB failure by closing the underlying SQL DB
			sqlDB, err := database.DB()
			if err == nil {
				sqlDB.Close()
			}

			req, _ := http.NewRequest("GET", "/healthz", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 503, w.Code)

			// 4. THE FIX: Re-initialize the database connection so subsequent tests work!
			// We call the same function from your db package that you use in main.go
			db.InitializeDatabase()
		})
	})
}
