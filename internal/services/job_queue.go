package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/redis/go-redis/v9"
)

// JobQueueService handles job queuing operations using Redis
type JobQueueService struct {
	redisClient *RedisClient
	client      *redis.Client
	ctx         context.Context
}

// Queue names
const (
	QueueReady      = "job_queue:ready"
	QueueProcessing = "job_queue:processing"
	QueueCompleted  = "job_queue:completed"
	QueueFailed     = "job_queue:failed"
	QueueRetrying   = "job_queue:retrying"
)

// NewJobQueueService creates a new job queue service
func NewJobQueueService(redisClient *RedisClient) *JobQueueService {
	return &JobQueueService{
		redisClient: redisClient,
		client:      redisClient.GetClient(),
		ctx:         redisClient.GetContext(),
	}
}

// EnqueueJob adds a job to the ready queue
func (jqs *JobQueueService) EnqueueJob(job *models.QueueJob) error {
	// Serialize the job
	jobData, err := job.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	// Add to ready queue
	if err := jqs.client.LPush(jqs.ctx, QueueReady, jobData).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	log.Printf("Enqueued job %s (JobID: %d) to ready queue", job.ID, job.JobID)
	return nil
}

// DequeueJob removes and returns a job from the ready queue
func (jqs *JobQueueService) DequeueJob(timeout time.Duration) (*models.QueueJob, error) {
	// Block until a job is available or timeout
	result, err := jqs.client.BRPop(jqs.ctx, timeout, QueueReady).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No job available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid result from Redis BRPop")
	}

	// Deserialize the job
	job, err := models.DeserializeQueueJob([]byte(result[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	// Move to processing queue
	if err := jqs.MoveToProcessing(job); err != nil {
		log.Printf("Warning: failed to move job %s to processing queue: %v", job.ID, err)
	}

	return job, nil
}

// MoveToProcessing moves a job from ready to processing queue
func (jqs *JobQueueService) MoveToProcessing(job *models.QueueJob) error {
	jobData, err := job.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	// Add to processing set with job ID as member
	if err := jqs.client.SAdd(jqs.ctx, QueueProcessing, job.ID).Err(); err != nil {
		return fmt.Errorf("failed to add job to processing queue: %w", err)
	}

	// Store job data with TTL (e.g., 6 hours for processing long-running jobs)
	if err := jqs.client.Set(jqs.ctx, fmt.Sprintf("job_data:%s", job.ID), jobData, 6*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store job data: %w", err)
	}

	return nil
}

// CompleteJob marks a job as completed and removes it from processing
func (jqs *JobQueueService) CompleteJob(jobID string, result *models.QueueJobResult) error {
	// Remove from processing queue
	if err := jqs.client.SRem(jqs.ctx, QueueProcessing, jobID).Err(); err != nil {
		log.Printf("Warning: failed to remove job %s from processing queue: %v", jobID, err)
	}

	// Remove job data
	if err := jqs.client.Del(jqs.ctx, fmt.Sprintf("job_data:%s", jobID)).Err(); err != nil {
		log.Printf("Warning: failed to remove job data for %s: %v", jobID, err)
	}

	// Add to completed queue
	resultData, err := result.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize job result: %w", err)
	}

	if err := jqs.client.LPush(jqs.ctx, QueueCompleted, resultData).Err(); err != nil {
		return fmt.Errorf("failed to add job result to completed queue: %w", err)
	}

	// Keep only last 1000 completed jobs
	if err := jqs.client.LTrim(jqs.ctx, QueueCompleted, 0, 999).Err(); err != nil {
		log.Printf("Warning: failed to trim completed queue: %v", err)
	}

	log.Printf("Completed job %s with status %s", jobID, result.Status)
	return nil
}

// FailJob marks a job as failed and handles retry logic
func (jqs *JobQueueService) FailJob(job *models.QueueJob, errorMsg string) error {
	// Remove from processing queue
	if err := jqs.client.SRem(jqs.ctx, QueueProcessing, job.ID).Err(); err != nil {
		log.Printf("Warning: failed to remove job %s from processing queue: %v", job.ID, err)
	}

	// Remove job data
	if err := jqs.client.Del(jqs.ctx, fmt.Sprintf("job_data:%s", job.ID)).Err(); err != nil {
		log.Printf("Warning: failed to remove job data for %s: %v", job.ID, err)
	}

	// Check if job should be retried
	if job.ShouldRetry() {
		// Increment retry count and schedule retry
		retryJob := job.IncrementRetry()
		retryDelay := retryJob.CalculateRetryDelay()

		// Schedule retry using Redis delayed execution
		retryTime := time.Now().Add(retryDelay)
		retryJobData, err := retryJob.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize retry job: %w", err)
		}

		// Use Redis sorted set for delayed execution
		score := float64(retryTime.Unix())
		if err := jqs.client.ZAdd(jqs.ctx, QueueRetrying, redis.Z{
			Score:  score,
			Member: retryJobData,
		}).Err(); err != nil {
			return fmt.Errorf("failed to schedule retry: %w", err)
		}

		log.Printf("Scheduled retry %d/%d for job %s in %v",
			retryJob.RetryCount, retryJob.MaxRetryCount, job.ID, retryDelay)
	} else {
		// Max retries exceeded, mark as permanently failed
		result := &models.QueueJobResult{
			JobID:      job.ID,
			Status:     models.QueueStatusFailed,
			Success:    false,
			Error:      errorMsg,
			RetryCount: job.RetryCount,
		}

		if err := jqs.CompleteJob(job.ID, result); err != nil {
			return fmt.Errorf("failed to mark job as permanently failed: %w", err)
		}

		log.Printf("Job %s permanently failed after %d retries", job.ID, job.RetryCount)
	}

	return nil
}

