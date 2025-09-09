package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manyu/job-scheduler/internal/models"
	"github.com/manyu/job-scheduler/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage is a simple mock for testing handlers
type MockStorage struct {
	jobs       map[uint]*models.Job
	schedules  map[uint]*models.JobSchedule
	executions map[uint][]*models.JobExecution
	nextID     uint
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		jobs:       make(map[uint]*models.Job),
		schedules:  make(map[uint]*models.JobSchedule),
		executions: make(map[uint][]*models.JobExecution),
		nextID:     1,
	}
}

func (m *MockStorage) CreateJob(job *models.Job) error {
	job.ID = m.nextID
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	m.jobs[m.nextID] = job
	m.nextID++
	return nil
}

func (m *MockStorage) GetJob(id uint) (*models.Job, error) {
	job, exists := m.jobs[id]
	if !exists || !job.IsActive {
		return nil, storage.ErrJobNotFound
	}
	return job, nil
}

func (m *MockStorage) GetAllJobs() ([]*models.Job, error) {
	var activeJobs []*models.Job
	for _, job := range m.jobs {
		if job.IsActive {
			activeJobs = append(activeJobs, job)
		}
	}
	return activeJobs, nil
}

func (m *MockStorage) CreateJobSchedule(schedule *models.JobSchedule) error {
	schedule.ID = m.nextID
	schedule.CreatedAt = time.Now()
	schedule.CreatedAt = time.Now()
	m.schedules[schedule.JobID] = schedule
	m.nextID++
	return nil
}

func (m *MockStorage) GetJobSchedule(jobID uint) (*models.JobSchedule, error) {
	schedule, exists := m.schedules[jobID]
	if !exists {
		return nil, assert.AnError
	}
	return schedule, nil
}

func (m *MockStorage) UpdateJobSchedule(jobID uint, nextExecutionTime time.Time) error {
	schedule, exists := m.schedules[jobID]
	if !exists {
		return assert.AnError
	}
	schedule.NextExecutionTime = nextExecutionTime
	schedule.CreatedAt = time.Now()
	return nil
}

func (m *MockStorage) DeleteJobSchedule(jobID uint) error {
	delete(m.schedules, jobID)
	return nil
}

func (m *MockStorage) GetJobsReadyForExecution(limit int) ([]*models.Job, []*models.JobSchedule, error) {
	// Simple implementation for testing
	return []*models.Job{}, []*models.JobSchedule{}, nil
}

func (m *MockStorage) CreateJobExecution(execution *models.JobExecution) error {
	execution.ID = m.nextID
	execution.CreatedAt = time.Now()
	execution.UpdatedAt = time.Now()
	m.executions[execution.JobID] = append(m.executions[execution.JobID], execution)
	m.nextID++
	return nil
}

func (m *MockStorage) UpdateJobExecution(execution *models.JobExecution) error {
	execution.UpdatedAt = time.Now()
	return nil
}

func (m *MockStorage) GetJobExecutions(jobID uint, limit int) ([]*models.JobExecution, error) {
	executions, exists := m.executions[jobID]
	if !exists {
		return []*models.JobExecution{}, nil
	}
	return executions, nil
}

func (m *MockStorage) GetJobExecutionInProgress(jobID uint) (*models.JobExecution, error) {
	// For testing purposes, always return nil (no execution in progress)
	return nil, nil
}

func TestJobHandler_CreateJob_Unit(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewJobHandler(mockStorage)

	// Test data
	jobData := map[string]interface{}{
		"description":   "Test job",
		"schedule":      "0 */5 * * * *",
		"api":           "https://httpbin.org/status/200",
		"type":          "AT_LEAST_ONCE",
		"isRecurring":   true,
		"maxRetryCount": 3,
	}

	jsonData, err := json.Marshal(jobData)
	require.NoError(t, err)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/jobs", handler.CreateJob)

	// Create request
	req, err := http.NewRequest("POST", "/jobs", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "id")
	assert.Equal(t, "Job created successfully", response["message"])
}

func TestJobHandler_GetJob_Unit(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewJobHandler(mockStorage)

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
	mockStorage.CreateJob(job)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jobs/:id", handler.GetJob)

	// Create request
	req, err := http.NewRequest("GET", "/jobs/1", nil)
	require.NoError(t, err)
	req.Header.Set("X-API-Key", "test-api-key")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Job
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, job.ID, response.ID)
	assert.Equal(t, job.Description, response.Description)
	assert.Equal(t, job.Schedule, response.Schedule)
	assert.Equal(t, job.API, response.API)
	assert.Equal(t, job.Type, response.Type)
}

func TestJobHandler_GetJob_NotFound_Unit(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewJobHandler(mockStorage)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jobs/:id", handler.GetJob)

	// Create request for non-existent job
	req, err := http.NewRequest("GET", "/jobs/999", nil)
	require.NoError(t, err)
	req.Header.Set("X-API-Key", "test-api-key")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "Job not found", response["error"])
}

func TestJobHandler_ListJobs_Unit(t *testing.T) {
	mockStorage := NewMockStorage()
	handler := NewJobHandler(mockStorage)

	// Create test jobs
	job1 := &models.Job{
		Description:   "Job 1",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}
	job2 := &models.Job{
		Description:   "Job 2",
		Schedule:      "0 */10 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_MOST_ONCE,
		IsRecurring:   false,
		MaxRetryCount: 1,
		IsActive:      true,
	}
	mockStorage.CreateJob(job1)
	mockStorage.CreateJob(job2)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/jobs", handler.ListJobs)

	// Create request
	req, err := http.NewRequest("GET", "/jobs", nil)
	require.NoError(t, err)
	req.Header.Set("X-API-Key", "test-api-key")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "jobs")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "limit")
	assert.Contains(t, response, "offset")

	jobs := response["jobs"].([]interface{})
	assert.Len(t, jobs, 2)
	assert.Equal(t, float64(2), response["total"])
}
