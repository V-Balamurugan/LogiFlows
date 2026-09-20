package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/redis/go-redis/v9"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// MockDB implements database.DB for testing
type MockDB struct {
	pingErr error
	latency time.Duration
}

func (m *MockDB) Ping(ctx context.Context) (time.Duration, error) {
	return m.latency, m.pingErr
}
func (m *MockDB) VerifyPostGIS(ctx context.Context) (string, error) {
	return "POSTGIS 3.4.0", nil
}
func (m *MockDB) Pool() *pgxpool.Pool { return nil }
func (m *MockDB) Close()              {}

// MockRedis implements redis.Cache for testing
type MockRedis struct {
	pingErr error
	latency time.Duration
}

func (m *MockRedis) Ping(ctx context.Context) (time.Duration, error) {
	return m.latency, m.pingErr
}
func (m *MockRedis) Client() *redis.Client { return nil }
func (m *MockRedis) Close() error          { return nil }

func TestLiveness_Returns200OK(t *testing.T) {
	h := health.NewHandler(&MockDB{}, &MockRedis{})

	r := gin.New()
	r.GET("/api/v1/health", h.Liveness)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp health.LivenessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode liveness response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
	if resp.Service != "logiflows-api" {
		t.Errorf("expected service 'logiflows-api', got '%s'", resp.Service)
	}
	if resp.Version != "v1" {
		t.Errorf("expected version 'v1', got '%s'", resp.Version)
	}
}

func TestReadiness_AllHealthy_Returns200(t *testing.T) {
	mockDB := &MockDB{latency: 1 * time.Millisecond}
	mockRedis := &MockRedis{latency: 500 * time.Microsecond}
	h := health.NewHandler(mockDB, mockRedis)

	r := gin.New()
	r.GET("/api/v1/readiness", h.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp health.ReadinessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode readiness response: %v", err)
	}

	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", resp.Status)
	}
	if resp.Checks["database"].Status != "UP" {
		t.Errorf("expected database check 'UP', got '%s'", resp.Checks["database"].Status)
	}
	if resp.Checks["redis"].Status != "UP" {
		t.Errorf("expected redis check 'UP', got '%s'", resp.Checks["redis"].Status)
	}
}

func TestReadiness_DegradedDB_Returns503(t *testing.T) {
	mockDB := &MockDB{pingErr: errors.New("dial tcp: connection refused")}
	mockRedis := &MockRedis{latency: 500 * time.Microsecond}
	h := health.NewHandler(mockDB, mockRedis)

	r := gin.New()
	r.GET("/api/v1/readiness", h.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 Service Unavailable, got %d", w.Code)
	}

	var resp health.ReadinessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode readiness response: %v", err)
	}

	if resp.Status != "unready" {
		t.Errorf("expected overall status 'unready', got '%s'", resp.Status)
	}
	if resp.Checks["database"].Status != "DOWN" {
		t.Errorf("expected database status 'DOWN', got '%s'", resp.Checks["database"].Status)
	}
	if resp.Checks["redis"].Status != "UP" {
		t.Errorf("expected redis status 'UP', got '%s'", resp.Checks["redis"].Status)
	}
}
