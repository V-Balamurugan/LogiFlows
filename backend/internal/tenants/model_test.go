package tenants_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/tenants"
)

func TestTenant_StatusConstants(t *testing.T) {
	if tenants.StatusActive != "ACTIVE" {
		t.Errorf("expected StatusActive to be 'ACTIVE', got %s", tenants.StatusActive)
	}
	if tenants.StatusSuspended != "SUSPENDED" {
		t.Errorf("expected StatusSuspended to be 'SUSPENDED', got %s", tenants.StatusSuspended)
	}
	if tenants.StatusDeactivated != "DEACTIVATED" {
		t.Errorf("expected StatusDeactivated to be 'DEACTIVATED', got %s", tenants.StatusDeactivated)
	}
}

func TestTenant_JSONSerialization(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Millisecond)

	tn := tenants.Tenant{
		ID:           id,
		Name:         "QuickFleet India",
		Slug:         "quickfleet-india",
		Status:       tenants.StatusActive,
		ContactEmail: "ops@quickfleet.in",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, err := json.Marshal(tn)
	if err != nil {
		t.Fatalf("failed to marshal Tenant: %v", err)
	}

	var decoded tenants.Tenant
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Tenant: %v", err)
	}

	if decoded.ID != id {
		t.Errorf("expected ID %s, got %s", id, decoded.ID)
	}
	if decoded.Name != "QuickFleet India" {
		t.Errorf("expected name 'QuickFleet India', got %s", decoded.Name)
	}
	if decoded.Slug != "quickfleet-india" {
		t.Errorf("expected slug 'quickfleet-india', got %s", decoded.Slug)
	}
}

func TestTenant_UpdateTenantRequest_JSON(t *testing.T) {
	newName := "Updated Fleet"
	newEmail := "newops@fleet.com"

	req := tenants.UpdateTenantRequest{
		Name:         &newName,
		ContactEmail: &newEmail,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal UpdateTenantRequest: %v", err)
	}

	var decoded tenants.UpdateTenantRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal UpdateTenantRequest: %v", err)
	}

	if decoded.Name == nil || *decoded.Name != newName {
		t.Errorf("expected name %s, got %v", newName, decoded.Name)
	}
	if decoded.ContactEmail == nil || *decoded.ContactEmail != newEmail {
		t.Errorf("expected email %s, got %v", newEmail, decoded.ContactEmail)
	}
}
