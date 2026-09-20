package response

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/middleware"
)

const (
	ServiceName = "logiflows-api"
	APIVersion  = "v1"
)

// SuccessEnvelope is the uniform envelope for successful responses.
type SuccessEnvelope struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// ErrorEnvelope is the uniform envelope for errors.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains the specific error details.
type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
	Details   any    `json:"details,omitempty"`
}

// Success writes a standard JSON success response.
func Success(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, SuccessEnvelope{
		Status:    "ok",
		Service:   ServiceName,
		Version:   APIVersion,
		Timestamp: time.Now().UTC(),
		Data:      data,
	})
}

// Error writes a structured error response with request correlation.
func Error(c *gin.Context, statusCode int, code, message string, details any) {
	reqID := middleware.GetRequestID(c)

	c.AbortWithStatusJSON(statusCode, ErrorEnvelope{
		Error: ErrorBody{
			Code:      code,
			Message:   message,
			RequestID: reqID,
			Details:   details,
		},
	})
}

// BadRequest helper.
func BadRequest(c *gin.Context, message string, details any) {
	Error(c, 400, "BAD_REQUEST", message, details)
}

// NotFound helper.
func NotFound(c *gin.Context, message string) {
	Error(c, 404, "NOT_FOUND", message, nil)
}

// InternalServerError helper.
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "An unexpected internal server error occurred"
	}
	Error(c, 500, "INTERNAL_SERVER_ERROR", message, nil)
}
