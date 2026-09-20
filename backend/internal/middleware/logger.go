package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger logs HTTP requests with method, path, status, latency, IP, and request ID.
func StructuredLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		reqID := GetRequestID(c)

		fullPath := path
		if query != "" {
			fullPath = path + "?" + query
		}

		attrs := []any{
			"http_method", method,
			"path", fullPath,
			"status", statusCode,
			"latency_ms", float64(latency.Microseconds()) / 1000.0,
			"client_ip", clientIP,
			"request_id", reqID,
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case statusCode >= 500:
			log.ErrorContext(c.Request.Context(), "HTTP Request Failed", attrs...)
		case statusCode >= 400:
			log.WarnContext(c.Request.Context(), "HTTP Client Error", attrs...)
		default:
			log.InfoContext(c.Request.Context(), "HTTP Request", attrs...)
		}
	}
}
