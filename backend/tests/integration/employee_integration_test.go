package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/tenants"
)

// TC-P2-EMP-001: Create Employee with Valid Role and Branch Association
func TestEmployee_Create_WithValidRoleAndBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Fleet Corp", "fleet_emp")

	// 1. Create a branch to assign the employee to
	branchPayload := branches.CreateBranchRequest{
		BranchCode:       fmt.Sprintf("BRN-%d", time.Now().UnixNano()%100000),
		Name:             "Gurgaon Delivery Hub",
		Address:          "Sector 18, Cyber City",
		City:             "Gurgaon",
		Latitude:         28.4595,
		Longitude:        77.0266,
		CoverageRadiusKM: 15,
	}
	bBody, _ := json.Marshal(branchPayload)
	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(bBody))
	reqB.Header.Set("Authorization", "Bearer "+token)
	reqB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusCreated {
		t.Fatalf("failed to create branch: %d %s", wB.Code, wB.Body.String())
	}
	var bResp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(wB.Body.Bytes(), &bResp)
	branchID := bResp.Data.ID

	// 2. Create Employee
	empCode := fmt.Sprintf("EMP-%d", time.Now().UnixNano()%100000)
	email := "driver.kumar@fleetcorp.com"
	phone := "+919876543210"
	license := "DL-04-2022-0091823"
	empPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    empCode,
		FirstName:       "Ramesh",
		LastName:        "Kumar",
		Email:           &email,
		Phone:           &phone,
		Designation:     "Senior Heavy Vehicle Driver",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
		LicenseNumber:   &license,
		BranchID:        &branchID,
	}
	eBody, _ := json.Marshal(empPayload)
	wE := httptest.NewRecorder()
	reqE := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(eBody))
	reqE.Header.Set("Authorization", "Bearer "+token)
	reqE.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wE, reqE)

	if wE.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for employee, got: %d %s", wE.Code, wE.Body.String())
	}

	var eResp struct {
		Data employees.Employee `json:"data"`
	}
	if err := json.Unmarshal(wE.Body.Bytes(), &eResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if eResp.Data.EmployeeCode != empCode {
		t.Errorf("expected employee code %s, got %s", empCode, eResp.Data.EmployeeCode)
	}
	if eResp.Data.OperationalRole != employees.OperationalRoleDriver {
		t.Errorf("expected role DRIVER, got %s", eResp.Data.OperationalRole)
	}
	if eResp.Data.BranchName == nil || *eResp.Data.BranchName != "Gurgaon Delivery Hub" {
		t.Errorf("expected branch name 'Gurgaon Delivery Hub', got %v", eResp.Data.BranchName)
	}
}

// TC-P2-EMP-002: Duplicate Employee Code Rejected Within Tenant
func TestEmployee_DuplicateCode_RejectedWithinTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Swift Delivery", "swift_emp")

	empCode := fmt.Sprintf("DUP-%d", time.Now().UnixNano()%100000)
	empPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    empCode,
		FirstName:       "Amit",
		LastName:        "Verma",
		Designation:     "Delivery Associate",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
	}
	body, _ := json.Marshal(empPayload)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(body))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first creation expected 201, got: %d", w1.Code)
	}

	// Attempt duplicate
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate employee code, got: %d %s", w2.Code, w2.Body.String())
	}
}

// TC-P2-EMP-003: Identical Employee Code Allowed In Different Tenants
func TestEmployee_IdenticalCode_AllowedInDifferentTenants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Alpha Logistics", "alpha_emp")
	tokenB, _, tenantIDB := registerTestCompany(t, router, "Beta Express", "beta_emp")

	sharedCode := fmt.Sprintf("SHARED-EMP-%d", time.Now().UnixNano()%100000)

	payloadA := employees.CreateEmployeeRequest{
		EmployeeCode:    sharedCode,
		FirstName:       "Alpha",
		LastName:        "Driver",
		Designation:     "Driver A",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
	}
	bodyA, _ := json.Marshal(payloadA)
	wA := httptest.NewRecorder()
	reqA := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(bodyA))
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	reqA.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusCreated {
		t.Fatalf("Tenant A creation failed: %d %s", wA.Code, wA.Body.String())
	}

	payloadB := employees.CreateEmployeeRequest{
		EmployeeCode:    sharedCode,
		FirstName:       "Beta",
		LastName:        "Driver",
		Designation:     "Driver B",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
	}
	bodyB, _ := json.Marshal(payloadB)
	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDB), bytes.NewReader(bodyB))
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	reqB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusCreated {
		t.Fatalf("Tenant B creation should succeed with same employee code across tenants: %d %s", wB.Code, wB.Body.String())
	}
}

