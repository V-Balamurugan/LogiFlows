package memberships

import (
	"time"

	"github.com/google/uuid"
)

// Role constants defined in LogiFlows RBAC architecture.
const (
	RolePlatformAdmin  = "PLATFORM_ADMIN"
	RoleTenantAdmin    = "TENANT_ADMIN"
	RoleTenantOperator = "TENANT_OPERATOR"
	RoleViewer         = "VIEWER"
)

// Status constants for memberships.
const (
	StatusActive      = "ACTIVE"
	StatusInvited     = "INVITED"
	StatusDeactivated = "DEACTIVATED"
)

// TenantMembership links a user to a tenant with an assigned role.
type TenantMembership struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemberDetails provides enriched tenant member info including user details.
type MemberDetails struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
