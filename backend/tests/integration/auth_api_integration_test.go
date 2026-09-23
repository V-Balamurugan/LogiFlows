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
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/middleware"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/server"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/users"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *auth.TokenService) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("APP_ENV", "test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	cfg.Database.MaxOpenConns = 5
	cfg.Database.MaxIdleConns = 1

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	log := logger.Init("test", "debug")
	userRepo := users.NewRepository(db.Pool())
	tenantRepo := tenants.NewRepository(db.Pool())
	membershipRepo := memberships.NewRepository(db.Pool())
	auditRepo := audit.NewRepository(db.Pool())
	tokenRepo := auth.NewRefreshTokenRepository(db.Pool())

	tokenService := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.Issuer)
	authService := auth.NewService(db.Pool(), userRepo, tenantRepo, membershipRepo, auditRepo, tokenRepo, tokenService)
	tenantService := tenants.NewService(db.Pool(), tenantRepo, membershipRepo, userRepo, auditRepo)

	branchRepo := branches.NewRepository(db.Pool())
	branchService := branches.NewService(branchRepo, auditRepo)
	branchHandler := branches.NewHandler(branchService)

	employeeRepo := employees.NewRepository(db.Pool())
	employeeService := employees.NewService(employeeRepo, branchRepo, auditRepo, userRepo, membershipRepo)
	employeeHandler := employees.NewHandler(employeeService)

	vehicleRepo := vehicles.NewRepository(db.Pool())
	vehicleService := vehicles.NewService(vehicleRepo, branchRepo, employeeRepo, auditRepo)
	vehicleHandler := vehicles.NewHandler(vehicleService)

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
		BranchHandler:    branchHandler,
		EmployeeHandler:  employeeHandler,
		VehicleHandler:   vehicleHandler,
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

func TestAuthAPI_TokenRefresh_Rotation_And_BreachDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping token refresh integration test in short mode")
	}

	router, _ := setupTestRouter(t)

	// 1. Register a new user and company
	uniqueSuffix := time.Now().UnixNano()
	email := fmt.Sprintf("refresh_tester_%d@example.com", uniqueSuffix)
	password := "SecureP@ssw0rd!2026"
	company := fmt.Sprintf("Refresh Logistics %d", uniqueSuffix)

	regPayload := auth.RegisterRequest{
		Email:       email,
		Password:    password,
		FullName:    "Refresh Tester",
		CompanyName: company,
	}
	body, _ := json.Marshal(regPayload)

	wReg := httptest.NewRecorder()
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	reqReg.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("registration failed: %d %s", wReg.Code, wReg.Body.String())
	}

	var regEnv response.SuccessEnvelope
	_ = json.Unmarshal(wReg.Body.Bytes(), &regEnv)
	regBytes, _ := json.Marshal(regEnv.Data)
	var regResp auth.AuthResponse
	_ = json.Unmarshal(regBytes, &regResp)

	if regResp.RefreshToken == "" {
		t.Fatal("expected non-empty refresh token in registration response")
	}

	originalRefreshToken := regResp.RefreshToken

	// 2. Perform valid token refresh
	refreshPayload := auth.RefreshRequest{
		RefreshToken: originalRefreshToken,
	}
	rBody, _ := json.Marshal(refreshPayload)

	wRefresh := httptest.NewRecorder()
	reqRefresh := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(rBody))
	reqRefresh.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRefresh, reqRefresh)

	if wRefresh.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on token refresh, got %d: %s", wRefresh.Code, wRefresh.Body.String())
	}

	var refreshEnv response.SuccessEnvelope
	_ = json.Unmarshal(wRefresh.Body.Bytes(), &refreshEnv)
	refBytes, _ := json.Marshal(refreshEnv.Data)
	var rotatedResp auth.AuthResponse
	_ = json.Unmarshal(refBytes, &rotatedResp)

	if rotatedResp.Token == "" {
		t.Error("expected non-empty new access token")
	}
	if rotatedResp.RefreshToken == "" {
		t.Error("expected non-empty rotated refresh token")
	}
	if rotatedResp.RefreshToken == originalRefreshToken {
		t.Error("expected rotated refresh token to differ from original (single-use rotation)")
	}

	// 3. Breach Detection: Attempting to reuse the revoked original refresh token
	wReplay := httptest.NewRecorder()
	reqReplay := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(rBody))
	reqReplay.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wReplay, reqReplay)

	if wReplay.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for revoked token replay, got %d: %s", wReplay.Code, wReplay.Body.String())
	}

	// 4. Verify that breach detection revoked all tokens for the user,
	// so the newer rotated token should now also be rejected
	rotatedPayload := auth.RefreshRequest{
		RefreshToken: rotatedResp.RefreshToken,
	}
	rotBody, _ := json.Marshal(rotatedPayload)

	wRotatedReplay := httptest.NewRecorder()
	reqRotatedReplay := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(rotBody))
	reqRotatedReplay.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRotatedReplay, reqRotatedReplay)

	if wRotatedReplay.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for rotated token after breach detection cascade, got %d", wRotatedReplay.Code)
	}
}

func TestAuthAPI_Logout_Revocation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping logout revocation integration test in short mode")
	}

	router, _ := setupTestRouter(t)

	// Register user
	uniqueSuffix := time.Now().UnixNano()
	email := fmt.Sprintf("logout_tester_%d@example.com", uniqueSuffix)
	password := "SecureP@ssw0rd!2026"
	company := fmt.Sprintf("Logout Corp %d", uniqueSuffix)

	regPayload := auth.RegisterRequest{
		Email:       email,
		Password:    password,
		FullName:    "Logout Tester",
		CompanyName: company,
	}
	body, _ := json.Marshal(regPayload)

	wReg := httptest.NewRecorder()
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	reqReg.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wReg, reqReg)

	var regEnv response.SuccessEnvelope
	_ = json.Unmarshal(wReg.Body.Bytes(), &regEnv)
	regBytes, _ := json.Marshal(regEnv.Data)
	var regResp auth.AuthResponse
	_ = json.Unmarshal(regBytes, &regResp)

	// Call POST /api/v1/auth/logout with the refresh token
	logoutPayload := auth.LogoutRequest{
		RefreshToken: regResp.RefreshToken,
	}
	lBody, _ := json.Marshal(logoutPayload)

	wLogout := httptest.NewRecorder()
	reqLogout := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(lBody))
	reqLogout.Header.Set("Authorization", "Bearer "+regResp.Token)
	reqLogout.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wLogout, reqLogout)

	if wLogout.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on logout, got %d: %s", wLogout.Code, wLogout.Body.String())
	}

	// Verify that the logged-out refresh token can no longer be used to refresh
	refreshPayload := auth.RefreshRequest{
		RefreshToken: regResp.RefreshToken,
	}
	rBody, _ := json.Marshal(refreshPayload)

	wRefresh := httptest.NewRecorder()
	reqRefresh := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(rBody))
	reqRefresh.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRefresh, reqRefresh)

	if wRefresh.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized on refreshing with logged-out token, got %d", wRefresh.Code)
	}
}
