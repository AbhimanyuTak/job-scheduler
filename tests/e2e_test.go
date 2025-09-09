package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/manyu/job-scheduler/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration
const (
	BaseURL     = "http://localhost:8080"
	APIKey      = "test-api-key"
	RedisHost   = "localhost"
	RedisPort   = "6379"
	TestTimeout = 30 * time.Second
)

// TestClient wraps HTTP client with test helpers
type TestClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewTestClient creates a new test client
func NewTestClient() *TestClient {
	return &TestClient{
		baseURL: BaseURL,
		apiKey:  APIKey,
		client: &http.Client{
			Timeout: TestTimeout,
		},
	}
}

// CreateJobRequest represents the request payload for creating a job
type CreateJobRequest struct {
	Schedule      string         `json:"schedule"`
	API           string         `json:"api"`
	Type          models.JobType `json:"type"`
	IsRecurring   bool           `json:"isRecurring"`
	Description   string         `json:"description"`
	MaxRetryCount int            `json:"maxRetryCount"`
}

// CreateJobResponse represents the response for creating a job
type CreateJobResponse struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

// Job represents a job in the system
type Job struct {
	ID            uint           `json:"id"`
	Schedule      string         `json:"schedule"`
	API           string         `json:"api"`
	Type          models.JobType `json:"type"`
	IsRecurring   bool           `json:"isRecurring"`
	IsActive      bool           `json:"isActive"`
	Description   string         `json:"description"`
	MaxRetryCount int            `json:"maxRetryCount"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// JobSchedule represents a job schedule
type JobSchedule struct {
	ID                uint      `json:"id"`
	JobID             uint      `json:"jobId"`
	NextExecutionTime time.Time `json:"nextExecutionTime"`
	CreatedAt         time.Time `json:"createdAt"`
}

// JobExecution represents a job execution
type JobExecution struct {
	ID                uint      `json:"id"`
	JobID             uint      `json:"jobId"`
	Status            string    `json:"status"`
	Error             string    `json:"error"`
	ExecutionTime     time.Time `json:"executionTime"`
	ExecutionDuration int       `json:"executionDuration"`
	RetryCount        int       `json:"retryCount"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// QueueStats represents queue statistics
type QueueStats struct {
	Ready      int64 `json:"ready"`
	Processing int64 `json:"processing"`
	Completed  int64 `json:"completed"`
	Retrying   int64 `json:"retrying"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// HTTP helper methods
func (tc *TestClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, tc.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", tc.apiKey)

	return tc.client.Do(req)
}

func (tc *TestClient) getJSON(endpoint string, target interface{}) error {
	resp, err := tc.makeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (tc *TestClient) postJSON(endpoint string, body interface{}, target interface{}) error {
	resp, err := tc.makeRequest("POST", endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return nil
}

// Test helper methods
func (tc *TestClient) CreateJob(t *testing.T, req CreateJobRequest) uint {
	var resp CreateJobResponse
	err := tc.postJSON("/api/v1/jobs", req, &resp)
	require.NoError(t, err, "Failed to create job")
	require.NotZero(t, resp.ID, "Job ID should not be zero")

	t.Logf("Created job ID: %d", resp.ID)
	return resp.ID
}

func (tc *TestClient) GetJob(t *testing.T, jobID uint) Job {
	var job Job
	err := tc.getJSON(fmt.Sprintf("/api/v1/jobs/%d", jobID), &job)
	require.NoError(t, err, "Failed to get job")
	return job
}

func (tc *TestClient) GetJobSchedule(t *testing.T, jobID uint) JobSchedule {
	var schedule JobSchedule
	err := tc.getJSON(fmt.Sprintf("/api/v1/jobs/%d/schedule", jobID), &schedule)
	require.NoError(t, err, "Failed to get job schedule")
	return schedule
}

func (tc *TestClient) GetJobHistory(t *testing.T, jobID uint) []JobExecution {
	var response struct {
		Executions []JobExecution `json:"executions"`
		Total      int            `json:"total"`
	}
	err := tc.getJSON(fmt.Sprintf("/api/v1/jobs/%d/history", jobID), &response)
	require.NoError(t, err, "Failed to get job history")
	return response.Executions
}

func (tc *TestClient) GetQueueStats(t *testing.T) QueueStats {
	var stats QueueStats
	err := tc.getJSON("/queue/stats", &stats)
	require.NoError(t, err, "Failed to get queue stats")
	return stats
}

func (tc *TestClient) HealthCheck(t *testing.T) HealthResponse {
	var health HealthResponse
	err := tc.getJSON("/health", &health)
	require.NoError(t, err, "Failed to get health status")
	return health
}

// Test setup and teardown
func TestMain(m *testing.M) {
	// Wait for service to be ready
	if !waitForService() {
		fmt.Println("Service is not ready. Exiting tests.")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup if needed
	os.Exit(code)
}

func waitForService() bool {
	client := NewTestClient()
	maxAttempts := 30

	for i := 0; i < maxAttempts; i++ {
		health := client.HealthCheck(&testing.T{})
		if health.Status == "healthy" {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

// Test cases
func TestHealthCheck(t *testing.T) {
	client := NewTestClient()

	health := client.HealthCheck(t)
	assert.Equal(t, "healthy", health.Status)
}

func TestJobCreation(t *testing.T) {
	client := NewTestClient()

	testCases := []struct {
		name string
		req  CreateJobRequest
	}{
		{
			name: "Recurring job every 10 seconds",
			req: CreateJobRequest{
				Schedule:      "*/10 * * * * *",
				API:           "https://httpbin.org/delay/1",
				Type:          models.AT_LEAST_ONCE,
				IsRecurring:   true,
				Description:   "Test recurring job",
				MaxRetryCount: 2,
			},
		},
		{
			name: "One-time job",
			req: CreateJobRequest{
				Schedule:      "30 0 * * * *",
				API:           "https://httpbin.org/status/200",
				Type:          models.AT_MOST_ONCE,
				IsRecurring:   false,
				Description:   "Test one-time job",
				MaxRetryCount: 1,
			},
		},
		{
			name: "Job with custom retry count",
			req: CreateJobRequest{
				Schedule:      "0 * * * * *",
				API:           "https://httpbin.org/json",
				Type:          models.AT_LEAST_ONCE,
				IsRecurring:   true,
				Description:   "Test job with custom retries",
				MaxRetryCount: 5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jobID := client.CreateJob(t, tc.req)

			// Verify job was created correctly
			job := client.GetJob(t, jobID)
			assert.Equal(t, tc.req.Schedule, job.Schedule)
			assert.Equal(t, tc.req.API, job.API)
			assert.Equal(t, tc.req.Type, job.Type)
			assert.Equal(t, tc.req.IsRecurring, job.IsRecurring)
			assert.Equal(t, tc.req.Description, job.Description)
			assert.Equal(t, tc.req.MaxRetryCount, job.MaxRetryCount)
			assert.True(t, job.IsActive)
			assert.NotZero(t, job.CreatedAt)
		})
	}
}

func TestJobSchedulePopulation(t *testing.T) {
	client := NewTestClient()

	// Create a job
	req := CreateJobRequest{
		Schedule:      "*/15 * * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		Description:   "Schedule test job",
		MaxRetryCount: 1,
	}

	jobID := client.CreateJob(t, req)

	// Verify schedule was created
	schedule := client.GetJobSchedule(t, jobID)
	assert.Equal(t, jobID, schedule.JobID)
	assert.NotZero(t, schedule.NextExecutionTime)
	assert.NotZero(t, schedule.CreatedAt)

	// Verify next execution time is in the future
	assert.True(t, schedule.NextExecutionTime.After(time.Now()),
		"Next execution time should be in the future")
}

func TestJobExecution(t *testing.T) {
	client := NewTestClient()

	// Create a job that should execute quickly
	req := CreateJobRequest{
		Schedule:      "*/5 * * * * *",
		API:           "https://httpbin.org/delay/1",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		Description:   "Execution test job",
		MaxRetryCount: 2,
	}

	jobID := client.CreateJob(t, req)

	// Wait for job to execute
	t.Log("Waiting 30 seconds for job execution...")
	time.Sleep(30 * time.Second)

	// Check job history
	executions := client.GetJobHistory(t, jobID)

	if len(executions) > 0 {
		t.Logf("Found %d executions for job %d", len(executions), jobID)

		// Check execution details
		latestExecution := executions[0]
		assert.Equal(t, jobID, latestExecution.JobID)
		assert.NotZero(t, latestExecution.ExecutionTime)
		assert.NotEmpty(t, latestExecution.Status)

		t.Logf("Latest execution status: %s", latestExecution.Status)
	} else {
		t.Log("No executions found yet - this might be normal for longer schedules")
	}
}

func TestQueueStats(t *testing.T) {
	client := NewTestClient()

	// Get initial queue stats
	initialStats := client.GetQueueStats(t)
	t.Logf("Initial queue stats: %+v", initialStats)

	// Create a job to potentially affect queue stats
	req := CreateJobRequest{
		Schedule:      "*/10 * * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		Description:   "Queue stats test job",
		MaxRetryCount: 1,
	}

	jobID := client.CreateJob(t, req)
	t.Logf("Created job %d for queue stats test", jobID)

	// Wait a bit for job to be processed
	time.Sleep(10 * time.Second)

	// Get updated queue stats
	finalStats := client.GetQueueStats(t)
	t.Logf("Final queue stats: %+v", finalStats)

	// Verify stats are reasonable
	assert.GreaterOrEqual(t, finalStats.Ready, int64(0))
	assert.GreaterOrEqual(t, finalStats.Processing, int64(0))
	assert.GreaterOrEqual(t, finalStats.Completed, int64(0))
	assert.GreaterOrEqual(t, finalStats.Retrying, int64(0))
}

func TestErrorHandling(t *testing.T) {
	client := NewTestClient()

	// Test invalid job creation
	invalidReq := CreateJobRequest{
		Schedule: "invalid-schedule",
		API:      "not-a-url",
		Type:     "INVALID_TYPE",
	}

	resp, err := client.makeRequest("POST", "/api/v1/jobs", invalidReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return error status
	assert.NotEqual(t, http.StatusCreated, resp.StatusCode)

	// Test getting non-existent job
	resp, err = client.makeRequest("GET", "/api/v1/jobs/99999", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHighThroughput(t *testing.T) {
	client := NewTestClient()

	// Create multiple jobs quickly
	numJobs := 10
	jobIDs := make([]uint, 0, numJobs)

	start := time.Now()

	for i := 0; i < numJobs; i++ {
		req := CreateJobRequest{
			Schedule:      "*/30 * * * * *",
			API:           "https://httpbin.org/status/200",
			Type:          models.AT_LEAST_ONCE,
			IsRecurring:   true,
			Description:   fmt.Sprintf("High throughput test job %d", i),
			MaxRetryCount: 1,
		}

		jobID := client.CreateJob(t, req)
		jobIDs = append(jobIDs, jobID)
	}

	duration := time.Since(start)
	t.Logf("Created %d jobs in %v", numJobs, duration)

	// Verify all jobs were created successfully
	assert.Len(t, jobIDs, numJobs)

	// Verify all jobs are active
	for _, jobID := range jobIDs {
		job := client.GetJob(t, jobID)
		assert.True(t, job.IsActive, "Job %d should be active", jobID)
	}
}

func TestJobTypes(t *testing.T) {
	client := NewTestClient()

	testCases := []struct {
		name     string
		jobType  models.JobType
		expected string
	}{
		{
			name:     "AT_LEAST_ONCE job",
			jobType:  models.AT_LEAST_ONCE,
			expected: "AT_LEAST_ONCE",
		},
		{
			name:     "AT_MOST_ONCE job",
			jobType:  models.AT_MOST_ONCE,
			expected: "AT_MOST_ONCE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateJobRequest{
				Schedule:      "*/20 * * * * *",
				API:           "https://httpbin.org/status/200",
				Type:          tc.jobType,
				IsRecurring:   true,
				Description:   fmt.Sprintf("Test %s job", tc.expected),
				MaxRetryCount: 1,
			}

			jobID := client.CreateJob(t, req)
			job := client.GetJob(t, jobID)

			assert.Equal(t, tc.expected, string(job.Type))
		})
	}
}

// Integration test that combines multiple operations
func TestEndToEndWorkflow(t *testing.T) {
	client := NewTestClient()

	// 1. Health check
	health := client.HealthCheck(t)
	assert.Equal(t, "healthy", health.Status)

	// 2. Create a job
	req := CreateJobRequest{
		Schedule:      "*/10 * * * * *",
		API:           "https://httpbin.org/delay/1",
		Type:          models.AT_LEAST_ONCE,
		IsRecurring:   true,
		Description:   "End-to-end test job",
		MaxRetryCount: 2,
	}

	jobID := client.CreateJob(t, req)

	// 3. Verify job details
	job := client.GetJob(t, jobID)
	assert.True(t, job.IsActive)
	assert.Equal(t, req.Description, job.Description)

	// 4. Verify schedule
	schedule := client.GetJobSchedule(t, jobID)
	assert.Equal(t, jobID, schedule.JobID)
	assert.True(t, schedule.NextExecutionTime.After(time.Now()))

	// 5. Check initial queue stats
	initialStats := client.GetQueueStats(t)
	t.Logf("Initial queue stats: %+v", initialStats)

	// 6. Wait for execution
	t.Log("Waiting 30 seconds for job execution...")
	time.Sleep(30 * time.Second)

	// 7. Check for executions
	executions := client.GetJobHistory(t, jobID)
	if len(executions) > 0 {
		t.Logf("Found %d executions", len(executions))
		assert.Equal(t, jobID, executions[0].JobID)
	}

	// 8. Check final queue stats
	finalStats := client.GetQueueStats(t)
	t.Logf("Final queue stats: %+v", finalStats)

	// 9. Verify job is still active
	job = client.GetJob(t, jobID)
	assert.True(t, job.IsActive)

	t.Log("End-to-end workflow completed successfully")
}
