package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/response"
)

func TestResponse_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payload := map[string]string{"message": "operation successful"}
	response.Success(c, http.StatusOK, payload)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var env response.SuccessEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if env.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", env.Status)
	}
	if env.Service != "logiflows-api" {
		t.Errorf("expected service 'logiflows-api', got %s", env.Service)
	}
	if env.Version != "v1" {
		t.Errorf("expected version 'v1', got %s", env.Version)
	}
}

func TestResponse_ErrorHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		invoke       func(c *gin.Context)
		expectedCode int
		expectedErr  string
	}{
		{
			name: "BadRequest",
			invoke: func(c *gin.Context) {
				response.BadRequest(c, "invalid field", nil)
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  "BAD_REQUEST",
		},
		{
			name: "Unauthorized default",
			invoke: func(c *gin.Context) {
				response.Unauthorized(c, "")
			},
			expectedCode: http.StatusUnauthorized,
			expectedErr:  "UNAUTHORIZED",
		},
		{
			name: "Forbidden custom code",
			invoke: func(c *gin.Context) {
				response.Forbidden(c, "CROSS_TENANT_ACCESS_DENIED", "cannot access other tenant")
			},
			expectedCode: http.StatusForbidden,
			expectedErr:  "CROSS_TENANT_ACCESS_DENIED",
		},
		{
			name: "NotFound",
			invoke: func(c *gin.Context) {
				response.NotFound(c, "item missing")
			},
			expectedCode: http.StatusNotFound,
			expectedErr:  "NOT_FOUND",
		},
		{
			name: "Conflict",
			invoke: func(c *gin.Context) {
				response.Conflict(c, "already exists")
			},
			expectedCode: http.StatusConflict,
			expectedErr:  "CONFLICT",
		},
		{
			name: "InternalServerError default",
			invoke: func(c *gin.Context) {
				response.InternalServerError(c, "")
			},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  "INTERNAL_SERVER_ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Request-ID", "req-test-uuid")
			c.Request = req

			tc.invoke(c)

			if w.Code != tc.expectedCode {
				t.Fatalf("expected HTTP status %d, got %d", tc.expectedCode, w.Code)
			}

			var errEnv response.ErrorEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &errEnv); err != nil {
				t.Fatalf("failed to decode error envelope: %v", err)
			}

			if errEnv.Error.Code != tc.expectedErr {
				t.Errorf("expected error code %s, got %s", tc.expectedErr, errEnv.Error.Code)
			}
			if errEnv.Error.RequestID != "req-test-uuid" {
				t.Errorf("expected request_id 'req-test-uuid', got %s", errEnv.Error.RequestID)
			}
		})
	}
}
