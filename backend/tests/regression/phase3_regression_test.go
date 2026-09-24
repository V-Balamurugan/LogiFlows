package regression_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

func TestRegression_Phase3_EmployeeLifecycle(t *testing.T) {
	router, _ := setupRegressionRouter(t)

	tokenA, _, tenantIDA := registerRegressionTenant(t, router, "P3EmpTenantA")
	_, _, tenantIDB := registerRegressionTenant(t, router, "P3EmpTenantB")

	// Create branch in Tenant A
	bPayload := branches.CreateBranchRequest{
		BranchCode:       fmt.Sprintf("P3-BR-%d", time.Now().UnixNano()%10000),
		Name:             "P3 Distribution Hub",
		Address:          "456 Express Way",
		City:             "Mumbai",
		Latitude:         19.0760,
		Longitude:        72.8777,
		CoverageRadiusKM: 15.0,
	}
	bBody, _ := json.Marshal(bPayload)
	bReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDA), bytes.NewReader(bBody))
	bReq.Header.Set("Authorization", "Bearer "+tokenA)
	bReq.Header.Set("Content-Type", "application/json")
	bW := httptest.NewRecorder()
	router.ServeHTTP(bW, bReq)
	if bW.Code != http.StatusCreated {
		t.Fatalf("failed to create branch: %s", bW.Body.String())
	}
	var bResp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(bW.Body.Bytes(), &bResp)
	branchIDA := bResp.Data.ID

	var driverIDA string

	// US-E01 & US-E02: Create Employee with Auto-Generated Code
	t.Run("US-E01 & US-E02: Auto-Generate Sequential Employee Code", func(t *testing.T) {
		empReq := employees.CreateEmployeeRequest{
			EmployeeCode:       "", // Auto-generate
			FirstName:          "Vikram",
			LastName:           "Singh",
			Designation:        "Commercial Driver",
			EmploymentType:     employees.EmploymentTypeFullTime,
			OperationalRole:    employees.OperationalRoleDriver,
			AvailabilityStatus: employees.AvailabilityStatusAvailable,
			VerificationStatus: employees.VerificationStatusVerified,
			BranchID:           &branchIDA,
		}
		body, _ := json.Marshal(empReq)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data employees.Employee `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Data.EmployeeCode == "" {
			t.Errorf("expected auto-generated employee code, got empty")
		}
		if resp.Data.AvailabilityStatus != employees.AvailabilityStatusAvailable {
			t.Errorf("expected AVAILABLE status, got %s", resp.Data.AvailabilityStatus)
		}
		driverIDA = resp.Data.ID.String()
	})

	// US-E03: Branch Association
	t.Run("US-E03: Associate Employee with Branch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantIDA, driverIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
		var resp struct {
			Data employees.Employee `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data.BranchID == nil || *resp.Data.BranchID != branchIDA {
			t.Errorf("expected branch ID %s, got %v", branchIDA, resp.Data.BranchID)
		}
	})

	// US-E04 & US-E07: Filter Employees
	t.Run("US-E04 & US-E07: Filter Employees by Role and Status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees?operational_role=DRIVER&status=ACTIVE", tenantIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}
		var resp struct {
			Data employees.EmployeeListResponse `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data.Total < 1 {
			t.Errorf("expected at least 1 driver, got %d", resp.Data.Total)
		}
	})

	// US-E05 & US-E06: Update Employee Status and Availability
	t.Run("US-E05 & US-E06: Update Status via PATCH /status", func(t *testing.T) {
		onLeave := employees.StatusOnLeave
		offDuty := employees.AvailabilityStatusOffDuty
		payload := employees.UpdateEmployeeStatusRequest{
			Status:             &onLeave,
			AvailabilityStatus: &offDuty,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/employees/%s/status", tenantIDA, driverIDA), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
		var resp struct {
			Data employees.Employee `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data.Status != employees.StatusOnLeave || resp.Data.AvailabilityStatus != employees.AvailabilityStatusOffDuty {
			t.Errorf("expected status ON_LEAVE and OFF_DUTY, got %s / %s", resp.Data.Status, resp.Data.AvailabilityStatus)
		}

		// Restore driver to ACTIVE & AVAILABLE
		active := employees.StatusActive
		avail := employees.AvailabilityStatusAvailable
		restorePayload := employees.UpdateEmployeeStatusRequest{
			Status:             &active,
			AvailabilityStatus: &avail,
		}
		rBody, _ := json.Marshal(restorePayload)
		rReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/employees/%s/status", tenantIDA, driverIDA), bytes.NewReader(rBody))
		rReq.Header.Set("Authorization", "Bearer "+tokenA)
		rReq.Header.Set("Content-Type", "application/json")
		rW := httptest.NewRecorder()
		router.ServeHTTP(rW, rReq)
		if rW.Code != http.StatusOK {
			t.Fatalf("failed to restore driver status")
		}
	})

	// US-E08: Available Drivers
	t.Run("US-E08: Query Available Drivers Endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/available-drivers", tenantIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}
		var resp struct {
			Data struct {
				Total int `json:"total"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data.Total < 1 {
			t.Errorf("expected at least 1 available driver, got %d", resp.Data.Total)
		}
	})

	// US-E09: Cross-Tenant Isolation
	t.Run("US-E09: Cross-Tenant Employee Access Denied with 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantIDB, driverIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-tenant employee access, got %d", w.Code)
		}
	})
}

func TestRegression_Phase3_VehicleFleetAndAssignmentLifecycle(t *testing.T) {
	router, _ := setupRegressionRouter(t)

	tokenA, _, tenantIDA := registerRegressionTenant(t, router, "P3VehTenantA")
	_, _, tenantIDB := registerRegressionTenant(t, router, "P3VehTenantB")

	var vehicleIDA string

	// 1. Create active driver in Tenant A
	drvReq := employees.CreateEmployeeRequest{
		EmployeeCode:       fmt.Sprintf("DRV-P3-%d", time.Now().UnixNano()%10000),
		FirstName:          "Arun",
		LastName:           "Kumar",
		Designation:        "Fleet Driver",
		EmploymentType:     employees.EmploymentTypeFullTime,
		OperationalRole:    employees.OperationalRoleDriver,
		AvailabilityStatus: employees.AvailabilityStatusAvailable,
	}
	dBody, _ := json.Marshal(drvReq)
	dReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(dBody))
	dReq.Header.Set("Authorization", "Bearer "+tokenA)
	dReq.Header.Set("Content-Type", "application/json")
	dW := httptest.NewRecorder()
	router.ServeHTTP(dW, dReq)
	if dW.Code != http.StatusCreated {
		t.Fatalf("failed to create driver: %s", dW.Body.String())
	}
	var dResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(dW.Body.Bytes(), &dResp)

	// US-V01 & US-V02: Register Vehicle with Type and Capacity
	t.Run("US-V01 & US-V02: Register Vehicle with Physical Capacity", func(t *testing.T) {
		vehReq := vehicles.CreateVehicleRequest{
			RegistrationNumber: fmt.Sprintf("TN-01-%d", time.Now().UnixNano()%100000),
			VehicleType:        vehicles.VehicleTypeElectricVan,
			MaxWeightKG:        1500,
			MaxVolumeCBM:       10.5,
			AvailabilityStatus: vehicles.AvailabilityStatusAvailable,
		}
		body, _ := json.Marshal(vehReq)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
		}
		var resp struct {
			Data vehicles.Vehicle `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data.MaxWeightKG != 1500 || resp.Data.MaxVolumeCBM != 10.5 {
			t.Errorf("capacity mismatch: %v kg, %v cbm", resp.Data.MaxWeightKG, resp.Data.MaxVolumeCBM)
		}
		vehicleIDA = resp.Data.ID.String()
	})

	// US-V04: View Available Vehicles
	t.Run("US-V04: Filter Available Vehicles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/vehicles?status=AVAILABLE", tenantIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}
		var resp struct {
			Data vehicles.VehicleListResponse `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data.Total < 1 {
			t.Errorf("expected at least 1 available vehicle")
		}
	})

	// US-V07: Driver Eligibility & Vehicle Assignment
	t.Run("US-V07: Driver Assignment Lifecycle and State Transitions", func(t *testing.T) {
		notes := "Express priority assignment"
		assignReq := vehicles.AssignVehicleRequest{
			DriverID: dResp.Data.ID,
			Notes:    &notes,
		}
		body, _ := json.Marshal(assignReq)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantIDA, vehicleIDA), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created for assignment, got %d: %s", w.Code, w.Body.String())
		}

		// Verify vehicle status is ASSIGNED
		vGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s", tenantIDA, vehicleIDA), nil)
		vGetReq.Header.Set("Authorization", "Bearer "+tokenA)
		vGetW := httptest.NewRecorder()
		router.ServeHTTP(vGetW, vGetReq)
		var vGetResp struct {
			Data vehicles.Vehicle `json:"data"`
		}
		_ = json.Unmarshal(vGetW.Body.Bytes(), &vGetResp)
		if vGetResp.Data.Status != vehicles.VehicleStatusAssigned {
			t.Errorf("expected vehicle status ASSIGNED, got %s", vGetResp.Data.Status)
		}
	})

	// US-V08: Double Booking Prevention
	t.Run("US-V08: Prevent Conflicting Active Assignment with 409", func(t *testing.T) {
		notes := "Conflicting assignment attempt"
		assignReq := vehicles.AssignVehicleRequest{
			DriverID: dResp.Data.ID,
			Notes:    &notes,
		}
		body, _ := json.Marshal(assignReq)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantIDA, vehicleIDA), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected 409 Conflict for double booking, got %d: %s", w.Code, w.Body.String())
		}
	})

	// US-V06: Change Vehicle Status after unassignment
	t.Run("US-V06: Unassign and Transition Status", func(t *testing.T) {
		uReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/unassign", tenantIDA, vehicleIDA), nil)
		uReq.Header.Set("Authorization", "Bearer "+tokenA)
		uW := httptest.NewRecorder()
		router.ServeHTTP(uW, uReq)
		if uW.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on unassign, got %d", uW.Code)
		}

		// Patch vehicle status to MAINTENANCE
		maint := vehicles.VehicleStatusMaintenance
		patchReq := vehicles.UpdateVehicleStatusRequest{
			Status: &maint,
		}
		pBody, _ := json.Marshal(patchReq)
		pReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/status", tenantIDA, vehicleIDA), bytes.NewReader(pBody))
		pReq.Header.Set("Authorization", "Bearer "+tokenA)
		pReq.Header.Set("Content-Type", "application/json")
		pW := httptest.NewRecorder()
		router.ServeHTTP(pW, pReq)
		if pW.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on status patch, got %d: %s", pW.Code, pW.Body.String())
		}
	})

	// US-V09: Cross-Tenant Vehicle Isolation
	t.Run("US-V09: Cross-Tenant Vehicle Access Denied with 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s", tenantIDB, vehicleIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-tenant vehicle access, got %d", w.Code)
		}
	})
}
