package audit_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/audit"
)

func TestAuditLog_StatusConstants(t *testing.T) {
	if audit.StatusSuccess != "SUCCESS" {
		t.Errorf("expected StatusSuccess to be 'SUCCESS', got %s", audit.StatusSuccess)
	}
	if audit.StatusFailure != "FAILURE" {
		t.Errorf("expected StatusFailure to be 'FAILURE', got %s", audit.StatusFailure)
	}
	if audit.StatusDenied != "DENIED" {
		t.Errorf("expected StatusDenied to be 'DENIED', got %s", audit.StatusDenied)
	}
}

func TestAuditLog_JSONSerialization(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	resourceID := "res-123"
	ip := "192.168.1.1"
	ua := "LogiFlows-Client/1.0"
	now := time.Now().UTC().Truncate(time.Millisecond)

	entry := audit.AuditLog{
		ID:           id,
		TenantID:     &tenantID,
		UserID:       &userID,
		Action:       "USER_LOGIN",
		ResourceType: "SESSION",
		ResourceID:   &resourceID,
		IPAddress:    &ip,
		UserAgent:    &ua,
		Status:       audit.StatusSuccess,
		Details: map[string]any{
			"auth_method": "password",
			"attempt":     1,
		},
		CreatedAt: now,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal AuditLog: %v", err)
	}

	var decoded audit.AuditLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal AuditLog: %v", err)
	}

	if decoded.ID != id {
		t.Errorf("expected ID %s, got %s", id, decoded.ID)
	}
	if decoded.Action != "USER_LOGIN" {
		t.Errorf("expected Action USER_LOGIN, got %s", decoded.Action)
	}
	if decoded.Status != audit.StatusSuccess {
		t.Errorf("expected Status %s, got %s", audit.StatusSuccess, decoded.Status)
	}
	if decoded.Details["auth_method"] != "password" {
		t.Errorf("expected details auth_method to be 'password', got %v", decoded.Details["auth_method"])
	}
}

func TestAuditLog_OmitEmptyFields(t *testing.T) {
	id := uuid.New()
	entry := audit.AuditLog{
		ID:           id,
		Action:       "SYSTEM_STARTUP",
		ResourceType: "SERVER",
		Status:       audit.StatusSuccess,
		Details:      map[string]any{},
		CreatedAt:    time.Now().UTC(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal minimal AuditLog: %v", err)
	}

	jsonStr := string(data)
	if jsonStr == "" {
		t.Errorf("expected non-empty JSON string")
	}

	// Verify omitted fields are not serialized
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal JSON map: %v", err)
	}

	if _, exists := m["tenant_id"]; exists {
		t.Errorf("tenant_id should be omitted when nil")
	}
	if _, exists := m["user_id"]; exists {
		t.Errorf("user_id should be omitted when nil")
	}
	if _, exists := m["resource_id"]; exists {
		t.Errorf("resource_id should be omitted when nil")
	}
}
