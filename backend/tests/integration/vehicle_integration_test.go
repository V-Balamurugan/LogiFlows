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
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

// TC-P2-VEH-001: Create Vehicle With Branch Association
func TestVehicle_Create_WithBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Fleet Masters", "fleet_veh")

	// 1. Create a branch
	branchPayload := branches.CreateBranchRequest{
		BranchCode:       fmt.Sprintf("BRN-V-%d", time.Now().UnixNano()%100000),
		Name:             "Noida Fleet Hub",
		Address:          "Sector 62",
		City:             "Noida",
		Latitude:         28.6280,
		Longitude:        77.3649,
		CoverageRadiusKM: 20,
	}
	bBody, _ := json.Marshal(branchPayload)
	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(bBody))
	reqB.Header.Set("Authorization", "Bearer "+token)
	reqB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wB, reqB)
	var bResp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(wB.Body.Bytes(), &bResp)
	branchID := bResp.Data.ID

	// 2. Create Vehicle
	regNum := fmt.Sprintf("UP-16-EV-%d", time.Now().UnixNano()%10000)
	makeModel := "Tata Ace EV"
	year := 2025
	vPayload := vehicles.CreateVehicleRequest{
		RegistrationNumber: regNum,
		VehicleType:        vehicles.VehicleTypeElectricVan,
		MakeModel:          &makeModel,
		Year:               &year,
		MaxWeightKG:        1000.0,
		MaxVolumeCBM:       5.5,
		BranchID:           &branchID,
	}
	vBody, _ := json.Marshal(vPayload)
	wV := httptest.NewRecorder()
	reqV := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantID), bytes.NewReader(vBody))
	reqV.Header.Set("Authorization", "Bearer "+token)
	reqV.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wV, reqV)

	if wV.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for vehicle, got: %d %s", wV.Code, wV.Body.String())
	}

	var vResp struct {
		Data vehicles.Vehicle `json:"data"`
	}
	if err := json.Unmarshal(wV.Body.Bytes(), &vResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if vResp.Data.RegistrationNumber != regNum {
		t.Errorf("expected registration number %s, got: %s", regNum, vResp.Data.RegistrationNumber)
	}
	if vResp.Data.VehicleType != vehicles.VehicleTypeElectricVan {
		t.Errorf("expected vehicle type ELECTRIC_VAN, got: %s", vResp.Data.VehicleType)
	}
	if vResp.Data.BranchName == nil || *vResp.Data.BranchName != "Noida Fleet Hub" {
		t.Errorf("expected branch name 'Noida Fleet Hub', got %v", vResp.Data.BranchName)
	}
}

// TC-P2-VEH-002: Duplicate Registration Number Rejected Within Tenant
func TestVehicle_DuplicateRegNum_RejectedWithinTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Rapid Trans", "rapid_veh")

	regNum := fmt.Sprintf("DL-10-TR-%d", time.Now().UnixNano()%10000)
	vPayload := vehicles.CreateVehicleRequest{
		RegistrationNumber: regNum,
		VehicleType:        vehicles.VehicleTypeTruck,
		MaxWeightKG:        3000.0,
		MaxVolumeCBM:       12.0,
	}
	body, _ := json.Marshal(vPayload)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantID), bytes.NewReader(body))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first creation expected 201, got: %d", w1.Code)
	}

	// Attempt duplicate
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantID), bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate registration number, got: %d %s", w2.Code, w2.Body.String())
	}
}

// TC-P2-VEH-003: Identical Registration Number Allowed In Different Tenants
func TestVehicle_IdenticalRegNum_AllowedInDifferentTenants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Alpha Fleet", "alpha_veh")
	tokenB, _, tenantIDB := registerTestCompany(t, router, "Beta Fleet", "beta_veh")

	sharedReg := fmt.Sprintf("SHARED-REG-%d", time.Now().UnixNano()%10000)

	payloadA := vehicles.CreateVehicleRequest{
		RegistrationNumber: sharedReg,
		VehicleType:        vehicles.VehicleTypeVan,
		MaxWeightKG:        800.0,
		MaxVolumeCBM:       4.0,
	}
	bodyA, _ := json.Marshal(payloadA)
	wA := httptest.NewRecorder()
	reqA := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantIDA), bytes.NewReader(bodyA))
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	reqA.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusCreated {
		t.Fatalf("Tenant A vehicle creation failed: %d %s", wA.Code, wA.Body.String())
	}

	payloadB := vehicles.CreateVehicleRequest{
		RegistrationNumber: sharedReg,
		VehicleType:        vehicles.VehicleTypeVan,
		MaxWeightKG:        800.0,
		MaxVolumeCBM:       4.0,
	}
	bodyB, _ := json.Marshal(payloadB)
	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantIDB), bytes.NewReader(bodyB))
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	reqB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusCreated {
		t.Fatalf("Tenant B vehicle creation should succeed across tenants: %d %s", wB.Code, wB.Body.String())
	}
}

