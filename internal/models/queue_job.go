package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// QueueJob represents a job in the Redis queue
type QueueJob struct {
	ID            string    `json:"id"`              // Unique queue job ID
	JobID         uint      `json:"job_id"`          // Original job ID from database
	API           string    `json:"api"`             // API endpoint to call
	MaxRetryCount int       `json:"max_retry_count"` // Maximum number of retries
	RetryCount    int       `json:"retry_count"`     // Current retry count
	CreatedAt     time.Time `json:"created_at"`      // When the job was created
	ScheduledAt   time.Time `json:"scheduled_at"`    // When the job should be executed
	Timeout       int       `json:"timeout"`         // Timeout in seconds (default 90)
	Type          JobType   `json:"type"`            // Job type (AT_MOST_ONCE, AT_LEAST_ONCE)
	IsRecurring   bool      `json:"is_recurring"`    // Whether this is a recurring job
	Schedule      string    `json:"schedule"`        // Cron schedule for recurring jobs
}

// QueueJobStatus represents the status of a job in the queue
type QueueJobStatus string

const (
	QueueStatusReady      QueueJobStatus = "ready"
	QueueStatusProcessing QueueJobStatus = "processing"
	QueueStatusCompleted  QueueJobStatus = "completed"
	QueueStatusFailed     QueueJobStatus = "failed"
	QueueStatusRetrying   QueueJobStatus = "retrying"
)

// QueueJobResult represents the result of a job execution
type QueueJobResult struct {
	JobID             string         `json:"job_id"`
	Status            QueueJobStatus `json:"status"`
	Success           bool           `json:"success"`
	Error             string         `json:"error,omitempty"`
	ExecutionTime     time.Time      `json:"execution_time"`
	ExecutionDuration time.Duration  `json:"execution_duration"`
	RetryCount        int            `json:"retry_count"`
	NextExecution     *time.Time     `json:"next_execution,omitempty"` // For recurring jobs
}

// Serialize converts a QueueJob to JSON bytes
func (qj *QueueJob) Serialize() ([]byte, error) {
	return json.Marshal(qj)
}

// DeserializeQueueJob creates a QueueJob from JSON bytes
func DeserializeQueueJob(data []byte) (*QueueJob, error) {
	var job QueueJob
	err := json.Unmarshal(data, &job)
	return &job, err
}

// Serialize converts a QueueJobResult to JSON bytes
func (qjr *QueueJobResult) Serialize() ([]byte, error) {
	return json.Marshal(qjr)
}

// DeserializeQueueJobResult creates a QueueJobResult from JSON bytes
func DeserializeQueueJobResult(data []byte) (*QueueJobResult, error) {
	var result QueueJobResult
	err := json.Unmarshal(data, &result)
	return &result, err
}

// NewQueueJob creates a new QueueJob from a database Job and JobSchedule
func NewQueueJob(job *Job, schedule *JobSchedule) *QueueJob {
	return &QueueJob{
		ID:            generateQueueJobID(job.ID),
		JobID:         job.ID,
		API:           job.API,
		MaxRetryCount: job.MaxRetryCount,
		RetryCount:    0,
		CreatedAt:     time.Now(),
		ScheduledAt:   schedule.NextExecutionTime,
		Timeout:       90, // Default 90 seconds for long-running tasks
		Type:          job.Type,
		IsRecurring:   job.IsRecurring,
		Schedule:      job.Schedule,
	}
}

// generateQueueJobID creates a unique ID for the queue job
func generateQueueJobID(jobID uint) string {
	return fmt.Sprintf("job_%d_%d", jobID, time.Now().UnixNano())
}

// ShouldRetry determines if the job should be retried based on its type and retry count
func (qj *QueueJob) ShouldRetry() bool {
	// Don't retry if we've exceeded max retry count
	if qj.RetryCount >= qj.MaxRetryCount {
		return false
	}

	// AT_MOST_ONCE jobs should not be retried on failure
	if qj.Type == AT_MOST_ONCE {
		return false
	}

	// AT_LEAST_ONCE jobs should be retried
	return qj.Type == AT_LEAST_ONCE
}

// IncrementRetry increments the retry count and returns a new QueueJob
func (qj *QueueJob) IncrementRetry() *QueueJob {
	newJob := *qj
	newJob.RetryCount++
	newJob.ID = generateQueueJobID(qj.JobID) // Generate new ID for retry
	return &newJob
}

// CalculateRetryDelay calculates the delay for the next retry using exponential backoff
func (qj *QueueJob) CalculateRetryDelay() time.Duration {
	// Exponential backoff: 2^retryCount seconds, max 5 minutes
	delaySeconds := 1 << qj.RetryCount // 1, 2, 4, 8, 16, 32, 64, 128, 256, 512 seconds
	maxDelay := 5 * time.Minute

	delay := time.Duration(delaySeconds) * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}
