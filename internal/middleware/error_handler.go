package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/manyu/job-scheduler/internal/errors"
)

// ErrorHandlerMiddleware provides consistent error handling across the API
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
			appErr := errors.ErrInternalServer.WithDetails(err)
			c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		} else if appErr, ok := recovered.(*errors.AppError); ok {
			log.Printf("Application error: %s", appErr.Error())
			c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		} else {
			log.Printf("Unknown panic: %v", recovered)
			appErr := errors.ErrInternalServer.WithDetails("Unknown error occurred")
			c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		}
		c.Abort()
	})
}

// HandleError handles errors consistently across handlers
func HandleError(c *gin.Context, err error) {
	log.Printf("Handler error: %v", err)

	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	// Default to internal server error for unknown errors
	appErr := errors.ErrInternalServer.WithDetails(err.Error())
	c.JSON(appErr.HTTPStatus, appErr.ToResponse())
}
