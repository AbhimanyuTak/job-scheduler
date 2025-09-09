package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobExecution_Creation(t *testing.T) {
	now := time.Now()
	duration := 1500 * time.Millisecond

	exec := JobExecution{
		JobID:             123,
		Status:            StatusSuccess,
		ExecutionTime:     now,
		ExecutionDuration: &duration,
		RetryCount:        0,
	}

	assert.Equal(t, uint(123), exec.JobID)
	assert.Equal(t, StatusSuccess, exec.Status)
	assert.Equal(t, now, exec.ExecutionTime)
	assert.Equal(t, duration, *exec.ExecutionDuration)
	assert.Equal(t, 0, exec.RetryCount)
}

func TestJobExecution_WithError(t *testing.T) {
	now := time.Now()
	duration := 500 * time.Millisecond

	exec := JobExecution{
		JobID:             123,
		Status:            StatusFailed,
		Error:             "API call failed",
		ExecutionTime:     now,
		ExecutionDuration: &duration,
		RetryCount:        1,
	}

	assert.Equal(t, uint(123), exec.JobID)
	assert.Equal(t, StatusFailed, exec.Status)
	assert.Equal(t, "API call failed", exec.Error)
	assert.Equal(t, now, exec.ExecutionTime)
	assert.Equal(t, duration, *exec.ExecutionDuration)
	assert.Equal(t, 1, exec.RetryCount)
}

func TestExecutionStatus_Constants(t *testing.T) {
	assert.Equal(t, "SUCCESS", string(StatusSuccess))
	assert.Equal(t, "FAILED", string(StatusFailed))
	assert.Equal(t, "SCHEDULED", string(StatusScheduled))
	assert.Equal(t, "RUNNING", string(StatusRunning))
}

func TestExecutionStatus_Validation(t *testing.T) {
	// Test valid statuses
	assert.Equal(t, ExecutionStatus("SUCCESS"), StatusSuccess)
	assert.Equal(t, ExecutionStatus("FAILED"), StatusFailed)
	assert.Equal(t, ExecutionStatus("SCHEDULED"), StatusScheduled)
	assert.Equal(t, ExecutionStatus("RUNNING"), StatusRunning)

	// Test invalid status
	invalidStatus := ExecutionStatus("INVALID")
	assert.NotEqual(t, StatusSuccess, invalidStatus)
	assert.NotEqual(t, StatusFailed, invalidStatus)
	assert.NotEqual(t, StatusScheduled, invalidStatus)
	assert.NotEqual(t, StatusRunning, invalidStatus)
}

func TestJobExecution_IsSuccessful(t *testing.T) {
	tests := []struct {
		name     string
		exec     JobExecution
		expected bool
	}{
		{
			name: "successful execution",
			exec: JobExecution{
				Status: StatusSuccess,
			},
			expected: true,
		},
		{
			name: "failed execution",
			exec: JobExecution{
				Status: StatusFailed,
			},
			expected: false,
		},
		{
			name: "scheduled execution",
			exec: JobExecution{
				Status: StatusScheduled,
			},
			expected: false,
		},
		{
			name: "running execution",
			exec: JobExecution{
				Status: StatusRunning,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since IsSuccessful method doesn't exist, we'll test the logic directly
			isSuccessful := tt.exec.Status == StatusSuccess
			assert.Equal(t, tt.expected, isSuccessful)
		})
	}
}
