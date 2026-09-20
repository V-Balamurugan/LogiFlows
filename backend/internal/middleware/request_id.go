package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/logger"
)

const (
	DefaultRequestIDHeader = "X-Request-ID"
	RequestIDKey           = "request_id"
)

// RequestID returns a middleware that injects or propagates a unique Request ID.
func RequestID(headerName string) gin.HandlerFunc {
	if strings.TrimSpace(headerName) == "" {
		headerName = DefaultRequestIDHeader
	}

	return func(c *gin.Context) {
		reqID := strings.TrimSpace(c.GetHeader(headerName))
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// Set in Gin context
		c.Set(RequestIDKey, reqID)

		// Set on response header
		c.Header(headerName, reqID)

		// Inject into standard Go context.Context for slog correlation
		ctx := logger.ContextWithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetRequestID extracts the Request ID from the Gin context or headers.
func GetRequestID(c *gin.Context) string {
	if val, exists := c.Get(RequestIDKey); exists {
		if id, ok := val.(string); ok && id != "" {
			return id
		}
	}
	if headerVal := c.GetHeader(DefaultRequestIDHeader); headerVal != "" {
		return headerVal
	}
	return ""
}