// ProcessRetryQueue moves ready retry jobs back to the ready queue
func (jqs *JobQueueService) ProcessRetryQueue() error {
	now := time.Now().Unix()

	// Get jobs that are ready for retry
	jobs, err := jqs.client.ZRangeByScore(jqs.ctx, QueueRetrying, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", now),
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get retry jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil
	}

	// Move jobs back to ready queue
	for _, jobData := range jobs {
		job, err := models.DeserializeQueueJob([]byte(jobData))
		if err != nil {
			log.Printf("Warning: failed to deserialize retry job: %v", err)
			continue
		}

		// Remove from retry queue
		if err := jqs.client.ZRem(jqs.ctx, QueueRetrying, jobData).Err(); err != nil {
			log.Printf("Warning: failed to remove job from retry queue: %v", err)
		}

		// Add back to ready queue
		if err := jqs.EnqueueJob(job); err != nil {
			log.Printf("Warning: failed to re-enqueue retry job: %v", err)
		}
	}

	if len(jobs) > 0 {
		log.Printf("Processed %d retry jobs", len(jobs))
	}

	return nil
}

// GetQueueStats returns statistics about the job queues
func (jqs *JobQueueService) GetQueueStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Get queue lengths
	readyLen, err := jqs.client.LLen(jqs.ctx, QueueReady).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get ready queue length: %w", err)
	}
	stats["ready"] = readyLen

	processingLen, err := jqs.client.SCard(jqs.ctx, QueueProcessing).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get processing queue length: %w", err)
	}
	stats["processing"] = processingLen

	completedLen, err := jqs.client.LLen(jqs.ctx, QueueCompleted).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get completed queue length: %w", err)
	}
	stats["completed"] = completedLen

	retryingLen, err := jqs.client.ZCard(jqs.ctx, QueueRetrying).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get retry queue length: %w", err)
	}
	stats["retrying"] = retryingLen

	return stats, nil
}

// CleanupStaleJobs removes jobs that have been in processing for too long
func (jqs *JobQueueService) CleanupStaleJobs(maxProcessingTime time.Duration) error {
	// Get all jobs in processing queue
	jobIDs, err := jqs.client.SMembers(jqs.ctx, QueueProcessing).Result()
	if err != nil {
		return fmt.Errorf("failed to get processing jobs: %w", err)
	}

	staleCount := 0
	for _, jobID := range jobIDs {
		// Check if job data exists and is stale
		exists, err := jqs.client.Exists(jqs.ctx, fmt.Sprintf("job_data:%s", jobID)).Result()
		if err != nil {
			log.Printf("Warning: failed to check job data for %s: %v", jobID, err)
			continue
		}

		if exists == 0 {
			// Job data doesn't exist, remove from processing queue
			if err := jqs.client.SRem(jqs.ctx, QueueProcessing, jobID).Err(); err != nil {
				log.Printf("Warning: failed to remove stale job %s: %v", jobID, err)
			} else {
				staleCount++
			}
		}
	}

	if staleCount > 0 {
		log.Printf("Cleaned up %d stale jobs from processing queue", staleCount)
	}

	return nil
}