// TC-P2-EMP-004: Cross-Tenant Branch Assignment Rejected
func TestEmployee_CrossTenantBranch_Rejected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Tenant A Logi", "ten_a_emp")
	tokenB, _, tenantIDB := registerTestCompany(t, router, "Tenant B Logi", "ten_b_emp")

	// Tenant A creates Branch A
	branchPayload := branches.CreateBranchRequest{
		BranchCode:       fmt.Sprintf("BRN-A-%d", time.Now().UnixNano()%100000),
		Name:             "Tenant A Branch",
		Address:          "Street A",
		City:             "Delhi",
		Latitude:         28.6139,
		Longitude:        77.2090,
		CoverageRadiusKM: 10,
	}
	bBody, _ := json.Marshal(branchPayload)
	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDA), bytes.NewReader(bBody))
	reqB.Header.Set("Authorization", "Bearer "+tokenA)
	reqB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wB, reqB)
	var bResp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(wB.Body.Bytes(), &bResp)
	branchAID := bResp.Data.ID

	// Tenant B attempts to create Employee with Tenant A's branch
	empPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    fmt.Sprintf("EMP-B-%d", time.Now().UnixNano()%100000),
		FirstName:       "Infiltrator",
		LastName:        "User",
		Designation:     "Driver",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
		BranchID:        &branchAID, // BELONGS TO TENANT A!
	}
	eBody, _ := json.Marshal(empPayload)
	wE := httptest.NewRecorder()
	reqE := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDB), bytes.NewReader(eBody))
	reqE.Header.Set("Authorization", "Bearer "+tokenB)
	reqE.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wE, reqE)

	// Must be rejected (either 400 or 403)
	if wE.Code != http.StatusBadRequest && wE.Code != http.StatusForbidden {
		t.Fatalf("expected 400 Bad Request or 403 Forbidden for cross-tenant branch assignment, got: %d %s", wE.Code, wE.Body.String())
	}
}

// TC-P2-EMP-005: Cross-Tenant Access Forbidden
func TestEmployee_CrossTenantAccess_Forbidden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Security Org A", "sec_org_a")
	tokenB, _, _ := registerTestCompany(t, router, "Security Org B", "sec_org_b")

	// Create employee in Org A
	empPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    fmt.Sprintf("EMP-SEC-%d", time.Now().UnixNano()%100000),
		FirstName:       "Private",
		LastName:        "Employee",
		Designation:     "Operator",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleOperator,
	}
	eBody, _ := json.Marshal(empPayload)
	wE := httptest.NewRecorder()
	reqE := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(eBody))
	reqE.Header.Set("Authorization", "Bearer "+tokenA)
	reqE.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wE, reqE)
	var eResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(wE.Body.Bytes(), &eResp)
	empAID := eResp.Data.ID

	// Tenant B attempts to read Tenant A's employee
	wGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantIDA, empAID), nil)
	reqGet.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on cross-tenant read attempt, got: %d %s", wGet.Code, wGet.Body.String())
	}
}

