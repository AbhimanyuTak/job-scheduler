package services

import (
	"context"
	"testing"
	"time"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/manyu/job-scheduler/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage for testing scheduler service
type MockSchedulerStorage struct {
	jobs      map[uint]*models.Job
	schedules map[uint]*models.JobSchedule
	nextID    uint
}

func NewMockSchedulerStorage() *MockSchedulerStorage {
	return &MockSchedulerStorage{
		jobs:      make(map[uint]*models.Job),
		schedules: make(map[uint]*models.JobSchedule),
		nextID:    1,
	}
}

func (m *MockSchedulerStorage) CreateJob(job *models.Job) error {
	job.ID = m.nextID
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	m.jobs[m.nextID] = job
	m.nextID++
	return nil
}

func (m *MockSchedulerStorage) GetJob(id uint) (*models.Job, error) {
	job, exists := m.jobs[id]
	if !exists || !job.IsActive {
		return nil, assert.AnError
	}
	return job, nil
}

func (m *MockSchedulerStorage) GetAllJobs() ([]*models.Job, error) {
	var activeJobs []*models.Job
	for _, job := range m.jobs {
		if job.IsActive {
			activeJobs = append(activeJobs, job)
		}
	}
	return activeJobs, nil
}

func (m *MockSchedulerStorage) CreateJobSchedule(schedule *models.JobSchedule) error {
	schedule.ID = m.nextID
	schedule.CreatedAt = time.Now()
	schedule.CreatedAt = time.Now()
	m.schedules[schedule.JobID] = schedule
	m.nextID++
	return nil
}

func (m *MockSchedulerStorage) GetJobSchedule(jobID uint) (*models.JobSchedule, error) {
	schedule, exists := m.schedules[jobID]
	if !exists {
		return nil, assert.AnError
	}
	return schedule, nil
}

func (m *MockSchedulerStorage) UpdateJobSchedule(jobID uint, nextExecutionTime time.Time) error {
	schedule, exists := m.schedules[jobID]
	if !exists {
		return assert.AnError
	}
	schedule.NextExecutionTime = nextExecutionTime
	schedule.CreatedAt = time.Now()
	return nil
}

func (m *MockSchedulerStorage) DeleteJobSchedule(jobID uint) error {
	delete(m.schedules, jobID)
	return nil
}

func (m *MockSchedulerStorage) GetJobsReadyForExecution(limit int) ([]*models.Job, []*models.JobSchedule, error) {
	var readyJobs []*models.Job
	var readySchedules []*models.JobSchedule
	now := time.Now()

	for _, job := range m.jobs {
		if !job.IsActive {
			continue
		}
		schedule, exists := m.schedules[job.ID]
		if !exists {
			continue
		}
		if schedule.NextExecutionTime.Before(now) || schedule.NextExecutionTime.Equal(now) {
			readyJobs = append(readyJobs, job)
			readySchedules = append(readySchedules, schedule)
		}
	}

	return readyJobs, readySchedules, nil
}

func (m *MockSchedulerStorage) CreateJobExecution(execution *models.JobExecution) error {
	execution.ID = m.nextID
	execution.CreatedAt = time.Now()
	execution.UpdatedAt = time.Now()
	m.nextID++
	return nil
}

func (m *MockSchedulerStorage) UpdateJobExecution(execution *models.JobExecution) error {
	execution.UpdatedAt = time.Now()
	return nil
}

func (m *MockSchedulerStorage) GetJobExecutions(jobID uint, limit int) ([]*models.JobExecution, error) {
	return []*models.JobExecution{}, nil
}

func (m *MockSchedulerStorage) GetJobExecutionInProgress(jobID uint) (*models.JobExecution, error) {
	// For testing purposes, always return nil (no execution in progress)
	return nil, nil
}

// MockJobQueue for testing scheduler service
type MockJobQueue struct {
	enqueuedJobs []*models.QueueJob
	stats        map[string]int64
}

func NewMockJobQueue() *MockJobQueue {
	return &MockJobQueue{
		enqueuedJobs: make([]*models.QueueJob, 0),
		stats: map[string]int64{
			"ready":      0,
			"processing": 0,
			"completed":  0,
			"retrying":   0,
		},
	}
}

func (m *MockJobQueue) EnqueueJob(job *models.QueueJob) error {
	m.enqueuedJobs = append(m.enqueuedJobs, job)
	m.stats["ready"]++
	return nil
}

func (m *MockJobQueue) DequeueJob(timeout time.Duration) (*models.QueueJob, error) {
	return nil, nil
}

func (m *MockJobQueue) CompleteJob(jobID string, result *models.QueueJobResult) error {
	m.stats["completed"]++
	return nil
}

func (m *MockJobQueue) GetQueueStats() (map[string]int64, error) {
	return m.stats, nil
}

func (m *MockJobQueue) ProcessRetryQueue() error {
	return nil
}

// MockRedisClient for testing
type MockRedisClient struct{}

func (m *MockRedisClient) GetClient() *redis.Client {
	return nil
}

func (m *MockRedisClient) GetContext() context.Context {
	return context.Background()
}

func (m *MockRedisClient) Close() error {
	return nil
}

func (m *MockRedisClient) Health() error {
	return nil
}

func TestSchedulerService_ProcessReadyJobs_Unit(t *testing.T) {
	mockStorage := NewMockSchedulerStorage()
	mockJobQueue := NewMockJobQueue()
	mockRedisClient := &MockRedisClient{}

	// Create scheduler service with mocks
	scheduler := &SchedulerService{
		storage:        mockStorage,
		jobQueue:       mockJobQueue,
		redisClient:    mockRedisClient,
		scheduleParser: utils.NewScheduleParser(),
	}

	// Create a test job that's ready for execution
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 0 */5 * * *", // 6-field format with seconds
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}
	mockStorage.CreateJob(job)

	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: time.Now().Add(-time.Minute), // Ready for execution
	}
	mockStorage.CreateJobSchedule(schedule)

	// Execute
	ctx := context.Background()
	err := scheduler.ProcessReadyJobs(ctx, 100)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, mockJobQueue.enqueuedJobs, 1)
	assert.Equal(t, job.ID, mockJobQueue.enqueuedJobs[0].JobID)
}

