package models

import (
	"time"

	"gorm.io/gorm"
)

type JobSchedule struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	JobID             uint           `json:"jobId" gorm:"not null;uniqueIndex;index"`
	NextExecutionTime time.Time      `json:"nextExecutionTime" gorm:"not null;index"`
	CreatedAt         time.Time      `json:"createdAt"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}
