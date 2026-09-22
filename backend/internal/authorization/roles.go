package authorization

import (
	"slices"

	"github.com/logiflows/logiflows/backend/internal/memberships"
)

// Re-export common roles for convenience.
const (
	RolePlatformAdmin  = memberships.RolePlatformAdmin
	RoleTenantAdmin    = memberships.RoleTenantAdmin
	RoleTenantOperator = memberships.RoleTenantOperator
	RoleViewer         = memberships.RoleViewer
)

// ValidRoles returns all authorized roles in LogiFlows.
var ValidRoles = []string{
	RolePlatformAdmin,
	RoleTenantAdmin,
	RoleTenantOperator,
	RoleViewer,
}

// IsValidRole checks if a given role string is a recognized system role.
func IsValidRole(role string) bool {
	return slices.Contains(ValidRoles, role)
}

// HasRole checks if the target role is included in allowedRoles, or if user is a Platform Admin.
func HasRole(userRole string, isPlatformAdmin bool, allowedRoles ...string) bool {
	if isPlatformAdmin || userRole == RolePlatformAdmin {
		return true
	}
	return slices.Contains(allowedRoles, userRole)
}

// CanManageTenant returns true if the user has administrative rights over the tenant.
func CanManageTenant(role string, isPlatformAdmin bool) bool {
	return HasRole(role, isPlatformAdmin, RoleTenantAdmin)
}

// CanOperateTenant returns true if the user can execute operational tasks (shipments, handovers).
func CanOperateTenant(role string, isPlatformAdmin bool) bool {
	return HasRole(role, isPlatformAdmin, RoleTenantAdmin, RoleTenantOperator)
}

// CanViewTenant returns true if the user has at least read-only viewing permissions.
func CanViewTenant(role string, isPlatformAdmin bool) bool {
	return HasRole(role, isPlatformAdmin, RoleTenantAdmin, RoleTenantOperator, RoleViewer)
}
