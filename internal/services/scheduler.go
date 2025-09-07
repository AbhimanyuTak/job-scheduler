package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/manyu/job-scheduler/internal/storage"
	"github.com/manyu/job-scheduler/internal/utils"
)

// SchedulerService handles job execution and scheduling
type SchedulerService struct {
	storage        *storage.PostgresStorage
	scheduleParser *utils.ScheduleParser
	httpClient     *http.Client
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(storage *storage.PostgresStorage) *SchedulerService {
	return &SchedulerService{
		storage:        storage,
		scheduleParser: utils.NewScheduleParser(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessReadyJobs processes jobs that are ready for execution
func (s *SchedulerService) ProcessReadyJobs(ctx context.Context, limit int) error {
	jobs, schedules, err := s.storage.GetJobsReadyForExecution(limit)
	if err != nil {
		return fmt.Errorf("failed to get ready jobs: %w", err)
	}

	for i, job := range jobs {
		schedule := schedules[i]
		if err := s.executeJob(ctx, job, schedule); err != nil {
			log.Printf("Failed to execute job %d: %v", job.ID, err)
			// Continue processing other jobs even if one fails
		}
	}

	return nil
}

// executeJob executes a single job and handles rescheduling
func (s *SchedulerService) executeJob(ctx context.Context, job *models.Job, schedule *models.JobSchedule) error {
	// Get the current retry count from previous executions
	retryCount := s.getCurrentRetryCount(job.ID)

	// Create job execution record
	execution := &models.JobExecution{
		JobID:         job.ID,
		Status:        models.StatusScheduled,
		ExecutionTime: time.Now(),
		RetryCount:    retryCount,
	}

	if err := s.storage.CreateJobExecution(execution); err != nil {
		return fmt.Errorf("failed to create execution record: %w", err)
	}

	// Update execution status to running
	execution.Status = models.StatusRunning
	if err := s.storage.UpdateJobExecution(execution); err != nil {
		log.Printf("Failed to update execution status to running for job %d: %v", job.ID, err)
	}

	// Execute the job
	startTime := time.Now()
	success := s.callJobAPI(ctx, job.API)
	executionDuration := time.Since(startTime)
	execution.ExecutionDuration = &executionDuration

	// Update execution status based on result
	if success {
		execution.Status = models.StatusSuccess
		log.Printf("Job %d executed successfully (attempt %d)", job.ID, retryCount+1)
	} else {
		execution.Status = models.StatusFailed
		execution.Error = "API call failed"
		log.Printf("Job %d failed (attempt %d/%d)", job.ID, retryCount+1, job.MaxRetryCount)
	}

	if err := s.storage.UpdateJobExecution(execution); err != nil {
		log.Printf("Failed to update execution status for job %d: %v", job.ID, err)
	}

	// Handle rescheduling based on job type and retry logic
	if err := s.handleJobRescheduling(job, schedule, success, retryCount); err != nil {
		log.Printf("Failed to reschedule job %d: %v", job.ID, err)
	}

	return nil
}

// callJobAPI makes HTTP call to the job's API endpoint
func (s *SchedulerService) callJobAPI(ctx context.Context, apiURL string) bool {
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if err != nil {
		log.Printf("Failed to create request for %s: %v", apiURL, err)
		return false
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to call API %s: %v", apiURL, err)
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx status codes as success
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// getCurrentRetryCount gets the current retry count for a job
func (s *SchedulerService) getCurrentRetryCount(jobID uint) int {
	// Get the most recent execution for this job
	executions, err := s.storage.GetJobExecutions(jobID, 1)
	if err != nil || len(executions) == 0 {
		return 0 // First execution
	}

	// Return the retry count from the most recent execution
	// This represents how many times we've already tried
	return executions[0].RetryCount
}

// handleJobRescheduling handles rescheduling logic for jobs after execution
func (s *SchedulerService) handleJobRescheduling(job *models.Job, schedule *models.JobSchedule, executionSuccess bool, retryCount int) error {
	// If execution was successful, handle normal rescheduling
	if executionSuccess {
		return s.handleSuccessfulExecution(job, schedule)
	}

	// If execution failed, handle retry logic
	return s.handleFailedExecution(job, schedule, retryCount)
}

// handleSuccessfulExecution handles rescheduling for successful executions
func (s *SchedulerService) handleSuccessfulExecution(job *models.Job, schedule *models.JobSchedule) error {
	// For non-recurring jobs, delete the schedule after successful execution
	if !job.IsRecurring {
		if err := s.storage.DeleteJobSchedule(job.ID); err != nil {
			return fmt.Errorf("failed to delete schedule for non-recurring job: %w", err)
		}
		log.Printf("Non-recurring job %d completed successfully, schedule deleted", job.ID)
		return nil
	}

	// For recurring jobs, calculate next execution time
	nextExecutionTime, err := s.scheduleParser.CalculateNextExecutionFromTime(job.Schedule, schedule.NextExecutionTime)
	if err != nil {
		return fmt.Errorf("failed to calculate next execution time: %w", err)
	}

	// Update the schedule with next execution time
	if err := s.storage.UpdateJobSchedule(job.ID, nextExecutionTime); err != nil {
		return fmt.Errorf("failed to update job schedule: %w", err)
	}

	log.Printf("Recurring job %d completed successfully, rescheduled for %v", job.ID, nextExecutionTime)
	return nil
}

// handleFailedExecution handles retry logic for failed executions
func (s *SchedulerService) handleFailedExecution(job *models.Job, schedule *models.JobSchedule, retryCount int) error {
	// Check if we should retry based on job type and retry count
	shouldRetry := s.shouldRetryJob(job, retryCount)

	if !shouldRetry {
		// Max retries exceeded or AT_MOST_ONCE job type
		if !job.IsRecurring {
			// For non-recurring jobs, delete the schedule after max retries
			if err := s.storage.DeleteJobSchedule(job.ID); err != nil {
				return fmt.Errorf("failed to delete schedule for failed non-recurring job: %w", err)
			}
			log.Printf("Non-recurring job %d failed after %d retries, schedule deleted", job.ID, retryCount+1)
		} else {
			// For recurring jobs, reschedule normally (next occurrence)
			nextExecutionTime, err := s.scheduleParser.CalculateNextExecutionFromTime(job.Schedule, schedule.NextExecutionTime)
			if err != nil {
				return fmt.Errorf("failed to calculate next execution time: %w", err)
			}
			if err := s.storage.UpdateJobSchedule(job.ID, nextExecutionTime); err != nil {
				return fmt.Errorf("failed to update job schedule: %w", err)
			}
			log.Printf("Recurring job %d failed after %d retries, rescheduled for next occurrence: %v", job.ID, retryCount+1, nextExecutionTime)
		}
		return nil
	}

	// Schedule retry with exponential backoff
	retryDelay := s.calculateRetryDelay(retryCount)
	nextRetryTime := time.Now().Add(retryDelay)

	if err := s.storage.UpdateJobSchedule(job.ID, nextRetryTime); err != nil {
		return fmt.Errorf("failed to schedule retry: %w", err)
	}

	log.Printf("Job %d failed, scheduling retry %d/%d in %v (at %v)",
		job.ID, retryCount+1, job.MaxRetryCount, retryDelay, nextRetryTime.Format("2006-01-02 15:04:05"))

	return nil
}

// shouldRetryJob determines if a job should be retried
func (s *SchedulerService) shouldRetryJob(job *models.Job, retryCount int) bool {
	// Don't retry if we've exceeded max retry count
	if retryCount >= job.MaxRetryCount {
		return false
	}

	// AT_MOST_ONCE jobs should not be retried on failure
	if job.Type == models.AT_MOST_ONCE {
		return false
	}

	// AT_LEAST_ONCE jobs should be retried
	return job.Type == models.AT_LEAST_ONCE
}

// calculateRetryDelay calculates the delay for the next retry using exponential backoff
func (s *SchedulerService) calculateRetryDelay(retryCount int) time.Duration {
	// Exponential backoff: 2^retryCount seconds, max 5 minutes
	delaySeconds := 1 << retryCount // 1, 2, 4, 8, 16, 32, 64, 128, 256, 512 seconds
	maxDelay := 5 * time.Minute

	delay := time.Duration(delaySeconds) * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// DeleteJobSchedule deletes a job schedule (helper method)
func (s *SchedulerService) DeleteJobSchedule(jobID uint) error {
	return s.storage.DeleteJobSchedule(jobID)
}

// GetJobSchedule retrieves a job schedule
func (s *SchedulerService) GetJobSchedule(jobID uint) (*models.JobSchedule, error) {
	return s.storage.GetJobSchedule(jobID)
}
