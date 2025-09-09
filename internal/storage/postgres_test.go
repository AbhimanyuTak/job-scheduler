package storage

import (
	"testing"
	"time"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Use SQLite for unit tests to avoid requiring PostgreSQL
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Job{}, &models.JobSchedule{}, &models.JobExecution{})
	require.NoError(t, err)

	return db
}

func TestNewPostgresStorage(t *testing.T) {
	// This test requires a running database, so we'll skip it in unit tests
	// and only run it in integration tests
	t.Skip("Requires running database - run in integration tests")
}

func TestPostgresStorage_CreateJob(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)
	assert.NotZero(t, job.ID)
	assert.NotZero(t, job.CreatedAt)
	assert.NotZero(t, job.UpdatedAt)
}

func TestPostgresStorage_GetJob(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a test job
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	// Get the job
	retrievedJob, err := storage.GetJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, job.Description, retrievedJob.Description)
	assert.Equal(t, job.Schedule, retrievedJob.Schedule)
	assert.Equal(t, job.API, retrievedJob.API)
	assert.Equal(t, job.Type, retrievedJob.Type)
	assert.Equal(t, job.IsRecurring, retrievedJob.IsRecurring)
	assert.Equal(t, job.MaxRetryCount, retrievedJob.MaxRetryCount)
	assert.Equal(t, job.IsActive, retrievedJob.IsActive)
}

func TestPostgresStorage_GetJob_NotFound(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Try to get a non-existent job
	_, err := storage.GetJob(999)
	assert.Error(t, err)
	assert.Equal(t, ErrJobNotFound, err)
}

func TestPostgresStorage_GetJob_Inactive(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create an inactive job
	job := &models.Job{
		Description:   "Inactive job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      false,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	// Try to get the inactive job
	_, err = storage.GetJob(job.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrJobNotFound, err)
}

func TestPostgresStorage_GetAllJobs(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create multiple jobs
	jobs := []*models.Job{
		{
			Description:   "Job 1",
			Schedule:      "0 */5 * * * *",
			API:           "https://httpbin.org/status/200",
			Type:          models.AT_LEAST_ONCE,
			IsRecurring:   true,
			MaxRetryCount: 3,
			IsActive:      true,
		},
		{
			Description:   "Job 2",
			Schedule:      "0 */10 * * * *",
			API:           "https://httpbin.org/status/200",
			Type:          models.AT_MOST_ONCE,
			IsRecurring:   false,
			MaxRetryCount: 1,
			IsActive:      true,
		},
		{
			Description:   "Inactive Job",
			Schedule:      "0 */15 * * * *",
			API:           "https://httpbin.org/status/200",
			Type:          models.AT_LEAST_ONCE,
			IsRecurring:   true,
			MaxRetryCount: 3,
			IsActive:      false,
		},
	}

	for _, job := range jobs {
		err := storage.CreateJob(job)
		require.NoError(t, err)
	}

	// Get all jobs
	allJobs, err := storage.GetAllJobs()
	require.NoError(t, err)
	assert.Len(t, allJobs, 2) // Only active jobs

	// Verify the jobs
	jobDescriptions := make([]string, len(allJobs))
	for i, job := range allJobs {
		jobDescriptions[i] = job.Description
	}
	assert.Contains(t, jobDescriptions, "Job 1")
	assert.Contains(t, jobDescriptions, "Job 2")
	assert.NotContains(t, jobDescriptions, "Inactive Job")
}

func TestPostgresStorage_CreateJobSchedule(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a job first
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	// Create a schedule
	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: time.Now().Add(time.Hour),
	}

	err = storage.CreateJobSchedule(schedule)
	require.NoError(t, err)
	assert.NotZero(t, schedule.ID)
	assert.NotZero(t, schedule.CreatedAt)
}

