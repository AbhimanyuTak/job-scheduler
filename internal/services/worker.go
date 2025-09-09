package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/manyu/job-scheduler/internal/storage"
)

// WorkerService handles job execution from the Redis queue
type WorkerService struct {
	jobQueue   *JobQueueService
	storage    *storage.PostgresStorage
	scheduler  SchedulerServiceInterface
	httpClient *http.Client
	workerPool chan struct{} // Semaphore for limiting concurrent workers
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	shutdown   bool
	shutdownMu sync.RWMutex
}

// NewWorkerService creates a new worker service
func NewWorkerService(jobQueue *JobQueueService, storage *storage.PostgresStorage, scheduler SchedulerServiceInterface) *WorkerService {
	ctx, cancel := context.WithCancel(context.Background())

	// Get worker configuration from environment
	maxWorkers := getEnvIntOrDefault("WORKER_POOL_SIZE", 10)
	httpTimeout := getEnvIntOrDefault("WORKER_HTTP_TIMEOUT", 90) // 90 seconds default

	return &WorkerService{
		jobQueue:  jobQueue,
		storage:   storage,
		scheduler: scheduler,
		httpClient: &http.Client{
			Timeout: time.Duration(httpTimeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		workerPool: make(chan struct{}, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the worker service
func (ws *WorkerService) Start() {
	log.Printf("Starting worker service with %d workers", cap(ws.workerPool))

	// Start retry queue processor
	ws.wg.Add(1)
	go ws.processRetryQueue()

	// Start main worker loop
	ws.wg.Add(1)
	go ws.workerLoop()
}

// Stop gracefully stops the worker service
func (ws *WorkerService) Stop() {
	ws.shutdownMu.Lock()
	ws.shutdown = true
	ws.shutdownMu.Unlock()

	log.Println("Stopping worker service...")
	ws.cancel()
	ws.wg.Wait()
	log.Println("Worker service stopped")
}

// IsShutdown checks if the worker service is shutting down
func (ws *WorkerService) IsShutdown() bool {
	ws.shutdownMu.RLock()
	defer ws.shutdownMu.RUnlock()
	return ws.shutdown
}

// workerLoop is the main worker loop that processes jobs from the queue
func (ws *WorkerService) workerLoop() {
	defer ws.wg.Done()

	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
			if ws.IsShutdown() {
				return
			}

			// Try to get a job from the queue
			job, err := ws.jobQueue.DequeueJob(1 * time.Second)
			if err != nil {
				log.Printf("Error dequeuing job: %v", err)
				continue
			}

			if job == nil {
				// No job available, continue
				continue
			}

			// Acquire a worker slot
			select {
			case ws.workerPool <- struct{}{}:
				// Got a worker slot, process the job
				ws.wg.Add(1)
				go ws.processJob(job)
			case <-ws.ctx.Done():
				// Context cancelled, put job back in queue if possible
				ws.jobQueue.EnqueueJob(job)
				return
			}
		}
	}
}

// processJob processes a single job
func (ws *WorkerService) processJob(job *models.QueueJob) {
	defer ws.wg.Done()
	defer func() { <-ws.workerPool }() // Release worker slot

	log.Printf("Processing job %s (JobID: %d, attempt %d/%d)",
		job.ID, job.JobID, job.RetryCount+1, job.MaxRetryCount+1)

	// Check if there's already an execution in progress for this job
	existingExecution, err := ws.storage.GetJobExecutionInProgress(job.JobID)
	if err != nil {
		log.Printf("Failed to check for existing execution for job %s: %v", job.ID, err)
		ws.jobQueue.FailJob(job, fmt.Sprintf("Failed to check for existing execution: %v", err))
		return
	}

	if existingExecution != nil {
		log.Printf("Job %s (JobID: %d) already has an execution in progress, skipping", job.ID, job.JobID)
		// Remove from processing queue since we're not processing it
		if err := ws.jobQueue.client.SRem(ws.jobQueue.ctx, "job_queue:processing", job.ID).Err(); err != nil {
			log.Printf("Warning: failed to remove job %s from processing queue: %v", job.ID, err)
		}
		return
	}

	// Create job execution record
	execution := &models.JobExecution{
		JobID:         job.JobID,
		Status:        models.StatusScheduled,
		ExecutionTime: time.Now(),
		RetryCount:    job.RetryCount,
	}

	if err := ws.storage.CreateJobExecution(execution); err != nil {
		log.Printf("Failed to create execution record for job %s: %v", job.ID, err)
		ws.jobQueue.FailJob(job, fmt.Sprintf("Failed to create execution record: %v", err))
		return
	}

	// Update execution status to running
	execution.Status = models.StatusRunning
	if err := ws.storage.UpdateJobExecution(execution); err != nil {
		log.Printf("Failed to update execution status to running for job %s: %v", job.ID, err)
	}

	// Execute the job
	startTime := time.Now()
	success := ws.callJobAPI(job.API)
	executionDuration := time.Since(startTime)
	execution.ExecutionDuration = &executionDuration

	// Update execution status based on result
	if success {
		execution.Status = models.StatusSuccess
		log.Printf("Job %s executed successfully (attempt %d)", job.ID, job.RetryCount+1)
	} else {
		execution.Status = models.StatusFailed
		execution.Error = "API call failed"
		log.Printf("Job %s failed (attempt %d/%d)", job.ID, job.RetryCount+1, job.MaxRetryCount+1)
	}

	if err := ws.storage.UpdateJobExecution(execution); err != nil {
		log.Printf("Failed to update execution status for job %s: %v", job.ID, err)
	}

	// Handle job completion or failure
	log.Printf("DEBUG: Job %s execution result: success=%v", job.ID, success)
	if success {
		ws.handleSuccessfulJob(job, execution)
	} else {
		ws.handleFailedJob(job, execution)
	}
}

// callJobAPI makes HTTP call to the job's API endpoint
func (ws *WorkerService) callJobAPI(apiURL string) bool {
	req, err := http.NewRequestWithContext(ws.ctx, "POST", apiURL, nil)
	if err != nil {
		log.Printf("Failed to create request for %s: %v", apiURL, err)
		return false
	}

	resp, err := ws.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to call API %s: %v", apiURL, err)
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx status codes as success
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// handleSuccessfulJob handles a successfully executed job
func (ws *WorkerService) handleSuccessfulJob(job *models.QueueJob, execution *models.JobExecution) {
	result := &models.QueueJobResult{
		JobID:             job.ID,
		Status:            models.QueueStatusCompleted,
		Success:           true,
		ExecutionTime:     execution.ExecutionTime,
		ExecutionDuration: *execution.ExecutionDuration,
		RetryCount:        job.RetryCount,
	}

	if err := ws.jobQueue.CompleteJob(job.ID, result); err != nil {
		log.Printf("Failed to complete job %s: %v", job.ID, err)
	}

	log.Printf("Job %s completed successfully", job.ID)
}

// handleFailedJob handles a failed job execution
func (ws *WorkerService) handleFailedJob(job *models.QueueJob, execution *models.JobExecution) {
	log.Printf("DEBUG: handleFailedJob called for job %s (JobID: %d)", job.ID, job.JobID)
	errorMsg := "API call failed"
	if execution.Error != "" {
		errorMsg = execution.Error
	}

	if err := ws.jobQueue.FailJob(job, errorMsg); err != nil {
		log.Printf("Failed to handle failed job %s: %v", job.ID, err)
	}

	// Notify scheduler about job failure
	log.Printf("Notifying scheduler about job failure %s (JobID: %d)", job.ID, job.JobID)
	if err := ws.scheduler.HandleJobCompletion(job.JobID, false); err != nil {
		log.Printf("Failed to notify scheduler about job failure %s: %v", job.ID, err)
	} else {
		log.Printf("Successfully notified scheduler about job failure %s", job.ID)
	}
}

// processRetryQueue processes jobs that are ready for retry
func (ws *WorkerService) processRetryQueue() {
	defer ws.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds for faster cleanup
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			if ws.IsShutdown() {
				return
			}

			if err := ws.jobQueue.ProcessRetryQueue(); err != nil {
				log.Printf("Error processing retry queue: %v", err)
			}

			// Cleanup stale jobs every 10 seconds
			if err := ws.jobQueue.CleanupStaleJobs(1 * time.Hour); err != nil {
				log.Printf("Error cleaning up stale jobs: %v", err)
			}
		}
	}
}

// GetStats returns worker statistics
func (ws *WorkerService) GetStats() map[string]interface{} {
	queueStats, err := ws.jobQueue.GetQueueStats()
	if err != nil {
		log.Printf("Failed to get queue stats: %v", err)
		queueStats = make(map[string]int64)
	}

	return map[string]interface{}{
		"active_workers": len(ws.workerPool),
		"max_workers":    cap(ws.workerPool),
		"queue_stats":    queueStats,
		"is_shutdown":    ws.IsShutdown(),
	}
}

// getEnvIntOrDefault gets an environment variable as integer or returns a default value
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
