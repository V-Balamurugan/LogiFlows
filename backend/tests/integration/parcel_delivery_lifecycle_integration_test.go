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
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/deliveries"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/parcels"
	"github.com/logiflows/logiflows/backend/internal/transfers"
)

// Helper to create test branch
func createTestBranch(t *testing.T, router http.Handler, token string, tenantID uuid.UUID, name, city string, lat, lng float64) uuid.UUID {
	reqPayload := branches.CreateBranchRequest{
		BranchCode:       fmt.Sprintf("BRN-%d", time.Now().UnixNano()%1000000),
		Name:             name,
		Address:          "Main Road",
		City:             city,
		Latitude:         lat,
		Longitude:        lng,
		CoverageRadiusKM: 25,
	}
	b, _ := json.Marshal(reqPayload)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID.String()), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for branch, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res.Data.ID
}

// Helper to create test driver
func createTestDriver(t *testing.T, router http.Handler, token string, tenantID uuid.UUID, branchID uuid.UUID) uuid.UUID {
	license := fmt.Sprintf("DL-%d", time.Now().UnixNano()%1000000)
	empPayload := employees.CreateEmployeeRequest{
		BranchID:        &branchID,
		FirstName:       "Ramesh",
		LastName:        "Kumar",
		Designation:     "Senior Delivery Courier",
		EmploymentType:  employees.EmploymentTypeFullTime,
		OperationalRole: employees.OperationalRoleDriver,
		LicenseNumber:   &license,
	}
	b, _ := json.Marshal(empPayload)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/employees", tenantID.String()), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for driver employee, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data employees.Employee `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res.Data.ID
}

// TC-P4-PARCEL-001: Parcel Creation, Auto Tracking Number, Details & Listing
func TestParcel_Create_Get_List_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Express Movers", "exp_mov")

	origBranchID := createTestBranch(t, router, token, tenantID, "Chennai Sorting Hub", "Chennai", 13.0827, 80.2707)
	destBranchID := createTestBranch(t, router, token, tenantID, "Bengaluru Delivery Hub", "Bengaluru", 12.9716, 77.5946)

	// 1. Create Parcel
	createReq := parcels.CreateParcelRequest{
		SenderName:          "Arun Electronics",
		SenderPhone:         "+91 9876543210",
		SenderAddress:       "10 Anna Salai, Chennai",
		ReceiverName:        "Priya Sharma",
		ReceiverPhone:       "+91 9123456789",
		ReceiverAddress:     "22 Indiranagar, Bengaluru",
		OriginBranchID:      origBranchID.String(),
		DestinationBranchID: destBranchID.String(),
		WeightKG:            3.5,
		DimensionsCM:        "30x20x15",
		ServiceType:         parcels.ServiceTypeExpress,
	}
	b, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantID), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for parcel, got %d: %s", w.Code, w.Body.String())
	}

	var createResp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	parcel := createResp.Data

	if parcel.ID == uuid.Nil {
		t.Fatalf("expected non-nil parcel ID")
	}
	if !bytes.HasPrefix([]byte(parcel.TrackingNumber), []byte("PKG-")) {
		t.Fatalf("expected tracking number format PKG-..., got %s", parcel.TrackingNumber)
	}
	if parcel.Status != parcels.StatusCreated {
		t.Fatalf("expected initial status CREATED, got %s", parcel.Status)
	}

	// 2. Get Parcel Details
	wGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s", tenantID, parcel.ID), nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for parcel get, got %d: %s", wGet.Code, wGet.Body.String())
	}

	// 3. Get Parcel Timeline
	wTime := httptest.NewRecorder()
	reqTime := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s/timeline", tenantID, parcel.ID), nil)
	reqTime.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wTime, reqTime)

	if wTime.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for parcel timeline, got %d: %s", wTime.Code, wTime.Body.String())
	}

	// 4. List Parcels with search
	wList := httptest.NewRecorder()
	reqList := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels?search=%s", tenantID, parcel.TrackingNumber), nil)
	reqList.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wList, reqList)

	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for parcel listing, got %d: %s", wList.Code, wList.Body.String())
	}
	var listResp struct {
		Data parcels.ParcelListResponse `json:"data"`
	}
	_ = json.Unmarshal(wList.Body.Bytes(), &listResp)
	if listResp.Data.Total == 0 {
		t.Fatalf("expected search to find parcel")
	}
}

