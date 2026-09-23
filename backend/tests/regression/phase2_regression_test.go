package regression_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

// TestRegression_Phase2_BranchManagement verifies that branch management invariants never regress.
func TestRegression_Phase2_BranchManagement(t *testing.T) {
	router, _ := setupRegressionRouter(t)

	tokenA, _, tenantIDA := registerRegressionTenant(t, router, "BranchOrgA")
	tokenB, _, _ := registerRegressionTenant(t, router, "BranchOrgB")

	var createdBranchID string

	t.Run("Create Branch with PostGIS Coordinates", func(t *testing.T) {
		payload := branches.CreateBranchRequest{
			BranchCode:       "REG-BR-01",
			Name:             "Regression Central Hub",
			Address:          "100 Innovation Blvd",
			City:             "New Delhi",
			Country:          "India",
			Latitude:         28.6139,
			Longitude:        77.2090,
			CoverageRadiusKM: 25.0,
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("regression failure: create branch returned %d: %s", w.Code, w.Body.String())
		}

		var env response.SuccessEnvelope
		_ = json.Unmarshal(w.Body.Bytes(), &env)
		dataBytes, _ := json.Marshal(env.Data)
		var b branches.Branch
		_ = json.Unmarshal(dataBytes, &b)
		createdBranchID = b.ID.String()

		if b.BranchCode != "REG-BR-01" || !b.IsActive {
			t.Errorf("regression failure: branch fields mismatch: %+v", b)
		}
	})

	t.Run("Duplicate Branch Code Rejected with 409", func(t *testing.T) {
		payload := branches.CreateBranchRequest{
			BranchCode:       "REG-BR-01", // Duplicate code
			Name:             "Duplicate Hub",
			Address:          "200 Innovation Blvd",
			City:             "New Delhi",
			Country:          "India",
			Latitude:         28.6140,
			Longitude:        77.2095,
			CoverageRadiusKM: 10.0,
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("regression failure: duplicate branch code allowed, got %d", w.Code)
		}
	})

	t.Run("Cross-Tenant Access Forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantIDA, createdBranchID), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("regression failure: cross-tenant branch access returned %d instead of 403", w.Code)
		}
	})

	t.Run("Soft Delete Branch Marks as Inactive", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantIDA, createdBranchID), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("regression failure: delete branch returned %d", w.Code)
		}

		// Verify GET returns inactive
		wGet := httptest.NewRecorder()
		reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantIDA, createdBranchID), nil)
		reqGet.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(wGet, reqGet)

		var env response.SuccessEnvelope
		_ = json.Unmarshal(wGet.Body.Bytes(), &env)
		dataBytes, _ := json.Marshal(env.Data)
		var b branches.Branch
		_ = json.Unmarshal(dataBytes, &b)

		if b.IsActive || b.Status != "INACTIVE" {
			t.Errorf("regression failure: expected inactive branch, got is_active=%v, status=%s", b.IsActive, b.Status)
		}
	})
}

