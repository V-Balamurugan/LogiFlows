package audit

import (
	"time"

	"github.com/google/uuid"
)

// Status constants for audit log entries.
const (
	StatusSuccess = "SUCCESS"
	StatusFailure = "FAILURE"
	StatusDenied  = "DENIED"
)

// AuditLog represents a security audit event record in LogiFlows.
type AuditLog struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     *uuid.UUID     `json:"tenant_id,omitempty"`
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   *string        `json:"resource_id,omitempty"`
	IPAddress    *string        `json:"ip_address,omitempty"`
	UserAgent    *string        `json:"user_agent,omitempty"`
	Status       string         `json:"status"`
	Details      map[string]any `json:"details"`
	CreatedAt    time.Time      `json:"created_at"`
}
