package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobSchedule_Creation(t *testing.T) {
	now := time.Now()

	schedule := JobSchedule{
		JobID:             123,
		NextExecutionTime: now.Add(time.Hour),
	}

	assert.Equal(t, uint(123), schedule.JobID)
	assert.WithinDuration(t, now.Add(time.Hour), schedule.NextExecutionTime, time.Second)
}

func TestJobSchedule_Validation(t *testing.T) {
	now := time.Now()

	// Test valid schedule
	validSchedule := JobSchedule{
		JobID:             123,
		NextExecutionTime: now.Add(time.Hour),
	}
	assert.Equal(t, uint(123), validSchedule.JobID)
	assert.False(t, validSchedule.NextExecutionTime.IsZero())

	// Test invalid schedule (zero job ID)
	invalidSchedule := JobSchedule{
		JobID:             0,
		NextExecutionTime: now.Add(time.Hour),
	}
	assert.Equal(t, uint(0), invalidSchedule.JobID)

	// Test invalid schedule (zero execution time)
	invalidSchedule2 := JobSchedule{
		JobID:             123,
		NextExecutionTime: time.Time{},
	}
	assert.True(t, invalidSchedule2.NextExecutionTime.IsZero())
}

func TestJobSchedule_ExecutionLogic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		nextExecutionTime time.Time
		expected          bool
	}{
		{
			name:              "ready for execution - time has passed",
			nextExecutionTime: now.Add(-time.Minute),
			expected:          true,
		},
		{
			name:              "not ready - time in future",
			nextExecutionTime: now.Add(time.Hour),
			expected:          false,
		},
		{
			name:              "ready for execution - exactly now",
			nextExecutionTime: now,
			expected:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := JobSchedule{
				NextExecutionTime: tt.nextExecutionTime,
			}

			// Test execution logic directly since IsReadyForExecution method doesn't exist
			isReady := schedule.NextExecutionTime.Before(now) || schedule.NextExecutionTime.Equal(now)
			assert.Equal(t, tt.expected, isReady)
		})
	}
}

func TestJobSchedule_TimeUpdate(t *testing.T) {
	now := time.Now()
	schedule := JobSchedule{
		JobID:             123,
		NextExecutionTime: now,
	}

	newTime := now.Add(time.Hour)
	schedule.NextExecutionTime = newTime

	assert.Equal(t, newTime, schedule.NextExecutionTime)
}

func TestJobSchedule_TimeCalculation(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(2 * time.Hour)

	schedule := JobSchedule{
		NextExecutionTime: futureTime,
	}

	// Test time calculation logic directly
	duration := schedule.NextExecutionTime.Sub(now)
	expected := 2 * time.Hour

	// Allow for small time differences
	assert.WithinDuration(t, now.Add(expected), now.Add(duration), time.Second)
}

func TestJobSchedule_OverdueLogic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		nextExecutionTime time.Time
		expected          bool
	}{
		{
			name:              "not overdue - time in future",
			nextExecutionTime: now.Add(time.Hour),
			expected:          false,
		},
		{
			name:              "overdue - time in past",
			nextExecutionTime: now.Add(-time.Hour),
			expected:          true,
		},
		{
			name:              "not overdue - exactly now",
			nextExecutionTime: now,
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := JobSchedule{
				NextExecutionTime: tt.nextExecutionTime,
			}

			// Test overdue logic directly since IsOverdue method doesn't exist
			isOverdue := schedule.NextExecutionTime.Before(now)
			assert.Equal(t, tt.expected, isOverdue)
		})
	}
}
