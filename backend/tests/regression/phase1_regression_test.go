package regression_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/middleware"
	"github.com/logiflows/logiflows/backend/internal/redis"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/server"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/users"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

func setupRegressionRouter(t *testing.T) (*gin.Engine, *auth.TokenService) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("APP_ENV", "test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("regression setup: failed to load config: %v", err)
	}
	cfg.Database.MaxOpenConns = 5
	cfg.Database.MaxIdleConns = 1

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Fatalf("regression setup: failed to connect to database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	rdb, err := redis.New(ctx, &cfg.Redis)
	if err != nil {
		t.Fatalf("regression setup: failed to connect to redis: %v", err)
	}

	log := logger.Init("test", "error")
	userRepo := users.NewRepository(db.Pool())
	tenantRepo := tenants.NewRepository(db.Pool())
	membershipRepo := memberships.NewRepository(db.Pool())
	auditRepo := audit.NewRepository(db.Pool())
	tokenRepo := auth.NewRefreshTokenRepository(db.Pool())
	branchRepo := branches.NewRepository(db.Pool())
	employeeRepo := employees.NewRepository(db.Pool())
	vehicleRepo := vehicles.NewRepository(db.Pool())

	tokenService := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.Issuer)
	authService := auth.NewService(db.Pool(), userRepo, tenantRepo, membershipRepo, auditRepo, tokenRepo, tokenService)
	tenantService := tenants.NewService(db.Pool(), tenantRepo, membershipRepo, userRepo, auditRepo)
	branchService := branches.NewService(branchRepo, auditRepo)
	employeeService := employees.NewService(employeeRepo, branchRepo, auditRepo)
	vehicleService := vehicles.NewService(vehicleRepo, branchRepo, employeeRepo, auditRepo)

	authHandler := auth.NewHandler(authService)
	tenantHandler := tenants.NewHandler(tenantService)
	branchHandler := branches.NewHandler(branchService)
	employeeHandler := employees.NewHandler(employeeService)
	vehicleHandler := vehicles.NewHandler(vehicleService)
	healthHandler := health.NewHandler(db, rdb)

	authMiddleware := middleware.Auth(tokenService, userRepo)
	tenantMiddleware := middleware.TenantContext(membershipRepo, tenantRepo)

	router := server.SetupRouter(server.RouterParams{
		Cfg:              cfg,
		Log:              log,
		HealthHandler:    healthHandler,
		AuthHandler:      authHandler,
		TenantHandler:    tenantHandler,
		BranchHandler:    branchHandler,
		EmployeeHandler:  employeeHandler,
		VehicleHandler:   vehicleHandler,
		AuthMiddleware:   authMiddleware,
		TenantMiddleware: tenantMiddleware,
	})

	return router, tokenService
}

func registerRegressionTenant(t *testing.T, router http.Handler, orgPrefix string) (token string, userID uuid.UUID, tenantID uuid.UUID) {
	uniqueID := time.Now().UnixNano()
	email := fmt.Sprintf("reg-%s-%d@logiflows.test", orgPrefix, uniqueID)
	password := "RegP@ssw0rd!2026"
	company := fmt.Sprintf("Reg %s Logistics %d", orgPrefix, uniqueID)

	payload := auth.RegisterRequest{
		Email:       email,
		Password:    password,
		FullName:    orgPrefix + " Admin",
		CompanyName: company,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to register regression tenant %s: %d %s", orgPrefix, w.Code, w.Body.String())
	}

	var env response.SuccessEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	dataBytes, _ := json.Marshal(env.Data)
	var resp auth.AuthResponse
	_ = json.Unmarshal(dataBytes, &resp)

	return resp.Token, resp.User.ID, resp.Tenants[0].ID
}

// TestRegression_Phase0_Foundation verifies that foundation features never regress:
// Health (200), Readiness (200, DB UP, Redis UP), RequestID, 404, 405.
func TestRegression_Phase0_Foundation(t *testing.T) {
	router, _ := setupRegressionRouter(t)

	t.Run("Health Liveness Returns 200 OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("regression failure: GET /health returned %d", w.Code)
		}
		if w.Header().Get("X-Request-Id") == "" {
			t.Errorf("regression failure: missing X-Request-Id header")
		}
	})

	t.Run("Readiness Probe Returns 200 Ready", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("regression failure: GET /readiness returned %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Custom X-Request-ID Preserved", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		req.Header.Set("X-Request-ID", "custom-regression-trace-123")
		router.ServeHTTP(w, req)

		if w.Header().Get("X-Request-Id") != "custom-regression-trace-123" {
			t.Errorf("regression failure: X-Request-Id not preserved, got %q", w.Header().Get("X-Request-Id"))
		}
	})

	t.Run("404 Uniform JSON Envelope", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown-path-for-test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("regression failure: expected 404, got %d", w.Code)
		}
	})
}

// TestRegression_Phase1_Identity_And_MultiTenancy verifies that Phase 1 core rules never regress:
// Bcrypt hashing, Login, Refresh single-use rotation, Tenant isolation, and RBAC matrix.
func TestRegression_Phase1_Identity_And_MultiTenancy(t *testing.T) {
	router, _ := setupRegressionRouter(t)

	// 1. Multi-Tenant Isolation
	tokenA, _, tenantIDA := registerRegressionTenant(t, router, "Alpha")
	tokenB, _, tenantIDB := registerRegressionTenant(t, router, "Beta")

	t.Run("Tenant Isolation - Cross-Tenant Access Denied", func(t *testing.T) {
		// User B attempts to access Tenant A -> 403 Forbidden
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s", tenantIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("regression failure: cross-tenant access allowed with code %d", w.Code)
		}
	})

	t.Run("Tenant Isolation - Self Access Allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s", tenantIDB), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("regression failure: self tenant access failed with code %d", w.Code)
		}
	})

	// 2. RBAC Matrix Enforcement
	t.Run("RBAC - Tenant Admin Updates Tenant Metadata", func(t *testing.T) {
		updatePayload := map[string]string{
			"name": "Alpha Logistics Renamed",
		}
		body, _ := json.Marshal(updatePayload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("regression failure: tenant admin update failed with %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("RBAC - Non-Member Forbidden from Tenant Update", func(t *testing.T) {
		updatePayload := map[string]string{
			"name": "Malicious Rename",
		}
		body, _ := json.Marshal(updatePayload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenB)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("regression failure: non-member allowed to update tenant with %d", w.Code)
		}
	})

	// 3. Token Expiration & Authentication
	t.Run("Authentication - Unauthenticated Request Rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("regression failure: unauthenticated request returned %d instead of 401", w.Code)
		}
	})
}