// TC-P2-EMP-006: Viewer Role Forbidden From Employee Mutation
func TestEmployee_ViewerRole_ForbiddenFromEmployeeMutation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	ownerToken, _, ownerTenantID := registerTestCompany(t, router, "Perm Corp", "perm_corp")

	// Register viewer user
	viewerEmail := fmt.Sprintf("viewer_%d@test.com", time.Now().UnixNano())
	regBody, _ := json.Marshal(auth.RegisterRequest{
		Email:       viewerEmail,
		Password:    "SecureP@ssw0rd!2026",
		FullName:    "Viewer User",
		CompanyName: "Viewer Org Standalone",
	})
	wReg := httptest.NewRecorder()
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wReg, reqReg)
	var regResp struct {
		Data auth.AuthResponse `json:"data"`
	}
	_ = json.Unmarshal(wReg.Body.Bytes(), &regResp)

	// Owner invites viewer with VIEWER role
	invBody, _ := json.Marshal(tenants.AddMemberRequest{
		Email: viewerEmail,
		Role:  memberships.RoleViewer,
	})
	wInv := httptest.NewRecorder()
	reqInv := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/members", ownerTenantID), bytes.NewReader(invBody))
	reqInv.Header.Set("Authorization", "Bearer "+ownerToken)
	reqInv.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wInv, reqInv)
	if wInv.Code != http.StatusCreated {
		t.Fatalf("failed to invite viewer: %d %s", wInv.Code, wInv.Body.String())
	}

	// Viewer attempts to create an employee -> 403 Forbidden
	empPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    "VW-EMP",
		FirstName:       "Unauthorized",
		LastName:        "Emp",
		Designation:     "Driver",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
	}
	eBody, _ := json.Marshal(empPayload)
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", ownerTenantID), bytes.NewReader(eBody))
	reqCreate.Header.Set("Authorization", "Bearer "+regResp.Data.Token)
	reqCreate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when VIEWER attempts employee creation, got: %d %s", wCreate.Code, wCreate.Body.String())
	}
}

// TC-P2-EMP-007: Update and Soft-Deactivate Employee
func TestEmployee_Update_And_SoftDeactivate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Logi Lifecyle", "logi_life")

	// 1. Create Employee
	empPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    fmt.Sprintf("EMP-LIFE-%d", time.Now().UnixNano()%100000),
		FirstName:       "Suresh",
		LastName:        "Raina",
		Designation:     "Junior Driver",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
	}
	body, _ := json.Marshal(empPayload)
	wC := httptest.NewRecorder()
	reqC := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(body))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wC, reqC)
	if wC.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got: %d %s", wC.Code, wC.Body.String())
	}
	var cResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(wC.Body.Bytes(), &cResp)
	empID := cResp.Data.ID

	// 2. Update Employee Promotion
	newDesignation := "Lead Driver & Dispatch Coordinator"
	newRole := employees.OperationalRoleDispatcher
	updatePayload := employees.UpdateEmployeeRequest{
		Designation:     &newDesignation,
		OperationalRole: &newRole,
	}
	uBody, _ := json.Marshal(updatePayload)
	wU := httptest.NewRecorder()
	reqU := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantID, empID), bytes.NewReader(uBody))
	reqU.Header.Set("Authorization", "Bearer "+token)
	reqU.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wU, reqU)
	if wU.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on update, got: %d %s", wU.Code, wU.Body.String())
	}
	var uResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(wU.Body.Bytes(), &uResp)
	if uResp.Data.Designation != newDesignation {
		t.Errorf("expected updated designation %s, got %s", newDesignation, uResp.Data.Designation)
	}

	// 3. Deactivate Employee
	wD := httptest.NewRecorder()
	reqD := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantID, empID), nil)
	reqD.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wD, reqD)
	if wD.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on delete, got: %d %s", wD.Code, wD.Body.String())
	}

	// 4. Verify status is TERMINATED and is_active is false
	wGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantID, empID), nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on get, got: %d %s", wGet.Code, wGet.Body.String())
	}
	var getResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(wGet.Body.Bytes(), &getResp)
	if getResp.Data.IsActive {
		t.Errorf("expected is_active = false after deactivation, got true")
	}
	if getResp.Data.Status != employees.StatusTerminated {
		t.Errorf("expected status 'TERMINATED', got %s", getResp.Data.Status)
	}
}