func TestSchedulerService_ProcessReadyJobs_NoJobs_Unit(t *testing.T) {
	mockStorage := NewMockSchedulerStorage()
	mockJobQueue := NewMockJobQueue()
	mockRedisClient := &MockRedisClient{}

	// Create scheduler service with mocks
	scheduler := &SchedulerService{
		storage:        mockStorage,
		jobQueue:       mockJobQueue,
		redisClient:    mockRedisClient,
		scheduleParser: utils.NewScheduleParser(),
	}

	// Execute with no jobs
	ctx := context.Background()
	err := scheduler.ProcessReadyJobs(ctx, 100)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, mockJobQueue.enqueuedJobs, 0)
}

func TestSchedulerService_GetQueueStats_Unit(t *testing.T) {
	mockStorage := NewMockSchedulerStorage()
	mockJobQueue := NewMockJobQueue()
	mockRedisClient := &MockRedisClient{}

	// Create scheduler service with mocks
	scheduler := &SchedulerService{
		storage:        mockStorage,
		jobQueue:       mockJobQueue,
		redisClient:    mockRedisClient,
		scheduleParser: utils.NewScheduleParser(),
	}

	// Execute
	stats, err := scheduler.GetQueueStats()

	// Assertions
	require.NoError(t, err)
	assert.Contains(t, stats, "ready")
	assert.Contains(t, stats, "processing")
	assert.Contains(t, stats, "completed")
	assert.Contains(t, stats, "retrying")
}

