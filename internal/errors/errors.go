package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application-specific error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// WithDetails adds details to an error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string, httpStatus int) *AppError {
	appErr := NewAppError(code, message, httpStatus)
	if err != nil {
		appErr.Details = err.Error()
	}
	return appErr
}

// Predefined application errors
var (
	// Validation errors
	ErrInvalidRequest  = NewAppError("INVALID_REQUEST", "Invalid request", http.StatusBadRequest)
	ErrInvalidJobType  = NewAppError("INVALID_JOB_TYPE", "Invalid job type. Must be AT_LEAST_ONCE or AT_MOST_ONCE", http.StatusBadRequest)
	ErrInvalidSchedule = NewAppError("INVALID_SCHEDULE", "Invalid schedule format", http.StatusBadRequest)

	// Resource errors
	ErrJobNotFound         = NewAppError("JOB_NOT_FOUND", "Job not found", http.StatusNotFound)
	ErrJobScheduleNotFound = NewAppError("JOB_SCHEDULE_NOT_FOUND", "Job schedule not found", http.StatusNotFound)

	// Server errors
	ErrInternalServer = NewAppError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrDatabaseError  = NewAppError("DATABASE_ERROR", "Database operation failed", http.StatusInternalServerError)
	ErrRedisError     = NewAppError("REDIS_ERROR", "Redis operation failed", http.StatusInternalServerError)
	ErrQueueError     = NewAppError("QUEUE_ERROR", "Queue operation failed", http.StatusInternalServerError)

	// Configuration errors
	ErrConfigError = NewAppError("CONFIG_ERROR", "Configuration error", http.StatusInternalServerError)
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// ToResponse converts an AppError to an ErrorResponse
func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Error:   e.Message,
		Code:    e.Code,
		Details: e.Details,
	}
}
