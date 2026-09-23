package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompany_GetCurrentAndBranchStatusUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	token, _, tenantID := registerTestCompany(t, router, "Status Corp", "status_corp")

	// 1. Test GET /api/v1/companies/current
	reqCurrent, _ := http.NewRequest(http.MethodGet, "/api/v1/companies/current", nil)
	reqCurrent.Header.Set("Authorization", "Bearer "+token)
	wCurrent := httptest.NewRecorder()
	router.ServeHTTP(wCurrent, reqCurrent)

	assert.Equal(t, http.StatusOK, wCurrent.Code)
	var respCurrent struct {
		Success bool           `json:"success"`
		Data    tenants.Tenant `json:"data"`
	}
	err := json.Unmarshal(wCurrent.Body.Bytes(), &respCurrent)
	require.NoError(t, err)
	assert.NotEmpty(t, respCurrent.Data.ID)
	assert.Equal(t, tenantID, respCurrent.Data.ID)

	// 2. Test GET /api/v1/tenants/current (alias)
	reqTenantCurrent, _ := http.NewRequest(http.MethodGet, "/api/v1/tenants/current", nil)
	reqTenantCurrent.Header.Set("Authorization", "Bearer "+token)
	wTenantCurrent := httptest.NewRecorder()
	router.ServeHTTP(wTenantCurrent, reqTenantCurrent)

	assert.Equal(t, http.StatusOK, wTenantCurrent.Code)

	// 3. Create a branch under the company
	branchCode := fmt.Sprintf("ST-%d", time.Now().UnixNano()%100000)
	createBody := map[string]interface{}{
		"branch_code":        branchCode,
		"name":               "Status Test Hub",
		"address":            "100 Status Way",
		"city":               "Coimbatore",
		"latitude":           11.0168,
		"longitude":          76.9558,
		"coverage_radius_km": 20.0,
	}
	bodyBytes, _ := json.Marshal(createBody)
	reqCreate, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/companies/%s/branches", tenantID), bytes.NewReader(bodyBytes))
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)

	require.Equal(t, http.StatusCreated, wCreate.Code)
	var respBranch struct {
		Success bool            `json:"success"`
		Data    branches.Branch `json:"data"`
	}
	err = json.Unmarshal(wCreate.Body.Bytes(), &respBranch)
	require.NoError(t, err)
	branchID := respBranch.Data.ID
	assert.Equal(t, "ACTIVE", respBranch.Data.Status)
	assert.True(t, respBranch.Data.IsActive)

	// 4. Test PATCH /api/v1/companies/:company_id/branches/:branch_id/status -> INACTIVE
	statusBody, _ := json.Marshal(map[string]string{
		"status": "INACTIVE",
	})
	reqStatus, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/companies/%s/branches/%s/status", tenantID, branchID), bytes.NewReader(statusBody))
	reqStatus.Header.Set("Authorization", "Bearer "+token)
	reqStatus.Header.Set("Content-Type", "application/json")
	wStatus := httptest.NewRecorder()
	router.ServeHTTP(wStatus, reqStatus)

	assert.Equal(t, http.StatusOK, wStatus.Code)
	var respUpdated struct {
		Success bool            `json:"success"`
		Data    branches.Branch `json:"data"`
	}
	err = json.Unmarshal(wStatus.Body.Bytes(), &respUpdated)
	require.NoError(t, err)
	assert.Equal(t, "INACTIVE", respUpdated.Data.Status)
	assert.False(t, respUpdated.Data.IsActive)

	// 5. Test PATCH status with invalid value -> 400 Bad Request
	badStatusBody, _ := json.Marshal(map[string]string{
		"status": "INVALID_STATUS_VALUE",
	})
	reqBadStatus, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/companies/%s/branches/%s/status", tenantID, branchID), bytes.NewReader(badStatusBody))
	reqBadStatus.Header.Set("Authorization", "Bearer "+token)
	reqBadStatus.Header.Set("Content-Type", "application/json")
	wBadStatus := httptest.NewRecorder()
	router.ServeHTTP(wBadStatus, reqBadStatus)

	assert.Equal(t, http.StatusBadRequest, wBadStatus.Code)

	// 6. Test PATCH status under another tenant -> 403 Forbidden
	_, _, otherTenantID := registerTestCompany(t, router, "Other Corp", "other_corp")
	reqForbidden, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/companies/%s/branches/%s/status", otherTenantID, branchID), bytes.NewReader(statusBody))
	reqForbidden.Header.Set("Authorization", "Bearer "+token)
	reqForbidden.Header.Set("Content-Type", "application/json")
	wForbidden := httptest.NewRecorder()
	router.ServeHTTP(wForbidden, reqForbidden)

	assert.Equal(t, http.StatusForbidden, wForbidden.Code)
}
