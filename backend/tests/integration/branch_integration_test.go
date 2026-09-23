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
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/tenants"
)

// TC-P2-BRN-001: Create Delivery Branch with PostGIS Coordinates
func TestBranch_Create_WithPostGIS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Atlas Logistics", "atlas_branch")

	uniqueID := time.Now().UnixNano()
	branchCode := fmt.Sprintf("BRN-DEL-%d", uniqueID%100000)

	payload := branches.CreateBranchRequest{
		BranchCode:       branchCode,
		Name:             "Delhi North Sorting Hub",
		Address:          "Plot 12, Industrial Area, GT Karnal Rd",
		City:             "Delhi",
		State:            "Delhi",
		PostalCode:       "110033",
		Country:          "India",
		Latitude:         28.7041,
		Longitude:        77.1025,
		CoverageRadiusKM: 25.0,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got: %d %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data branches.Branch `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Data.BranchCode != branchCode {
		t.Errorf("expected branch code %s, got: %s", branchCode, resp.Data.BranchCode)
	}
	if resp.Data.Latitude != 28.7041 || resp.Data.Longitude != 77.1025 {
		t.Errorf("expected coordinates 28.7041, 77.1025, got: %f, %f", resp.Data.Latitude, resp.Data.Longitude)
	}
	if !resp.Data.IsActive || resp.Data.Status != branches.StatusActive {
		t.Errorf("expected active status, got: is_active=%v, status=%s", resp.Data.IsActive, resp.Data.Status)
	}
}

// TC-P2-BRN-002: Reject Duplicate Branch Code Within Same Tenant
func TestBranch_DuplicateCode_RejectedWithinTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Chronos Freight", "chronos_dup")

	branchCode := fmt.Sprintf("DUP-%d", time.Now().UnixNano()%100000)

	payload := branches.CreateBranchRequest{
		BranchCode:       branchCode,
		Name:             "Main Depot",
		Address:          "Main St 1",
		City:             "Mumbai",
		Latitude:         18.9220,
		Longitude:        72.8347,
		CoverageRadiusKM: 15.0,
	}
	body, _ := json.Marshal(payload)

	// First creation: 201
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(body))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected first creation 201, got: %d", w1.Code)
	}

	// Second creation with same branch code: 409 Conflict
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate branch code, got: %d %s", w2.Code, w2.Body.String())
	}
}

// Identical branch codes are permitted across different tenants
func TestBranch_IdenticalCode_AllowedInDifferentTenants(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Alpha Express", "alpha_branch")
	tokenB, _, tenantIDB := registerTestCompany(t, router, "Beta Express", "beta_branch")

	sharedCode := fmt.Sprintf("HQ-%d", time.Now().UnixNano()%100000)

	payloadA := branches.CreateBranchRequest{
		BranchCode: sharedCode,
		Name:       "Alpha Headquarters",
		Address:    "Alpha Street",
		City:       "Bengaluru",
		Latitude:   12.9716,
		Longitude:  77.5946,
	}
	bodyA, _ := json.Marshal(payloadA)

	wA := httptest.NewRecorder()
	reqA := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDA), bytes.NewReader(bodyA))
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	reqA.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusCreated {
		t.Fatalf("tenant A creation failed: %d %s", wA.Code, wA.Body.String())
	}

	payloadB := branches.CreateBranchRequest{
		BranchCode: sharedCode,
		Name:       "Beta Headquarters",
		Address:    "Beta Street",
		City:       "Chennai",
		Latitude:   13.0827,
		Longitude:  80.2707,
	}
	bodyB, _ := json.Marshal(payloadB)

	wB := httptest.NewRecorder()
	reqB := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDB), bytes.NewReader(bodyB))
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	reqB.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusCreated {
		t.Fatalf("tenant B creation failed for same branch code in different tenant: %d %s", wB.Code, wB.Body.String())
	}
}

// TC-P2-BRN-003: Enforce Multi-Tenant Isolation on Branch Retrieval
func TestBranch_CrossTenantAccess_Forbidden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Tenant Alpha", "t_alpha")
	tokenB, _, tenantIDB := registerTestCompany(t, router, "Tenant Beta", "t_beta")

	// Create branch in Tenant A
	payload := branches.CreateBranchRequest{
		BranchCode: fmt.Sprintf("SEC-%d", time.Now().UnixNano()%100000),
		Name:       "Secret Branch",
		Address:    "Secret Rd",
		City:       "Delhi",
		Latitude:   28.6139,
		Longitude:  77.2090,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantIDA), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create branch: %d", w.Code)
	}

	var resp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	branchA_ID := resp.Data.ID

	// User B attempts to access Tenant A's branch: 403 Forbidden
	wCross := httptest.NewRecorder()
	reqCross := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantIDA, branchA_ID), nil)
	reqCross.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(wCross, reqCross)

	if wCross.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on cross-tenant branch access, got: %d", wCross.Code)
	}

	// User B attempts to use Tenant B path with Tenant A branch ID: 404 Not Found (scoped by tenant_id)
	wIDOR := httptest.NewRecorder()
	reqIDOR := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantIDB, branchA_ID), nil)
	reqIDOR.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(wIDOR, reqIDOR)

	if wIDOR.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found when referencing foreign branch within own tenant path, got: %d", wIDOR.Code)
	}
}

// TC-P2-BRN-004: Branch Listing, Spatial Filtering, Update & Soft Deactivation
func TestBranch_List_And_SpatialFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Geo Logistics", "geo_log")

	// 1. Create Delhi Hub (28.7041, 77.1025)
	code1 := fmt.Sprintf("DEL-%d", time.Now().UnixNano()%10000)
	body1, _ := json.Marshal(branches.CreateBranchRequest{
		BranchCode: code1,
		Name:       "Delhi Hub",
		Address:    "North Delhi",
		City:       "Delhi",
		Latitude:   28.7041,
		Longitude:  77.1025,
	})
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(body1))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("failed to create hub 1: %d %s", w1.Code, w1.Body.String())
	}
	var created1 struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(w1.Body.Bytes(), &created1)

	// 2. Create Gurgaon Hub (28.4595, 77.0266) ~30km away
	code2 := fmt.Sprintf("GGN-%d", time.Now().UnixNano()%10000)
	body2, _ := json.Marshal(branches.CreateBranchRequest{
		BranchCode: code2,
		Name:       "Gurgaon Hub",
		Address:    "Cyber City",
		City:       "Gurgaon",
		Latitude:   28.4595,
		Longitude:  77.0266,
	})
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(body2))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("failed to create hub 2: %d", w2.Code)
	}

	// 3. Create Mumbai Hub (19.0760, 72.8777) ~1150km away
	code3 := fmt.Sprintf("BOM-%d", time.Now().UnixNano()%10000)
	body3, _ := json.Marshal(branches.CreateBranchRequest{
		BranchCode: code3,
		Name:       "Mumbai Port Hub",
		Address:    "South Mumbai",
		City:       "Mumbai",
		Latitude:   19.0760,
		Longitude:  72.8777,
	})
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", tenantID), bytes.NewReader(body3))
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("failed to create hub 3: %d", w3.Code)
	}

	// 4. Spatial query: near Delhi (lat: 28.70, lng: 77.10, radius: 45km)
	wSpatial := httptest.NewRecorder()
	reqSpatial := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches?near_lat=28.70&near_lng=77.10&radius_km=45", tenantID), nil)
	reqSpatial.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wSpatial, reqSpatial)

	if wSpatial.Code != http.StatusOK {
		t.Fatalf("spatial query failed: %d %s", wSpatial.Code, wSpatial.Body.String())
	}

	var spatialResp struct {
		Data branches.BranchListResponse `json:"data"`
	}
	if err := json.Unmarshal(wSpatial.Body.Bytes(), &spatialResp); err != nil {
		t.Fatalf("failed to decode spatial response: %v", err)
	}

	if spatialResp.Data.Total != 2 {
		t.Fatalf("expected exactly 2 branches within 45km of Delhi, got %d", spatialResp.Data.Total)
	}
	if spatialResp.Data.Branches[0].DistanceKM == nil {
		t.Error("expected distance_km populated for spatial query")
	}

	// 5. Update Branch 1
	newName := "Delhi Super Hub Updated"
	updateBody, _ := json.Marshal(branches.UpdateBranchRequest{
		Name: &newName,
	})
	wUpdate := httptest.NewRecorder()
	reqUpdate := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantID, created1.Data.ID), bytes.NewReader(updateBody))
	reqUpdate.Header.Set("Authorization", "Bearer "+token)
	reqUpdate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wUpdate, reqUpdate)

	if wUpdate.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", wUpdate.Code, wUpdate.Body.String())
	}

	// 6. Soft-deactivate Branch 1
	wDelete := httptest.NewRecorder()
	reqDelete := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantID, created1.Data.ID), nil)
	reqDelete.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wDelete, reqDelete)

	if wDelete.Code != http.StatusOK {
		t.Fatalf("delete failed: %d %s", wDelete.Code, wDelete.Body.String())
	}

	// Verify status after soft-delete
	wGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/branches/%s", tenantID, created1.Data.ID), nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(wGet, reqGet)

	var getResp struct {
		Data branches.Branch `json:"data"`
	}
	_ = json.Unmarshal(wGet.Body.Bytes(), &getResp)
	if getResp.Data.IsActive || getResp.Data.Status != branches.StatusInactive {
		t.Errorf("expected deactivated status, got is_active=%v, status=%s", getResp.Data.IsActive, getResp.Data.Status)
	}
}

// Viewer role cannot create branch
func TestBranch_ViewerRole_ForbiddenFromBranchCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	ownerToken, _, ownerTenantID := registerTestCompany(t, router, "Invite Owner", "inv_owner")
	viewerEmail := fmt.Sprintf("invited_viewer_%d@test.com", time.Now().UnixNano())
	regViewerBody, _ := json.Marshal(auth.RegisterRequest{
		Email:       viewerEmail,
		Password:    "SecureP@ssw0rd!2026",
		FullName:    "Invited Viewer",
		CompanyName: "Viewer Standalone",
	})
	wReg := httptest.NewRecorder()
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regViewerBody))
	reqReg.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wReg, reqReg)

	var regResp struct {
		Data auth.AuthResponse `json:"data"`
	}
	_ = json.Unmarshal(wReg.Body.Bytes(), &regResp)

	// Owner invites viewer with role VIEWER
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

	// Viewer attempts to create branch: 403 Forbidden
	branchBody, _ := json.Marshal(branches.CreateBranchRequest{
		BranchCode: "VIEW-01",
		Name:       "Viewer Unauthorized Hub",
		Address:    "No Perms St",
		City:       "Delhi",
		Latitude:   28.6139,
		Longitude:  77.2090,
	})
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/branches", ownerTenantID), bytes.NewReader(branchBody))
	reqCreate.Header.Set("Authorization", "Bearer "+regResp.Data.Token)
	reqCreate.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when Viewer attempts branch creation, got: %d %s", wCreate.Code, wCreate.Body.String())
	}
}
