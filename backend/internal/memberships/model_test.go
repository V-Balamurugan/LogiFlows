package memberships_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/memberships"
)

func TestMembership_Constants(t *testing.T) {
	if memberships.RoleTenantAdmin != "TENANT_ADMIN" {
		t.Errorf("expected TENANT_ADMIN, got %s", memberships.RoleTenantAdmin)
	}
	if memberships.RoleTenantOperator != "TENANT_OPERATOR" {
		t.Errorf("expected TENANT_OPERATOR, got %s", memberships.RoleTenantOperator)
	}
	if memberships.StatusActive != "ACTIVE" {
		t.Errorf("expected StatusActive to be 'ACTIVE', got %s", memberships.StatusActive)
	}
	if memberships.StatusInvited != "INVITED" {
		t.Errorf("expected StatusInvited to be 'INVITED', got %s", memberships.StatusInvited)
	}
	if memberships.StatusDeactivated != "DEACTIVATED" {
		t.Errorf("expected StatusDeactivated to be 'DEACTIVATED', got %s", memberships.StatusDeactivated)
	}
}

func TestMembership_JSONSerialization(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Millisecond)

	m := memberships.TenantMembership{
		ID:        id,
		TenantID:  tenantID,
		UserID:    userID,
		Role:      memberships.RoleTenantAdmin,
		Status:    memberships.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal TenantMembership: %v", err)
	}

	var decoded memberships.TenantMembership
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal TenantMembership: %v", err)
	}

	if decoded.ID != id || decoded.TenantID != tenantID || decoded.UserID != userID {
		t.Errorf("decoded UUIDs do not match original")
	}
	if decoded.Role != memberships.RoleTenantAdmin {
		t.Errorf("expected role %s, got %s", memberships.RoleTenantAdmin, decoded.Role)
	}
}

func TestMemberDetails_JSONSerialization(t *testing.T) {
	details := memberships.MemberDetails{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		Email:     "driver@logiflows.com",
		FullName:  "Ramesh Kumar",
		Role:      memberships.RoleTenantOperator,
		Status:    memberships.StatusActive,
		CreatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(details)
	if err != nil {
		t.Fatalf("failed to marshal MemberDetails: %v", err)
	}

	var decoded memberships.MemberDetails
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal MemberDetails: %v", err)
	}

	if decoded.Email != "driver@logiflows.com" {
		t.Errorf("expected email driver@logiflows.com, got %s", decoded.Email)
	}
	if decoded.FullName != "Ramesh Kumar" {
		t.Errorf("expected full name Ramesh Kumar, got %s", decoded.FullName)
	}
}
