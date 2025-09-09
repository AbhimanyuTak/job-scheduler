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
