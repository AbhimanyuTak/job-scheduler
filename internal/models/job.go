package models

import (
	"time"

	"gorm.io/gorm"
)

type JobType string

const (
	AT_LEAST_ONCE JobType = "AT_LEAST_ONCE"
	AT_MOST_ONCE  JobType = "AT_MOST_ONCE"
)

type Job struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Schedule      string         `json:"schedule" gorm:"size:100;not null"`
	API           string         `json:"api" gorm:"type:text;not null"`
	Type          JobType        `json:"type" gorm:"size:20;not null"`
	IsRecurring   bool           `json:"isRecurring" gorm:"default:false"`
	IsActive      bool           `json:"isActive" gorm:"default:true;index"`
	Description   string         `json:"description" gorm:"type:text"`
	MaxRetryCount int            `json:"maxRetryCount" gorm:"default:3"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
