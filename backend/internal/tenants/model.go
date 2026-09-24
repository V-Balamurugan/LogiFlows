package tenants

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive      = "ACTIVE"
	StatusSuspended   = "SUSPENDED"
	StatusDeactivated = "DEACTIVATED"
)

// Tenant represents an onboarding business company in LogiFlows.
type Tenant struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Status       string    `json:"status"`
	ContactEmail string    `json:"contact_email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
