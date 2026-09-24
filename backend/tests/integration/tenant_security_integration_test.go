package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/tenants"
)

func registerTestCompany(t *testing.T, router http.Handler, name, emailPrefix string) (string, uuid.UUID, uuid.UUID) {
	uniqueSuffix := time.Now().UnixNano()
	email := fmt.Sprintf("%s_%d@example.com", emailPrefix, uniqueSuffix)
	password := "SecureP@ssw0rd!2026"
	company := fmt.Sprintf("%s Corp %d", name, uniqueSuffix)

	payload := auth.RegisterRequest{
		Email:       email,
		Password:    password,
		FullName:    name + " Admin",
		CompanyName: company,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to register test company %s: %d %s", name, w.Code, w.Body.String())
	}

	var env response.SuccessEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	dataBytes, _ := json.Marshal(env.Data)
	var resp auth.AuthResponse
	_ = json.Unmarshal(dataBytes, &resp)

	return resp.Token, resp.User.ID, resp.Tenants[0].ID
}

func TestSecurity_CrossTenantAccess_Forbidden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cross-tenant security test in short mode")
	}

	router, _ := setupTestRouter(t)

	// Register Company Alpha
	tokenAlpha, _, tenantIDAlpha := registerTestCompany(t, router, "Alpha", "admin_alpha")

	// Register Company Beta
	tokenBeta, _, tenantIDBeta := registerTestCompany(t, router, "Beta", "admin_beta")

	// 1. Alpha Admin accesses Alpha Tenant (Should succeed with 200 OK)
	wAlphaSelf := httptest.NewRecorder()
	reqAlphaSelf := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s", tenantIDAlpha), nil)
	reqAlphaSelf.Header.Set("Authorization", "Bearer "+tokenAlpha)
	router.ServeHTTP(wAlphaSelf, reqAlphaSelf)

	if wAlphaSelf.Code != http.StatusOK {
		t.Fatalf("expected 200 OK accessing self tenant, got %d: %s", wAlphaSelf.Code, wAlphaSelf.Body.String())
	}

	// 2. Alpha Admin attempts to access Beta Tenant resources (Cross-Tenant Attack)
	// Must be rejected with 403 Forbidden!
	wCross := httptest.NewRecorder()
	reqCross := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s", tenantIDBeta), nil)
	reqCross.Header.Set("Authorization", "Bearer "+tokenAlpha)
	router.ServeHTTP(wCross, reqCross)

	if wCross.Code != http.StatusForbidden {
		t.Errorf("SECURITY BREACH: Alpha user was able to query Beta tenant! Expected 403 Forbidden, got %d", wCross.Code)
	}

	var errResp response.ErrorEnvelope
	_ = json.Unmarshal(wCross.Body.Bytes(), &errResp)
	if errResp.Error.Code != "CROSS_TENANT_ACCESS_DENIED" {
		t.Errorf("expected error code CROSS_TENANT_ACCESS_DENIED, got: %s", errResp.Error.Code)
	}

	// 3. Beta Admin attempts to list Alpha Tenant members
	// Must be rejected with 403 Forbidden!
	wCrossMembers := httptest.NewRecorder()
	reqCrossMembers := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/members", tenantIDAlpha), nil)
	reqCrossMembers.Header.Set("Authorization", "Bearer "+tokenBeta)
	router.ServeHTTP(wCrossMembers, reqCrossMembers)

	if wCrossMembers.Code != http.StatusForbidden {
		t.Errorf("SECURITY BREACH: Beta user was able to query Alpha members! Expected 403 Forbidden, got %d", wCrossMembers.Code)
	}
}