// TC-P4-PARCEL-002: Multi-Tenant Isolation
func TestParcel_MultiTenant_Isolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantA := registerTestCompany(t, router, "Alpha Logistics", "alpha_iso")
	tokenB, _, tenantB := registerTestCompany(t, router, "Beta Freight", "beta_iso")

	origA := createTestBranch(t, router, tokenA, tenantA, "Alpha Hub 1", "Chennai", 13.0827, 80.2707)
	destA := createTestBranch(t, router, tokenA, tenantA, "Alpha Hub 2", "Coimbatore", 11.0168, 76.9558)
	origB := createTestBranch(t, router, tokenB, tenantB, "Beta Hub 1", "Madurai", 9.9252, 78.1198)

	// Create parcel in Tenant A
	createReq := parcels.CreateParcelRequest{
		SenderName:          "Sender A",
		SenderPhone:         "+91 9000000001",
		SenderAddress:       "Chennai",
		ReceiverName:        "Receiver A",
		ReceiverPhone:       "+91 9000000002",
		ReceiverAddress:     "Coimbatore",
		OriginBranchID:      origA.String(),
		DestinationBranchID: destA.String(),
		WeightKG:            1.5,
		DimensionsCM:        "10x10x10",
		ServiceType:         parcels.ServiceTypeStandard,
	}
	b, _ := json.Marshal(createReq)
	wA := httptest.NewRecorder()
	reqA := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantA), bytes.NewReader(b))
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	reqA.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for Tenant A parcel, got %d: %s", wA.Code, wA.Body.String())
	}
	var resA struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wA.Body.Bytes(), &resA)
	parcelA := resA.Data

	// 1. Tenant B attempts to read Tenant A's parcel -> Expect 404 Not Found
	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s", tenantB, parcelA.ID), nil)
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for cross-tenant parcel access, got %d: %s", wB.Code, wB.Body.String())
	}

	// 2. Tenant A attempts to use Tenant B's branch in parcel creation -> Expect 403 Forbidden
	crossBranchReq := createReq
	crossBranchReq.DestinationBranchID = origB.String()
	bCross, _ := json.Marshal(crossBranchReq)
	wCross := httptest.NewRecorder()
	reqCross := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantA), bytes.NewReader(bCross))
	reqCross.Header.Set("Authorization", "Bearer "+tokenA)
	reqCross.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCross, reqCross)
	if wCross.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when creating parcel with foreign tenant branch, got %d: %s", wCross.Code, wCross.Body.String())
	}
}

// TC-P4-PARCEL-003: State Machine Validation & Transition Enforcement
func TestParcel_Status_StateMachine_Enforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Rapid Courier", "rapid_fsm")

	origID := createTestBranch(t, router, token, tenantID, "Hub Alpha", "Chennai", 13.08, 80.27)
	destID := createTestBranch(t, router, token, tenantID, "Hub Beta", "Bengaluru", 12.97, 77.59)

	// Create parcel
	createReq := parcels.CreateParcelRequest{
		SenderName:          "Sender",
		SenderPhone:         "+91 9000000001",
		SenderAddress:       "Chennai",
		ReceiverName:        "Receiver",
		ReceiverPhone:       "+91 9000000002",
		ReceiverAddress:     "Bengaluru",
		OriginBranchID:      origID.String(),
		DestinationBranchID: destID.String(),
		WeightKG:            2.0,
		DimensionsCM:        "20x20x20",
		ServiceType:         parcels.ServiceTypeStandard,
	}
	b, _ := json.Marshal(createReq)
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantID), bytes.NewReader(b))
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	reqCreate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCreate, reqCreate)
	var createResp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)
	parcelID := createResp.Data.ID

	// 1. Valid Transition: CREATED -> BOOKED
	trans1 := parcels.UpdateParcelStatusRequest{Status: parcels.StatusBooked}
	b1, _ := json.Marshal(trans1)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s/status", tenantID, parcelID), bytes.NewReader(b1))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for CREATED -> BOOKED, got %d: %s", w1.Code, w1.Body.String())
	}

	// 2. Illegal Transition: BOOKED -> DELIVERED (Must be rejected)
	illegalTrans := parcels.UpdateParcelStatusRequest{Status: parcels.StatusDelivered}
	bIllegal, _ := json.Marshal(illegalTrans)
	wIllegal := httptest.NewRecorder()
	reqIllegal := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s/status", tenantID, parcelID), bytes.NewReader(bIllegal))
	reqIllegal.Header.Set("Authorization", "Bearer "+token)
	reqIllegal.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wIllegal, reqIllegal)
	if wIllegal.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for illegal transition BOOKED -> DELIVERED, got %d: %s", wIllegal.Code, wIllegal.Body.String())
	}

	// 3. Valid Transition: BOOKED -> RECEIVED_AT_ORIGIN_BRANCH
	trans2 := parcels.UpdateParcelStatusRequest{Status: parcels.StatusReceivedAtOriginBranch}
	b2, _ := json.Marshal(trans2)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s/status", tenantID, parcelID), bytes.NewReader(b2))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for BOOKED -> RECEIVED_AT_ORIGIN_BRANCH, got %d: %s", w2.Code, w2.Body.String())
	}
}

