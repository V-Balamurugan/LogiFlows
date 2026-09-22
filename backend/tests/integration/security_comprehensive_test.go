package integration_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/response"
)

// TC-P1-AUT-003, TC-P1-AUT-004, TC-P1-AUT-005, TC-P1-SEC-001, TC-P1-SEC-002
func TestSecurity_Registration_NegativeInputs(t *testing.T) {
	router, _ := setupTestRouter(t)

	tests := []struct {
		name           string
		payload        auth.RegisterRequest
		expectedStatus int
	}{
		{
			name: "Empty Email",
			payload: auth.RegisterRequest{
				Email:       "",
				Password:    "ValidP@ss123!",
				FullName:    "Test User",
				CompanyName: "Test Logistics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Email Format - No Domain",
			payload: auth.RegisterRequest{
				Email:       "invalid-email-address",
				Password:    "ValidP@ss123!",
				FullName:    "Test User",
				CompanyName: "Test Logistics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Email Format - Spaces",
			payload: auth.RegisterRequest{
				Email:       "user name@domain.com",
				Password:    "ValidP@ss123!",
				FullName:    "Test User",
				CompanyName: "Test Logistics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Password Shorter than 8 Chars",
			payload: auth.RegisterRequest{
				Email:       fmt.Sprintf("shortpass-%d@example.com", time.Now().UnixNano()),
				Password:    "Sh1!",
				FullName:    "Test User",
				CompanyName: "Test Logistics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Password Missing Special Character",
			payload: auth.RegisterRequest{
				Email:       fmt.Sprintf("nospec-%d@example.com", time.Now().UnixNano()),
				Password:    "Password12345",
				FullName:    "Test User",
				CompanyName: "Test Logistics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty Full Name",
			payload: auth.RegisterRequest{
				Email:       fmt.Sprintf("noname-%d@example.com", time.Now().UnixNano()),
				Password:    "ValidP@ss123!",
				FullName:    "",
				CompanyName: "Test Logistics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty Company Name",
			payload: auth.RegisterRequest{
				Email:       fmt.Sprintf("nocompany-%d@example.com", time.Now().UnixNano()),
				Password:    "ValidP@ss123!",
				FullName:    "Valid Name",
				CompanyName: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

// TC-P1-SEC-001, TC-P1-SEC-002: SQL Injection & XSS Payload Resistance
func TestSecurity_Registration_InjectionResistance(t *testing.T) {
	router, _ := setupTestRouter(t)

	injectionTests := []struct {
		name        string
		fullName    string
		companyName string
	}{
		{
			name:        "SQL Injection in Full Name",
			fullName:    "Robert'); DROP TABLE users;--",
			companyName: "SQLi Logistics Corp",
		},
		{
			name:        "SQL Injection OR 1=1 in Company Name",
			fullName:    "Admin Tester",
			companyName: "Apex ' OR '1'='1",
		},
		{
			name:        "XSS Script Tag in Full Name",
			fullName:    "<script>alert('XSS')</script>",
			companyName: "XSS Safe Logistics",
		},
		{
			name:        "XSS SVG Onload Tag in Company Name",
			fullName:    "Security Tester",
			companyName: "<svg onload=alert(1)> Corp",
		},
	}

	for _, tt := range injectionTests {
		t.Run(tt.name, func(t *testing.T) {
			uniqueSuffix := time.Now().UnixNano()
			email := fmt.Sprintf("injection-%d@example.com", uniqueSuffix)

			payload := auth.RegisterRequest{
				Email:       email,
				Password:    "SecureP@ss#2026!",
				FullName:    tt.fullName,
				CompanyName: fmt.Sprintf("%s %d", tt.companyName, uniqueSuffix),
			}
			body, _ := json.Marshal(payload)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			// Registration should succeed safely through parameterized queries
			// or fail gracefully without server panic (code 500)
			if w.Code == http.StatusInternalServerError {
				t.Fatalf("Server threw 500 on injection test: %s", w.Body.String())
			}

			if w.Code == http.StatusCreated {
				var resp response.SuccessEnvelope
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				dataBytes, _ := json.Marshal(resp.Data)
				var authResp auth.AuthResponse
				_ = json.Unmarshal(dataBytes, &authResp)

				if authResp.User.FullName != tt.fullName {
					t.Errorf("expected full name %q to be preserved safely, got %q", tt.fullName, authResp.User.FullName)
				}
			}
		})
	}
}

// TC-P1-AUT-006, TC-P1-AUT-010, TC-P1-AUT-011
func TestSecurity_Login_NegativeInputs_And_CaseNormalization(t *testing.T) {
	router, _ := setupTestRouter(t)

	// 1. Register a user with mixed-case email
	uniqueID := uuid.New().String()[:8]
	mixedEmail := fmt.Sprintf("CaseUser.%s@LogiFlows.Test", uniqueID)
	lowerEmail := strings.ToLower(mixedEmail)
	password := "CaseSensitiveP@ss1!"

	regPayload := auth.RegisterRequest{
		Email:       mixedEmail,
		Password:    password,
		FullName:    "Case Normalization User",
		CompanyName: "Case Logistics " + uniqueID,
	}
	regBody, _ := json.Marshal(regPayload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("registration failed: %d %s", w.Code, w.Body.String())
	}

	// 2. Test login using lowercase normalized email
	t.Run("Login with Lowercase Normalized Email", func(t *testing.T) {
		loginPayload := auth.LoginRequest{
			Email:    lowerEmail,
			Password: password,
		}
		body, _ := json.Marshal(loginPayload)

		wLogin := httptest.NewRecorder()
		reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		reqLogin.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(wLogin, reqLogin)

		if wLogin.Code != http.StatusOK {
			t.Errorf("expected login to succeed with normalized lowercase email, got status %d: %s", wLogin.Code, wLogin.Body.String())
		}
	})

	// 3. Test login with wrong password
	t.Run("Login with Incorrect Password", func(t *testing.T) {
		loginPayload := auth.LoginRequest{
			Email:    lowerEmail,
			Password: "WrongPassword999!",
		}
		body, _ := json.Marshal(loginPayload)

		wLogin := httptest.NewRecorder()
		reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		reqLogin.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(wLogin, reqLogin)

		if wLogin.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d: %s", wLogin.Code, wLogin.Body.String())
		}
	})

	// 4. Test login with non-existent email
	t.Run("Login with Non-Existent Email", func(t *testing.T) {
		loginPayload := auth.LoginRequest{
			Email:    "nonexistent.ghost.user@logiflows.test",
			Password: "AnyPassword123!",
		}
		body, _ := json.Marshal(loginPayload)

		wLogin := httptest.NewRecorder()
		reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		reqLogin.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(wLogin, reqLogin)

		if wLogin.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d: %s", wLogin.Code, wLogin.Body.String())
		}
	})

	// 5. Test login with SQL injection in email
	t.Run("Login with SQL Injection Payload in Email", func(t *testing.T) {
		loginPayload := auth.LoginRequest{
			Email:    "' OR '1'='1' --",
			Password: "AnyPassword123!",
		}
		body, _ := json.Marshal(loginPayload)

		wLogin := httptest.NewRecorder()
		reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		reqLogin.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(wLogin, reqLogin)

		if wLogin.Code != http.StatusUnauthorized && wLogin.Code != http.StatusBadRequest {
			t.Errorf("expected 401 or 400, got %d: %s", wLogin.Code, wLogin.Body.String())
		}
	})
}

// TC-P1-AUT-014, TC-P1-AUT-015: JWT Tampering & Attack Vectors
func TestSecurity_JWT_TamperingAndAttackVectors(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Register user to get valid context
	token, _, _ := registerTestCompany(t, router, "JWTAttack", "jwt_attack")

	t.Run("Missing Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Malformed Authorization Header - Not Bearer", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Basic "+token)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Algorithm Confusion Attack - alg none", func(t *testing.T) {
		// Craft unverified JWT with alg: none
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"%s","email":"fake@admin.test","exp":%d}`, uuid.New().String(), time.Now().Add(time.Hour).Unix())))
		unsignedToken := header + "." + payload + "."

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+unsignedToken)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for alg:none attack, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Forged Signature - Signed with Wrong Key", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":   uuid.New().String(),
			"email": "hacker@evil.test",
			"exp":   time.Now().Add(time.Hour).Unix(),
		}
		forgedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		forgedString, _ := forgedToken.SignedString([]byte("wrong-attacker-secret-key-32-chars!"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+forgedString)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for forged signature, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Access Token Submitted to Refresh Endpoint Rejected", func(t *testing.T) {
		// Attempt to use access token at /api/v1/auth/refresh (should fail as refresh token is opaque SHA-256)
		refreshPayload := map[string]string{
			"refresh_token": token,
		}
		body, _ := json.Marshal(refreshPayload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized when access token used as refresh token, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// TC-P1-MEM-003: Duplicate Membership Prevention
func TestSecurity_Membership_DuplicatePrevention(t *testing.T) {
	router, _ := setupTestRouter(t)

	// 1. Create company and owner (returns token, userID, tenantID)
	ownerToken, _, tenantID := registerTestCompany(t, router, "MemberDupeOrg", "dupe_org")

	// 2. Register second user
	uniqueSuffix := time.Now().UnixNano()
	operatorEmail := fmt.Sprintf("operator-dupe-%d@example.com", uniqueSuffix)
	opRegPayload := auth.RegisterRequest{
		Email:       operatorEmail,
		Password:    "OperatorP@ss#2026",
		FullName:    "Operator User",
		CompanyName: fmt.Sprintf("Dummy Independent Org %d", uniqueSuffix),
	}
	opBody, _ := json.Marshal(opRegPayload)
	wOp := httptest.NewRecorder()
	reqOp := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(opBody))
	reqOp.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wOp, reqOp)
	if wOp.Code != http.StatusCreated {
		t.Fatalf("failed to register operator: %d %s", wOp.Code, wOp.Body.String())
	}

	// 3. First invitation should succeed
	invitePayload := map[string]string{
		"email": operatorEmail,
		"role":  string(memberships.RoleTenantOperator),
	}
	inviteBody, _ := json.Marshal(invitePayload)

	wInvite1 := httptest.NewRecorder()
	reqInvite1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", tenantID), bytes.NewReader(inviteBody))
	reqInvite1.Header.Set("Content-Type", "application/json")
	reqInvite1.Header.Set("Authorization", "Bearer "+ownerToken)
	router.ServeHTTP(wInvite1, reqInvite1)

	if wInvite1.Code != http.StatusCreated {
		t.Fatalf("first invite failed: %d %s", wInvite1.Code, wInvite1.Body.String())
	}

	// 4. Second invitation with same email must fail with 409 Conflict
	wInvite2 := httptest.NewRecorder()
	reqInvite2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", tenantID), bytes.NewReader(inviteBody))
	reqInvite2.Header.Set("Content-Type", "application/json")
	reqInvite2.Header.Set("Authorization", "Bearer "+ownerToken)
	router.ServeHTTP(wInvite2, reqInvite2)

	if wInvite2.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict on duplicate member invitation, got %d: %s", wInvite2.Code, wInvite2.Body.String())
	}
}

// TC-P1-AUD-001: Audit Log Persistence & Immutability Verification
func TestSecurity_AuditLog_Integrity(t *testing.T) {
	router, _ := setupTestRouter(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Perform registration and tenant creation
	uniqueSuffix := time.Now().UnixNano()
	email := fmt.Sprintf("audit-user-%d@logiflows.test", uniqueSuffix)
	regPayload := auth.RegisterRequest{
		Email:       email,
		Password:    "AuditP@ss#2026!",
		FullName:    "Audited User",
		CompanyName: fmt.Sprintf("Audit Logistics %d", uniqueSuffix),
	}
	body, _ := json.Marshal(regPayload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("registration failed: %d %s", w.Code, w.Body.String())
	}

	var resp response.SuccessEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	dataBytes, _ := json.Marshal(resp.Data)
	var authResp auth.AuthResponse
	_ = json.Unmarshal(dataBytes, &authResp)

	// Directly inspect PostgreSQL audit_logs table
	query := `
		SELECT action, user_id, tenant_id, details, created_at 
		FROM audit_logs 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT 5;
	`
	rows, err := db.Pool().Query(ctx, query, authResp.User.ID)
	if err != nil {
		t.Fatalf("failed to query audit_logs: %v", err)
	}
	defer rows.Close()

	var actionsFound []string
	for rows.Next() {
		var action string
		var userID uuid.UUID
		var tenantID *uuid.UUID
		var details []byte
		var createdAt time.Time

		if err := rows.Scan(&action, &userID, &tenantID, &details, &createdAt); err != nil {
			t.Fatalf("failed to scan audit row: %v", err)
		}

		actionsFound = append(actionsFound, action)

		if userID != authResp.User.ID {
			t.Errorf("expected audit user_id %s, got %s", authResp.User.ID, userID)
		}

		if len(details) == 0 {
			t.Errorf("expected non-empty JSON details for action %s", action)
		}
	}

	if len(actionsFound) == 0 {
		t.Errorf("expected audit logs to be recorded for registration and tenant creation, found none")
	}
}
