package server

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/middleware"
	"github.com/logiflows/logiflows/backend/internal/response"
)

// SetupRouter configures the Gin engine, global middleware, and API routes.
func SetupRouter(cfg *config.Config, log *slog.Logger, healthHandler *health.Handler) *gin.Engine {
	if strings.ToLower(cfg.App.Env) == "production" || strings.ToLower(cfg.App.Env) == "staging" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global Middlewares
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID(cfg.App.RequestIDHeader))
	r.Use(middleware.StructuredLogger(log))
	r.Use(middleware.Recovery(log))

	// Handle 404 Not Found uniformly
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "The requested resource was not found")
	})

	// Handle 405 Method Not Allowed
	r.NoMethod(func(c *gin.Context) {
		response.Error(c, 405, "METHOD_NOT_ALLOWED", "HTTP method is not supported on this endpoint", nil)
	})

	// API v1 Route Group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Liveness)
		v1.GET("/readiness", healthHandler.Readiness)
	}

	return r
}
