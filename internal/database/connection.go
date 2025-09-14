package database

import (
	"fmt"
	"log"

	"github.com/manyu/job-scheduler/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dsn string) error {
	// Connect to database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return nil
}

func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Auto-migrate the schema
	err := DB.AutoMigrate(
		&models.Job{},
		&models.JobSchedule{},
		&models.JobExecution{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	log.Println("Database schema migrated successfully")
	return nil
}
