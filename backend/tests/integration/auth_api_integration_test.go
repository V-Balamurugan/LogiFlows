package integration_test

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
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/middleware"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/server"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/users"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *auth.TokenService) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("APP_ENV", "test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	log := logger.Init("test", "debug")
	userRepo := users.NewRepository(db.Pool())
	tenantRepo := tenants.NewRepository(db.Pool())
	membershipRepo := memberships.NewRepository(db.Pool())
	auditRepo := audit.NewRepository(db.Pool())

	tokenService := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.Issuer)
	authService := auth.NewService(db.Pool(), userRepo, tenantRepo, membershipRepo, auditRepo, tokenService)
	tenantService := tenants.NewService(db.Pool(), tenantRepo, membershipRepo, userRepo, auditRepo)

	authHandler := auth.NewHandler(authService)
	tenantHandler := tenants.NewHandler(tenantService)
	healthHandler := health.NewHandler(db, nil)

	authMiddleware := middleware.Auth(tokenService, userRepo)
	tenantMiddleware := middleware.TenantContext(membershipRepo, tenantRepo)

	router := server.SetupRouter(server.RouterParams{
		Cfg:              cfg,
		Log:              log,
		HealthHandler:    healthHandler,
		AuthHandler:      authHandler,
		TenantHandler:    tenantHandler,
		AuthMiddleware:   authMiddleware,
		TenantMiddleware: tenantMiddleware,
	})

	return router, tokenService
}

func TestAuthAPI_Register_Login_Me_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping auth API integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	uniqueSuffix := time.Now().UnixNano()
	testEmail := fmt.Sprintf("tenantadmin_%d@quicklogistics.com", uniqueSuffix)
	testPassword := "ComplexP@ssw0rd!2026"
	companyName := fmt.Sprintf("Quick Logistics %d", uniqueSuffix)

	// 1. Test Registration
	regPayload := auth.RegisterRequest{
		Email:       testEmail,
		Password:    testPassword,
		FullName:    "Alice Logistics Admin",
		PhoneNumber: "+1-555-0199",
		CompanyName: companyName,
	}
	body, _ := json.Marshal(regPayload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on registration, got %d: %s", w.Code, w.Body.String())
	}

	var regEnvelope response.SuccessEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &regEnvelope); err != nil {
		t.Fatalf("failed to parse registration response: %v", err)
	}

	dataBytes, _ := json.Marshal(regEnvelope.Data)
	var authResp auth.AuthResponse
	_ = json.Unmarshal(dataBytes, &authResp)

	if authResp.Token == "" {
		t.Errorf("expected access token in registration response")
	}
	if authResp.User.Email != testEmail {
		t.Errorf("expected user email %s, got %s", testEmail, authResp.User.Email)
	}
	if len(authResp.Tenants) == 0 {
		t.Fatalf("expected created tenant summary in response")
	}
	if authResp.Tenants[0].Role != memberships.RoleTenantAdmin {
		t.Errorf("expected TENANT_ADMIN role for creator, got %s", authResp.Tenants[0].Role)
	}

	// 2. Test Duplicate Registration Rejection
	wDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	reqDup.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wDup, reqDup)

	if wDup.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict on duplicate registration, got %d", wDup.Code)
	}

	// 3. Test Login Success
	loginPayload := auth.LoginRequest{
		Email:    testEmail,
		Password: testPassword,
	}
	loginBody, _ := json.Marshal(loginPayload)

	wLogin := httptest.NewRecorder()
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on login, got %d: %s", wLogin.Code, wLogin.Body.String())
	}

	// 4. Test Login Invalid Password
	badLoginPayload := auth.LoginRequest{
		Email:    testEmail,
		Password: "IncorrectPassword123!",
	}
	badBody, _ := json.Marshal(badLoginPayload)

	wBadLogin := httptest.NewRecorder()
	reqBadLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badBody))
	reqBadLogin.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wBadLogin, reqBadLogin)

	if wBadLogin.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized on invalid password, got %d", wBadLogin.Code)
	}

	// 5. Test GET /api/v1/auth/me with valid Bearer token
	wMe := httptest.NewRecorder()
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+authResp.Token)
	router.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on /auth/me, got %d: %s", wMe.Code, wMe.Body.String())
	}

	// 6. Test GET /api/v1/auth/me without token
	wNoToken := httptest.NewRecorder()
	reqNoToken := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	router.ServeHTTP(wNoToken, reqNoToken)

	if wNoToken.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized when token is missing, got %d", wNoToken.Code)
	}

	// 7. Test GET /api/v1/auth/me with corrupted token
	wCorrupt := httptest.NewRecorder()
	reqCorrupt := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqCorrupt.Header.Set("Authorization", "Bearer corrupted.jwt.token.string")
	router.ServeHTTP(wCorrupt, reqCorrupt)

	if wCorrupt.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for corrupted token, got %d", wCorrupt.Code)
	}
}

func TestAuthAPI_ExpiredToken_Rejected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping expired token test in short mode")
	}

	router, _ := setupTestRouter(t)

	cfg, _ := config.Load()
	expiredService := auth.NewTokenService(cfg.JWT.Secret, -1*time.Minute, cfg.JWT.Issuer)
	expiredToken, _, _ := expiredService.GenerateAccessToken(uuid.New(), "expired@logiflows.com", false)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for expired token, got %d", w.Code)
	}
}
