package database

import (
	"fmt"
	"log"

	"github.com/manyu/job-scheduler/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseService wraps the database connection and operations
type DatabaseService struct {
	db *gorm.DB
}

// NewDatabaseService creates a new database service
func NewDatabaseService(dsn string) (*DatabaseService, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	service := &DatabaseService{db: db}

	// Auto-migrate the schema
	if err := service.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	return service, nil
}

// GetDB returns the underlying GORM database instance
func (ds *DatabaseService) GetDB() *gorm.DB {
	return ds.db
}

// AutoMigrate runs database migrations
func (ds *DatabaseService) AutoMigrate() error {
	err := ds.db.AutoMigrate(
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

// Close closes the database connection
func (ds *DatabaseService) Close() error {
	if ds.db != nil {
		sqlDB, err := ds.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Health checks if the database is healthy
func (ds *DatabaseService) Health() error {
	if ds.db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	sqlDB, err := ds.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}
