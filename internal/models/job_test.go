package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJob_Creation(t *testing.T) {
	job := Job{
		Description:   "Test job",
		Schedule:      "0 */5 * * * *",
		API:           "https://httpbin.org/status/200",
		Type:          AT_LEAST_ONCE,
		IsRecurring:   true,
		MaxRetryCount: 3,
		IsActive:      true,
	}

	assert.Equal(t, "Test job", job.Description)
	assert.Equal(t, "0 */5 * * * *", job.Schedule)
	assert.Equal(t, "https://httpbin.org/status/200", job.API)
	assert.Equal(t, AT_LEAST_ONCE, job.Type)
	assert.True(t, job.IsRecurring)
	assert.Equal(t, 3, job.MaxRetryCount)
	assert.True(t, job.IsActive)
}

func TestJob_OneTimeJob(t *testing.T) {
	job := Job{
		Description:   "One-time job",
		Schedule:      "0 0 12 * * *",
		API:           "https://httpbin.org/status/200",
		Type:          AT_MOST_ONCE,
		IsRecurring:   false,
		MaxRetryCount: 1,
		IsActive:      true,
	}

	assert.Equal(t, "One-time job", job.Description)
	assert.Equal(t, "0 0 12 * * *", job.Schedule)
	assert.Equal(t, "https://httpbin.org/status/200", job.API)
	assert.Equal(t, AT_MOST_ONCE, job.Type)
	assert.False(t, job.IsRecurring)
	assert.Equal(t, 1, job.MaxRetryCount)
	assert.True(t, job.IsActive)
}

func TestJobType_Constants(t *testing.T) {
	assert.Equal(t, "AT_LEAST_ONCE", string(AT_LEAST_ONCE))
	assert.Equal(t, "AT_MOST_ONCE", string(AT_MOST_ONCE))
}

func TestJobType_Validation(t *testing.T) {
	// Test valid job types
	assert.Equal(t, JobType("AT_LEAST_ONCE"), AT_LEAST_ONCE)
	assert.Equal(t, JobType("AT_MOST_ONCE"), AT_MOST_ONCE)

	// Test invalid job type
	invalidType := JobType("INVALID")
	assert.NotEqual(t, AT_LEAST_ONCE, invalidType)
	assert.NotEqual(t, AT_MOST_ONCE, invalidType)
}

func TestJob_ActiveStatus(t *testing.T) {
	activeJob := Job{IsActive: true}
	inactiveJob := Job{IsActive: false}

	assert.True(t, activeJob.IsActive)
	assert.False(t, inactiveJob.IsActive)
}