// TC-P4-DELIVERY-001: Delivery Task Dispatch, Conflict Prevention & POD Completion
func TestDelivery_Task_Assignment_Conflict_Prevention_And_POD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Apex Delivery", "apex_pod")

	origID := createTestBranch(t, router, token, tenantID, "Apex Hub 1", "Chennai", 13.08, 80.27)
	destID := createTestBranch(t, router, token, tenantID, "Apex Hub 2", "Bengaluru", 12.97, 77.59)
	driverID := createTestDriver(t, router, token, tenantID, origID)

	// Create parcel and check-in at origin
	createReq := parcels.CreateParcelRequest{
		SenderName:          "Sender",
		SenderPhone:         "+91 9000000001",
		SenderAddress:       "Chennai",
		ReceiverName:        "Receiver",
		ReceiverPhone:       "+91 9000000002",
		ReceiverAddress:     "Bengaluru",
		OriginBranchID:      origID.String(),
		DestinationBranchID: destID.String(),
		WeightKG:            1.8,
		DimensionsCM:        "15x15x15",
		ServiceType:         parcels.ServiceTypeSameDay,
	}
	b, _ := json.Marshal(createReq)
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantID), bytes.NewReader(b))
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	reqCreate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCreate, reqCreate)
	var cResp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &cResp)
	parcelID := cResp.Data.ID

	// Check-in parcel to RECEIVED_AT_ORIGIN_BRANCH so it is eligible for delivery
	statusReq := parcels.UpdateParcelStatusRequest{Status: parcels.StatusReceivedAtOriginBranch}
	bSt, _ := json.Marshal(statusReq)
	wSt := httptest.NewRecorder()
	reqSt := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s/status", tenantID, parcelID), bytes.NewReader(bSt))
	reqSt.Header.Set("Authorization", "Bearer "+token)
	reqSt.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wSt, reqSt)

	// 1. Create Delivery Task
	taskReq := deliveries.CreateDeliveryTaskRequest{
		ParcelID:         parcelID.String(),
		AssignedDriverID: driverID.String(),
	}
	bTask, _ := json.Marshal(taskReq)
	wTask := httptest.NewRecorder()
	reqTask := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/deliveries", tenantID), bytes.NewReader(bTask))
	reqTask.Header.Set("Authorization", "Bearer "+token)
	reqTask.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wTask, reqTask)

	if wTask.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for delivery task, got %d: %s", wTask.Code, wTask.Body.String())
	}
	var tResp struct {
		Data deliveries.DeliveryTask `json:"data"`
	}
	_ = json.Unmarshal(wTask.Body.Bytes(), &tResp)
	deliveryTaskID := tResp.Data.ID

	// 2. Concurrency Conflict: Attempt to create another active delivery task for the same parcel -> Expect 409 Conflict
	wDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/deliveries", tenantID), bytes.NewReader(bTask))
	reqDup.Header.Set("Authorization", "Bearer "+token)
	reqDup.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wDup, reqDup)

	if wDup.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate delivery task for same parcel, got %d: %s", wDup.Code, wDup.Body.String())
	}

	// 3. Record Delivery Attempt (CUSTOMER_UNAVAILABLE)
	attemptNotes := "Door locked, no response to phone"
	attemptReq := deliveries.RecordDeliveryAttemptRequest{
		Outcome: deliveries.OutcomeCustomerUnavailable,
		Notes:   &attemptNotes,
	}
	bAtt, _ := json.Marshal(attemptReq)
	wAtt := httptest.NewRecorder()
	reqAtt := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/deliveries/%s/attempt", tenantID, deliveryTaskID), bytes.NewReader(bAtt))
	reqAtt.Header.Set("Authorization", "Bearer "+token)
	reqAtt.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wAtt, reqAtt)

	if wAtt.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for delivery attempt, got %d: %s", wAtt.Code, wAtt.Body.String())
	}

	// 4. Submit Proof of Delivery (Delivered successfully)
	podReq := deliveries.SubmitDeliveryProofRequest{
		ProofType:     deliveries.ProofTypeRecipientSignature,
		RecipientName: "Priya Sharma",
	}
	bPOD, _ := json.Marshal(podReq)
	wPOD := httptest.NewRecorder()
	reqPOD := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/deliveries/%s/proof", tenantID, deliveryTaskID), bytes.NewReader(bPOD))
	reqPOD.Header.Set("Authorization", "Bearer "+token)
	reqPOD.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPOD, reqPOD)

	if wPOD.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for proof of delivery submission, got %d: %s", wPOD.Code, wPOD.Body.String())
	}

	// Verify parcel status is now DELIVERED
	wCheck := httptest.NewRecorder()
	reqCheck := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s", tenantID, parcelID), nil)
	reqCheck.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wCheck, reqCheck)
	var checkResp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wCheck.Body.Bytes(), &checkResp)
	if checkResp.Data.Status != parcels.StatusDelivered {
		t.Fatalf("expected parcel status DELIVERED after POD, got %s", checkResp.Data.Status)
	}
}

