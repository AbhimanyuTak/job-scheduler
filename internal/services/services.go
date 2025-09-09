package services

import (
	"context"
	"time"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/redis/go-redis/v9"
)

// SchedulerServiceInterface defines the interface for scheduler operations
type SchedulerServiceInterface interface {
	ProcessReadyJobs(ctx context.Context, limit int) error
	GetQueueStats() (map[string]int64, error)
	HandleJobCompletion(jobID uint, success bool) error
}

// JobQueueServiceInterface defines the interface for queue operations
type JobQueueServiceInterface interface {
	EnqueueJob(job *models.QueueJob) error
	DequeueJob(timeout time.Duration) (*models.QueueJob, error)
	CompleteJob(jobID string, result *models.QueueJobResult) error
	GetQueueStats() (map[string]int64, error)
	ProcessRetryQueue() error
}

// RedisClientInterface defines the interface for Redis operations
type RedisClientInterface interface {
	GetClient() *redis.Client
	GetContext() context.Context
	Close() error
	Health() error
}
