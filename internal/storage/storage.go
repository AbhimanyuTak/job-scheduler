package storage

import (
	"time"

	"github.com/manyu/job-scheduler/internal/models"
)

// Storage defines the interface for data persistence operations
type Storage interface {
	// Job operations
	CreateJob(job *models.Job) error
	GetJob(id uint) (*models.Job, error)
	GetAllJobs() ([]*models.Job, error)

	// Job schedule operations
	CreateJobSchedule(schedule *models.JobSchedule) error
	GetJobSchedule(jobID uint) (*models.JobSchedule, error)
	UpdateJobSchedule(jobID uint, nextExecutionTime time.Time) error
	DeleteJobSchedule(jobID uint) error
	GetJobsReadyForExecution(limit int) ([]*models.Job, []*models.JobSchedule, error)

	// Job execution operations
	CreateJobExecution(execution *models.JobExecution) error
	UpdateJobExecution(execution *models.JobExecution) error
	GetJobExecutions(jobID uint, limit int) ([]*models.JobExecution, error)
	GetJobExecutionInProgress(jobID uint) (*models.JobExecution, error)
}
