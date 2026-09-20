package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestID_GeneratesNewUUIDWhenMissing(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestID("X-Request-ID"))
	r.GET("/test", func(c *gin.Context) {
		reqID := middleware.GetRequestID(c)
		c.String(http.StatusOK, reqID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	respHeaderID := w.Header().Get("X-Request-ID")
	if respHeaderID == "" {
		t.Fatalf("expected X-Request-ID header in response")
	}

	if _, err := uuid.Parse(respHeaderID); err != nil {
		t.Errorf("expected generated request ID to be a valid UUID, got: %s", respHeaderID)
	}

	bodyID := w.Body.String()
	if bodyID != respHeaderID {
		t.Errorf("expected body request ID to match header ID: %s vs %s", bodyID, respHeaderID)
	}
}

func TestRequestID_PreservesExistingHeader(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestID("X-Request-ID"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, middleware.GetRequestID(c))
	})

	customID := "custom-trace-id-abc-123"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", customID)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	respHeaderID := w.Header().Get("X-Request-ID")
	if respHeaderID != customID {
		t.Errorf("expected preserved header %s, got: %s", customID, respHeaderID)
	}
}

func TestRecovery_HandlesPanicGracefully(t *testing.T) {
	var buf bytes.Buffer
	log := logger.InitWithWriter("test", "error", &buf)

	r := gin.New()
	r.Use(middleware.RequestID("X-Request-ID"))
	r.Use(middleware.Recovery(log))
	r.GET("/panic", func(c *gin.Context) {
		panic("simulated critical runtime failure")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	if w.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID in error response")
	}
}
