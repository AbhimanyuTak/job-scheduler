package storage

import (
	"errors"
	"time"

	"github.com/manyu/job-scheduler/internal/database"
	"github.com/manyu/job-scheduler/internal/models"
	"gorm.io/gorm"
)

type PostgresStorage struct {
	db *gorm.DB
}

func NewPostgresStorage() *PostgresStorage {
	return &PostgresStorage{
		db: database.DB,
	}
}

// Job operations
func (s *PostgresStorage) CreateJob(job *models.Job) error {
	result := s.db.Create(job)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *PostgresStorage) GetJob(id uint) (*models.Job, error) {
	var job models.Job
	result := s.db.Where("is_active = ?", true).First(&job, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrJobNotFound
		}
		return nil, result.Error
	}
	return &job, nil
}

func (s *PostgresStorage) GetAllJobs() ([]*models.Job, error) {
	var jobs []*models.Job
	result := s.db.Where("is_active = ?", true).Find(&jobs)
	if result.Error != nil {
		return nil, result.Error
	}
	return jobs, nil
}

// JobSchedule operations
func (s *PostgresStorage) CreateJobSchedule(schedule *models.JobSchedule) error {
	result := s.db.Create(schedule)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *PostgresStorage) GetJobSchedule(jobID uint) (*models.JobSchedule, error) {
	var schedule models.JobSchedule
	result := s.db.Where("job_id = ?", jobID).First(&schedule)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrJobScheduleNotFound
		}
		return nil, result.Error
	}
	return &schedule, nil
}

func (s *PostgresStorage) UpdateJobSchedule(jobID uint, nextExecutionTime time.Time) error {
	result := s.db.Model(&models.JobSchedule{}).
		Where("job_id = ?", jobID).
		Update("next_execution_time", nextExecutionTime)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrJobScheduleNotFound
	}
	return nil
}

func (s *PostgresStorage) DeleteJobSchedule(jobID uint) error {
	result := s.db.Where("job_id = ?", jobID).Delete(&models.JobSchedule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrJobScheduleNotFound
	}
	return nil
}

func (s *PostgresStorage) GetJobsReadyForExecution(limit int) ([]*models.Job, []*models.JobSchedule, error) {
	var results []struct {
		models.Job
		models.JobSchedule
	}

	// Use JOIN to get only schedules for active jobs that are ready for execution
	result := s.db.Table("job_schedules").
		Select("jobs.*, job_schedules.*").
		Joins("JOIN jobs ON job_schedules.job_id = jobs.id").
		Where("job_schedules.next_execution_time <= ? AND jobs.is_active = ? AND job_schedules.deleted_at IS NULL", time.Now(), true).
		Order("job_schedules.next_execution_time ASC").
		Limit(limit).
		Scan(&results)

	if result.Error != nil {
		return nil, nil, result.Error
	}

	if len(results) == 0 {
		return []*models.Job{}, []*models.JobSchedule{}, nil
	}

	// Separate jobs and schedules
	var jobs []*models.Job
	var schedules []*models.JobSchedule

	for _, result := range results {
		job := result.Job
		schedule := result.JobSchedule
		jobs = append(jobs, &job)
		schedules = append(schedules, &schedule)
	}

	return jobs, schedules, nil
}

// JobExecution operations
func (s *PostgresStorage) CreateJobExecution(execution *models.JobExecution) error {
	result := s.db.Create(execution)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *PostgresStorage) UpdateJobExecution(execution *models.JobExecution) error {
	result := s.db.Save(execution)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *PostgresStorage) GetJobExecutions(jobID uint, limit int) ([]*models.JobExecution, error) {
	var executions []*models.JobExecution
	query := s.db.Where("job_id = ?", jobID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&executions)
	if result.Error != nil {
		return nil, result.Error
	}
	return executions, nil
}

func (s *PostgresStorage) GetJobExecutionInProgress(jobID uint) (*models.JobExecution, error) {
	var execution models.JobExecution
	result := s.db.Where("job_id = ? AND status IN (?)", jobID, []string{"SCHEDULED", "RUNNING"}).
		Order("created_at DESC").
		First(&execution)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No execution in progress
		}
		return nil, result.Error
	}
	return &execution, nil
}

// Error definitions
var (
	ErrJobNotFound         = errors.New("job not found")
	ErrJobScheduleNotFound = errors.New("job schedule not found")
)
