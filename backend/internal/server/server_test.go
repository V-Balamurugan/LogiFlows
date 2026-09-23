package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/server"
)

func TestServer_SetupRouter_404And405(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		App: config.AppConfig{
			Env:             "test",
			RequestIDHeader: "X-Request-ID",
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	healthHandler := health.NewHandler(nil, nil)

	r := server.SetupRouter(server.RouterParams{
		Cfg:              cfg,
		Log:              logger,
		HealthHandler:    healthHandler,
		AuthHandler:      nil,
		TenantHandler:    nil,
		AuthMiddleware:   func(c *gin.Context) { c.Next() },
		TenantMiddleware: func(c *gin.Context) { c.Next() },
	})

	// 1. Verify 404 uniform envelope
	w404 := httptest.NewRecorder()
	req404 := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	r.ServeHTTP(w404, req404)

	if w404.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", w404.Code)
	}

	// 2. Verify health liveness route exists
	wHealth := httptest.NewRecorder()
	reqHealth := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	r.ServeHTTP(wHealth, reqHealth)

	if wHealth.Code != http.StatusOK {
		t.Errorf("expected 200 OK on health endpoint, got %d", wHealth.Code)
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		App: config.AppConfig{
			Name: "test-server",
			Env:  "test",
			Host: "127.0.0.1",
			Port: 0, // OS chooses available ephemeral port
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := gin.New()

	srv := server.New(cfg, logger, engine)
	errChan := srv.Start()

	// Give the server a moment to bind
	time.Sleep(50 * time.Millisecond)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}

	// Wait for server goroutine to terminate cleanly
	select {
	case err, ok := <-errChan:
		if ok && err != nil {
			t.Errorf("server reported unexpected error on shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("server shutdown timed out")
	}
}
