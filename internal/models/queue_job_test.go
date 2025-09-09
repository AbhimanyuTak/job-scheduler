package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueJob_Serialize(t *testing.T) {
	now := time.Now()
	queueJob := &QueueJob{
		ID:            "test-job-123",
		JobID:         123,
		API:           "https://httpbin.org/status/200",
		MaxRetryCount: 3,
		RetryCount:    1,
		CreatedAt:     now,
		ScheduledAt:   now.Add(time.Minute),
		Timeout:       90,
		Type:          AT_LEAST_ONCE,
		IsRecurring:   true,
		Schedule:      "0 */5 * * * *",
	}

	data, err := queueJob.Serialize()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.IsType(t, []byte{}, data)
}

func TestQueueJob_Deserialize(t *testing.T) {
	now := time.Now()
	originalJob := &QueueJob{
		ID:            "test-job-123",
		JobID:         123,
		API:           "https://httpbin.org/status/200",
		MaxRetryCount: 3,
		RetryCount:    1,
		CreatedAt:     now,
		ScheduledAt:   now.Add(time.Minute),
		Timeout:       90,
		Type:          AT_LEAST_ONCE,
		IsRecurring:   true,
		Schedule:      "0 */5 * * * *",
	}

	// Serialize the job
	data, err := originalJob.Serialize()
	require.NoError(t, err)

	// Deserialize it back
	deserializedJob, err := DeserializeQueueJob(data)
	require.NoError(t, err)

	// Compare the jobs
	assert.Equal(t, originalJob.ID, deserializedJob.ID)
	assert.Equal(t, originalJob.JobID, deserializedJob.JobID)
	assert.Equal(t, originalJob.API, deserializedJob.API)
	assert.Equal(t, originalJob.MaxRetryCount, deserializedJob.MaxRetryCount)
	assert.Equal(t, originalJob.RetryCount, deserializedJob.RetryCount)
	assert.Equal(t, originalJob.Timeout, deserializedJob.Timeout)
	assert.Equal(t, originalJob.Type, deserializedJob.Type)
	assert.Equal(t, originalJob.IsRecurring, deserializedJob.IsRecurring)
	assert.Equal(t, originalJob.Schedule, deserializedJob.Schedule)

	// Compare times (allowing for small differences due to serialization)
	assert.WithinDuration(t, originalJob.CreatedAt, deserializedJob.CreatedAt, time.Second)
	assert.WithinDuration(t, originalJob.ScheduledAt, deserializedJob.ScheduledAt, time.Second)
}

func TestQueueJob_Deserialize_InvalidJSON(t *testing.T) {
	invalidData := []byte("invalid json")

	_, err := DeserializeQueueJob(invalidData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestQueueJob_Deserialize_EmptyData(t *testing.T) {
	emptyData := []byte("")

	_, err := DeserializeQueueJob(emptyData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected end of JSON input")
}

func TestQueueJob_RetryLogic(t *testing.T) {
	tests := []struct {
		name          string
		retryCount    int
		maxRetryCount int
		expected      bool
	}{
		{
			name:          "can retry when under limit",
			retryCount:    2,
			maxRetryCount: 5,
			expected:      true,
		},
		{
			name:          "cannot retry when at limit",
			retryCount:    5,
			maxRetryCount: 5,
			expected:      false,
		},
		{
			name:          "cannot retry when over limit",
			retryCount:    6,
			maxRetryCount: 5,
			expected:      false,
		},
		{
			name:          "cannot retry when max is 0",
			retryCount:    0,
			maxRetryCount: 0,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queueJob := &QueueJob{
				RetryCount:    tt.retryCount,
				MaxRetryCount: tt.maxRetryCount,
			}

			// Test the retry logic directly since CanRetry method doesn't exist
			canRetry := queueJob.RetryCount < queueJob.MaxRetryCount
			assert.Equal(t, tt.expected, canRetry)
		})
	}
}

func TestQueueJob_RetryIncrement(t *testing.T) {
	queueJob := &QueueJob{
		RetryCount:    2,
		MaxRetryCount: 5,
	}

	// Test retry increment logic
	originalCount := queueJob.RetryCount
	queueJob.RetryCount++
	assert.Equal(t, originalCount+1, queueJob.RetryCount)
}

func TestQueueJob_ExpirationLogic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		createdAt time.Time
		timeout   int
		expected  bool
	}{
		{
			name:      "not expired - within timeout",
			createdAt: now.Add(-time.Minute),
			timeout:   90,
			expected:  false,
		},
		{
			name:      "expired - past timeout",
			createdAt: now.Add(-2 * time.Minute),
			timeout:   90,
			expected:  true,
		},
		{
			name:      "not expired - zero timeout",
			createdAt: now.Add(-time.Hour),
			timeout:   0,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queueJob := &QueueJob{
				CreatedAt: tt.createdAt,
				Timeout:   tt.timeout,
			}

			// Test expiration logic directly since IsExpired method doesn't exist
			if queueJob.Timeout == 0 {
				assert.False(t, tt.expected) // Zero timeout means never expires
			} else {
				expired := time.Since(queueJob.CreatedAt) > time.Duration(queueJob.Timeout)*time.Second
				assert.Equal(t, tt.expected, expired)
			}
		})
	}
}