// TestRegression_Phase2_EmployeeManagement verifies that employee creation, validation, and role enforcement never regress.
func TestRegression_Phase2_EmployeeManagement(t *testing.T) {
	router, _ := setupRegressionRouter(t)

	tokenA, _, tenantIDA := registerRegressionTenant(t, router, "EmpOrgA")
	tokenB, _, tenantIDB := registerRegressionTenant(t, router, "EmpOrgB")

	// Create branch in Tenant B
	bReq := branches.CreateBranchRequest{
		BranchCode: "BR-B-01",
		Name:       "Branch B",
		Address:    "Road B",
		City:       "City B",
		Country:    "India",
		Latitude:   28.5,
		Longitude:  77.2,
	}
	bBody, _ := json.Marshal(bReq)
	bRec := httptest.NewRecorder()
	bHttpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDB), bytes.NewReader(bBody))
	bHttpReq.Header.Set("Content-Type", "application/json")
	bHttpReq.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(bRec, bHttpReq)

	var bEnv response.SuccessEnvelope
	_ = json.Unmarshal(bRec.Body.Bytes(), &bEnv)
	bData, _ := json.Marshal(bEnv.Data)
	var branchB branches.Branch
	_ = json.Unmarshal(bData, &branchB)

	var empIDA string

	t.Run("Create Employee with Driver Role", func(t *testing.T) {
		lic := "DL-REG-2026-99"
		email := "driver.reg@logiflows.test"
		payload := employees.CreateEmployeeRequest{
			EmployeeCode:    "DRV-001",
			FirstName:       "Ramesh",
			LastName:        "Kumar",
			Email:           &email,
			Designation:     "Senior Delivery Driver",
			EmploymentType:  "FULL_TIME",
			OperationalRole: "DRIVER",
			LicenseNumber:   &lic,
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("regression failure: create employee returned %d: %s", w.Code, w.Body.String())
		}

		var env response.SuccessEnvelope
		_ = json.Unmarshal(w.Body.Bytes(), &env)
		dataBytes, _ := json.Marshal(env.Data)
		var emp employees.Employee
		_ = json.Unmarshal(dataBytes, &emp)
		empIDA = emp.ID.String()

		if emp.EmployeeCode != "DRV-001" || emp.OperationalRole != "DRIVER" {
			t.Errorf("regression failure: employee fields mismatch: %+v", emp)
		}
	})

	t.Run("Duplicate Employee Code Rejected within Tenant", func(t *testing.T) {
		payload := employees.CreateEmployeeRequest{
			EmployeeCode:    "DRV-001", // Duplicate code
			FirstName:       "Suresh",
			LastName:        "Verma",
			Designation:     "Driver",
			EmploymentType:  "FULL_TIME",
			OperationalRole: "DRIVER",
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("regression failure: duplicate employee code allowed, got %d", w.Code)
		}
	})

	t.Run("Cross-Tenant Branch Assignment Rejected", func(t *testing.T) {
		branchBID := branchB.ID
		payload := employees.CreateEmployeeRequest{
			EmployeeCode:    "DRV-002",
			FirstName:       "Anil",
			LastName:        "Sharma",
			Designation:     "Driver",
			EmploymentType:  "FULL_TIME",
			OperationalRole: "DRIVER",
			BranchID:        &branchBID, // Belongs to Tenant B!
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("regression failure: cross-tenant branch assignment should be 400, got %d", w.Code)
		}
	})

	t.Run("Cross-Tenant Employee Read Forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantIDA, empIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("regression failure: cross-tenant employee read returned %d instead of 403", w.Code)
		}
	})

	t.Run("Soft Deactivate Employee", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantIDA, empIDA), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("regression failure: delete employee returned %d", w.Code)
		}

		wGet := httptest.NewRecorder()
		reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/%s", tenantIDA, empIDA), nil)
		reqGet.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(wGet, reqGet)

		var env response.SuccessEnvelope
		_ = json.Unmarshal(wGet.Body.Bytes(), &env)
		dataBytes, _ := json.Marshal(env.Data)
		var emp employees.Employee
		_ = json.Unmarshal(dataBytes, &emp)

		if emp.IsActive || emp.Status != "TERMINATED" {
			t.Errorf("regression failure: expected terminated employee, got is_active=%v, status=%s", emp.IsActive, emp.Status)
		}
	})
}

