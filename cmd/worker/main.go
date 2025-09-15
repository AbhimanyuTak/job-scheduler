package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/manyu/job-scheduler/internal/config"
	"github.com/manyu/job-scheduler/internal/database"
	"github.com/manyu/job-scheduler/internal/redis"
	"github.com/manyu/job-scheduler/internal/services"
	"github.com/manyu/job-scheduler/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database service
	dbService, err := database.NewDatabaseService(cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbService.Close()

	// Initialize Redis client with config
	redisClient, err := redis.NewRedisClient(cfg.Redis.GetRedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize PostgreSQL storage
	postgresStorage := storage.NewPostgresStorage(dbService)

	// Initialize job queue service
	jobQueue := services.NewJobQueueService(redisClient)

	// Initialize scheduler service
	schedulerService := services.NewSchedulerService(postgresStorage, redisClient)

	// Initialize worker service
	workerService := services.NewWorkerService(jobQueue, postgresStorage, schedulerService)

	// Start worker service
	workerService.Start()

	log.Println("Worker service started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	<-sigChan
	log.Println("Received shutdown signal")

	// Stop worker service gracefully
	workerService.Stop()

	log.Println("Worker service shutdown complete")
}