func TestSchedulerService_HandleJobCompletion_Unit(t *testing.T) {
	mockStorage := NewMockSchedulerStorage()
	mockJobQueue := NewMockJobQueue()
	mockRedisClient := &MockRedisClient{}

	// Create scheduler service with mocks
	scheduler := &SchedulerService{
		storage:        mockStorage,
		jobQueue:       mockJobQueue,
		redisClient:    mockRedisClient,
		scheduleParser: utils.NewScheduleParser(),
	}

	// Create a test job
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 0 */5 * * *", // 6-field format with seconds
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}
	mockStorage.CreateJob(job)

	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: time.Now().Add(-time.Minute),
	}
	mockStorage.CreateJobSchedule(schedule)

	// Execute
	err := scheduler.HandleJobCompletion(job.ID, true)

	// Assertions
	require.NoError(t, err)

	// Verify schedule was updated
	updatedSchedule, err := mockStorage.GetJobSchedule(job.ID)
	require.NoError(t, err)
	assert.True(t, updatedSchedule.NextExecutionTime.After(time.Now()))
}

func TestSchedulerService_HandleJobCompletion_AT_MOST_ONCE_Unit(t *testing.T) {
	mockStorage := NewMockSchedulerStorage()
	mockJobQueue := NewMockJobQueue()
	mockRedisClient := &MockRedisClient{}

	// Create scheduler service with mocks
	scheduler := &SchedulerService{
		storage:        mockStorage,
		jobQueue:       mockJobQueue,
		redisClient:    mockRedisClient,
		scheduleParser: utils.NewScheduleParser(),
	}

	// Create a test AT_MOST_ONCE job
	job := &models.Job{
		Description:   "Test AT_MOST_ONCE job",
		Schedule:      "0 0 */5 * * *", // 6-field format with seconds
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_MOST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}
	mockStorage.CreateJob(job)

	originalTime := time.Now().Add(-time.Minute)
	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: originalTime,
	}
	mockStorage.CreateJobSchedule(schedule)

	// Test successful execution - should reschedule for recurring AT_MOST_ONCE jobs
	err := scheduler.HandleJobCompletion(job.ID, true)
	assert.NoError(t, err)

	// Verify schedule was updated (not deleted) for successful recurring AT_MOST_ONCE job
	updatedSchedule, err := mockStorage.GetJobSchedule(job.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedSchedule)
	assert.True(t, updatedSchedule.NextExecutionTime.After(originalTime))
}

func TestSchedulerService_HandleJobCompletion_AT_MOST_ONCE_Failure_Unit(t *testing.T) {
	mockStorage := NewMockSchedulerStorage()
	mockJobQueue := NewMockJobQueue()
	mockRedisClient := &MockRedisClient{}

	// Create scheduler service with mocks
	scheduler := &SchedulerService{
		storage:        mockStorage,
		jobQueue:       mockJobQueue,
		redisClient:    mockRedisClient,
		scheduleParser: utils.NewScheduleParser(),
	}

	// Create a test AT_MOST_ONCE job
	job := &models.Job{
		Description:   "Test AT_MOST_ONCE job",
		Schedule:      "0 0 */5 * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_MOST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}
	mockStorage.CreateJob(job)

	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: time.Now().Add(-time.Minute),
	}
	mockStorage.CreateJobSchedule(schedule)

	// Test failed execution - should reschedule for recurring AT_MOST_ONCE jobs
	err := scheduler.HandleJobCompletion(job.ID, false)
	assert.NoError(t, err)

	// Verify schedule was updated (rescheduled) for failed recurring AT_MOST_ONCE job
	updatedSchedule, err := mockStorage.GetJobSchedule(job.ID)
	assert.NoError(t, err) // Should not return error because schedule was rescheduled
	assert.NotNil(t, updatedSchedule)
	// The next execution time should be in the future
	assert.True(t, updatedSchedule.NextExecutionTime.After(time.Now()))
}

