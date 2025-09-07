package utils

import (
	"testing"
	"time"
)

func TestScheduleParser(t *testing.T) {
	parser := NewScheduleParser()

	tests := []struct {
		name          string
		schedule      string
		shouldBeValid bool
		expectedError string
	}{
		{
			name:          "Valid daily schedule",
			schedule:      "0 0 12 * * *",
			shouldBeValid: true,
		},
		{
			name:          "Valid weekday schedule",
			schedule:      "30 0 9 * * MON-FRI",
			shouldBeValid: true,
		},
		{
			name:          "Valid every 15 minutes",
			schedule:      "0 */15 * * * *",
			shouldBeValid: true,
		},
		{
			name:          "Invalid schedule - too few fields",
			schedule:      "0 0 12 * *",
			shouldBeValid: false,
		},
		{
			name:          "Invalid schedule - invalid field",
			schedule:      "0 0 25 * * *",
			shouldBeValid: false,
		},
		{
			name:          "Invalid schedule - malformed",
			schedule:      "invalid cron",
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := parser.IsValidSchedule(tt.schedule)
			if isValid != tt.shouldBeValid {
				t.Errorf("IsValidSchedule() = %v, want %v", isValid, tt.shouldBeValid)
			}

			if tt.shouldBeValid {
				// Test that we can calculate next execution time
				nextTime, err := parser.CalculateNextExecutionFromNow(tt.schedule)
				if err != nil {
					t.Errorf("CalculateNextExecutionFromNow() error = %v", err)
				}
				if nextTime.IsZero() {
					t.Error("CalculateNextExecutionFromNow() returned zero time")
				}
				if nextTime.Before(time.Now()) {
					t.Error("CalculateNextExecutionFromNow() returned time in the past")
				}
			}
		})
	}
}

func TestCalculateNextExecution(t *testing.T) {
	parser := NewScheduleParser()

	// Test with a specific time
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		schedule      string
		baseTime      time.Time
		expectedAfter time.Time
	}{
		{
			name:          "Every hour at minute 0",
			schedule:      "0 0 * * * *",
			baseTime:      baseTime,
			expectedAfter: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			name:          "Every day at noon",
			schedule:      "0 0 12 * * *",
			baseTime:      baseTime,
			expectedAfter: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTime, err := parser.CalculateNextExecutionFromTime(tt.schedule, tt.baseTime)
			if err != nil {
				t.Errorf("CalculateNextExecutionFromTime() error = %v", err)
			}
			if !nextTime.After(tt.expectedAfter.Add(-time.Second)) || !nextTime.Before(tt.expectedAfter.Add(time.Second)) {
				t.Errorf("CalculateNextExecutionFromTime() = %v, expected around %v", nextTime, tt.expectedAfter)
			}
		})
	}
}
