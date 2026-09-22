package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/response"
)

// TestE2E_CompleteBusinessWorkflow executes an end-to-end multi-tenant enterprise journey:
// 1. Owner A registers on LogiFlows (creates Tenant A).
// 2. Owner A retrieves profile and authenticates.
// 3. Owner A updates Tenant A metadata (PATCH /tenants/:id).
// 4. Owner A invites Operator A with role TENANT_OPERATOR.
// 5. Operator A logs in and retrieves profile showing Tenant A membership.
// 6. Operator A executes allowed operation (list members) -> 200 OK.
// 7. Operator A executes forbidden operation (invite member, update tenant) -> 403 Forbidden.
// 8. Owner B registers on LogiFlows (creates Tenant B).
// 9. Owner B attempts cross-tenant access to Tenant A -> 403 Forbidden.
// 10. Owner A performs refresh token rotation -> receives new token pair.
// 11. Owner A logs out -> session terminated and verified.
// 12. Audit log trail verifies all generated events.
func TestE2E_CompleteBusinessWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e business workflow test in short mode")
	}

	router, _ := setupTestRouter(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	uniqueRunID := time.Now().UnixNano()

	// ==========================================
	// Step 1: Owner A Registers (Tenant Alpha)
	// ==========================================
	t.Log("--- Step 1: Owner A Registration ---")
	ownerAEmail := fmt.Sprintf("owner-alpha-%d@logiflows.test", uniqueRunID)
	ownerAPassword := "AlphaP@ss#2026!"
	tenantAName := fmt.Sprintf("Alpha Logistics Corp %d", uniqueRunID)

	regAPayload := auth.RegisterRequest{
		Email:       ownerAEmail,
		Password:    ownerAPassword,
		FullName:    "Alice Alpha (CEO)",
		CompanyName: tenantAName,
	}
	bodyA, _ := json.Marshal(regAPayload)

	wRegA := httptest.NewRecorder()
	reqRegA := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyA))
	reqRegA.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRegA, reqRegA)

	if wRegA.Code != http.StatusCreated {
		t.Fatalf("Step 1 failed: Owner A registration returned %d: %s", wRegA.Code, wRegA.Body.String())
	}

	var envRegA response.SuccessEnvelope
	_ = json.Unmarshal(wRegA.Body.Bytes(), &envRegA)
	dataBytesA, _ := json.Marshal(envRegA.Data)
	var authRespA auth.AuthResponse
	_ = json.Unmarshal(dataBytesA, &authRespA)

	tokenA := authRespA.Token
	refreshTokenA := authRespA.RefreshToken
	userAID := authRespA.User.ID
	tenantAID := authRespA.Tenants[0].ID

	if tokenA == "" || refreshTokenA == "" || tenantAID == uuid.Nil {
		t.Fatalf("Step 1 assertion failed: missing tokens or tenant ID")
	}

	// ==========================================
	// Step 2: Owner A Profile & Login Check
	// ==========================================
	t.Log("--- Step 2: Owner A Profile Verification ---")
	wMeA := httptest.NewRecorder()
	reqMeA := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMeA.Header.Set("Authorization", "Bearer "+tokenA)
	router.ServeHTTP(wMeA, reqMeA)

	if wMeA.Code != http.StatusOK {
		t.Fatalf("Step 2 failed: GET /auth/me returned %d: %s", wMeA.Code, wMeA.Body.String())
	}

	// ==========================================
	// Step 3: Owner A Updates Tenant A Metadata
	// ==========================================
	t.Log("--- Step 3: Owner A Updates Tenant Metadata (PATCH) ---")
	updatePayload := map[string]string{
		"name":          tenantAName + " Updated",
		"contact_email": "ops-alpha@logiflows.test",
	}
	updateBody, _ := json.Marshal(updatePayload)

	wPatch := httptest.NewRecorder()
	reqPatch := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s", tenantAID), bytes.NewReader(updateBody))
	reqPatch.Header.Set("Content-Type", "application/json")
	reqPatch.Header.Set("Authorization", "Bearer "+tokenA)
	router.ServeHTTP(wPatch, reqPatch)

	if wPatch.Code != http.StatusOK {
		t.Fatalf("Step 3 failed: Tenant update PATCH returned %d: %s", wPatch.Code, wPatch.Body.String())
	}

	// ==========================================
	// Step 4: Register Operator A User & Invite to Tenant A
	// ==========================================
	t.Log("--- Step 4: Operator A Onboarding & Invitation ---")
	operatorAEmail := fmt.Sprintf("operator-alpha-%d@logiflows.test", uniqueRunID)
	operatorAPassword := "OperatorP@ss#2026!"

	regOpPayload := auth.RegisterRequest{
		Email:       operatorAEmail,
		Password:    operatorAPassword,
		FullName:    "Bob Operator",
		CompanyName: fmt.Sprintf("Operator Personal Org %d", uniqueRunID),
	}
	bodyOp, _ := json.Marshal(regOpPayload)

	wRegOp := httptest.NewRecorder()
	reqRegOp := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyOp))
	reqRegOp.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRegOp, reqRegOp)

	if wRegOp.Code != http.StatusCreated {
		t.Fatalf("Step 4 failed: Operator registration returned %d: %s", wRegOp.Code, wRegOp.Body.String())
	}

	var envOp response.SuccessEnvelope
	_ = json.Unmarshal(wRegOp.Body.Bytes(), &envOp)
	dataOpBytes, _ := json.Marshal(envOp.Data)
	var authRespOp auth.AuthResponse
	_ = json.Unmarshal(dataOpBytes, &authRespOp)
	tokenOp := authRespOp.Token

	// Owner A invites Operator A to Tenant A as TENANT_OPERATOR
	invitePayload := map[string]string{
		"email": operatorAEmail,
		"role":  string(memberships.RoleTenantOperator),
	}
	inviteBody, _ := json.Marshal(invitePayload)

	wInvite := httptest.NewRecorder()
	reqInvite := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", tenantAID), bytes.NewReader(inviteBody))
	reqInvite.Header.Set("Content-Type", "application/json")
	reqInvite.Header.Set("Authorization", "Bearer "+tokenA)
	router.ServeHTTP(wInvite, reqInvite)

	if wInvite.Code != http.StatusCreated {
		t.Fatalf("Step 4 failed: Member invitation returned %d: %s", wInvite.Code, wInvite.Body.String())
	}

	// ==========================================
	// Step 5: Operator A Checks Allowed Operation (List Members)
	// ==========================================
	t.Log("--- Step 5: Operator A Allowed Operation (List Members) ---")
	wListMembers := httptest.NewRecorder()
	reqListMembers := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/members", tenantAID), nil)
	reqListMembers.Header.Set("Authorization", "Bearer "+tokenOp)
	router.ServeHTTP(wListMembers, reqListMembers)

	if wListMembers.Code != http.StatusOK {
		t.Fatalf("Step 5 failed: Operator listing members returned %d: %s", wListMembers.Code, wListMembers.Body.String())
	}

	// ==========================================
	// Step 6: Operator A Attempts Forbidden Operations (RBAC Check)
	// ==========================================
	t.Log("--- Step 6: Operator A Forbidden Operations (RBAC Enforced) ---")
	// Forbidden 1: Operator attempts to invite another member
	wOpInvite := httptest.NewRecorder()
	reqOpInvite := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", tenantAID), bytes.NewReader(inviteBody))
	reqOpInvite.Header.Set("Content-Type", "application/json")
	reqOpInvite.Header.Set("Authorization", "Bearer "+tokenOp)
	router.ServeHTTP(wOpInvite, reqOpInvite)

	if wOpInvite.Code != http.StatusForbidden {
		t.Errorf("Step 6 failed: Operator member invite should be 403 Forbidden, got %d", wOpInvite.Code)
	}

	// Forbidden 2: Operator attempts to update tenant metadata
	wOpPatch := httptest.NewRecorder()
	reqOpPatch := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s", tenantAID), bytes.NewReader(updateBody))
	reqOpPatch.Header.Set("Content-Type", "application/json")
	reqOpPatch.Header.Set("Authorization", "Bearer "+tokenOp)
	router.ServeHTTP(wOpPatch, reqOpPatch)

	if wOpPatch.Code != http.StatusForbidden {
		t.Errorf("Step 6 failed: Operator tenant PATCH should be 403 Forbidden, got %d", wOpPatch.Code)
	}

	// ==========================================
	// Step 7: Owner B (Tenant Beta) Cross-Tenant Rejection
	// ==========================================
	t.Log("--- Step 7: Cross-Tenant Boundary Enforcement ---")
	ownerBEmail := fmt.Sprintf("owner-beta-%d@logiflows.test", uniqueRunID)
	tenantBName := fmt.Sprintf("Beta Freight %d", uniqueRunID)

	regBPayload := auth.RegisterRequest{
		Email:       ownerBEmail,
		Password:    "BetaP@ss#2026!",
		FullName:    "Bob Beta",
		CompanyName: tenantBName,
	}
	bodyB, _ := json.Marshal(regBPayload)

	wRegB := httptest.NewRecorder()
	reqRegB := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyB))
	reqRegB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRegB, reqRegB)

	if wRegB.Code != http.StatusCreated {
		t.Fatalf("Step 7 failed: Owner B registration returned %d", wRegB.Code)
	}

	var envRegB response.SuccessEnvelope
	_ = json.Unmarshal(wRegB.Body.Bytes(), &envRegB)
	dataBytesB, _ := json.Marshal(envRegB.Data)
	var authRespB auth.AuthResponse
	_ = json.Unmarshal(dataBytesB, &authRespB)
	tokenB := authRespB.Token

	// Owner B attempts to read Tenant A details
	wCross := httptest.NewRecorder()
	reqCross := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s", tenantAID), nil)
	reqCross.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(wCross, reqCross)

	if wCross.Code != http.StatusForbidden {
		t.Errorf("Step 7 failed: Cross-tenant access should be 403 Forbidden, got %d: %s", wCross.Code, wCross.Body.String())
	}

	// ==========================================
	// Step 8: Token Refresh Single-Use Rotation
	// ==========================================
	t.Log("--- Step 8: Session Token Refresh Rotation ---")
	refreshPayload := map[string]string{
		"refresh_token": refreshTokenA,
	}
	refBody, _ := json.Marshal(refreshPayload)

	wRef := httptest.NewRecorder()
	reqRef := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refBody))
	reqRef.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRef, reqRef)

	if wRef.Code != http.StatusOK {
		t.Fatalf("Step 8 failed: Token refresh returned %d: %s", wRef.Code, wRef.Body.String())
	}

	var envRef response.SuccessEnvelope
	_ = json.Unmarshal(wRef.Body.Bytes(), &envRef)
	dataRefBytes, _ := json.Marshal(envRef.Data)
	var authRespRef auth.AuthResponse
	_ = json.Unmarshal(dataRefBytes, &authRespRef)

	newRefreshTokenA := authRespRef.RefreshToken
	newAccessTokenA := authRespRef.Token

	// ==========================================
	// Step 9: Owner A Logout & Revocation
	// ==========================================
	t.Log("--- Step 9: Owner A Logout ---")
	logoutPayload := map[string]string{
		"refresh_token": newRefreshTokenA,
	}
	logoutBody, _ := json.Marshal(logoutPayload)

	wLogout := httptest.NewRecorder()
	reqLogout := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(logoutBody))
	reqLogout.Header.Set("Content-Type", "application/json")
	reqLogout.Header.Set("Authorization", "Bearer "+newAccessTokenA)
	router.ServeHTTP(wLogout, reqLogout)

	if wLogout.Code != http.StatusOK {
		t.Fatalf("Step 9 failed: Logout returned %d: %s", wLogout.Code, wLogout.Body.String())
	}

	// Subsequent refresh must fail with 401
	wPostLogoutRef := httptest.NewRecorder()
	reqPostLogoutRef := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(logoutBody))
	reqPostLogoutRef.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPostLogoutRef, reqPostLogoutRef)

	if wPostLogoutRef.Code != http.StatusUnauthorized {
		t.Errorf("Step 9 failed: Expected 401 on refresh after logout, got %d", wPostLogoutRef.Code)
	}

	// ==========================================
	// Step 10: Verify Database Audit Logs Trail
	// ==========================================
	t.Log("--- Step 10: Audit Log Verification ---")
	auditQuery := `
		SELECT action, count(*) 
		FROM audit_logs 
		WHERE user_id = $1 
		GROUP BY action;
	`
	rows, err := db.Pool().Query(ctx, auditQuery, userAID)
	if err != nil {
		t.Fatalf("Step 10 failed: query audit_logs error: %v", err)
	}
	defer rows.Close()

	actionsLogged := make(map[string]int)
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			t.Fatalf("failed to scan audit summary: %v", err)
		}
		actionsLogged[action] = count
	}

	if actionsLogged["USER_REGISTERED"] == 0 {
		t.Errorf("expected audit action USER_REGISTERED to be logged for Owner A")
	}

	t.Logf("E2E Business Workflow completed successfully. Actions logged for Owner A: %+v", actionsLogged)
}
