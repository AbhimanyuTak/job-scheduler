package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/manyu/job-scheduler/internal/database"
	"github.com/manyu/job-scheduler/internal/services"
	"github.com/manyu/job-scheduler/internal/storage"
)

func main() {
	// Load environment variables from config file
	loadConfig()

	// Initialize database connection
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database schema
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Redis client
	redisClient, err := services.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize PostgreSQL storage
	postgresStorage := storage.NewPostgresStorage()

	// Initialize job queue service
	jobQueue := services.NewJobQueueService(redisClient)

	// Initialize worker service
	workerService := services.NewWorkerService(jobQueue, postgresStorage)

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

func loadConfig() {
	// Try to load from config.env file
	file, err := os.Open("config.env")
	if err != nil {
		log.Println("config.env file not found, using system environment variables")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading config.env: %v", err)
	} else {
		log.Println("Configuration loaded from config.env")
	}
}
