package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"my-project/models" // UNCOMMENT THIS LINE after we create the models package
)

// DB is the global database instance (Equivalent to AppDataSource)
var DB *gorm.DB

// InitializeDatabase connects to Postgres and performs migrations
func InitializeDatabase() {
	// 1. Build Connection String (DSN)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DBHOST"),
		os.Getenv("DBUSER"),
		os.Getenv("DBPASSWORD"),
		os.Getenv("DBNAME"),
		os.Getenv("DBPORT"),
	)

	// 2. Configure Logger
	// Equivalent to: logging: !isTestEnv
	var gormLogger logger.Interface
	if os.Getenv("GO_ENV") == "test" {
		gormLogger = logger.Discard // Silent during tests
	} else {
		gormLogger = logger.Default.LogMode(logger.Info) // Standard logging
	}

	// 3. Connect to Database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		// Equivalent to: logger.error("Error during Data Source initialization:", err);
		log.Fatalf("Error during Data Source initialization: %v", err)
	}

	log.Println("PostgreSQL Data Source has been initialized!")

	// 4. Auto Migration (Equivalent to synchronize: true)
	// This automatically creates/updates tables based on your structs.
	// UNCOMMENT THE BLOCK BELOW once you provide the Model files.

	err = DB.AutoMigrate(
		&models.HealthCheck{},
		&models.User{},
		&models.Product{},
		&models.Image{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

}
