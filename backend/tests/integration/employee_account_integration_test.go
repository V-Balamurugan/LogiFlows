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
	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

// TC-P3-EMP-ACC-001: Transactional Employee Account Creation and Login
func TestEmployee_CreateWithAccount_And_Login(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	adminToken, _, tenantID := registerTestCompany(t, router, "Logistics Pro Corp", "logi_pro")

	// 1. Create a Branch
	bPayload := branches.CreateBranchRequest{
		BranchCode:       fmt.Sprintf("BRN-ACC-%d", time.Now().UnixNano()%100000),
		Name:             "Cyber Hub Logistics Depot",
		Address:          "DLF Phase 2",
		City:             "Gurgaon",
		Latitude:         28.4900,
		Longitude:        77.0900,
		CoverageRadiusKM: 15,
	}
	bBody, _ := json.Marshal(bPayload)
	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(bBody))
	reqB.Header.Set("Authorization", "Bearer "+adminToken)
	reqB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for branch, got: %d %s", wB.Code, wB.Body.String())
	}
	var bResp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(wB.Body.Bytes(), &bResp)
	branchID := bResp.Data.ID

	// 2. Create Employee with Login Account
	empEmail := fmt.Sprintf("driver.vikram.%d@logiflows.internal", time.Now().UnixNano())
	empPassword := "SecureDriver@2026!"
	empPayload := employees.CreateEmployeeWithAccountRequest{
		FirstName:          "Vikram",
		LastName:           "Sharma",
		Email:              empEmail,
		Designation:        "Senior Delivery Lead",
		EmploymentType:     employees.EmploymentTypeFullTime,
		OperationalRole:    employees.OperationalRoleDriver,
		AvailabilityStatus: employees.AvailabilityStatusAvailable,
		VerificationStatus: employees.VerificationStatusVerified,
		BranchID:           &branchID,
		Password:           empPassword,
		SystemRole:         "EMPLOYEE",
	}

	eBody, _ := json.Marshal(empPayload)
	wE := httptest.NewRecorder()
	reqE := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees/with-account", tenantID), bytes.NewReader(eBody))
	reqE.Header.Set("Authorization", "Bearer "+adminToken)
	reqE.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wE, reqE)

	if wE.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for employee with account, got: %d %s", wE.Code, wE.Body.String())
	}

	var empResp struct {
		Data employees.Employee `json:"data"`
	}
	if err := json.Unmarshal(wE.Body.Bytes(), &empResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if empResp.Data.UserID == nil {
		t.Fatal("expected employee UserID to be populated")
	}
	if empResp.Data.FirstName != "Vikram" || empResp.Data.LastName != "Sharma" {
		t.Fatalf("unexpected employee name: %s %s", empResp.Data.FirstName, empResp.Data.LastName)
	}
	if empResp.Data.OperationalRole != employees.OperationalRoleDriver {
		t.Fatalf("unexpected operational role: %s", empResp.Data.OperationalRole)
	}
	empID := empResp.Data.ID

	// 3. Query Account Status
	wStatus := httptest.NewRecorder()
	reqStatus := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/%s/account-status", tenantID, empID), nil)
	reqStatus.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(wStatus, reqStatus)

	if wStatus.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for account status, got: %d %s", wStatus.Code, wStatus.Body.String())
	}
	var statusResp struct {
		Data employees.EmployeeAccountStatusResponse `json:"data"`
	}
	_ = json.Unmarshal(wStatus.Body.Bytes(), &statusResp)

	if !statusResp.Data.HasAccount {
		t.Fatal("expected has_account to be true")
	}
	if statusResp.Data.UserEmail == nil || *statusResp.Data.UserEmail != empEmail {
		t.Fatalf("unexpected user email: %v", statusResp.Data.UserEmail)
	}
	if statusResp.Data.SystemRole == nil || *statusResp.Data.SystemRole != "EMPLOYEE" {
		t.Fatalf("unexpected system role: %v", statusResp.Data.SystemRole)
	}

	// 4. Authenticate as the newly created Employee
	loginPayload := auth.LoginRequest{
		Email:    empEmail,
		Password: empPassword,
	}
	lBody, _ := json.Marshal(loginPayload)
	wLogin := httptest.NewRecorder()
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(lBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for employee login, got: %d %s", wLogin.Code, wLogin.Body.String())
	}
	var loginResp struct {
		Data auth.AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(wLogin.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	empAccessToken := loginResp.Data.Token
	if empAccessToken == "" {
		t.Fatal("expected non-empty access token for logged in employee")
	}

	// 5. Query /employees/me with Employee token
	wMe := httptest.NewRecorder()
	reqMe := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/me", tenantID), nil)
	reqMe.Header.Set("Authorization", "Bearer "+empAccessToken)
	router.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /employees/me, got: %d %s", wMe.Code, wMe.Body.String())
	}
	var meResp struct {
		Data employees.EmployeeMeResponse `json:"data"`
	}
	_ = json.Unmarshal(wMe.Body.Bytes(), &meResp)
	if meResp.Data.Employee.ID != empID {
		t.Fatalf("expected employee ID %s, got: %s", empID, meResp.Data.Employee.ID)
	}
	if meResp.Data.SystemRole != "EMPLOYEE" {
		t.Fatalf("expected system role EMPLOYEE, got: %s", meResp.Data.SystemRole)
	}

	// 6. Branch Asset Listing: Employees & Vehicles
	wBranchEmps := httptest.NewRecorder()
	reqBranchEmps := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches/%s/employees", tenantID, branchID), nil)
	reqBranchEmps.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(wBranchEmps, reqBranchEmps)
	if wBranchEmps.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for branch employees, got: %d %s", wBranchEmps.Code, wBranchEmps.Body.String())
	}
	var branchEmpsResp struct {
		Data struct {
			Employees []branches.BranchEmployeeSummary `json:"employees"`
			Total     int                              `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wBranchEmps.Body.Bytes(), &branchEmpsResp)
	if branchEmpsResp.Data.Total < 1 {
		t.Fatalf("expected at least 1 branch employee, got %d", branchEmpsResp.Data.Total)
	}

	// Create and assign a vehicle to verify branch vehicles listing
	regNum := fmt.Sprintf("DL-1C-EV-%d", time.Now().UnixNano()%10000)
	vPayload := vehicles.CreateVehicleRequest{
		RegistrationNumber: regNum,
		VehicleType:        vehicles.VehicleTypeElectricVan,
		MaxWeightKG:        750.0,
		MaxVolumeCBM:       4.0,
		BranchID:           &branchID,
	}
	vBody, _ := json.Marshal(vPayload)
	wV := httptest.NewRecorder()
	reqV := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantID), bytes.NewReader(vBody))
	reqV.Header.Set("Authorization", "Bearer "+adminToken)
	reqV.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wV, reqV)
	if wV.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for vehicle, got: %d %s", wV.Code, wV.Body.String())
	}

	wBranchVehs := httptest.NewRecorder()
	reqBranchVehs := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches/%s/vehicles", tenantID, branchID), nil)
	reqBranchVehs.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(wBranchVehs, reqBranchVehs)
	if wBranchVehs.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for branch vehicles, got: %d %s", wBranchVehs.Code, wBranchVehs.Body.String())
	}
	var branchVehsResp struct {
		Data struct {
			Vehicles []branches.BranchVehicleSummary `json:"vehicles"`
			Total    int                             `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wBranchVehs.Body.Bytes(), &branchVehsResp)
	if branchVehsResp.Data.Total < 1 {
		t.Fatalf("expected at least 1 branch vehicle, got %d", branchVehsResp.Data.Total)
	}
}

