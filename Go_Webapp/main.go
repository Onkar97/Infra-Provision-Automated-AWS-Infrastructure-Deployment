package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// Import local packages
	"my-project/db"
	"my-project/logs"
	"my-project/middleware"
	"my-project/routes"
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 2. Initialize Logger
	logs.InitLogger()

	// ---------------------------------------------------------
	// 3. Initialize Metrics (ADD THIS BLOCK)
	// ---------------------------------------------------------
	// Check your logs/metrics.go file.
	// If the function is named 'Init', call logs.Init()
	// If you renamed it to 'InitMetrics', call logs.InitMetrics()
	logs.Init()

	// Safety check: Close the connection when main exits
	if logs.Client != nil {
		defer logs.Client.Close()
	}

	// 4. Connect to Database
	db.InitializeDatabase()

	// 5. Initialize Router
	r := gin.New()

	// 6. Global Middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.SetHeaders())
	r.Use(middleware.SetAPITimer())

	// 7. Routes
	routes.RegisterHealthRoutes(r)

	v1User := r.Group("/v1/user")
	routes.RegisterUserRoutes(v1User)

	v1Product := r.Group("/v1/product")
	routes.RegisterProductRoutes(v1Product)
	routes.RegisterImageRoutes(v1Product)

	// 8. Error Handling (404)
	r.NoRoute(middleware.OtherRoutes())

	// 9. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logs.Info("Server running on port " + port)
	if err := r.Run(":" + port); err != nil {
		logs.Fatal("Server failed to start: " + err.Error())
	}
}