func TestSecurity_RBAC_RolePermissionEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping RBAC role enforcement test in short mode")
	}

	router, _ := setupTestRouter(t)

	// 1. Create company with Tenant Admin
	tokenAdmin, _, tenantID := registerTestCompany(t, router, "LogiAdminCo", "admin_logi")

	// 2. Create second user to be added as an Operator
	uniqueSuffix := time.Now().UnixNano()
	operatorEmail := fmt.Sprintf("operator_%d@logiadmin.com", uniqueSuffix)
	operatorPass := "SecureP@ssw0rd!2026"

	// Register operator with their own dummy registration first
	tokenOp, opUserID, _ := registerTestCompany(t, router, "OpDummy", "dummy_op")
	_ = opUserID

	// 3. Admin adds Operator to LogiAdminCo with TENANT_OPERATOR role
	addPayload := tenants.AddMemberRequest{
		Email: operatorEmail,
		Role:  memberships.RoleTenantOperator,
	}
	addBody, _ := json.Marshal(addPayload)

	wAdd := httptest.NewRecorder()
	reqAdd := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", tenantID), bytes.NewReader(addBody))
	reqAdd.Header.Set("Authorization", "Bearer "+tokenAdmin)
	reqAdd.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wAdd, reqAdd)

	// Note: operator was registered with dummy company, so their email exists in users table!
	// Let's create user with operatorEmail directly to be sure:
	// We'll register them through the API:
	regOpPayload := auth.RegisterRequest{
		Email:       operatorEmail,
		Password:    operatorPass,
		FullName:    "Bob Operator",
		CompanyName: fmt.Sprintf("Operator Co %d", uniqueSuffix),
	}
	regOpBody, _ := json.Marshal(regOpPayload)
	wRegOp := httptest.NewRecorder()
	reqRegOp := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regOpBody))
	reqRegOp.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRegOp, reqRegOp)

	var regOpEnv response.SuccessEnvelope
	_ = json.Unmarshal(wRegOp.Body.Bytes(), &regOpEnv)
	regOpDataBytes, _ := json.Marshal(regOpEnv.Data)
	var regOpResp auth.AuthResponse
	_ = json.Unmarshal(regOpDataBytes, &regOpResp)
	tokenBob := regOpResp.Token

	// Admin adds Bob to LogiAdminCo as TENANT_OPERATOR
	wAddReal := httptest.NewRecorder()
	reqAddReal := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", tenantID), bytes.NewReader(addBody))
	reqAddReal.Header.Set("Authorization", "Bearer "+tokenAdmin)
	reqAddReal.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wAddReal, reqAddReal)

	if wAddReal.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created adding operator, got %d: %s", wAddReal.Code, wAddReal.Body.String())
	}

	// 4. Bob (TENANT_OPERATOR) lists members of LogiAdminCo -> Allowed (200 OK)
	wListMembers := httptest.NewRecorder()
	reqListMembers := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/members", tenantID), nil)
	reqListMembers.Header.Set("Authorization", "Bearer "+tokenBob)
	router.ServeHTTP(wListMembers, reqListMembers)

	if wListMembers.Code != http.StatusOK {
		t.Errorf("expected 200 OK for operator listing members, got %d: %s", wListMembers.Code, wListMembers.Body.String())
	}

	// 5. Bob (TENANT_OPERATOR) attempts to add another member -> MUST BE REJECTED with 403 Forbidden!
	// (Only TENANT_ADMIN has permission to add members)
	invPayload := tenants.AddMemberRequest{
		Email: "someone_else@logiadmin.com",
		Role:  memberships.RoleViewer,
	}
	invBody, _ := json.Marshal(invPayload)

	wOpAddMember := httptest.NewRecorder()
	reqOpAddMember := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", tenantID), bytes.NewReader(invBody))
	reqOpAddMember.Header.Set("Authorization", "Bearer "+tokenBob)
	reqOpAddMember.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wOpAddMember, reqOpAddMember)

	if wOpAddMember.Code != http.StatusForbidden {
		t.Errorf("SECURITY BREACH: Operator was able to execute Admin operation! Expected 403 Forbidden, got %d", wOpAddMember.Code)
	}

	_ = tokenOp
}

func TestSecurity_InvalidTenantID_Rejected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping invalid tenant ID test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, _ := registerTestCompany(t, router, "TestInv", "admin_inv")

	// Non-existent UUID
	randomUUID := uuid.New().String()
	wNotFound := httptest.NewRecorder()
	reqNotFound := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s", randomUUID), nil)
	reqNotFound.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for non-existent tenant, got %d", wNotFound.Code)
	}

	// Malformed UUID string
	wBadUUID := httptest.NewRecorder()
	reqBadUUID := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/not-a-valid-uuid", nil)
	reqBadUUID.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wBadUUID, reqBadUUID)

	if wBadUUID.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for malformed UUID, got %d", wBadUUID.Code)
	}
}

func TestSecurity_TenantUpdate_PATCH(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tenant update security test in short mode")
	}

	router, _ := setupTestRouter(t)

	// Register Company Gamma
	tokenAdminGamma, _, tenantIDGamma := registerTestCompany(t, router, "Gamma", "admin_gamma")
	// Register Company Delta
	tokenAdminDelta, _, _ := registerTestCompany(t, router, "Delta", "admin_delta")

	// 1. Valid update by Gamma Admin
	newName := "Gamma Global Logistics"
	newEmail := "ops@gammaglobal.com"
	updatePayload := tenants.UpdateTenantRequest{
		Name:         &newName,
		ContactEmail: &newEmail,
	}
	upBody, _ := json.Marshal(updatePayload)

	wValid := httptest.NewRecorder()
	reqValid := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s", tenantIDGamma), bytes.NewReader(upBody))
	reqValid.Header.Set("Authorization", "Bearer "+tokenAdminGamma)
	reqValid.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on tenant update, got %d: %s", wValid.Code, wValid.Body.String())
	}

	var updateEnv response.SuccessEnvelope
	_ = json.Unmarshal(wValid.Body.Bytes(), &updateEnv)
	upBytes, _ := json.Marshal(updateEnv.Data)
	var updatedTenant tenants.Tenant
	_ = json.Unmarshal(upBytes, &updatedTenant)

	if updatedTenant.Name != newName {
		t.Errorf("expected updated name %s, got %s", newName, updatedTenant.Name)
	}
	if updatedTenant.ContactEmail != newEmail {
		t.Errorf("expected updated email %s, got %s", newEmail, updatedTenant.ContactEmail)
	}

	// 2. Cross-tenant update attempt: Delta Admin attempts to PATCH Gamma's tenant
	wCross := httptest.NewRecorder()
	reqCross := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s", tenantIDGamma), bytes.NewReader(upBody))
	reqCross.Header.Set("Authorization", "Bearer "+tokenAdminDelta)
	reqCross.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCross, reqCross)

	if wCross.Code != http.StatusForbidden {
		t.Errorf("SECURITY BREACH: Cross-tenant PATCH succeeded! Expected 403 Forbidden, got %d", wCross.Code)
	}
}
