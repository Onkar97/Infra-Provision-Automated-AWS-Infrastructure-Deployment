// tests/main_test.go
package tests

import (
	"log"
	"os"
	"testing"

	"my-project/db"
	"my-project/logs"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	// 1. Load .env
	_ = godotenv.Load("../.env")

	// 2. Initialize DB
	db.InitializeDatabase()

	// Verify DB is open
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatal("Failed to get generic database object: ", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Database connection is dead: ", err)
	}

	// 3. Init Metrics
	os.Setenv("APP_ENV", "test")
	logs.Init()

	// 4. Run Tests
	exitVal := m.Run()

	// 5. Exit
	os.Exit(exitVal)
}