// TC-P4-TRANSFER-001: Branch Transfer Manifest Dispatch & Receiving
func TestBranch_Transfer_Custody_Workflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Interstate Cargo", "cargo_trf")

	origID := createTestBranch(t, router, token, tenantID, "Hub Chennai", "Chennai", 13.08, 80.27)
	destID := createTestBranch(t, router, token, tenantID, "Hub Hyderabad", "Hyderabad", 17.38, 78.48)

	// Create parcel
	createReq := parcels.CreateParcelRequest{
		SenderName:          "Sender",
		SenderPhone:         "+91 9000000001",
		SenderAddress:       "Chennai",
		ReceiverName:        "Receiver",
		ReceiverPhone:       "+91 9000000002",
		ReceiverAddress:     "Hyderabad",
		OriginBranchID:      origID.String(),
		DestinationBranchID: destID.String(),
		WeightKG:            5.0,
		DimensionsCM:        "50x40x30",
		ServiceType:         parcels.ServiceTypeStandard,
	}
	b, _ := json.Marshal(createReq)
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantID), bytes.NewReader(b))
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	reqCreate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCreate, reqCreate)
	var cResp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &cResp)
	parcelID := cResp.Data.ID

	// 1. Create Transfer Manifest
	trfReq := transfers.CreateBranchTransferRequest{
		SourceBranchID:      origID.String(),
		DestinationBranchID: destID.String(),
		ParcelIDs:           []string{parcelID.String()},
	}
	bTrf, _ := json.Marshal(trfReq)
	wTrf := httptest.NewRecorder()
	reqTrf := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/transfers", tenantID), bytes.NewReader(bTrf))
	reqTrf.Header.Set("Authorization", "Bearer "+token)
	reqTrf.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wTrf, reqTrf)

	if wTrf.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for branch transfer, got %d: %s", wTrf.Code, wTrf.Body.String())
	}
	var trfResp struct {
		Data transfers.BranchTransfer `json:"data"`
	}
	_ = json.Unmarshal(wTrf.Body.Bytes(), &trfResp)
	transferID := trfResp.Data.ID

	// 2. Dispatch Transfer
	wDisp := httptest.NewRecorder()
	reqDisp := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/transfers/%s/dispatch", tenantID, transferID), bytes.NewReader([]byte("{}")))
	reqDisp.Header.Set("Authorization", "Bearer "+token)
	reqDisp.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wDisp, reqDisp)

	if wDisp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for dispatch, got %d: %s", wDisp.Code, wDisp.Body.String())
	}

	// Verify parcel is IN_TRANSIT
	wP1 := httptest.NewRecorder()
	reqP1 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s", tenantID, parcelID), nil)
	reqP1.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wP1, reqP1)
	var p1Resp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wP1.Body.Bytes(), &p1Resp)
	if p1Resp.Data.Status != parcels.StatusInTransit {
		t.Fatalf("expected parcel status IN_TRANSIT, got %s", p1Resp.Data.Status)
	}

	// 3. Receive Transfer at Destination Branch
	wRecv := httptest.NewRecorder()
	reqRecv := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/transfers/%s/receive", tenantID, transferID), bytes.NewReader([]byte("{}")))
	reqRecv.Header.Set("Authorization", "Bearer "+token)
	reqRecv.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wRecv, reqRecv)

	if wRecv.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for receiving, got %d: %s", wRecv.Code, wRecv.Body.String())
	}

	// Verify parcel is RECEIVED_AT_TRANSFER_BRANCH with current_branch_id = destID
	wP2 := httptest.NewRecorder()
	reqP2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s", tenantID, parcelID), nil)
	reqP2.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wP2, reqP2)
	var p2Resp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wP2.Body.Bytes(), &p2Resp)
	if p2Resp.Data.Status != parcels.StatusReceivedAtTransferBranch {
		t.Fatalf("expected parcel status RECEIVED_AT_TRANSFER_BRANCH, got %s", p2Resp.Data.Status)
	}
	if p2Resp.Data.CurrentBranchID == nil || *p2Resp.Data.CurrentBranchID != destID {
		t.Fatalf("expected parcel current_branch_id to be destination branch ID")
	}
}

