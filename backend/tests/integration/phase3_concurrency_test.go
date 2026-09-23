package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

func TestPhase3_Concurrent_EmployeeCodeGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Async Fleet Corp", "async_fleet")

	// Create test branch
	branchPayload := branches.CreateBranchRequest{
		BranchCode:       fmt.Sprintf("ASY-%d", time.Now().UnixNano()%100000),
		Name:             "Async Central Hub",
		Address:          "Sector 21",
		City:             "Gurgaon",
		Latitude:         28.4595,
		Longitude:        77.0266,
		CoverageRadiusKM: 10,
	}
	bBody, _ := json.Marshal(branchPayload)
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(bBody))
	reqB.Header.Set("Authorization", "Bearer "+token)
	reqB.Header.Set("Content-Type", "application/json")
	wB := httptest.NewRecorder()
	router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusCreated {
		t.Fatalf("failed to create branch: %s", wB.Body.String())
	}
	var bResp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(wB.Body.Bytes(), &bResp)
	branchID := bResp.Data.ID

	concurrency := 10
	var wg sync.WaitGroup
	generatedCodes := make([]string, concurrency)
	errorsList := make([]error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			payload := employees.CreateEmployeeRequest{
				EmployeeCode:    "", // Trigger automatic sequence generation
				FirstName:       fmt.Sprintf("Driver%d", idx),
				LastName:        "Concur",
				Designation:     "Delivery Driver",
				EmploymentType:  employees.EmploymentTypeFullTime,
				OperationalRole: employees.OperationalRoleDriver,
				BranchID:        &branchID,
			}
			b, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				errorsList[idx] = fmt.Errorf("unexpected status %d: %s", w.Code, w.Body.String())
				return
			}

			var resp struct {
				Data employees.Employee `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				errorsList[idx] = err
				return
			}

			generatedCodes[idx] = resp.Data.EmployeeCode
		}(i)
	}

	wg.Wait()

	// Verify no errors occurred
	for i, err := range errorsList {
		if err != nil {
			t.Fatalf("goroutine %d failed: %v", i, err)
		}
	}

	// Verify all codes are unique and match EMP-XXXX format
	uniqueMap := make(map[string]bool)
	for _, code := range generatedCodes {
		if code == "" {
			t.Fatalf("empty employee code generated")
		}
		if uniqueMap[code] {
			t.Fatalf("duplicate employee code detected: %s", code)
		}
		uniqueMap[code] = true
	}

	if len(uniqueMap) != concurrency {
		t.Fatalf("expected %d unique codes, got %d", concurrency, len(uniqueMap))
	}
}

func TestPhase3_DriverEligibility_Validation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Eligible Fleet Corp", "elig_fleet")

	// 1. Create a non-driver employee (e.g. OPERATOR)
	opPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    fmt.Sprintf("OP-%d", time.Now().UnixNano()%100000),
		FirstName:       "Sarah",
		LastName:        "Connor",
		Designation:     "Control Room Operator",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleOperator,
	}
	opBody, _ := json.Marshal(opPayload)
	opReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(opBody))
	opReq.Header.Set("Content-Type", "application/json")
	opReq.Header.Set("Authorization", "Bearer "+token)
	opW := httptest.NewRecorder()
	router.ServeHTTP(opW, opReq)

	if opW.Code != http.StatusCreated {
		t.Fatalf("failed to create operator: %s", opW.Body.String())
	}
	var opResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(opW.Body.Bytes(), &opResp)
	operatorID := opResp.Data.ID

	// 2. Create vehicle
	vehPayload := vehicles.CreateVehicleRequest{
		RegistrationNumber: fmt.Sprintf("EV-%d", time.Now().UnixNano()%100000),
		VehicleType:        vehicles.VehicleTypeElectricVan,
		MaxWeightKG:        800,
		MaxVolumeCBM:       4.5,
	}
	vehBody, _ := json.Marshal(vehPayload)
	vReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantID), bytes.NewReader(vehBody))
	vReq.Header.Set("Content-Type", "application/json")
	vReq.Header.Set("Authorization", "Bearer "+token)
	vW := httptest.NewRecorder()
	router.ServeHTTP(vW, vReq)

	if vW.Code != http.StatusCreated {
		t.Fatalf("failed to create vehicle: %s", vW.Body.String())
	}
	var vehResp struct {
		Data vehicles.Vehicle `json:"data"`
	}
	_ = json.Unmarshal(vW.Body.Bytes(), &vehResp)
	vehicleID := vehResp.Data.ID

	// 3. Attempt to assign non-driver to vehicle -> MUST REJECT WITH 400
	assignPayload := vehicles.AssignVehicleRequest{
		DriverID: operatorID,
	}
	aBody, _ := json.Marshal(assignPayload)
	aReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantID, vehicleID), bytes.NewReader(aBody))
	aReq.Header.Set("Content-Type", "application/json")
	aReq.Header.Set("Authorization", "Bearer "+token)
	aW := httptest.NewRecorder()
	router.ServeHTTP(aW, aReq)

	if aW.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request when assigning non-driver, got: %d: %s", aW.Code, aW.Body.String())
	}
}

func TestPhase3_EmployeeStatusUpdate_And_AvailableDrivers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Status Fleet Corp", "stat_fleet")

	// 1. Create a driver
	drvPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    fmt.Sprintf("DRV-%d", time.Now().UnixNano()%100000),
		FirstName:       "Marcus",
		LastName:        "Wright",
		Designation:     "Fleet Driver",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
	}
	b, _ := json.Marshal(drvPayload)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create driver: %s", w.Body.String())
	}
	var drvResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &drvResp)
	driverID := drvResp.Data.ID

	// 2. Query available drivers -> must include Marcus
	availReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/available-drivers", tenantID), nil)
	availReq.Header.Set("Authorization", "Bearer "+token)
	availW := httptest.NewRecorder()
	router.ServeHTTP(availW, availReq)

	if availW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from available-drivers, got: %d: %s", availW.Code, availW.Body.String())
	}
	var availResp struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(availW.Body.Bytes(), &availResp)
	if availResp.Data.Total < 1 {
		t.Fatalf("expected at least 1 available driver, got %d", availResp.Data.Total)
	}

	// 3. Update driver status to ON_LEAVE and availability to OFF_DUTY via PATCH /status
	onLeave := employees.StatusOnLeave
	offDuty := employees.AvailabilityStatusOffDuty
	statusPayload := employees.UpdateEmployeeStatusRequest{
		Status:             &onLeave,
		AvailabilityStatus: &offDuty,
	}
	sBody, _ := json.Marshal(statusPayload)
	sReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/employees/%s/status", tenantID, driverID), bytes.NewReader(sBody))
	sReq.Header.Set("Content-Type", "application/json")
	sReq.Header.Set("Authorization", "Bearer "+token)
	sW := httptest.NewRecorder()
	router.ServeHTTP(sW, sReq)

	if sW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on status patch, got %d: %s", sW.Code, sW.Body.String())
	}

	// 4. Re-query available drivers -> Marcus should now be excluded!
	availReq2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/employees/available-drivers", tenantID), nil)
	availReq2.Header.Set("Authorization", "Bearer "+token)
	availW2 := httptest.NewRecorder()
	router.ServeHTTP(availW2, availReq2)

	var availResp2 struct {
		Data struct {
			Drivers []employees.Employee `json:"drivers"`
		} `json:"data"`
	}
	_ = json.Unmarshal(availW2.Body.Bytes(), &availResp2)
	for _, d := range availResp2.Data.Drivers {
		if d.ID == driverID {
			t.Fatalf("driver with OFF_DUTY status should not be returned in available drivers list")
		}
	}
}

func TestPhase3_VehicleStatus_And_AssignmentHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "History Fleet Corp", "hist_fleet")

	// 1. Create a driver
	drvPayload := employees.CreateEmployeeRequest{
		EmployeeCode:    fmt.Sprintf("DRV-%d", time.Now().UnixNano()%100000),
		FirstName:       "Elena",
		LastName:        "Rostova",
		Designation:     "Senior Driver",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
	}
	b, _ := json.Marshal(drvPayload)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var drvResp struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &drvResp)
	driverID := drvResp.Data.ID

	// 2. Create a vehicle
	vehPayload := vehicles.CreateVehicleRequest{
		RegistrationNumber: fmt.Sprintf("KA-%d", time.Now().UnixNano()%100000),
		VehicleType:        vehicles.VehicleTypeVan,
		MaxWeightKG:        1200,
		MaxVolumeCBM:       6.0,
	}
	vBody, _ := json.Marshal(vehPayload)
	vReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles", tenantID), bytes.NewReader(vBody))
	vReq.Header.Set("Content-Type", "application/json")
	vReq.Header.Set("Authorization", "Bearer "+token)
	vW := httptest.NewRecorder()
	router.ServeHTTP(vW, vReq)
	var vehResp struct {
		Data vehicles.Vehicle `json:"data"`
	}
	_ = json.Unmarshal(vW.Body.Bytes(), &vehResp)
	vehicleID := vehResp.Data.ID

	// 3. Assign driver to vehicle
	notes := "Route 42 morning run"
	assignPayload := vehicles.AssignVehicleRequest{
		DriverID: driverID,
		Notes:    &notes,
	}
	aBody, _ := json.Marshal(assignPayload)
	aReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assign", tenantID, vehicleID), bytes.NewReader(aBody))
	aReq.Header.Set("Content-Type", "application/json")
	aReq.Header.Set("Authorization", "Bearer "+token)
	aW := httptest.NewRecorder()
	router.ServeHTTP(aW, aReq)
	if aW.Code != http.StatusCreated {
		t.Fatalf("failed to assign driver: %s", aW.Body.String())
	}

	// 4. Query assignments endpoint: GET /api/v1/tenants/{tenant_id}/assignments
	listReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/assignments", tenantID), nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from assignments endpoint, got %d: %s", listW.Code, listW.Body.String())
	}

	// 5. Query vehicle assignment history endpoint
	vehHistReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/assignments", tenantID, vehicleID), nil)
	vehHistReq.Header.Set("Authorization", "Bearer "+token)
	vehHistW := httptest.NewRecorder()
	router.ServeHTTP(vehHistW, vehHistReq)

	if vehHistW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from vehicle assignments endpoint, got %d: %s", vehHistW.Code, vehHistW.Body.String())
	}

	// 6. Unassign driver
	unassignReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/unassign", tenantID, vehicleID), nil)
	unassignReq.Header.Set("Authorization", "Bearer "+token)
	unassignW := httptest.NewRecorder()
	router.ServeHTTP(unassignW, unassignReq)

	if unassignW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on unassign, got %d: %s", unassignW.Code, unassignW.Body.String())
	}

	// 7. Update vehicle status to MAINTENANCE via PATCH /status
	maintStatus := vehicles.VehicleStatusMaintenance
	maintAvail := vehicles.AvailabilityStatusMaintenance
	vehStatusPayload := vehicles.UpdateVehicleStatusRequest{
		Status:             &maintStatus,
		AvailabilityStatus: &maintAvail,
	}
	vsBody, _ := json.Marshal(vehStatusPayload)
	vsReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/vehicles/%s/status", tenantID, vehicleID), bytes.NewReader(vsBody))
	vsReq.Header.Set("Content-Type", "application/json")
	vsReq.Header.Set("Authorization", "Bearer "+token)
	vsW := httptest.NewRecorder()
	router.ServeHTTP(vsW, vsReq)

	if vsW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on vehicle status patch, got %d: %s", vsW.Code, vsW.Body.String())
	}
}