// TC-P2-VEH-004: Driver Assignment Lifecycle and Conflict Prevention
func TestVehicle_DriverAssignment_Lifecycle_And_ConflictPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Express Movers", "exp_mov")

	// Helper to create vehicle
	createVeh := func(reg string) uuid.UUID {
		p := vehicles.CreateVehicleRequest{
			RegistrationNumber: reg,
			VehicleType:        vehicles.VehicleTypeVan,
			MaxWeightKG:        600.0,
			MaxVolumeCBM:       3.5,
		}
		b, _ := json.Marshal(p)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantID), bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		var resp struct {
			Data vehicles.Vehicle `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return resp.Data.ID
	}

	// Helper to create driver
	createDriver := func(code, name string) uuid.UUID {
		p := employees.CreateEmployeeRequest{
			EmployeeCode:    code,
			FirstName:       name,
			LastName:        "Driver",
			Designation:     "Driver",
			EmploymentType:  employees.EmploymentTypeFullTime,
			OperationalRole: employees.OperationalRoleDriver,
		}
		b, _ := json.Marshal(p)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		var resp struct {
			Data employees.Employee `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return resp.Data.ID
	}

	uID := time.Now().UnixNano()
	veh1ID := createVeh(fmt.Sprintf("V1-%d", uID%10000))
	veh2ID := createVeh(fmt.Sprintf("V2-%d", uID%10000))

	driver1ID := createDriver(fmt.Sprintf("D1-%d", uID%10000), "Ravi")
	driver2ID := createDriver(fmt.Sprintf("D2-%d", uID%10000), "Vikram")

	// 1. Assign Driver 1 to Vehicle 1
	assignPayload := vehicles.AssignVehicleRequest{
		DriverID: driver1ID,
	}
	aBody, _ := json.Marshal(assignPayload)
	wA1 := httptest.NewRecorder()
	reqA1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantID, veh1ID), bytes.NewReader(aBody))
	reqA1.Header.Set("Authorization", "Bearer "+token)
	reqA1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA1, reqA1)
	if wA1.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for initial assignment, got: %d %s", wA1.Code, wA1.Body.String())
	}

	// 2. CONFLICT CHECK A: Attempt to assign Driver 1 to Vehicle 2 (Driver already busy)
	wA2 := httptest.NewRecorder()
	reqA2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantID, veh2ID), bytes.NewReader(aBody))
	reqA2.Header.Set("Authorization", "Bearer "+token)
	reqA2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA2, reqA2)
	if wA2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when assigning already assigned driver to another vehicle, got: %d %s", wA2.Code, wA2.Body.String())
	}

	// 3. CONFLICT CHECK B: Attempt to assign Driver 2 to Vehicle 1 (Vehicle already busy)
	assign2Payload := vehicles.AssignVehicleRequest{
		DriverID: driver2ID,
	}
	a2Body, _ := json.Marshal(assign2Payload)
	wA3 := httptest.NewRecorder()
	reqA3 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantID, veh1ID), bytes.NewReader(a2Body))
	reqA3.Header.Set("Authorization", "Bearer "+token)
	reqA3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA3, reqA3)
	if wA3.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when assigning driver to already assigned vehicle, got: %d %s", wA3.Code, wA3.Body.String())
	}

	// 4. Unassign Driver 1 from Vehicle 1
	wUn := httptest.NewRecorder()
	reqUn := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/unassign", tenantID, veh1ID), nil)
	reqUn.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wUn, reqUn)
	if wUn.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on unassign, got: %d %s", wUn.Code, wUn.Body.String())
	}

	// 5. Now assign Driver 2 to Vehicle 1 (Must SUCCEED after unassignment)
	wA4 := httptest.NewRecorder()
	reqA4 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantID, veh1ID), bytes.NewReader(a2Body))
	reqA4.Header.Set("Authorization", "Bearer "+token)
	reqA4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA4, reqA4)
	if wA4.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created when assigning newly freed vehicle, got: %d %s", wA4.Code, wA4.Body.String())
	}
}

// TC-P2-VEH-005: Cross-Tenant Access Forbidden
func TestVehicle_CrossTenantAccess_Forbidden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Security Fleet A", "sec_veh_a")
	tokenB, _, _ := registerTestCompany(t, router, "Security Fleet B", "sec_veh_b")

	// Create vehicle in Tenant A
	vPayload := vehicles.CreateVehicleRequest{
		RegistrationNumber: fmt.Sprintf("SEC-%d", time.Now().UnixNano()%10000),
		VehicleType:        vehicles.VehicleTypeVan,
		MaxWeightKG:        500.0,
		MaxVolumeCBM:       3.0,
	}
	body, _ := json.Marshal(vPayload)
	wV := httptest.NewRecorder()
	reqV := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantIDA), bytes.NewReader(body))
	reqV.Header.Set("Authorization", "Bearer "+tokenA)
	reqV.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wV, reqV)
	var resp struct {
		Data vehicles.Vehicle `json:"data"`
	}
	_ = json.Unmarshal(wV.Body.Bytes(), &resp)
	vehAID := resp.Data.ID

	// Tenant B attempts to read Tenant A's vehicle
	wGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s", tenantIDA, vehAID), nil)
	reqGet.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for cross-tenant vehicle read, got: %d %s", wGet.Code, wGet.Body.String())
	}
}

// TC-P2-VEH-006: Viewer Role Forbidden From Vehicle Mutation
func TestVehicle_ViewerRole_ForbiddenFromMutation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	ownerToken, _, ownerTenantID := registerTestCompany(t, router, "Perm Fleet Corp", "perm_fleet")

	viewerEmail := fmt.Sprintf("viewer_veh_%d@test.com", time.Now().UnixNano())
	regBody, _ := json.Marshal(auth.RegisterRequest{
		Email:       viewerEmail,
		Password:    "SecureP@ssw0rd!2026",
		FullName:    "Viewer Fleet User",
		CompanyName: "Viewer Fleet Standalone",
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

	// Viewer attempts to create a vehicle
	vPayload := vehicles.CreateVehicleRequest{
		RegistrationNumber: "UNAUTH-01",
		VehicleType:        vehicles.VehicleTypeVan,
		MaxWeightKG:        500.0,
		MaxVolumeCBM:       3.0,
	}
	body, _ := json.Marshal(vPayload)
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", ownerTenantID), bytes.NewReader(body))
	reqCreate.Header.Set("Authorization", "Bearer "+regResp.Data.Token)
	reqCreate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when VIEWER attempts vehicle creation, got: %d %s", wCreate.Code, wCreate.Body.String())
	}
}