func TestPostgresStorage_GetJobSchedule(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a job and schedule
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	nextExecutionTime := time.Now().Add(time.Hour)
	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: nextExecutionTime,
	}

	err = storage.CreateJobSchedule(schedule)
	require.NoError(t, err)

	// Get the schedule
	retrievedSchedule, err := storage.GetJobSchedule(job.ID)
	require.NoError(t, err)
	assert.Equal(t, schedule.ID, retrievedSchedule.ID)
	assert.Equal(t, job.ID, retrievedSchedule.JobID)
	assert.WithinDuration(t, nextExecutionTime, retrievedSchedule.NextExecutionTime, time.Second)
}

func TestPostgresStorage_GetJobSchedule_NotFound(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Try to get a non-existent schedule
	_, err := storage.GetJobSchedule(999)
	assert.Error(t, err)
	assert.Equal(t, ErrJobScheduleNotFound, err)
}

func TestPostgresStorage_UpdateJobSchedule(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a job and schedule
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	nextExecutionTime := time.Now().Add(time.Hour)
	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: nextExecutionTime,
	}

	err = storage.CreateJobSchedule(schedule)
	require.NoError(t, err)

	// Update the schedule
	newExecutionTime := time.Now().Add(2 * time.Hour)
	err = storage.UpdateJobSchedule(job.ID, newExecutionTime)
	require.NoError(t, err)

	// Verify the update
	updatedSchedule, err := storage.GetJobSchedule(job.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, newExecutionTime, updatedSchedule.NextExecutionTime, time.Second)
}

func TestPostgresStorage_UpdateJobSchedule_NotFound(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Try to update a non-existent schedule
	err := storage.UpdateJobSchedule(999, time.Now().Add(time.Hour))
	assert.Error(t, err)
	assert.Equal(t, ErrJobScheduleNotFound, err)
}

func TestPostgresStorage_DeleteJobSchedule(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a job and schedule
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: time.Now().Add(time.Hour),
	}

	err = storage.CreateJobSchedule(schedule)
	require.NoError(t, err)

	// Delete the schedule
	err = storage.DeleteJobSchedule(job.ID)
	require.NoError(t, err)

	// Verify the schedule was deleted
	_, err = storage.GetJobSchedule(job.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrJobScheduleNotFound, err)
}

func TestPostgresStorage_DeleteJobSchedule_NotFound(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Try to delete a non-existent schedule
	err := storage.DeleteJobSchedule(999)
	assert.Error(t, err)
	assert.Equal(t, ErrJobScheduleNotFound, err)
}

