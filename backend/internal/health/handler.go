package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/redis"
)

const (
	ServiceName = "logiflows-api"
	APIVersion  = "v1"
)

// Handler provides HTTP handlers for liveness and readiness probes.
type Handler struct {
	db    database.DB
	redis redis.Cache
}

// NewHandler creates a new health and readiness handler.
func NewHandler(db database.DB, redis redis.Cache) *Handler {
	return &Handler{
		db:    db,
		redis: redis,
	}
}

// LivenessResponse represents the payload for GET /api/v1/health.
type LivenessResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// DependencyCheck reports the individual status of a dependency.
type DependencyCheck struct {
	Status    string  `json:"status"`
	LatencyMS float64 `json:"latency_ms,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// ReadinessResponse represents the payload for GET /api/v1/readiness.
type ReadinessResponse struct {
	Status    string                     `json:"status"`
	Service   string                     `json:"service"`
	Version   string                     `json:"version"`
	Timestamp time.Time                  `json:"timestamp"`
	Checks    map[string]DependencyCheck `json:"checks"`
}

// Liveness handles GET /api/v1/health.
func (h *Handler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, LivenessResponse{
		Status:    "ok",
		Service:   ServiceName,
		Version:   APIVersion,
		Timestamp: time.Now().UTC(),
	})
}

// Readiness handles GET /api/v1/readiness.
func (h *Handler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]DependencyCheck)
	isReady := true

	// Check Database
	if h.db != nil {
		latency, err := h.db.Ping(ctx)
		if err != nil {
			isReady = false
			checks["database"] = DependencyCheck{
				Status: "DOWN",
				Error:  err.Error(),
			}
		} else {
			checks["database"] = DependencyCheck{
				Status:    "UP",
				LatencyMS: float64(latency.Microseconds()) / 1000.0,
			}
		}
	} else {
		isReady = false
		checks["database"] = DependencyCheck{
			Status: "DOWN",
			Error:  "database client uninitialized",
		}
	}

	// Check Redis
	if h.redis != nil {
		latency, err := h.redis.Ping(ctx)
		if err != nil {
			isReady = false
			checks["redis"] = DependencyCheck{
				Status: "DOWN",
				Error:  err.Error(),
			}
		} else {
			checks["redis"] = DependencyCheck{
				Status:    "UP",
				LatencyMS: float64(latency.Microseconds()) / 1000.0,
			}
		}
	} else {
		isReady = false
		checks["redis"] = DependencyCheck{
			Status: "DOWN",
			Error:  "redis client uninitialized",
		}
	}

	statusCode := http.StatusOK
	overallStatus := "ready"
	if !isReady {
		statusCode = http.StatusServiceUnavailable
		overallStatus = "unready"
	}

	c.JSON(statusCode, ReadinessResponse{
		Status:    overallStatus,
		Service:   ServiceName,
		Version:   APIVersion,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	})
}
