package utils

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// ScheduleParser handles CRON schedule parsing and next execution time calculation
type ScheduleParser struct {
	cronParser cron.Parser
}

// NewScheduleParser creates a new schedule parser with second precision
func NewScheduleParser() *ScheduleParser {
	// Create a parser that includes seconds (6 fields: second minute hour day month weekday)
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	return &ScheduleParser{
		cronParser: parser,
	}
}

// ParseSchedule validates and parses a CRON schedule string
func (sp *ScheduleParser) ParseSchedule(schedule string) (cron.Schedule, error) {
	return sp.cronParser.Parse(schedule)
}

// CalculateNextExecution calculates the next execution time for a given schedule
func (sp *ScheduleParser) CalculateNextExecution(schedule string, fromTime time.Time) (time.Time, error) {
	cronSchedule, err := sp.ParseSchedule(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid schedule format: %w", err)
	}

	nextTime := cronSchedule.Next(fromTime)
	return nextTime, nil
}

// ValidateSchedule validates if a schedule string is valid
func (sp *ScheduleParser) ValidateSchedule(schedule string) error {
	_, err := sp.ParseSchedule(schedule)
	return err
}

// IsValidSchedule checks if a schedule string is valid without returning the error
func (sp *ScheduleParser) IsValidSchedule(schedule string) bool {
	return sp.ValidateSchedule(schedule) == nil
}

// GetScheduleDescription returns a human-readable description of the schedule
func (sp *ScheduleParser) GetScheduleDescription(schedule string) (string, error) {
	_, err := sp.ParseSchedule(schedule)
	if err != nil {
		return "", err
	}

	// For now, return the original schedule string
	// In a more sophisticated implementation, we could parse and describe the schedule
	return schedule, nil
}

// CalculateNextExecutionFromNow calculates the next execution time from the current time
func (sp *ScheduleParser) CalculateNextExecutionFromNow(schedule string) (time.Time, error) {
	return sp.CalculateNextExecution(schedule, time.Now().UTC())
}

// CalculateNextExecutionFromTime calculates the next execution time from a specific time
func (sp *ScheduleParser) CalculateNextExecutionFromTime(schedule string, fromTime time.Time) (time.Time, error) {
	return sp.CalculateNextExecution(schedule, fromTime)
}
