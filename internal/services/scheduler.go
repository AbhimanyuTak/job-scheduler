package services

import (
	"context"
	"fmt"
	"log"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/manyu/job-scheduler/internal/storage"
	"github.com/manyu/job-scheduler/internal/utils"
)

// SchedulerService handles job scheduling and queue management
type SchedulerService struct {
	storage        storage.Storage
	scheduleParser *utils.ScheduleParser
	jobQueue       JobQueueServiceInterface
	redisClient    RedisClientInterface
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(storage storage.Storage, redisClient RedisClientInterface) *SchedulerService {
	jobQueue := NewJobQueueService(redisClient)
	return &SchedulerService{
		storage:        storage,
		scheduleParser: utils.NewScheduleParser(),
		jobQueue:       jobQueue,
		redisClient:    redisClient,
	}
}

// ProcessReadyJobs processes jobs that are ready for execution by enqueueing them
func (s *SchedulerService) ProcessReadyJobs(ctx context.Context, limit int) error {
	jobs, schedules, err := s.storage.GetJobsReadyForExecution(limit)
	if err != nil {
		return fmt.Errorf("failed to get ready jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil
	}

	// Process retry queue first
	if err := s.jobQueue.ProcessRetryQueue(); err != nil {
		log.Printf("Error processing retry queue: %v", err)
	}

	// Enqueue jobs for worker processing
	enqueuedCount := 0
	for i, job := range jobs {
		schedule := schedules[i]

		// Create queue job
		queueJob := models.NewQueueJob(job, schedule)

		// Enqueue the job
		if err := s.jobQueue.EnqueueJob(queueJob); err != nil {
			log.Printf("Failed to enqueue job %d: %v", job.ID, err)
			continue
		}

		enqueuedCount++
	}

	if enqueuedCount > 0 {
		log.Printf("Enqueued %d jobs for processing", enqueuedCount)
	}

	return nil
}

// GetQueueStats returns queue statistics
func (s *SchedulerService) GetQueueStats() (map[string]int64, error) {
	return s.jobQueue.GetQueueStats()
}

// HandleJobCompletion handles job completion from workers
func (s *SchedulerService) HandleJobCompletion(jobID uint, success bool) error {
	// Get the job and schedule
	job, err := s.storage.GetJob(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	schedule, err := s.storage.GetJobSchedule(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job schedule: %w", err)
	}

	if success {
		return s.handleSuccessfulExecution(job, schedule)
	} else {
		return s.handleFailedExecution(job, schedule)
	}
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

// handleFailedExecution handles rescheduling for failed executions
func (s *SchedulerService) handleFailedExecution(job *models.Job, schedule *models.JobSchedule) error {
	// For non-recurring jobs, delete the schedule after failure
	if !job.IsRecurring {
		if err := s.storage.DeleteJobSchedule(job.ID); err != nil {
			return fmt.Errorf("failed to delete schedule for failed non-recurring job: %w", err)
		}
		log.Printf("Non-recurring job %d failed, schedule deleted", job.ID)
		return nil
	}

	// For recurring jobs, reschedule normally (next occurrence)
	nextExecutionTime, err := s.scheduleParser.CalculateNextExecutionFromTime(job.Schedule, schedule.NextExecutionTime)
	if err != nil {
		return fmt.Errorf("failed to calculate next execution time: %w", err)
	}
	if err := s.storage.UpdateJobSchedule(job.ID, nextExecutionTime); err != nil {
		return fmt.Errorf("failed to update job schedule: %w", err)
	}
	log.Printf("Recurring job %d failed, rescheduled for next occurrence: %v", job.ID, nextExecutionTime)
	return nil
}

// DeleteJobSchedule deletes a job schedule (helper method)
func (s *SchedulerService) DeleteJobSchedule(jobID uint) error {
	return s.storage.DeleteJobSchedule(jobID)
}

// GetJobSchedule retrieves a job schedule
func (s *SchedulerService) GetJobSchedule(jobID uint) (*models.JobSchedule, error) {
	return s.storage.GetJobSchedule(jobID)
}