func TestPostgresStorage_GetJobsReadyForExecution(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create jobs with different execution times
	now := time.Now()

	// Job 1: Ready for execution
	job1 := &models.Job{
		Description:   "Ready job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}
	err := storage.CreateJob(job1)
	require.NoError(t, err)

	schedule1 := &models.JobSchedule{
		JobID:             job1.ID,
		NextExecutionTime: now.Add(-time.Minute), // Past time
	}
	err = storage.CreateJobSchedule(schedule1)
	require.NoError(t, err)

	// Job 2: Not ready for execution
	job2 := &models.Job{
		Description:   "Future job",
		Schedule:      "0 */10 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_MOST_ONCE,
		IsRecurring:   false,
		MaxRetryCount: 1,
		IsActive:      true,
	}
	err = storage.CreateJob(job2)
	require.NoError(t, err)

	schedule2 := &models.JobSchedule{
		JobID:             job2.ID,
		NextExecutionTime: now.Add(time.Hour), // Future time
	}
	err = storage.CreateJobSchedule(schedule2)
	require.NoError(t, err)

	// Job 3: Inactive job
	job3 := &models.Job{
		Description:   "Inactive job",
		Schedule:      "0 */15 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      false,
	}
	err = storage.CreateJob(job3)
	require.NoError(t, err)

	schedule3 := &models.JobSchedule{
		JobID:             job3.ID,
		NextExecutionTime: now.Add(-time.Minute), // Past time but inactive
	}
	err = storage.CreateJobSchedule(schedule3)
	require.NoError(t, err)

	// Get jobs ready for execution
	jobs, schedules, err := storage.GetJobsReadyForExecution(10)
	require.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Len(t, schedules, 1)
	assert.Equal(t, job1.ID, jobs[0].ID)
	assert.Equal(t, schedule1.ID, schedules[0].ID)
}

func TestPostgresStorage_CreateJobExecution(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a job first
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	// Create an execution
	execution := &models.JobExecution{
		JobID:             job.ID,
		Status:            models.StatusSuccess,
		ExecutionTime:     time.Now(),
		ExecutionDuration: func() *time.Duration { d := 1500 * time.Millisecond; return &d }(),
	}

	err = storage.CreateJobExecution(execution)
	require.NoError(t, err)
	assert.NotZero(t, execution.ID)
	assert.NotZero(t, execution.CreatedAt)
	assert.NotZero(t, execution.UpdatedAt)
}

func TestPostgresStorage_UpdateJobExecution(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a job and execution
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	execution := &models.JobExecution{
		JobID:             job.ID,
		Status:            models.StatusSuccess,
		ExecutionTime:     time.Now(),
		ExecutionDuration: func() *time.Duration { d := 1500 * time.Millisecond; return &d }(),
	}

	err = storage.CreateJobExecution(execution)
	require.NoError(t, err)

	// Update the execution
	execution.Status = models.StatusFailed
	execution.Error = "API call failed"
	execution.ExecutionDuration = func() *time.Duration { d := 2000 * time.Millisecond; return &d }()

	err = storage.UpdateJobExecution(execution)
	require.NoError(t, err)

	// Verify the update
	var updatedExecution models.JobExecution
	err = db.First(&updatedExecution, execution.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.StatusFailed, updatedExecution.Status)
	assert.Equal(t, "API call failed", updatedExecution.Error)
	assert.Equal(t, 2000*time.Millisecond, *updatedExecution.ExecutionDuration)
}

func TestPostgresStorage_GetJobExecutions(t *testing.T) {
	db := setupTestDB(t)
	storage := &PostgresStorage{db: db}

	// Create a job
	job := &models.Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	err := storage.CreateJob(job)
	require.NoError(t, err)

	// Create multiple executions
	executions := []*models.JobExecution{
		{
			JobID:             job.ID,
			Status:            models.StatusSuccess,
			ExecutionTime:     time.Now().Add(-2 * time.Hour),
			ExecutionDuration: func() *time.Duration { d := 1500 * time.Millisecond; return &d }(),
		},
		{
			JobID:             job.ID,
			Status:            models.StatusFailed,
			Error:             "API call failed",
			ExecutionTime:     time.Now().Add(-time.Hour),
			ExecutionDuration: func() *time.Duration { d := 2000 * time.Millisecond; return &d }(),
		},
		{
			JobID:             job.ID,
			Status:            models.StatusSuccess,
			ExecutionTime:     time.Now(),
			ExecutionDuration: func() *time.Duration { d := 1000 * time.Millisecond; return &d }(),
		},
	}

	for _, execution := range executions {
		err = storage.CreateJobExecution(execution)
		require.NoError(t, err)
	}

	// Get all executions
	allExecutions, err := storage.GetJobExecutions(job.ID, 0)
	require.NoError(t, err)
	assert.Len(t, allExecutions, 3)

	// Verify they are ordered by created_at DESC (newest first)
	assert.Equal(t, executions[2].ID, allExecutions[0].ID)
	assert.Equal(t, executions[1].ID, allExecutions[1].ID)
	assert.Equal(t, executions[0].ID, allExecutions[2].ID)

	// Get limited executions
	limitedExecutions, err := storage.GetJobExecutions(job.ID, 2)
	require.NoError(t, err)
	assert.Len(t, limitedExecutions, 2)
	assert.Equal(t, executions[2].ID, limitedExecutions[0].ID)
	assert.Equal(t, executions[1].ID, limitedExecutions[1].ID)
}