func TestSchedulerService_HandleJobCompletion_AT_MOST_ONCE_NonRecurring_Unit(t *testing.T) {
	mockStorage := NewMockSchedulerStorage()
	mockJobQueue := NewMockJobQueue()
	mockRedisClient := &MockRedisClient{}

	// Create scheduler service with mocks
	scheduler := &SchedulerService{
		storage:        mockStorage,
		jobQueue:       mockJobQueue,
		redisClient:    mockRedisClient,
		scheduleParser: utils.NewScheduleParser(),
	}

	// Create a test non-recurring AT_MOST_ONCE job
	job := &models.Job{
		Description:   "Test non-recurring AT_MOST_ONCE job",
		Schedule:      "0 0 */5 * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_MOST_ONCE,
		IsRecurring:   false, // Non-recurring
		MaxRetryCount: 3,
		IsActive:      true,
	}
	mockStorage.CreateJob(job)

	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: time.Now().Add(-time.Minute),
	}
	mockStorage.CreateJobSchedule(schedule)

	// Test successful execution - should delete schedule for non-recurring jobs
	err := scheduler.HandleJobCompletion(job.ID, true)
	assert.NoError(t, err)

	// Verify schedule was deleted for successful non-recurring job
	_, err = mockStorage.GetJobSchedule(job.ID)
	assert.Error(t, err) // Should return error because schedule was deleted

	// Recreate schedule for failure test
	mockStorage.CreateJobSchedule(schedule)

	// Test failed execution - should also delete schedule for non-recurring jobs
	err = scheduler.HandleJobCompletion(job.ID, false)
	assert.NoError(t, err)

	// Verify schedule was deleted for failed non-recurring job
	_, err = mockStorage.GetJobSchedule(job.ID)
	assert.Error(t, err) // Should return error because schedule was deleted
}

func TestQueueJob_AT_MOST_ONCE_ShouldNotRetry(t *testing.T) {
	// Test that AT_MOST_ONCE jobs should not be retried
	job := &models.QueueJob{
		ID:            "test-job-1",
		JobID:         1,
		API:           "https://httpbin.org/status/200",
		MaxRetryCount: 3,
		RetryCount:    0,
		Type:          models.AT_MOST_ONCE,
		IsRecurring:   true,
		Schedule:      "0 0 */5 * * *",
		CreatedAt:     time.Now(),
		ScheduledAt:   time.Now(),
		Timeout:       90,
	}

	// AT_MOST_ONCE jobs should not retry even if retry count is below max
	assert.False(t, job.ShouldRetry())

	// Even with retry count > 0, AT_MOST_ONCE jobs should not retry
	job.RetryCount = 1
	assert.False(t, job.ShouldRetry())

	// Even with retry count at max, AT_MOST_ONCE jobs should not retry
	job.RetryCount = job.MaxRetryCount
	assert.False(t, job.ShouldRetry())
}

func TestQueueJob_AT_LEAST_ONCE_ShouldRetry(t *testing.T) {
	// Test that AT_LEAST_ONCE jobs should be retried when appropriate
	job := &models.QueueJob{
		ID:            "test-job-2",
		JobID:         2,
		API:           "https://httpbin.org/status/200",
		MaxRetryCount: 3,
		RetryCount:    0,
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		Schedule:      "0 0 */5 * * *",
		CreatedAt:     time.Now(),
		ScheduledAt:   time.Now(),
		Timeout:       90,
	}

	// AT_LEAST_ONCE jobs should retry when retry count is below max
	assert.True(t, job.ShouldRetry())

	// AT_LEAST_ONCE jobs should retry when retry count is below max
	job.RetryCount = 1
	assert.True(t, job.ShouldRetry())

	// AT_LEAST_ONCE jobs should not retry when retry count reaches max
	job.RetryCount = job.MaxRetryCount
	assert.False(t, job.ShouldRetry())
}
