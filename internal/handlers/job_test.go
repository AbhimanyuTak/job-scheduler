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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreateJob(job *models.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockStorage) CreateJobWithSchedule(job *models.Job, schedule *models.JobSchedule) error {
	args := m.Called(job, schedule)
	return args.Error(0)
}

func (m *MockStorage) GetJob(id uint) (*models.Job, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockStorage) GetAllJobs() ([]*models.Job, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Job), args.Error(1)
}

func (m *MockStorage) CreateJobSchedule(schedule *models.JobSchedule) error {
	args := m.Called(schedule)
	return args.Error(0)
}

func (m *MockStorage) GetJobSchedule(jobID uint) (*models.JobSchedule, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobSchedule), args.Error(1)
}

func (m *MockStorage) UpdateJobSchedule(jobID uint, nextExecutionTime time.Time) error {
	args := m.Called(jobID, nextExecutionTime)
	return args.Error(0)
}

func (m *MockStorage) CreateJobExecution(execution *models.JobExecution) error {
	args := m.Called(execution)
	return args.Error(0)
}

func (m *MockStorage) UpdateJobExecution(execution *models.JobExecution) error {
	args := m.Called(execution)
	return args.Error(0)
}

func (m *MockStorage) GetJobExecutions(jobID uint, limit int) ([]*models.JobExecution, error) {
	args := m.Called(jobID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JobExecution), args.Error(1)
}

func (m *MockStorage) DeleteJobSchedule(jobID uint) error {
	args := m.Called(jobID)
	return args.Error(0)
}

func (m *MockStorage) GetJobsReadyForExecution(limit int) ([]*models.Job, []*models.JobSchedule, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*models.Job), args.Get(1).([]*models.JobSchedule), args.Error(2)
}

func (m *MockStorage) GetJobExecutionInProgress(jobID uint) (*models.JobExecution, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobExecution), args.Error(1)
}

func TestJobHandler_CreateJob_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockStorage := new(MockStorage)
	handler := NewJobHandler(mockStorage)

	// Mock expectations
	mockStorage.On("CreateJobWithSchedule", mock.AnythingOfType("*models.Job"), mock.AnythingOfType("*models.JobSchedule")).Return(nil)

	// Test data
	reqBody := CreateJobRequest{
		API:           "http://example.com/webhook",
		Type:          models.AT_LEAST_ONCE,
		Schedule:      "0 */5 * * * *", // Every 5 minutes
		IsRecurring:   true,
		Description:   "Test job description",
		MaxRetryCount: 3,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.CreateJob(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreateJobResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Job created successfully", response.Message)
	// Note: ID is set by database, so we can't predict it in unit tests

	mockStorage.AssertExpectations(t)
}

func TestJobHandler_CreateJob_InvalidJobType(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockStorage := new(MockStorage)
	handler := NewJobHandler(mockStorage)

	// Test data with invalid job type
	reqBody := CreateJobRequest{
		API:           "http://example.com/webhook",
		Type:          "INVALID_TYPE",
		Schedule:      "0 */5 * * * *",
		IsRecurring:   true,
		Description:   "Test job description",
		MaxRetryCount: 3,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.CreateJob(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_JOB_TYPE", response["code"])
	assert.Equal(t, "Invalid job type. Must be AT_LEAST_ONCE or AT_MOST_ONCE", response["error"])
}

func TestJobHandler_CreateJob_InvalidSchedule(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockStorage := new(MockStorage)
	handler := NewJobHandler(mockStorage)

	// Test data with invalid schedule
	reqBody := CreateJobRequest{
		API:           "http://example.com/webhook",
		Type:          models.AT_LEAST_ONCE,
		Schedule:      "invalid cron expression",
		IsRecurring:   true,
		Description:   "Test job description",
		MaxRetryCount: 3,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.CreateJob(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_SCHEDULE", response["code"])
}

func TestJobHandler_GetJob_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockStorage := new(MockStorage)
	handler := NewJobHandler(mockStorage)

	// Mock data
	expectedJob := &models.Job{
		ID:            1,
		API:           "http://example.com/webhook",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		Description:   "Test job description",
		MaxRetryCount: 3,
		IsActive:      true,
	}

	mockStorage.On("GetJob", uint(1)).Return(expectedJob, nil)

	req, _ := http.NewRequest("GET", "/api/v1/jobs/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Execute
	handler.GetJob(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Job
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedJob.ID, response.ID)
	assert.Equal(t, expectedJob.API, response.API)

	mockStorage.AssertExpectations(t)
}

func TestJobHandler_GetJob_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockStorage := new(MockStorage)
	handler := NewJobHandler(mockStorage)

	// Mock storage to return not found error
	mockStorage.On("GetJob", uint(999)).Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/api/v1/jobs/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	// Execute
	handler.GetJob(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockStorage.AssertExpectations(t)
}
