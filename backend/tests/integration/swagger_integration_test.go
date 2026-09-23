package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSwagger_Integration_UI_And_Spec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping swagger integration test in short mode")
	}

	router, _ := setupTestRouter(t)

	// 1. Verify Swagger UI HTML Endpoint
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on /swagger/index.html, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "swagger-ui") {
			t.Errorf("expected 'swagger-ui' in Swagger HTML body")
		}
	}

	// 2. Verify Swagger doc.json Endpoint
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on /swagger/doc.json, got %d", w.Code)
		}

		var spec map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
			t.Fatalf("failed to decode swagger spec JSON: %v", err)
		}

		info, ok := spec["info"].(map[string]interface{})
		if !ok || info["title"] != "LogiFlows API" {
			t.Errorf("expected Swagger info.title 'LogiFlows API', got %v", info["title"])
		}

		paths, ok := spec["paths"].(map[string]interface{})
		if !ok {
			t.Fatalf("missing paths in swagger spec")
		}

		// Ensure key endpoints are documented
		endpoints := []string{"/api/v1/health", "/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/tenants"}
		for _, ep := range endpoints {
			if _, exists := paths[ep]; !exists {
				t.Errorf("expected endpoint %s in Swagger doc", ep)
			}
		}
	}

	// 3. Verify /docs Redirect
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Fatalf("expected 301 on /docs, got %d", w.Code)
		}
		if loc := w.Header().Get("Location"); loc != "/swagger/index.html" {
			t.Errorf("expected redirect Location '/swagger/index.html', got %s", loc)
		}
	}
}