// TestRegression_Phase2_VehicleAndFleetAssignment verifies fleet management, driver assignment, and double-booking conflict prevention.
func TestRegression_Phase2_VehicleAndFleetAssignment(t *testing.T) {
	router, _ := setupRegressionRouter(t)

	tokenA, _, tenantIDA := registerRegressionTenant(t, router, "FleetOrgA")
	tokenB, _, _ := registerRegressionTenant(t, router, "FleetOrgB")

	var vehicle1ID, vehicle2ID uuid.UUID
	var driver1ID, driver2ID uuid.UUID

	t.Run("Register Electric Delivery Vans", func(t *testing.T) {
		makeModel := "Tata Ace EV"
		yr := 2026
		payload1 := vehicles.CreateVehicleRequest{
			RegistrationNumber: "DL-01-EV-7001",
			VehicleType:        "ELECTRIC_VAN",
			MakeModel:          &makeModel,
			Year:               &yr,
			MaxWeightKG:        800,
			MaxVolumeCBM:       5.0,
		}
		body1, _ := json.Marshal(payload1)

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantIDA), bytes.NewReader(body1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("regression failure: create vehicle 1 returned %d: %s", w1.Code, w1.Body.String())
		}

		var env1 response.SuccessEnvelope
		_ = json.Unmarshal(w1.Body.Bytes(), &env1)
		data1, _ := json.Marshal(env1.Data)
		var v1 vehicles.Vehicle
		_ = json.Unmarshal(data1, &v1)
		vehicle1ID = v1.ID

		payload2 := vehicles.CreateVehicleRequest{
			RegistrationNumber: "DL-01-EV-7002",
			VehicleType:        "ELECTRIC_VAN",
			MaxWeightKG:        800,
			MaxVolumeCBM:       5.0,
		}
		body2, _ := json.Marshal(payload2)
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantIDA), bytes.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w2, req2)

		var env2 response.SuccessEnvelope
		_ = json.Unmarshal(w2.Body.Bytes(), &env2)
		data2, _ := json.Marshal(env2.Data)
		var v2 vehicles.Vehicle
		_ = json.Unmarshal(data2, &v2)
		vehicle2ID = v2.ID
	})

	t.Run("Create Drivers for Fleet Assignment", func(t *testing.T) {
		for i, code := range []string{"FLT-DRV-1", "FLT-DRV-2"} {
			lic := fmt.Sprintf("LIC-FLT-%d", i+1)
			empReq := employees.CreateEmployeeRequest{
				EmployeeCode:    code,
				FirstName:       fmt.Sprintf("Driver%d", i+1),
				LastName:        "Singh",
				Designation:     "Delivery Executive",
				EmploymentType:  "FULL_TIME",
				OperationalRole: "DRIVER",
				LicenseNumber:   &lic,
			}
			b, _ := json.Marshal(empReq)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantIDA), bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tokenA)
			router.ServeHTTP(w, req)

			var env response.SuccessEnvelope
			_ = json.Unmarshal(w.Body.Bytes(), &env)
			data, _ := json.Marshal(env.Data)
			var emp employees.Employee
			_ = json.Unmarshal(data, &emp)

			if i == 0 {
				driver1ID = emp.ID
			} else {
				driver2ID = emp.ID
			}
		}
	})

	t.Run("Driver Assignment Lifecycle & Conflict Prevention (409)", func(t *testing.T) {
		// 1. Assign Driver 1 to Vehicle 1 -> 201 Created
		assignReq1 := vehicles.AssignVehicleRequest{
			DriverID: driver1ID,
		}
		b1, _ := json.Marshal(assignReq1)
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantIDA, vehicle1ID), bytes.NewReader(b1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("regression failure: initial driver assignment returned %d: %s", w1.Code, w1.Body.String())
		}

		// 2. Attempt to assign Driver 1 to Vehicle 2 -> 409 Conflict (Driver already assigned)
		assignReq2 := vehicles.AssignVehicleRequest{
			DriverID: driver1ID,
		}
		b2, _ := json.Marshal(assignReq2)
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantIDA, vehicle2ID), bytes.NewReader(b2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusConflict {
			t.Errorf("regression failure: expected 409 when double-assigning driver, got %d: %s", w2.Code, w2.Body.String())
		}

		// 3. Attempt to assign Driver 2 to Vehicle 1 -> 409 Conflict (Vehicle already assigned)
		assignReq3 := vehicles.AssignVehicleRequest{
			DriverID: driver2ID,
		}
		b3, _ := json.Marshal(assignReq3)
		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantIDA, vehicle1ID), bytes.NewReader(b3))
		req3.Header.Set("Content-Type", "application/json")
		req3.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w3, req3)

		if w3.Code != http.StatusConflict {
			t.Errorf("regression failure: expected 409 when double-assigning vehicle, got %d: %s", w3.Code, w3.Body.String())
		}

		// 4. Unassign Driver 1 from Vehicle 1 -> 200 OK
		wUnassign := httptest.NewRecorder()
		reqUnassign := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/unassign", tenantIDA, vehicle1ID), nil)
		reqUnassign.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(wUnassign, reqUnassign)

		if wUnassign.Code != http.StatusOK {
			t.Fatalf("regression failure: unassign vehicle returned %d: %s", wUnassign.Code, wUnassign.Body.String())
		}

		// 5. Driver 1 can now be assigned to Vehicle 2 cleanly -> 201 Created
		w4 := httptest.NewRecorder()
		req4 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantIDA, vehicle2ID), bytes.NewReader(b2))
		req4.Header.Set("Content-Type", "application/json")
		req4.Header.Set("Authorization", "Bearer "+tokenA)
		router.ServeHTTP(w4, req4)

		if w4.Code != http.StatusCreated {
			t.Errorf("regression failure: subsequent assignment after unassign returned %d: %s", w4.Code, w4.Body.String())
		}
	})

	t.Run("Cross-Tenant Vehicle Access Forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s", tenantIDA, vehicle1ID), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("regression failure: cross-tenant vehicle read returned %d instead of 403", w.Code)
		}
	})
}
