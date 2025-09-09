package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/manyu/job-scheduler/internal/models"
	"github.com/manyu/job-scheduler/internal/storage"
	"github.com/manyu/job-scheduler/internal/utils"
)

type JobHandler struct {
	storage        *storage.PostgresStorage
	scheduleParser *utils.ScheduleParser
}

func NewJobHandler(storage *storage.PostgresStorage) *JobHandler {
	return &JobHandler{
		storage:        storage,
		scheduleParser: utils.NewScheduleParser(),
	}
}

// CreateJobRequest represents the request payload for creating a job
type CreateJobRequest struct {
	Schedule      string         `json:"schedule" binding:"required"`
	API           string         `json:"api" binding:"required"`
	Type          models.JobType `json:"type" binding:"required"`
	IsRecurring   bool           `json:"isRecurring"`
	Description   string         `json:"description"`
	MaxRetryCount int            `json:"maxRetryCount"`
}

// CreateJobResponse represents the response for creating a job
type CreateJobResponse struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

// CreateJob handles POST /jobs
func (h *JobHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Validate job type
	if req.Type != models.AT_LEAST_ONCE && req.Type != models.AT_MOST_ONCE {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job type. Must be AT_LEAST_ONCE or AT_MOST_ONCE",
		})
		return
	}

	// Validate CRON schedule format
	if err := h.scheduleParser.ValidateSchedule(req.Schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid schedule format",
			"details": err.Error(),
		})
		return
	}

	// Set default values
	if req.MaxRetryCount == 0 {
		req.MaxRetryCount = 3
	}

	// Create job model
	job := &models.Job{
		Schedule:      req.Schedule,
		API:           req.API,
		Type:          req.Type,
		IsRecurring:   req.IsRecurring,
		Description:   req.Description,
		MaxRetryCount: req.MaxRetryCount,
		IsActive:      true,
	}

	// Save job to database
	if err := h.storage.CreateJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create job",
			"details": err.Error(),
		})
		return
	}

	// Create job schedule for both recurring and non-recurring jobs
	// For non-recurring jobs, we still need to schedule the first execution
	nextExecutionTime, err := h.scheduleParser.CalculateNextExecutionFromNow(req.Schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to calculate next execution time",
			"details": err.Error(),
		})
		return
	}

	schedule := &models.JobSchedule{
		JobID:             job.ID,
		NextExecutionTime: nextExecutionTime,
	}

	if err := h.storage.CreateJobSchedule(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create job schedule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, CreateJobResponse{
		ID:      job.ID,
		Message: "Job created successfully",
	})
}

// GetJob handles GET /jobs/:id
func (h *JobHandler) GetJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID",
		})
		return
	}

	job, err := h.storage.GetJob(uint(id))
	if err != nil {
		if err == storage.ErrJobNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Job not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs handles GET /jobs
func (h *JobHandler) ListJobs(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	jobs, err := h.storage.GetAllJobs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get jobs",
			"details": err.Error(),
		})
		return
	}

	// Apply pagination
	total := len(jobs)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedJobs := jobs[start:end]

	c.JSON(http.StatusOK, gin.H{
		"jobs":   paginatedJobs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetJobHistory handles GET /jobs/:id/history
func (h *JobHandler) GetJobHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	executions, err := h.storage.GetJobExecutions(uint(id), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get job history",
			"details": err.Error(),
		})
		return
	}

	// Apply limit
	if limit < len(executions) {
		executions = executions[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"total":      len(executions),
		"limit":      limit,
	})
}

// GetJobSchedule handles GET /jobs/:id/schedule
func (h *JobHandler) GetJobSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID",
		})
		return
	}

	schedule, err := h.storage.GetJobSchedule(uint(id))
	if err != nil {
		if err == storage.ErrJobScheduleNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Job schedule not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get job schedule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schedule)
}