// TC-P4-TRACKING-001: Public Tracking and Barcode Scanning Verification
func TestQR_Scan_And_Public_Tracking(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Omni Track", "omni_trk")

	origID := createTestBranch(t, router, token, tenantID, "Origin Hub", "Kochi", 9.9312, 76.2673)
	destID := createTestBranch(t, router, token, tenantID, "Dest Hub", "Trivandrum", 8.5241, 76.9366)

	// Create parcel
	createReq := parcels.CreateParcelRequest{
		SenderName:          "Confidential Sender Corp",
		SenderPhone:         "+91 9999999999",
		SenderAddress:       "Secret House 42, Marine Drive, Kochi",
		ReceiverName:        "VIP Recipient",
		ReceiverPhone:       "+91 8888888888",
		ReceiverAddress:     "Palace Road, Kowdiar, Trivandrum",
		OriginBranchID:      origID.String(),
		DestinationBranchID: destID.String(),
		WeightKG:            1.2,
		DimensionsCM:        "12x12x12",
		ServiceType:         parcels.ServiceTypeOvernight,
	}
	b, _ := json.Marshal(createReq)
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantID), bytes.NewReader(b))
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	reqCreate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCreate, reqCreate)
	var cResp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &cResp)
	parcel := cResp.Data

	// 1. Get QR code payload
	wQR := httptest.NewRecorder()
	reqQR := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/parcels/%s/qr", tenantID, parcel.ID), nil)
	reqQR.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wQR, reqQR)
	if wQR.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for QR get, got %d: %s", wQR.Code, wQR.Body.String())
	}
	var qrResp struct {
		Data struct {
			QRPayload string `json:"qr_payload"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wQR.Body.Bytes(), &qrResp)
	qrPayload := qrResp.Data.QRPayload

	// 2. Scan Parcel via Scanner Endpoint
	scanReq := parcels.ScanParcelRequest{
		QRPayload: &qrPayload,
	}
	bScan, _ := json.Marshal(scanReq)
	wScan := httptest.NewRecorder()
	reqScan := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels/scan", tenantID), bytes.NewReader(bScan))
	reqScan.Header.Set("Authorization", "Bearer "+token)
	reqScan.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wScan, reqScan)

	if wScan.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for barcode scan, got %d: %s", wScan.Code, wScan.Body.String())
	}

	// 3. Public Customer Tracking (NO Authorization Bearer Token)
	wPub := httptest.NewRecorder()
	reqPub := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tracking/%s", parcel.TrackingNumber), nil)
	// Notice: NO Authorization header set!
	router.ServeHTTP(wPub, reqPub)

	if wPub.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for unauthenticated public tracking, got %d: %s", wPub.Code, wPub.Body.String())
	}

	var pubResp struct {
		Data parcels.PublicTrackingResponse `json:"data"`
	}
	_ = json.Unmarshal(wPub.Body.Bytes(), &pubResp)
	info := pubResp.Data

	if info.TrackingNumber != parcel.TrackingNumber {
		t.Fatalf("expected tracking number %s, got %s", parcel.TrackingNumber, info.TrackingNumber)
	}
	if info.OriginCity != "Kochi" {
		t.Fatalf("expected origin city Kochi, got %s", info.OriginCity)
	}
	if info.DestinationCity != "Trivandrum" {
		t.Fatalf("expected destination city Trivandrum, got %s", info.DestinationCity)
	}

	// Security Verification: Ensure response body does NOT leak sensitive customer phone number or secret house address
	bodyStr := wPub.Body.String()
	if bytes.Contains([]byte(bodyStr), []byte("9999999999")) {
		t.Fatalf("SECURITY VIOLATION: public tracking response leaks customer sender phone number")
	}
	if bytes.Contains([]byte(bodyStr), []byte("8888888888")) {
		t.Fatalf("SECURITY VIOLATION: public tracking response leaks customer receiver phone number")
	}
	if bytes.Contains([]byte(bodyStr), []byte("Secret House 42")) {
		t.Fatalf("SECURITY VIOLATION: public tracking response leaks customer full street address")
	}
}