// TC-P3-EMP-ACC-002: Duplicate Email Conflict & Transaction Rollback
func TestEmployee_CreateWithAccount_DuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	adminToken, _, tenantID := registerTestCompany(t, router, "Express Carriers", "exp_carrier")

	dupEmail := fmt.Sprintf("dispatcher.anita.%d@logiflows.internal", time.Now().UnixNano())
	payload1 := employees.CreateEmployeeWithAccountRequest{
		FirstName:       "Anita",
		LastName:        "Deshmukh",
		Email:           dupEmail,
		Designation:     "Hub Dispatcher",
		OperationalRole: employees.OperationalRoleDispatcher,
		Password:        "SecureAnita@2026!",
		SystemRole:      "TENANT_OPERATOR",
	}

	b1, _ := json.Marshal(payload1)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees/with-account", tenantID), bytes.NewReader(b1))
	req1.Header.Set("Authorization", "Bearer "+adminToken)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on first employee, got: %d %s", w1.Code, w1.Body.String())
	}

	// Second creation with the exact same email in the same tenant must return 409 Conflict
	payload2 := employees.CreateEmployeeWithAccountRequest{
		FirstName:       "Anita",
		LastName:        "Duplicate",
		Email:           dupEmail,
		Designation:     "Duplicate Profile",
		OperationalRole: employees.OperationalRoleOperator,
		Password:        "SecureSecond@2026!",
	}
	b2, _ := json.Marshal(payload2)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees/with-account", tenantID), bytes.NewReader(b2))
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate email, got: %d %s", w2.Code, w2.Body.String())
	}
}

// TC-P3-EMP-ACC-003: Invitation Flow with Auto-Generated Password
func TestEmployee_CreateWithAccount_InvitationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	adminToken, _, tenantID := registerTestCompany(t, router, "Rapid Transports", "rapid_trans")

	inviteEmail := fmt.Sprintf("supervisor.raj.%d@logiflows.internal", time.Now().UnixNano())
	payload := employees.CreateEmployeeWithAccountRequest{
		FirstName:       "Rajesh",
		LastName:        "Kumar",
		Email:           inviteEmail,
		Designation:     "Shift Supervisor",
		OperationalRole: employees.OperationalRoleSupervisor,
		SendInvite:      true, // No password specified, invite flow enabled
	}

	b, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees/with-account", tenantID), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for employee via invite flow, got: %d %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.UserID == nil {
		t.Fatal("expected UserID to be created and linked for invite flow")
	}
}
