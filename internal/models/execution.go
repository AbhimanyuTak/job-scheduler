package models

import (
	"time"

	"gorm.io/gorm"
)

type ExecutionStatus string

const (
	StatusScheduled ExecutionStatus = "SCHEDULED"
	StatusRunning   ExecutionStatus = "RUNNING"
	StatusSuccess   ExecutionStatus = "SUCCESS"
	StatusFailed    ExecutionStatus = "FAILED"
)

type JobExecution struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	JobID             uint            `json:"jobId" gorm:"not null;index"`
	Status            ExecutionStatus `json:"status" gorm:"size:20;not null;index"`
	Error             string          `json:"error,omitempty" gorm:"type:text"`
	ExecutionTime     time.Time       `json:"executionTime" gorm:"not null;index"`
	ExecutionDuration *time.Duration  `json:"executionDuration,omitempty"`
	RetryCount        int             `json:"retryCount" gorm:"default:0"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt  `json:"-" gorm:"index"`
}
