package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics and logs the stack trace securely.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID := GetRequestID(c)
				stack := string(debug.Stack())

				log.ErrorContext(
					c.Request.Context(),
					"Internal server panic recovered",
					"error", fmt.Sprintf("%v", r),
					"stack", stack,
					"request_id", reqID,
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":       "INTERNAL_SERVER_ERROR",
						"message":    "An unexpected server error occurred",
						"request_id": reqID,
					},
				})
			}
		}()

		c.Next()
	}
}
