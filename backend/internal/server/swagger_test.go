package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/internal/server"
)

func setupSwaggerTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	log := logger.Init("test", "error")
	cfg := &config.Config{
		App: config.AppConfig{
			Name:            "LogiFlows Test API",
			Env:             "test",
			Host:            "localhost",
			Port:            8080,
			LogLevel:        "error",
			RequestIDHeader: "X-Request-ID",
		},
	}

	return server.SetupRouter(server.RouterParams{
		Cfg: cfg,
		Log: log,
	})
}

func TestSwagger_UI_Endpoint(t *testing.T) {
	router := setupSwaggerTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /swagger/index.html, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %q", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "swagger-ui") && !strings.Contains(body, "SwaggerUIBundle") {
		t.Errorf("expected swagger-ui markup in response body")
	}
}

func TestSwagger_DocJSON_Endpoint(t *testing.T) {
	router := setupSwaggerTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /swagger/doc.json, got %d", w.Code)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("failed to parse /swagger/doc.json as valid JSON: %v", err)
	}

	// 1. Verify basic swagger spec attributes
	if swaggerVer, ok := spec["swagger"].(string); !ok || swaggerVer != "2.0" {
		t.Errorf("expected swagger version '2.0', got %v", spec["swagger"])
	}

	// 2. Verify API metadata
	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing or invalid 'info' object in swagger spec")
	}
	if title, ok := info["title"].(string); !ok || title != "LogiFlows API" {
		t.Errorf("expected info.title 'LogiFlows API', got %v", info["title"])
	}
	if ver, ok := info["version"].(string); !ok || ver != "1.0" {
		t.Errorf("expected info.version '1.0', got %v", info["version"])
	}

	// 3. Verify security definitions
	secDefs, ok := spec["securityDefinitions"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'securityDefinitions' object in swagger spec")
	}
	if _, ok := secDefs["BearerAuth"]; !ok {
		t.Errorf("expected 'BearerAuth' in securityDefinitions")
	}

	// 4. Verify required API endpoints exist in spec
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'paths' object in swagger spec")
	}

	expectedPaths := []string{
		"/api/v1/health",
		"/api/v1/readiness",
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/auth/me",
		"/api/v1/auth/logout",
		"/api/v1/tenants",
		"/api/v1/tenants/{tenant_id}",
		"/api/v1/tenants/{tenant_id}/members",
	}

	for _, p := range expectedPaths {
		if _, exists := paths[p]; !exists {
			t.Errorf("expected path %q to be defined in swagger spec", p)
		}
	}
}

func TestSwagger_DocsRedirect(t *testing.T) {
	router := setupSwaggerTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301 Moved Permanently for /docs, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/swagger/index.html" {
		t.Errorf("expected redirect Location '/swagger/index.html', got %q", location)
	}
}

func TestSwagger_NonExistentAsset_Returns404(t *testing.T) {
	router := setupSwaggerTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/not-a-valid-swagger-asset.js", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for non-existent swagger asset, got %d", w.Code)
	}
}
